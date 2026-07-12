package favicon

import (
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/kodehat/portkey/internal/build"
	"github.com/kodehat/portkey/internal/metrics"
)

func TestMain(m *testing.M) {
	build.LoadBuildDetails("test")
	metrics.Load()
	os.Exit(m.Run())
}

func TestNormalizeHostname(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		{"github.com", "github.com"},
		{"www.github.com", "github.com"},
		{"GITHUB.COM", "github.com"},
		{"WWW.GITHUB.COM", "github.com"},
		{"  github.com  ", "github.com"},
		{"sub.domain.com", "sub.domain.com"},
		{"www.sub.domain.com", "sub.domain.com"},
		{"", ""},
	}

	for _, tt := range tests {
		got := NormalizeHostname(tt.input)
		if got != tt.expected {
			t.Errorf("NormalizeHostname(%q) = %q, want %q", tt.input, got, tt.expected)
		}
	}
}

func TestNew(t *testing.T) {
	dir := t.TempDir()
	cacheDir := filepath.Join(dir, "favicons")
	c := New(cacheDir)

	if c.dir != cacheDir {
		t.Errorf("New().dir = %q, want %q", c.dir, cacheDir)
	}

	// Verify directory was created.
	if _, err := os.Stat(cacheDir); os.IsNotExist(err) {
		t.Errorf("cache directory %q was not created", cacheDir)
	}

	if c.client == nil {
		t.Error("New().client is nil")
	}

	if c.failures == nil {
		t.Error("New().failures is nil")
	}
}

func TestCachePath(t *testing.T) {
	dir := t.TempDir()
	c := New(dir)

	path := c.cachePath("github.com")
	expected := filepath.Join(dir, "github.com.png")
	if path != expected {
		t.Errorf("cachePath(%q) = %q, want %q", "github.com", path, expected)
	}
}

func TestServeHTTP_EmptyDomain(t *testing.T) {
	c := New(t.TempDir())
	w := httptest.NewRecorder()
	r := httptest.NewRequest("GET", "/_/favicon", nil)
	c.ServeHTTP(w, r)

	if w.Code != http.StatusNotFound {
		t.Errorf("expected 404 for empty domain, got %d", w.Code)
	}
}

func TestServeHTTP_NoDomainParam(t *testing.T) {
	c := New(t.TempDir())
	w := httptest.NewRecorder()
	r := httptest.NewRequest("GET", "/_/favicon?domain=", nil)
	c.ServeHTTP(w, r)

	if w.Code != http.StatusNotFound {
		t.Errorf("expected 404 for missing domain, got %d", w.Code)
	}
}

func TestServeHTTP_CacheHitServesFile(t *testing.T) {
	c := New(t.TempDir())

	// Write a cached favicon.
	path := c.cachePath("github.com")
	if err := os.WriteFile(path, []byte("fake-png-data"), 0644); err != nil {
		t.Fatal(err)
	}

	w := httptest.NewRecorder()
	r := httptest.NewRequest("GET", "/_/favicon?domain=github.com", nil)
	c.ServeHTTP(w, r)

	if w.Code != http.StatusOK {
		t.Errorf("expected 200 for cache hit, got %d", w.Code)
	}
	if w.Body.String() != "fake-png-data" {
		t.Errorf("expected cached content, got %q", w.Body.String())
	}
}

func TestServeHTTP_CacheHitNormalizesDomain(t *testing.T) {
	c := New(t.TempDir())

	// Write cache under normalized name.
	path := c.cachePath("github.com")
	if err := os.WriteFile(path, []byte("data"), 0644); err != nil {
		t.Fatal(err)
	}

	// Request with www. prefix.
	w := httptest.NewRecorder()
	r := httptest.NewRequest("GET", "/_/favicon?domain=www.github.com", nil)
	c.ServeHTTP(w, r)

	if w.Code != http.StatusOK {
		t.Errorf("expected 200 after normalization, got %d", w.Code)
	}
}

func TestServeHTTP_CacheMissShowsFallback(t *testing.T) {
	c := New(t.TempDir())

	// Use a client with an extremely short timeout so the fetch fails quickly.
	c.client = &http.Client{Timeout: 1 * time.Nanosecond}

	w := httptest.NewRecorder()
	r := httptest.NewRequest("GET", "/_/favicon?domain=any-unreachable.example", nil)
	c.ServeHTTP(w, r)

	if w.Code != http.StatusOK {
		t.Errorf("expected 200 with fallback, got %d", w.Code)
	}
	if ct := w.Header().Get("Content-Type"); ct != "image/svg+xml" {
		t.Errorf("expected SVG content type, got %q", ct)
	}
	if !strings.Contains(w.Body.String(), "<svg") {
		t.Error("expected SVG in fallback response body")
	}

	c.mu.RLock()
	failAt, failed := c.failures["any-unreachable.example"]
	c.mu.RUnlock()
	if !failed {
		t.Error("expected failed domain to be recorded")
	}
	if failAt.IsZero() {
		t.Error("expected non-zero failure timestamp")
	}
}

func TestServeHTTP_FailureBackoff(t *testing.T) {
	c := New(t.TempDir())

	c.mu.Lock()
	c.failures["bad.example.com"] = time.Now()
	c.mu.Unlock()

	w := httptest.NewRecorder()
	r := httptest.NewRequest("GET", "/_/favicon?domain=bad.example.com", nil)
	c.ServeHTTP(w, r)

	if w.Code != http.StatusOK {
		t.Errorf("expected 200 with fallback during backoff, got %d", w.Code)
	}
	if ct := w.Header().Get("Content-Type"); ct != "image/svg+xml" {
		t.Errorf("expected SVG content type during backoff, got %q", ct)
	}
}

func TestServeHTTP_ServeDefaultHeaders(t *testing.T) {
	c := New(t.TempDir())

	// Ensure fetch fails by using a short timeout client.
	c.client = &http.Client{Timeout: 1 * time.Nanosecond}

	w := httptest.NewRecorder()
	r := httptest.NewRequest("GET", "/_/favicon?domain=any.example", nil)
	c.ServeHTTP(w, r)

	if cc := w.Header().Get("Cache-Control"); cc != "public, max-age=3600" {
		t.Errorf("expected Cache-Control on default, got %q", cc)
	}
}

func TestServeHTTP_CacheHitSetsCacheControl(t *testing.T) {
	c := New(t.TempDir())

	path := c.cachePath("example.com")
	if err := os.WriteFile(path, []byte("data"), 0644); err != nil {
		t.Fatal(err)
	}

	w := httptest.NewRecorder()
	r := httptest.NewRequest("GET", "/_/favicon?domain=example.com", nil)
	c.ServeHTTP(w, r)

	// http.ServeFile sets its own headers, but the handler shouldn't interfere.
	if w.Code != http.StatusOK {
		t.Errorf("expected 200, got %d", w.Code)
	}
}

func TestInit(t *testing.T) {
	dir := t.TempDir()
	Init(dir)

	if C == nil {
		t.Fatal("Init() did not set global C")
	}
	if C.dir != dir {
		t.Errorf("Init() set dir = %q, want %q", C.dir, dir)
	}
}

func TestDomainFromURL(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		{"https://github.com", "github.com"},
		{"http://github.com", "github.com"},
		{"https://github.com/foo/bar", "github.com"},
		{"https://github.com:8080/path", "github.com"},
		{"https://www.example.com", "www.example.com"},
		{"ftp://example.com", ""},
		{"/relative/path", ""},
		{"", ""},
	}

	for _, tt := range tests {
		got := DomainFromURL(tt.input)
		if got != tt.expected {
			t.Errorf("DomainFromURL(%q) = %q, want %q", tt.input, got, tt.expected)
		}
	}
}

func TestIsValidHostname_InvalidChars(t *testing.T) {
	invalid := []string{
		"GITHUB.COM",
		"github_com",
		"github com",
		"github.com/path",
		"github@com",
	}
	for _, d := range invalid {
		if isValidHostname(d) {
			t.Errorf("isValidHostname(%q) = true, want false", d)
		}
	}
}

func TestIsValidHostname_Empty(t *testing.T) {
	if isValidHostname("") {
		t.Error("isValidHostname(\"\") = true, want false")
	}
}

func TestCachePath_PathTraversal(t *testing.T) {
	dir := t.TempDir()
	c := New(dir)

	malicious := "../etc/passwd"
	got := c.cachePath(malicious)
	expected := filepath.Join(dir, "invalid.png")
	if got != expected {
		t.Errorf("cachePath(%q) = %q, want %q (traversal protection)", malicious, got, expected)
	}
}

func TestNew_PanicsOnInvalidDir(t *testing.T) {
	dir := t.TempDir()
	// Create a file where the dir should be so MkdirAll fails.
	filePath := filepath.Join(dir, "notadir")
	if err := os.WriteFile(filePath, []byte("x"), 0644); err != nil {
		t.Fatal(err)
	}

	defer func() {
		if r := recover(); r == nil {
			t.Fatal("expected New() to panic for uncreateable cache dir")
		}
	}()

	New(filepath.Join(filePath, "subdir"))
}

// roundTripFunc lets tests intercept http.Client requests without a real server.
type roundTripFunc func(*http.Request) (*http.Response, error)

func (f roundTripFunc) RoundTrip(r *http.Request) (*http.Response, error) { return f(r) }

func TestFetchAndSave_Success(t *testing.T) {
	dir := t.TempDir()
	c := New(dir)
	c.client = &http.Client{
		Transport: roundTripFunc(func(r *http.Request) (*http.Response, error) {
			return &http.Response{
				StatusCode: http.StatusOK,
				Body:       io.NopCloser(strings.NewReader("PNG-DATA")),
			}, nil
		}),
	}

	path := filepath.Join(dir, "test.com.png")
	if err := c.fetchAndSave("test.com", path); err != nil {
		t.Fatalf("expected success, got %v", err)
	}

	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("file not written: %v", err)
	}
	if string(data) != "PNG-DATA" {
		t.Fatalf("expected PNG-DATA, got %q", string(data))
	}
}

func TestFetchAndSave_NonOKStatus(t *testing.T) {
	dir := t.TempDir()
	c := New(dir)
	c.client = &http.Client{
		Transport: roundTripFunc(func(r *http.Request) (*http.Response, error) {
			return &http.Response{
				StatusCode: http.StatusNotFound,
				Body:       io.NopCloser(strings.NewReader("")),
			}, nil
		}),
	}

	path := filepath.Join(dir, "missing.com.png")
	if err := c.fetchAndSave("missing.com", path); err == nil {
		t.Fatal("expected error for non-200 status")
	}
}

func TestFetchAndSave_CreateFails(t *testing.T) {
	dir := t.TempDir()
	c := New(dir)
	c.client = &http.Client{
		Transport: roundTripFunc(func(r *http.Request) (*http.Response, error) {
			return &http.Response{
				StatusCode: http.StatusOK,
				Body:       io.NopCloser(strings.NewReader("data")),
			}, nil
		}),
	}

	// Path inside a non-existent subdirectory so os.Create fails.
	path := filepath.Join(dir, "nonexistent", "test.com.png")
	if err := c.fetchAndSave("test.com", path); err == nil {
		t.Fatal("expected error when temp file create fails")
	}
}

func TestFetchAndSave_CopyFails(t *testing.T) {
	dir := t.TempDir()
	c := New(dir)
	c.client = &http.Client{
		Transport: roundTripFunc(func(r *http.Request) (*http.Response, error) {
			return &http.Response{
				StatusCode: http.StatusOK,
				Body:       io.NopCloser(&errReader{}),
			}, nil
		}),
	}

	path := filepath.Join(dir, "copyerr.com.png")
	if err := c.fetchAndSave("copyerr.com", path); err == nil {
		t.Fatal("expected error when copy fails")
	}
	// Temp file should be cleaned up.
	if _, err := os.Stat(path + ".tmp"); !os.IsNotExist(err) {
		t.Error("expected temp file to be removed after copy error")
	}
}

// errReader always returns an error on Read.
type errReader struct{}

func (errReader) Read([]byte) (int, error) { return 0, io.ErrUnexpectedEOF }

func TestRefresh_Success(t *testing.T) {
	dir := t.TempDir()
	c := New(dir)
	c.client = &http.Client{
		Transport: roundTripFunc(func(r *http.Request) (*http.Response, error) {
			return &http.Response{
				StatusCode: http.StatusOK,
				Body:       io.NopCloser(strings.NewReader("REFRESHED")),
			}, nil
		}),
	}

	path := filepath.Join(dir, "refresh.com.png")
	c.refresh("refresh.com", path)

	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("expected refreshed file: %v", err)
	}
	if string(data) != "REFRESHED" {
		t.Fatalf("expected REFRESHED, got %q", string(data))
	}
}

func TestRefresh_Failure(t *testing.T) {
	dir := t.TempDir()
	c := New(dir)
	c.client = &http.Client{Timeout: 1 * time.Nanosecond}

	path := filepath.Join(dir, "fail.com.png")
	c.refresh("fail.com", path)

	c.mu.RLock()
	_, failed := c.failures["fail.com"]
	c.mu.RUnlock()
	if !failed {
		t.Fatal("expected refresh failure to be recorded")
	}
}

func TestServeHTTP_StaleFileTriggersRefresh(t *testing.T) {
	dir := t.TempDir()
	c := New(dir)
	c.client = &http.Client{Timeout: 1 * time.Nanosecond}

	path := c.cachePath("stale.com")
	if err := os.WriteFile(path, []byte("old-data"), 0644); err != nil {
		t.Fatal(err)
	}
	// Set mod time beyond CacheTTL so the stale branch triggers.
	oldTime := time.Now().Add(-(CacheTTL + 24*time.Hour))
	if err := os.Chtimes(path, oldTime, oldTime); err != nil {
		t.Fatal(err)
	}

	w := httptest.NewRecorder()
	r := httptest.NewRequest("GET", "/_/favicon?domain=stale.com", nil)
	c.ServeHTTP(w, r)

	if w.Code != http.StatusOK {
		t.Errorf("expected 200 for stale cache hit, got %d", w.Code)
	}
	// Allow background goroutine to complete.
	time.Sleep(20 * time.Millisecond)
}
