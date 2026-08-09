// Package favicon provides a disk-backed cache for favicons discovered from
// target websites. Favicons are cached by normalized hostname and served from
// local disk on subsequent requests.
package favicon

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"net/url"
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

	// defaultVemetricHost is the favicon service used as the relay target in
	// proxied mode when faviconServiceURL is not configured. It mirrors
	// favifetch's own built-in default fallback host.
	defaultVemetricHost = "favicon.vemetric.com"

	// proxyFaviconSize is the requested icon size (in pixels) when relaying to
	// a favicon service in proxied mode. It matches favifetch's default
	// fallback size.
	proxyFaviconSize = 64

	// maxProxyImageSize caps how many bytes are read from a favicon service
	// response in proxied mode, matching favifetch's MaxImageSize.
	maxProxyImageSize = 5 * 1024 * 1024
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
// When cacheEnabled is false, no cache directory is created and the cache stays
// disk-free (every fetch is served directly).
func Init(cacheDir string, cacheEnabled bool, logger *slog.Logger) {
	C = New(cacheDir, cacheEnabled, logger)
}

// New creates a new Cache with the given disk directory. Pass nil for logger
// to suppress all log output (useful in tests). When cacheEnabled is false, the
// directory is not created (nothing is ever written to disk) and an uncreateable
// cacheDir does not fail startup.
func New(cacheDir string, cacheEnabled bool, logger *slog.Logger) *Cache {
	if cacheEnabled {
		if cacheDir == "" {
			panic("favicon cache is enabled but cacheDir is empty")
		}
		if err := os.MkdirAll(cacheDir, 0755); err != nil {
			panic(fmt.Errorf("failed to create favicon cache directory %s: %w", cacheDir, err))
		}
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

// cachePath returns the on-disk path for a cached favicon with the given format.
func (c *Cache) cachePath(domain string, format favifetch.DetectedFormat) string {
	ext := extensionForFormat(format)
	p := filepath.Join(c.dir, domain+ext)
	// filepath.Join cleans the path, but we must ensure the result stays within
	// the cache directory to prevent path traversal.
	if !strings.HasPrefix(p, filepath.Clean(c.dir)+string(filepath.Separator)) && p != filepath.Clean(c.dir) {
		return filepath.Join(c.dir, "invalid"+ext)
	}
	return p
}

// findCachedPath looks for a cached favicon file with any supported format.
// Formats are tried in preference order (SVG > PNG > ICO > WebP > JPEG > GIF > BMP).
// Returns the file path and detected format, or empty string if nothing is cached.
func (c *Cache) findCachedPath(domain string) (string, favifetch.DetectedFormat) {
	formats := []favifetch.DetectedFormat{
		favifetch.FormatSVG,
		favifetch.FormatPNG,
		favifetch.FormatICO,
		favifetch.FormatWebP,
		favifetch.FormatJPEG,
		favifetch.FormatGIF,
		favifetch.FormatBMP,
	}
	for _, f := range formats {
		p := c.cachePath(domain, f)
		if _, err := os.Stat(p); err == nil {
			return p, f
		}
	}
	return "", favifetch.FormatUnknown
}

// cleanupOtherFormats removes cached favicon files for formats other than the
// one being kept. Best-effort: errors are silently ignored.
func (c *Cache) cleanupOtherFormats(domain string, keep favifetch.DetectedFormat) {
	formats := []favifetch.DetectedFormat{
		favifetch.FormatSVG,
		favifetch.FormatPNG,
		favifetch.FormatICO,
		favifetch.FormatWebP,
		favifetch.FormatJPEG,
		favifetch.FormatGIF,
		favifetch.FormatBMP,
	}
	for _, f := range formats {
		if f == keep {
			continue
		}
		os.Remove(c.cachePath(domain, f))
	}
}

// extensionForFormat returns the file extension (with leading dot) for a detected format.
func extensionForFormat(f favifetch.DetectedFormat) string {
	switch f {
	case favifetch.FormatSVG:
		return ".svg"
	case favifetch.FormatPNG:
		return ".png"
	case favifetch.FormatICO:
		return ".ico"
	case favifetch.FormatWebP:
		return ".webp"
	case favifetch.FormatJPEG:
		return ".jpg"
	case favifetch.FormatGIF:
		return ".gif"
	case favifetch.FormatBMP:
		return ".bmp"
	default:
		return ".png"
	}
}

// mimeForFormat returns the MIME type for a detected format.
func mimeForFormat(f favifetch.DetectedFormat) string {
	switch f {
	case favifetch.FormatSVG:
		return "image/svg+xml"
	case favifetch.FormatPNG:
		return "image/png"
	case favifetch.FormatICO:
		return "image/x-icon"
	case favifetch.FormatWebP:
		return "image/webp"
	case favifetch.FormatJPEG:
		return "image/jpeg"
	case favifetch.FormatGIF:
		return "image/gif"
	case favifetch.FormatBMP:
		return "image/bmp"
	default:
		return "image/png"
	}
}

// fetchOptions returns the standard favifetch options applied to every request.
func (c *Cache) fetchOptions() []favifetch.Option {
	opts := []favifetch.Option{
		// Browser mode returns the regular tab icon a Chromium browser would use.
		// It must not be combined with resizing or format conversion.
		favifetch.WithMode(favifetch.ModeBrowser),
		favifetch.WithFallbackAPI(true),
		favifetch.WithTimeout(5 * time.Second),
		favifetch.WithHTTPClient(c.client),
	}
	// Allow overriding the self-hostable favicon service used as the
	// last-resort fallback. favifetch expects a bare host (HTTPS is implicit);
	// a full URL is reduced to its host so config.yml values such as
	// "https://favicon.vemetric.com" work as-is.
	if host := faviconServiceHost(config.C.Favicon.ServiceURL); host != "" {
		opts = append(opts, favifetch.WithVemetricAPIHost(host))
	}
	return opts
}

// faviconServiceHost extracts the host (with optional port) from the configured
// faviconServiceURL. favifetch's fallback API option expects a bare host and
// always uses HTTPS, so a full URL like "https://favicon.vemetric.com" is
// reduced to "favicon.vemetric.com". A bare host string is returned unchanged
// (a trailing path or query, if present, is stripped). Returns "" when
// serviceURL is empty.
func faviconServiceHost(serviceURL string) string {
	if serviceURL = strings.TrimSpace(serviceURL); serviceURL == "" {
		return ""
	}
	// A scheme separator means it's a full URL — parse it and keep the host.
	if strings.Contains(serviceURL, "://") {
		if u, err := url.Parse(serviceURL); err == nil && u.Host != "" {
			return u.Host
		}
	}
	// Otherwise treat the value as a bare host. Strip any path/query suffix.
	host := serviceURL
	if idx := strings.IndexByte(host, '/'); idx >= 0 {
		host = host[:idx]
	}
	if idx := strings.IndexByte(host, '?'); idx >= 0 {
		host = host[:idx]
	}
	return strings.TrimSpace(host)
}

// fetchedIcon holds the bytes and detected format of a favicon, regardless of
// whether it was obtained via direct discovery (favifetch) or by relaying to
// a favicon service (proxied mode).
type fetchedIcon struct {
	data   []byte
	format favifetch.DetectedFormat
	source string
}

// fetch obtains the favicon for the given domain using the configured
// FaviconMode: direct discovery via favifetch, or by relaying to a
// Vemetric-compatible favicon service.
func (c *Cache) fetch(ctx context.Context, domain string) (fetchedIcon, error) {
	if config.C.Favicon.Mode == config.FaviconModeProxied {
		return c.fetchProxied(ctx, domain)
	}
	return c.fetchDirect(ctx, domain)
}

// fetchDirect discovers and fetches the favicon via the favifetch library,
// which parses the target website's HTML, manifest, and common fallback paths.
func (c *Cache) fetchDirect(ctx context.Context, domain string) (fetchedIcon, error) {
	result, err := favifetch.Fetch(ctx, HttpsPrefix+domain, c.fetchOptions()...)
	if err != nil {
		return fetchedIcon{}, err
	}
	return fetchedIcon{data: result.Data, format: result.Format, source: result.Source}, nil
}

// fetchProxied relays the favicon request to a Vemetric-compatible favicon
// service (configured via faviconServiceURL). The service performs all
// discovery and returns the image bytes; portkey just relays them. The result
// is cached like a direct fetch so subsequent requests are served from disk.
func (c *Cache) fetchProxied(ctx context.Context, domain string) (fetchedIcon, error) {
	host := faviconServiceHost(config.C.Favicon.ServiceURL)
	if host == "" {
		host = defaultVemetricHost
	}
	target := fmt.Sprintf("https://%s/%s?size=%d", host, domain, proxyFaviconSize)

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, target, nil)
	if err != nil {
		return fetchedIcon{}, fmt.Errorf("build proxy request for %s: %w", domain, err)
	}
	resp, err := c.client.Do(req)
	if err != nil {
		return fetchedIcon{}, fmt.Errorf("proxy favicon for %s: %w", domain, err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fetchedIcon{}, fmt.Errorf("proxy favicon for %s: service returned %s", domain, resp.Status)
	}

	data, err := io.ReadAll(io.LimitReader(resp.Body, maxProxyImageSize))
	if err != nil {
		return fetchedIcon{}, fmt.Errorf("read proxied favicon for %s: %w", domain, err)
	}

	format := detectFormatFromContentType(resp.Header.Get(ContentTypeHeader))
	if format == favifetch.FormatUnknown {
		format = detectFormatFromMagic(data)
	}
	if format == favifetch.FormatUnknown {
		format = favifetch.FormatPNG // safest default for caching and serving
	}
	return fetchedIcon{data: data, format: format, source: "proxied"}, nil
}

// detectFormatFromContentType maps a Content-Type header value (including any
// parameters such as "; charset=utf-8") to a DetectedFormat.
func detectFormatFromContentType(ct string) favifetch.DetectedFormat {
	if i := strings.IndexByte(ct, ';'); i >= 0 {
		ct = ct[:i]
	}
	switch strings.ToLower(strings.TrimSpace(ct)) {
	case "image/png", "image/jpg":
		return favifetch.FormatPNG
	case "image/svg+xml":
		return favifetch.FormatSVG
	case "image/x-icon", "image/vnd.microsoft.icon":
		return favifetch.FormatICO
	case "image/webp":
		return favifetch.FormatWebP
	case "image/jpeg":
		return favifetch.FormatJPEG
	case "image/gif":
		return favifetch.FormatGIF
	case "image/bmp":
		return favifetch.FormatBMP
	default:
		return favifetch.FormatUnknown
	}
}

// detectFormatFromMagic sniffs the image format from the leading magic bytes
// of raw image data. It is a fallback for proxied responses that arrive
// without a recognizable Content-Type header.
func detectFormatFromMagic(data []byte) favifetch.DetectedFormat {
	switch {
	case len(data) >= 8 && bytes.HasPrefix(data, []byte{0x89, 'P', 'N', 'G', 0x0D, 0x0A, 0x1A, 0x0A}):
		return favifetch.FormatPNG
	case len(data) >= 3 && bytes.HasPrefix(data, []byte{0xFF, 0xD8, 0xFF}):
		return favifetch.FormatJPEG
	case len(data) >= 12 && bytes.HasPrefix(data, []byte("RIFF")) && bytes.Equal(data[8:12], []byte("WEBP")):
		return favifetch.FormatWebP
	case len(data) >= 6 && (bytes.HasPrefix(data, []byte("GIF87a")) || bytes.HasPrefix(data, []byte("GIF89a"))):
		return favifetch.FormatGIF
	case len(data) >= 4 && bytes.Equal(data[:4], []byte{0x00, 0x00, 0x01, 0x00}):
		return favifetch.FormatICO
	case len(data) >= 2 && bytes.HasPrefix(data, []byte("BM")):
		return favifetch.FormatBMP
	case hasSVGMarker(data):
		return favifetch.FormatSVG
	default:
		return favifetch.FormatUnknown
	}
}

// hasSVGMarker reports whether the data looks like SVG markup by searching for
// an "<svg" tag within the first 512 bytes.
func hasSVGMarker(data []byte) bool {
	prefix := data
	if len(prefix) > 512 {
		prefix = prefix[:512]
	}
	return bytes.Contains(bytes.ToLower(prefix), []byte("<svg"))
}

// ServeHTTP handles a favicon request. It serves from cache if available and
// discovers favicons from the target website on cache miss. Failed requests
// receive a not-found response so the client can render its inline fallback.
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

	// When caching is disabled, fetch and serve without touching disk.
	if !config.C.Favicon.CacheEnabled {
		c.logDebug("favicon cache disabled, fetching directly", "domain", domain)
		c.serveDirect(w, r, domain)
		return
	}

	// Cache hit — findCachedPath tries formats in preference order (SVG first).
	path, format := c.findCachedPath(domain)
	if path != "" {
		info, err := os.Stat(path)
		if err == nil {
			metrics.M.FaviconCacheHits.Inc()
			c.logDebug("favicon cache hit", "domain", domain, "format", format.String(), "age", time.Since(info.ModTime()))
			w.Header().Set(ContentTypeHeader, mimeForFormat(format))
			http.ServeFile(w, r, path)
			// Stale — refresh in background, but don't block the response.
			if time.Since(info.ModTime()) > CacheTTL {
				c.logDebug("favicon cache stale, refreshing in background", "domain", domain)
				go c.refresh(domain)
			}
			return
		}
	}

	// Cache miss — fetch synchronously.
	c.logDebug("favicon cache miss, fetching", "domain", domain)
	metrics.M.FaviconCacheMisses.Inc()
	savedPath, savedFormat, err := c.fetchAndSave(r.Context(), domain)
	if err != nil {
		c.logWarn("favicon fetch failed, serving default", "domain", domain, "error", err)
		c.mu.Lock()
		c.failures[domain] = time.Now()
		c.mu.Unlock()
		metrics.M.FaviconFetchFailures.Inc()
		c.serveDefault(w)
		return
	}
	c.logDebug("favicon cached successfully", "domain", domain, "format", savedFormat.String())
	metrics.M.FaviconCacheSize.Inc()
	w.Header().Set(ContentTypeHeader, mimeForFormat(savedFormat))
	http.ServeFile(w, r, savedPath)
}

// serveDirect discovers, fetches, and serves a favicon without touching the disk
// cache. Used when FaviconCacheDisabled is true.
func (c *Cache) serveDirect(w http.ResponseWriter, r *http.Request, domain string) {
	icon, err := c.fetch(r.Context(), domain)
	if err != nil {
		c.logWarn("favicon direct fetch failed, serving default", "domain", domain, "error", err)
		c.mu.Lock()
		c.failures[domain] = time.Now()
		c.mu.Unlock()
		metrics.M.FaviconFetchFailures.Inc()
		c.serveDefault(w)
		return
	}

	c.logDebug("favicon served directly", "domain", domain, "format", icon.format.String(), "size", len(icon.data), "source", icon.source)
	w.Header().Set(ContentTypeHeader, mimeForFormat(icon.format))
	w.Header().Set("Cache-Control", "public, max-age=86400")
	w.Write(icon.data)
}

// refresh fetches a favicon in the background and updates the cache.
func (c *Cache) refresh(domain string) {
	c.logDebug("favicon background refresh started", "domain", domain)
	if _, _, err := c.fetchAndSave(context.Background(), domain); err != nil {
		c.logWarn("favicon background refresh failed", "domain", domain, "error", err)
		c.mu.Lock()
		c.failures[domain] = time.Now()
		c.mu.Unlock()
		metrics.M.FaviconFetchFailures.Inc()
	} else {
		c.logDebug("favicon background refresh complete", "domain", domain)
	}
}

// fetchAndSave discovers, fetches, and writes the favicon atomically to the
// cache file. Returns the saved file path and the detected format.
func (c *Cache) fetchAndSave(ctx context.Context, domain string) (string, favifetch.DetectedFormat, error) {
	// domain has already been validated by isValidHostname (alphanumeric, dots, hyphens only).
	icon, err := c.fetch(ctx, domain)
	if err != nil {
		c.logWarn("favicon fetch failed", "domain", domain, "error", err)
		return "", favifetch.FormatUnknown, fmt.Errorf("fetch favicon for %s: %w", domain, err)
	}

	c.logDebug("favicon fetched", "domain", domain, "format", icon.format.String(), "source", icon.source, "size", len(icon.data))

	// Clean up old files of other formats before writing the new one.
	c.cleanupOtherFormats(domain, icon.format)

	path := c.cachePath(domain, icon.format)
	tmpPath := path + ".tmp"
	f, err := os.Create(tmpPath)
	if err != nil {
		return "", favifetch.FormatUnknown, fmt.Errorf("create temp %s: %w", tmpPath, err)
	}

	if _, err := f.Write(icon.data); err != nil {
		f.Close()
		os.Remove(tmpPath)
		return "", favifetch.FormatUnknown, fmt.Errorf("write %s: %w", tmpPath, err)
	}
	f.Close()

	if err := os.Rename(tmpPath, path); err != nil {
		os.Remove(tmpPath)
		return "", favifetch.FormatUnknown, fmt.Errorf("rename %s -> %s: %w", tmpPath, path, err)
	}
	return path, icon.format, nil
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

// serveDefault signals that no favicon is available. The portal image's error
// handler then replaces it with the inline fallback SVG.
func (c *Cache) serveDefault(w http.ResponseWriter) {
	w.Header().Set("Cache-Control", "no-store")
	http.Error(w, http.StatusText(http.StatusNotFound), http.StatusNotFound)
}
