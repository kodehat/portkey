package favicon

import (
	"io"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	faviconlib "go.deanishe.net/favicon"

	"github.com/kodehat/portkey/internal/build"
	"github.com/kodehat/portkey/internal/config"
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
	c := New(cacheDir, nil)

	if c.dir != cacheDir {
		t.Errorf("New().dir = %q, want %q", c.dir, cacheDir)
	}

	if _, err := os.Stat(cacheDir); os.IsNotExist(err) {
		t.Errorf("cache directory %q was not created", cacheDir)
	}

	if c.client == nil {
		t.Error("New().client is nil")
	}

	if c.finder == nil {
		t.Error("New().finder is nil")
	}

	if c.failures == nil {
		t.Error("New().failures is nil")
	}
}

func TestNew_WithLogger(t *testing.T) {
	logger := slog.New(slog.NewTextHandler(io.Discard, nil))
	c := New(t.TempDir(), logger)
	if c.logger != logger {
		t.Error("expected logger to be set")
	}
}

func TestCachePath(t *testing.T) {
	dir := t.TempDir()
	c := New(dir, nil)

	path := c.cachePath("github.com")
	expected := filepath.Join(dir, "github.com.png")
	if path != expected {
		t.Errorf("cachePath(%q) = %q, want %q", "github.com", path, expected)
	}
}

func TestServeHTTP_EmptyDomain(t *testing.T) {
	c := New(t.TempDir(), nil)
	w := httptest.NewRecorder()
	r := httptest.NewRequest("GET", "/_/favicon", nil)
	c.ServeHTTP(w, r)

	if w.Code != http.StatusNotFound {
		t.Errorf("expected 404 for empty domain, got %d", w.Code)
	}
}

func TestServeHTTP_NoDomainParam(t *testing.T) {
	c := New(t.TempDir(), nil)
	w := httptest.NewRecorder()
	r := httptest.NewRequest("GET", "/_/favicon?domain=", nil)
	c.ServeHTTP(w, r)

	if w.Code != http.StatusNotFound {
		t.Errorf("expected 404 for missing domain, got %d", w.Code)
	}
}

func TestServeHTTP_CacheHitServesFile(t *testing.T) {
	c := New(t.TempDir(), nil)

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
	c := New(t.TempDir(), nil)

	path := c.cachePath("github.com")
	if err := os.WriteFile(path, []byte("data"), 0644); err != nil {
		t.Fatal(err)
	}

	w := httptest.NewRecorder()
	r := httptest.NewRequest("GET", "/_/favicon?domain=www.github.com", nil)
	c.ServeHTTP(w, r)

	if w.Code != http.StatusOK {
		t.Errorf("expected 200 after normalization, got %d", w.Code)
	}
}

func TestServeHTTP_CacheMissShowsFallback(t *testing.T) {
	c := New(t.TempDir(), nil)

	// Failing transport — both finder and download fail.
	failTransport := roundTripFunc(func(r *http.Request) (*http.Response, error) {
		return nil, io.ErrUnexpectedEOF
	})
	c.client = &http.Client{Transport: failTransport, Timeout: 10 * time.Second}
	c.finder = faviconlib.New(faviconlib.WithClient(c.client))

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
	c := New(t.TempDir(), nil)

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
	c := New(t.TempDir(), nil)

	failTransport := roundTripFunc(func(r *http.Request) (*http.Response, error) {
		return nil, io.ErrUnexpectedEOF
	})
	c.client = &http.Client{Transport: failTransport, Timeout: 10 * time.Second}
	c.finder = faviconlib.New(faviconlib.WithClient(c.client))

	w := httptest.NewRecorder()
	r := httptest.NewRequest("GET", "/_/favicon?domain=any.example", nil)
	c.ServeHTTP(w, r)

	if cc := w.Header().Get("Cache-Control"); cc != "public, max-age=3600" {
		t.Errorf("expected Cache-Control on default, got %q", cc)
	}
}

func TestServeHTTP_CacheHitSetsCacheControl(t *testing.T) {
	c := New(t.TempDir(), nil)

	path := c.cachePath("example.com")
	if err := os.WriteFile(path, []byte("data"), 0644); err != nil {
		t.Fatal(err)
	}

	w := httptest.NewRecorder()
	r := httptest.NewRequest("GET", "/_/favicon?domain=example.com", nil)
	c.ServeHTTP(w, r)

	if w.Code != http.StatusOK {
		t.Errorf("expected 200, got %d", w.Code)
	}
}

func TestInit(t *testing.T) {
	dir := t.TempDir()
	Init(dir, nil)

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
	c := New(dir, nil)

	malicious := "../etc/passwd"
	got := c.cachePath(malicious)
	expected := filepath.Join(dir, "invalid.png")
	if got != expected {
		t.Errorf("cachePath(%q) = %q, want %q (traversal protection)", malicious, got, expected)
	}
}

func TestNew_PanicsOnInvalidDir(t *testing.T) {
	dir := t.TempDir()
	filePath := filepath.Join(dir, "notadir")
	if err := os.WriteFile(filePath, []byte("x"), 0644); err != nil {
		t.Fatal(err)
	}

	defer func() {
		if r := recover(); r == nil {
			t.Fatal("expected New() to panic for uncreateable cache dir")
		}
	}()

	New(filepath.Join(filePath, "subdir"), nil)
}

// roundTripFunc lets tests intercept http.Client requests without a real server.
type roundTripFunc func(*http.Request) (*http.Response, error)

func (f roundTripFunc) RoundTrip(r *http.Request) (*http.Response, error) { return f(r) }

// testServerWithFavicon starts a test HTTP server that responds with a homepage
// (containing a favicon <link>) at / and PNG data at /favicon.png. Returns the
// server, plus a client and finder that route all requests through it.
func testServerWithFavicon(t *testing.T) (*httptest.Server, *http.Client, *faviconlib.Finder) {
	t.Helper()

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/" || r.URL.Path == "" {
			w.Header().Set("Content-Type", "text/html")
			w.Write([]byte(`<html><head><link rel="icon" type="image/png" href="/favicon.png"></head></html>`))
			return
		}
		w.Header().Set("Content-Type", "image/png")
		w.Write([]byte("FAVICON-DATA"))
	}))

	transport := roundTripFunc(func(r *http.Request) (*http.Response, error) {
		u := server.URL + r.URL.Path
		if r.URL.RawQuery != "" {
			u += "?" + r.URL.RawQuery
		}
		req, _ := http.NewRequest(r.Method, u, r.Body)
		for k, vs := range r.Header {
			for _, v := range vs {
				req.Header.Add(k, v)
			}
		}
		return server.Client().Do(req)
	})

	client := &http.Client{Transport: transport, Timeout: 10 * time.Second}
	finder := faviconlib.New(faviconlib.WithClient(client))
	return server, client, finder
}

func TestFetchAndSave_Success(t *testing.T) {
	server, client, finder := testServerWithFavicon(t)
	defer server.Close()

	dir := t.TempDir()
	c := New(dir, nil)
	c.client = client
	c.finder = finder

	path := filepath.Join(dir, "test.com.png")
	if err := c.fetchAndSave("test.com", path); err != nil {
		t.Fatalf("expected success, got %v", err)
	}

	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("file not written: %v", err)
	}
	if string(data) != "FAVICON-DATA" {
		t.Fatalf("expected FAVICON-DATA, got %q", string(data))
	}
}

func TestFetchAndSave_NonOKStatus(t *testing.T) {
	// Server where favicon download returns 404.
	failServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/" || r.URL.Path == "" {
			w.Header().Set("Content-Type", "text/html")
			w.Write([]byte(`<html><head><link rel="icon" type="image/png" href="/favicon.png"></head></html>`))
			return
		}
		w.WriteHeader(http.StatusNotFound)
	}))
	defer failServer.Close()

	transport := roundTripFunc(func(r *http.Request) (*http.Response, error) {
		u := failServer.URL + r.URL.Path
		if r.URL.RawQuery != "" {
			u += "?" + r.URL.RawQuery
		}
		req, _ := http.NewRequest(r.Method, u, r.Body)
		for k, vs := range r.Header {
			for _, v := range vs {
				req.Header.Add(k, v)
			}
		}
		return failServer.Client().Do(req)
	})

	dir := t.TempDir()
	c := New(dir, nil)
	c.client = &http.Client{Transport: transport, Timeout: 10 * time.Second}
	c.finder = faviconlib.New(faviconlib.WithClient(c.client))

	path := filepath.Join(dir, "notfound.com.png")
	if err := c.fetchAndSave("notfound.com", path); err == nil {
		t.Fatal("expected error for non-200 download status")
	}
}

func TestFetchAndSave_CreateFails(t *testing.T) {
	server, client, finder := testServerWithFavicon(t)
	defer server.Close()

	dir := t.TempDir()
	c := New(dir, nil)
	c.client = client
	c.finder = finder

	path := filepath.Join(dir, "nonexistent", "test.com.png")
	if err := c.fetchAndSave("test.com", path); err == nil {
		t.Fatal("expected error when temp file create fails")
	}
}

func TestFetchAndSave_CopyFails(t *testing.T) {
	server, client, finder := testServerWithFavicon(t)
	defer server.Close()

	dir := t.TempDir()
	c := New(dir, nil)
	c.client = client
	c.finder = finder

	// Override client so download body errors on read, but finder still works.
	origTransport := client.Transport
	errorTransport := roundTripFunc(func(r *http.Request) (*http.Response, error) {
		if r.URL.Path == "/" || r.URL.Path == "" {
			return origTransport.RoundTrip(r)
		}
		return &http.Response{
			StatusCode: http.StatusOK,
			Body:       io.NopCloser(&errReader{}),
		}, nil
	})
	c.client = &http.Client{Transport: errorTransport, Timeout: 10 * time.Second}

	path := filepath.Join(dir, "copyerr.com.png")
	if err := c.fetchAndSave("copyerr.com", path); err == nil {
		t.Fatal("expected error when copy fails")
	}
	if _, err := os.Stat(path + ".tmp"); !os.IsNotExist(err) {
		t.Error("expected temp file to be removed after copy error")
	}
}

type errReader struct{}

func (errReader) Read([]byte) (int, error) { return 0, io.ErrUnexpectedEOF }

func TestRefresh_Success(t *testing.T) {
	server, client, finder := testServerWithFavicon(t)
	defer server.Close()

	dir := t.TempDir()
	c := New(dir, nil)
	c.client = client
	c.finder = finder

	path := filepath.Join(dir, "refresh.com.png")
	c.refresh("refresh.com", path)

	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("expected refreshed file: %v", err)
	}
	if string(data) != "FAVICON-DATA" {
		t.Fatalf("expected FAVICON-DATA, got %q", string(data))
	}
}

func TestRefresh_Failure(t *testing.T) {
	dir := t.TempDir()
	c := New(dir, nil)

	failTransport := roundTripFunc(func(r *http.Request) (*http.Response, error) {
		return nil, io.ErrUnexpectedEOF
	})
	c.client = &http.Client{Transport: failTransport, Timeout: 10 * time.Second}
	c.finder = faviconlib.New(faviconlib.WithClient(c.client))

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
	c := New(dir, nil)

	path := c.cachePath("stale.com")
	if err := os.WriteFile(path, []byte("old-data"), 0644); err != nil {
		t.Fatal(err)
	}
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
	time.Sleep(20 * time.Millisecond)
}

func TestServeHTTP_CacheDisabled(t *testing.T) {
	server, client, finder := testServerWithFavicon(t)
	defer server.Close()

	orig := config.C.FaviconCacheDisabled
	config.C.FaviconCacheDisabled = true
	defer func() { config.C.FaviconCacheDisabled = orig }()

	dir := t.TempDir()
	c := New(dir, nil)
	c.client = client
	c.finder = finder

	w := httptest.NewRecorder()
	r := httptest.NewRequest("GET", "/_/favicon?domain=test.com", nil)
	c.ServeHTTP(w, r)

	if w.Code != http.StatusOK {
		t.Errorf("expected 200 with cache disabled, got %d", w.Code)
	}
	if w.Body.String() != "FAVICON-DATA" {
		t.Errorf("expected FAVICON-DATA, got %q", w.Body.String())
	}
	if ct := w.Header().Get("Content-Type"); ct != "image/png" {
		t.Errorf("expected Content-Type image/png, got %q", ct)
	}
}

func TestSelectBestIcon(t *testing.T) {
	icons := []*faviconlib.Icon{
		{URL: "/favicon.ico", MimeType: "image/x-icon", Width: 16, Height: 16},
		{URL: "/icon-32.png", MimeType: "image/png", Width: 32, Height: 32},
		{URL: "/icon-64.png", MimeType: "image/png", Width: 64, Height: 64},
		{URL: "/icon-128.png", MimeType: "image/png", Width: 128, Height: 128},
		{URL: "/og.png", MimeType: "image/png", Width: 1200, Height: 630},
	}

	best := selectBestIcon(icons)
	if best == nil {
		t.Fatal("expected non-nil best icon")
	}
	if best.URL != "/icon-64.png" {
		t.Errorf("expected /icon-64.png as best, got %q", best.URL)
	}
}

func TestSelectBestIcon_Empty(t *testing.T) {
	if got := selectBestIcon(nil); got != nil {
		t.Error("expected nil for empty icon list")
	}
	if got := selectBestIcon([]*faviconlib.Icon{}); got != nil {
		t.Error("expected nil for empty icon list")
	}
}

func TestLogDebug_NilLogger(t *testing.T) {
	c := New(t.TempDir(), nil)
	// Must not panic.
	c.logDebug("test", "key", "val")
}

func TestLogWarn_NilLogger(t *testing.T) {
	c := New(t.TempDir(), nil)
	// Must not panic.
	c.logWarn("test", "key", "val")
}
