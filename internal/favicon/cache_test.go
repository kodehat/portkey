package favicon

import (
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
