package favicon

import (
	"bytes"
	"context"
	"io"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/kodehat/favifetch"

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

func TestFetchOptions_UseBrowserMode(t *testing.T) {
	c := New(t.TempDir(), nil)
	opts := favifetch.DefaultOptions(c.fetchOptions()...)

	if opts.Mode != favifetch.ModeBrowser {
		t.Errorf("fetch mode = %v, want %v", opts.Mode, favifetch.ModeBrowser)
	}
	if opts.Size != 0 {
		t.Errorf("fetch size = %d, want 0 because browser mode cannot resize", opts.Size)
	}
}

func TestCachePath(t *testing.T) {
	dir := t.TempDir()
	c := New(dir, nil)

	path := c.cachePath("github.com", favifetch.FormatPNG)
	expected := filepath.Join(dir, "github.com.png")
	if path != expected {
		t.Errorf("cachePath(%q, FormatPNG) = %q, want %q", "github.com", path, expected)
	}

	svgPath := c.cachePath("github.com", favifetch.FormatSVG)
	svgExpected := filepath.Join(dir, "github.com.svg")
	if svgPath != svgExpected {
		t.Errorf("cachePath(%q, FormatSVG) = %q, want %q", "github.com", svgPath, svgExpected)
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

	path := c.cachePath("github.com", favifetch.FormatPNG)
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

	path := c.cachePath("github.com", favifetch.FormatPNG)
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

func TestServeHTTP_CacheMissSignalsFallback(t *testing.T) {
	c := New(t.TempDir(), nil)

	failTransport := roundTripFunc(func(r *http.Request) (*http.Response, error) {
		return nil, io.ErrUnexpectedEOF
	})
	c.client = &http.Client{Transport: failTransport, Timeout: 10 * time.Second}

	w := httptest.NewRecorder()
	r := httptest.NewRequest("GET", "/_/favicon?domain=any-unreachable.example", nil)
	c.ServeHTTP(w, r)

	if w.Code != http.StatusNotFound {
		t.Errorf("expected 404 to trigger the inline fallback, got %d", w.Code)
	}
	if cc := w.Header().Get("Cache-Control"); cc != "no-store" {
		t.Errorf("expected Cache-Control: no-store, got %q", cc)
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

	if w.Code != http.StatusNotFound {
		t.Errorf("expected 404 to trigger the inline fallback during backoff, got %d", w.Code)
	}
}

func TestServeHTTP_ServeDefaultHeaders(t *testing.T) {
	c := New(t.TempDir(), nil)

	failTransport := roundTripFunc(func(r *http.Request) (*http.Response, error) {
		return nil, io.ErrUnexpectedEOF
	})
	c.client = &http.Client{Transport: failTransport, Timeout: 10 * time.Second}

	w := httptest.NewRecorder()
	r := httptest.NewRequest("GET", "/_/favicon?domain=any.example", nil)
	c.ServeHTTP(w, r)

	if cc := w.Header().Get("Cache-Control"); cc != "no-store" {
		t.Errorf("expected Cache-Control: no-store, got %q", cc)
	}
}

func TestServeHTTP_CacheHitSetsCacheControl(t *testing.T) {
	c := New(t.TempDir(), nil)

	path := c.cachePath("example.com", favifetch.FormatPNG)
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
	got := c.cachePath(malicious, favifetch.FormatPNG)
	expected := filepath.Join(dir, "invalid.png")
	if got != expected {
		t.Errorf("cachePath(%q, FormatPNG) = %q, want %q (traversal protection)", malicious, got, expected)
	}

	// SVG extension should also be protected.
	gotSVG := c.cachePath(malicious, favifetch.FormatSVG)
	expectedSVG := filepath.Join(dir, "invalid.svg")
	if gotSVG != expectedSVG {
		t.Errorf("cachePath(%q, FormatSVG) = %q, want %q (traversal protection)", malicious, gotSVG, expectedSVG)
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

// validPNG is a minimal 1x1 pixel PNG used in tests to satisfy favifetch's image
// validation.
var validPNG = []byte{0x89, 0x50, 0x4E, 0x47, 0x0D, 0x0A, 0x1A, 0x0A,
	0x00, 0x00, 0x00, 0x0D, 0x49, 0x48, 0x44, 0x52,
	0x00, 0x00, 0x00, 0x01, 0x00, 0x00, 0x00, 0x01,
	0x08, 0x02, 0x00, 0x00, 0x00, 0x90, 0x77, 0x53,
	0xDE, 0x00, 0x00, 0x00, 0x0C, 0x49, 0x44, 0x41,
	0x54, 0x08, 0xD7, 0x63, 0xF8, 0xCF, 0xC0, 0x00,
	0x00, 0x01, 0x01, 0x00, 0x05, 0x18, 0xD8, 0x73,
	0xE4, 0x00, 0x00, 0x00, 0x00, 0x49, 0x45, 0x4E,
	0x44, 0xAE, 0x42, 0x60, 0x82}

// roundTripFunc lets tests intercept http.Client requests without a real server.
type roundTripFunc func(*http.Request) (*http.Response, error)

func (f roundTripFunc) RoundTrip(r *http.Request) (*http.Response, error) { return f(r) }

// testServerWithFavicon starts a test HTTP server that responds with a homepage
// (containing a favicon <link>) at / and PNG data at /favicon.png. Returns the
// server and an http.Client that routes all requests through it.
func testServerWithFavicon(t *testing.T) (*httptest.Server, *http.Client) {
	t.Helper()

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/" || r.URL.Path == "" {
			w.Header().Set("Content-Type", "text/html")
			w.Write([]byte(`<html><head><link rel="icon" type="image/png" href="/favicon.png"></head></html>`))
			return
		}
		w.Header().Set("Content-Type", "image/png")
		w.Write(validPNG)
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

	return server, &http.Client{Transport: transport, Timeout: 10 * time.Second}
}

func TestFetchAndSave_Success(t *testing.T) {
	server, client := testServerWithFavicon(t)
	defer server.Close()

	dir := t.TempDir()
	c := New(dir, nil)
	c.client = client

	path, format, err := c.fetchAndSave(context.Background(), "test.com")
	if err != nil {
		t.Fatalf("expected success, got %v", err)
	}
	if format != favifetch.FormatPNG {
		t.Errorf("expected FormatPNG, got %v", format)
	}
	expectedPath := filepath.Join(dir, "test.com.png")
	if path != expectedPath {
		t.Errorf("expected path %q, got %q", expectedPath, path)
	}

	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("file not readable: %v", err)
	}
	// The PNG may be resized (WithSize(64)), so we can't compare exact bytes.
	// Verify it's valid image data by checking it starts with PNG header.
	if !bytes.HasPrefix(data, []byte{0x89, 0x50, 0x4E, 0x47}) {
		t.Fatalf("expected PNG data, got %x", data[:min(16, len(data))])
	}
}

func TestFetchAndSave_NonOKStatus(t *testing.T) {
	// Server where the homepage has no favicon link — favifetch will try
	// fallback paths like /favicon.ico, which will also 404.
	failServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
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

	_, _, err := c.fetchAndSave(context.Background(), "notfound.com")
	if err == nil {
		t.Fatal("expected error for non-200 download status")
	}
}

func TestFetchAndSave_CreateFails(t *testing.T) {
	server, client := testServerWithFavicon(t)
	defer server.Close()

	dir := t.TempDir()
	c := New(dir, nil)
	c.client = client

	// Poison cachePath with a non-existent subdirectory.
	c.dir = filepath.Join(dir, "nonexistent")

	_, _, err := c.fetchAndSave(context.Background(), "test.com")
	if err == nil {
		t.Fatal("expected error when temp file create fails")
	}
}

func TestRefresh_Success(t *testing.T) {
	server, client := testServerWithFavicon(t)
	defer server.Close()

	dir := t.TempDir()
	c := New(dir, nil)
	c.client = client

	c.refresh("refresh.com")

	path := filepath.Join(dir, "refresh.com.png")
	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("expected refreshed file: %v", err)
	}
	if !bytes.HasPrefix(data, []byte{0x89, 0x50, 0x4E, 0x47}) {
		t.Fatalf("expected PNG data, got %x", data[:min(16, len(data))])
	}
}

func TestRefresh_Failure(t *testing.T) {
	dir := t.TempDir()
	c := New(dir, nil)

	failTransport := roundTripFunc(func(r *http.Request) (*http.Response, error) {
		return nil, io.ErrUnexpectedEOF
	})
	c.client = &http.Client{Transport: failTransport, Timeout: 10 * time.Second}

	c.refresh("fail.com")

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

	path := c.cachePath("stale.com", favifetch.FormatPNG)
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
	server, client := testServerWithFavicon(t)
	defer server.Close()

	orig := config.C.FaviconCacheDisabled
	config.C.FaviconCacheDisabled = true
	defer func() { config.C.FaviconCacheDisabled = orig }()

	dir := t.TempDir()
	c := New(dir, nil)
	c.client = client

	w := httptest.NewRecorder()
	r := httptest.NewRequest("GET", "/_/favicon?domain=test.com", nil)
	c.ServeHTTP(w, r)

	if w.Code != http.StatusOK {
		t.Errorf("expected 200 with cache disabled, got %d", w.Code)
	}
	if !bytes.Equal(w.Body.Bytes(), validPNG) {
		t.Errorf("expected valid PNG data, got %x", w.Body.Bytes())
	}
	if ct := w.Header().Get("Content-Type"); ct != "image/png" {
		t.Errorf("expected Content-Type image/png, got %q", ct)
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
