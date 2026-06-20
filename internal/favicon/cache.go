// Package favicon provides a disk-backed cache for favicons fetched from
// a remote favicon service. Favicons are cached by normalized hostname and
// served from local disk on subsequent requests.
package favicon

import (
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"github.com/kodehat/portkey/internal/metrics"
)

const (
	// CacheTTL is how long a cached favicon is considered fresh.
	CacheTTL = 7 * 24 * time.Hour

	// FailureRetryAfter is how long to wait before retrying a failed domain.
	FailureRetryAfter = 1 * time.Hour

	// RemoteServiceURL is the default remote favicon service.
	RemoteServiceURL = "https://favicon.vemetric.com"
)

// C is the global favicon cache. Initialized by Init().
var C *Cache

// Cache caches favicons on disk by normalized hostname.
type Cache struct {
	dir      string
	client   *http.Client
	failures map[string]time.Time
	mu       sync.RWMutex
}

// Init initializes the global favicon cache with the given disk directory.
// Must be called before C is used (typically during application startup).
func Init(cacheDir string) {
	C = New(cacheDir)
}

// New creates a new Cache with the given disk directory.
func New(cacheDir string) *Cache {
	if err := os.MkdirAll(cacheDir, 0755); err != nil {
		panic(fmt.Errorf("failed to create favicon cache directory %s: %w", cacheDir, err))
	}
	return &Cache{
		dir:      cacheDir,
		client:   &http.Client{Timeout: 10 * time.Second},
		failures: make(map[string]time.Time),
	}
}

// DomainFromURL extracts the hostname from an absolute URL.
// Returns empty string for non-HTTP URLs.
func DomainFromURL(rawURL string) string {
	if !strings.HasPrefix(rawURL, "http") {
		return ""
	}
	// net/url is not imported; use a simple string-based approach.
	raw := strings.TrimPrefix(rawURL, "https://")
	raw = strings.TrimPrefix(raw, "http://")
	if idx := strings.IndexByte(raw, '/'); idx >= 0 {
		raw = raw[:idx]
	}
	if idx := strings.IndexByte(raw, ':'); idx >= 0 {
		raw = raw[:idx]
	}
	return raw
}

// NormalizeHostname lowercases a domain and strips the www. prefix.
func NormalizeHostname(domain string) string {
	domain = strings.ToLower(strings.TrimSpace(domain))
	domain = strings.TrimPrefix(domain, "www.")
	return domain
}

// isValidHostname returns true when domain consists only of characters that are
// legal in a DNS hostname (ASCII letters, digits, dots, and hyphens). This
// rejects path-traversal sequences such as "../" before the value is used as
// part of a file-system path or URL.
func isValidHostname(domain string) bool {
	if domain == "" {
		return false
	}
	for _, r := range domain {
		if !((r >= 'a' && r <= 'z') || (r >= '0' && r <= '9') || r == '.' || r == '-') {
			return false
		}
	}
	return true
}

func (c *Cache) cachePath(domain string) string {
	// filepath.Join cleans the path, but we must ensure the result stays within
	// the cache directory to prevent path traversal.
	p := filepath.Join(c.dir, domain+".png")
	if !strings.HasPrefix(p, filepath.Clean(c.dir)+string(filepath.Separator)) && p != filepath.Clean(c.dir) {
		return filepath.Join(c.dir, "invalid.png")
	}
	return p
}

// ServeHTTP handles a favicon request. It serves from cache if available,
// fetches from the remote service on cache miss, and returns a default
// fallback icon on failure.
func (c *Cache) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	domain := NormalizeHostname(r.URL.Query().Get("domain"))
	if domain == "" || !isValidHostname(domain) {
		http.NotFound(w, r)
		return
	}

	// Check failure backoff — don't hammer failing hosts.
	c.mu.RLock()
	failAt, failed := c.failures[domain]
	c.mu.RUnlock()
	if failed && time.Since(failAt) < FailureRetryAfter {
		c.serveDefault(w)
		return
	}

	path := c.cachePath(domain)

	// Cache hit.
	if info, err := os.Stat(path); err == nil {
		metrics.M.FaviconCacheHits.Inc()
		http.ServeFile(w, r, path)
		// Stale — refresh in background, but don't block the response.
		if time.Since(info.ModTime()) > CacheTTL {
			go c.refresh(domain, path)
		}
		return
	}

	// Cache miss — fetch synchronously.
	metrics.M.FaviconCacheMisses.Inc()
	if err := c.fetchAndSave(domain, path); err != nil {
		c.mu.Lock()
		c.failures[domain] = time.Now()
		c.mu.Unlock()
		metrics.M.FaviconFetchFailures.Inc()
		c.serveDefault(w)
		return
	}
	metrics.M.FaviconCacheSize.Inc()
	http.ServeFile(w, r, path)
}

// refresh fetches a favicon in the background and updates the cache.
func (c *Cache) refresh(domain, path string) {
	if err := c.fetchAndSave(domain, path); err != nil {
		c.mu.Lock()
		c.failures[domain] = time.Now()
		c.mu.Unlock()
		metrics.M.FaviconFetchFailures.Inc()
	}
}

// fetchAndSave downloads a favicon from the remote service and writes it
// to the cache file atomically (write to .tmp, then rename).
func (c *Cache) fetchAndSave(domain, path string) error {
	// domain has already been validated by isValidHostname (alphanumeric, dots, hyphens only).
	fetchURL := fmt.Sprintf("%s/%s?size=64&format=png", RemoteServiceURL, domain)
	resp, err := c.client.Get(fetchURL)
	if err != nil {
		return fmt.Errorf("fetch %s: %w", fetchURL, err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("fetch %s: unexpected status %d", fetchURL, resp.StatusCode)
	}

	tmpPath := path + ".tmp"
	f, err := os.Create(tmpPath)
	if err != nil {
		return fmt.Errorf("create temp %s: %w", tmpPath, err)
	}

	if _, err := io.Copy(f, resp.Body); err != nil {
		f.Close()
		os.Remove(tmpPath)
		return fmt.Errorf("write %s: %w", tmpPath, err)
	}
	f.Close()

	if err := os.Rename(tmpPath, path); err != nil {
		os.Remove(tmpPath)
		return fmt.Errorf("rename %s -> %s: %w", tmpPath, path, err)
	}
	return nil
}

// serveDefault writes a minimal SVG circle as a fallback favicon.
func (c *Cache) serveDefault(w http.ResponseWriter) {
	w.Header().Set("Content-Type", "image/svg+xml")
	w.Header().Set("Cache-Control", "public, max-age=3600")
	w.Write([]byte(`<svg xmlns="http://www.w3.org/2000/svg" viewBox="0 0 24 24" fill="%2394a3b8"><circle cx="12" cy="12" r="8"/></svg>`))
}
