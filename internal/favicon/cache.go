// Package favicon provides a disk-backed cache for favicons discovered from
// target websites. Favicons are cached by normalized hostname and served from
// local disk on subsequent requests.
package favicon

import (
	"context"
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"github.com/kodehat/favifetch"

	"github.com/kodehat/portkey/internal/config"
	"github.com/kodehat/portkey/internal/metrics"
)

const (
	// CacheTTL is how long a cached favicon is considered fresh.
	CacheTTL = 7 * 24 * time.Hour

	// FailureRetryAfter is how long to wait before retrying a failed domain.
	FailureRetryAfter = 1 * time.Hour

	// HttpsPrefix is the prefix used for fetching favicons from target websites.
	HttpsPrefix = "https://"

	// ContentTypeHeader is the HTTP header for content type.
	ContentTypeHeader = "Content-Type"

	// MimeTypePng is the MIME type for PNG images.
	MimeTypePng = "image/png"
)

// C is the global favicon cache. Initialized by Init().
var C *Cache

// Cache caches favicons on disk by normalized hostname.
type Cache struct {
	dir      string
	client   *http.Client
	logger   *slog.Logger
	failures map[string]time.Time
	mu       sync.RWMutex
}

// Init initializes the global favicon cache with the given disk directory and
// logger. Must be called before C is used (typically during application startup).
func Init(cacheDir string, logger *slog.Logger) {
	C = New(cacheDir, logger)
}

// New creates a new Cache with the given disk directory. Pass nil for logger
// to suppress all log output (useful in tests).
func New(cacheDir string, logger *slog.Logger) *Cache {
	if err := os.MkdirAll(cacheDir, 0755); err != nil {
		panic(fmt.Errorf("failed to create favicon cache directory %s: %w", cacheDir, err))
	}
	return &Cache{
		dir:      cacheDir,
		client:   &http.Client{Timeout: 10 * time.Second},
		logger:   logger,
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
	raw := strings.TrimPrefix(rawURL, HttpsPrefix)
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

// fetchOptions returns the standard favifetch options applied to every request.
func (c *Cache) fetchOptions() []favifetch.Option {
	return []favifetch.Option{
		favifetch.WithHTTPClient(c.client),
		favifetch.WithFormat(favifetch.TargetPNG),
		favifetch.WithSize(64),
	}
}

// ServeHTTP handles a favicon request. It serves from cache if available,
// discovers favicons from the target website on cache miss, and returns a
// default fallback icon on failure.
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
		c.logDebug("favicon backoff, serving default", "domain", domain, "retryIn", FailureRetryAfter-time.Since(failAt))
		c.serveDefault(w)
		return
	}

	path := c.cachePath(domain)

	// When caching is disabled, fetch and serve without touching disk.
	if config.C.FaviconCacheDisabled {
		c.logDebug("favicon cache disabled, fetching directly", "domain", domain)
		c.serveDirect(w, r, domain)
		return
	}

	// Cache hit.
	if info, err := os.Stat(path); err == nil {
		metrics.M.FaviconCacheHits.Inc()
		c.logDebug("favicon cache hit", "domain", domain, "age", time.Since(info.ModTime()))
		w.Header().Set(ContentTypeHeader, MimeTypePng)
		http.ServeFile(w, r, path)
		// Stale — refresh in background, but don't block the response.
		if time.Since(info.ModTime()) > CacheTTL {
			c.logDebug("favicon cache stale, refreshing in background", "domain", domain)
			go c.refresh(domain, path)
		}
		return
	}

	// Cache miss — fetch synchronously.
	c.logDebug("favicon cache miss, fetching", "domain", domain)
	metrics.M.FaviconCacheMisses.Inc()
	if err := c.fetchAndSave(r.Context(), domain, path); err != nil {
		c.logWarn("favicon fetch failed, serving default", "domain", domain, "error", err)
		c.mu.Lock()
		c.failures[domain] = time.Now()
		c.mu.Unlock()
		metrics.M.FaviconFetchFailures.Inc()
		c.serveDefault(w)
		return
	}
	c.logDebug("favicon cached successfully", "domain", domain)
	metrics.M.FaviconCacheSize.Inc()
	w.Header().Set(ContentTypeHeader, MimeTypePng)
	http.ServeFile(w, r, path)
}

// serveDirect discovers, converts to PNG, and serves a favicon without touching
// the disk cache. Used when FaviconCacheDisabled is true.
func (c *Cache) serveDirect(w http.ResponseWriter, r *http.Request, domain string) {
	result, err := favifetch.Fetch(r.Context(), HttpsPrefix+domain, c.fetchOptions()...)
	if err != nil {
		c.logWarn("favicon direct fetch failed, serving default", "domain", domain, "error", err)
		c.mu.Lock()
		c.failures[domain] = time.Now()
		c.mu.Unlock()
		metrics.M.FaviconFetchFailures.Inc()
		c.serveDefault(w)
		return
	}

	c.logDebug("favicon served directly", "domain", domain, "size", result.Size, "source", result.Source)
	w.Header().Set(ContentTypeHeader, MimeTypePng)
	w.Header().Set("Cache-Control", "public, max-age=86400")
	w.Write(result.Data)
}

// refresh fetches a favicon in the background and updates the cache.
func (c *Cache) refresh(domain, path string) {
	c.logDebug("favicon background refresh started", "domain", domain)
	if err := c.fetchAndSave(context.Background(), domain, path); err != nil {
		c.logWarn("favicon background refresh failed", "domain", domain, "error", err)
		c.mu.Lock()
		c.failures[domain] = time.Now()
		c.mu.Unlock()
		metrics.M.FaviconFetchFailures.Inc()
	} else {
		c.logDebug("favicon background refresh complete", "domain", domain)
	}
}

// fetchAndSave discovers, converts to PNG, and writes the favicon atomically to
// the cache file.
func (c *Cache) fetchAndSave(ctx context.Context, domain, path string) error {
	// domain has already been validated by isValidHostname (alphanumeric, dots, hyphens only).
	result, err := favifetch.Fetch(ctx, HttpsPrefix+domain, c.fetchOptions()...)
	if err != nil {
		c.logWarn("favifetch fetch failed", "domain", domain, "error", err)
		return fmt.Errorf("fetch favicon for %s: %w", domain, err)
	}

	c.logDebug("favifetch succeeded", "domain", domain, "format", result.Format, "source", result.Source, "size", result.Size)

	tmpPath := path + ".tmp"
	f, err := os.Create(tmpPath)
	if err != nil {
		return fmt.Errorf("create temp %s: %w", tmpPath, err)
	}

	if _, err := f.Write(result.Data); err != nil {
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

// logDebug logs a debug message if the cache has a logger configured.
func (c *Cache) logDebug(msg string, args ...any) {
	if c.logger != nil && c.logger.Enabled(context.TODO(), slog.LevelDebug) {
		c.logger.Debug(msg, args...)
	}
}

// logWarn logs a warning message if the cache has a logger configured.
func (c *Cache) logWarn(msg string, args ...any) {
	if c.logger != nil {
		c.logger.Warn(msg, args...)
	}
}

// serveDefault writes a "no entry" SVG as a fallback favicon.
func (c *Cache) serveDefault(w http.ResponseWriter) {
	w.Header().Set(ContentTypeHeader, "image/svg+xml")
	w.Header().Set("Cache-Control", "public, max-age=3600")
	w.Write([]byte(`<svg xmlns="http://www.w3.org/2000/svg" fill="none" viewBox="0 0 24 24" stroke-width="1.5" stroke="%2394a3b8"><path stroke-linecap="round" stroke-linejoin="round" d="M18.364 18.364A9 9 0 0 0 5.636 5.636m12.728 12.728A9 9 0 0 1 5.636 5.636m12.728 12.728L5.636 5.636"/></svg>`))
}
