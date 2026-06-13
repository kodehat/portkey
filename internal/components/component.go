package components

import (
	"net/url"
	"strings"

	"github.com/kodehat/portkey/internal/config"
)

//go:generate go tool templ generate .

// domainFromURL extracts the hostname from an absolute URL.
// Returns empty string for relative URLs.
func domainFromURL(rawURL string) string {
	if !strings.HasPrefix(rawURL, "http") {
		return ""
	}
	u, err := url.Parse(rawURL)
	if err != nil {
		return ""
	}
	return u.Hostname()
}

// isIconImage returns true if the icon value should be rendered as an <img> tag
// (absolute URL, relative path, or data URI) rather than as emoji text.
func isIconImage(icon string) bool {
	return strings.HasPrefix(icon, "http") || strings.HasPrefix(icon, "/") || strings.HasPrefix(icon, "data:")
}

// faviconURL returns a favicon URL for the given link.
// When caching is enabled (default), returns a local cache URL handled by
// the favicon cache handler. When disabled, returns a direct remote URL.
func faviconURL(rawURL string) string {
	domain := domainFromURL(rawURL)
	if domain == "" {
		return ""
	}
	if config.C.FaviconCacheDisabled {
		base := strings.TrimRight(config.C.FaviconServiceURL, "/")
		return base + "/" + domain + "?size=64&format=png"
	}
	return config.C.ContextPath + "/_/favicon?domain=" + domain
}

// gridClasses holds Tailwind class strings for 1-12 column grid layouts.
// Each entry appears as a literal so Tailwind v4 detects the classes.
var gridClasses = [...]string{
	1:  "max-md:flex max-md:flex-col max-md:gap-5 md:grid md:grid-cols-1 md:gap-8 w-full",
	2:  "max-md:flex max-md:flex-col max-md:gap-5 md:grid md:grid-cols-2 md:gap-8 w-full",
	3:  "max-md:flex max-md:flex-col max-md:gap-5 md:grid md:grid-cols-3 md:gap-8 w-full",
	4:  "max-md:flex max-md:flex-col max-md:gap-5 md:grid md:grid-cols-4 md:gap-8 w-full",
	5:  "max-md:flex max-md:flex-col max-md:gap-5 md:grid md:grid-cols-5 md:gap-8 w-full",
	6:  "max-md:flex max-md:flex-col max-md:gap-5 md:grid md:grid-cols-6 md:gap-8 w-full",
	7:  "max-md:flex max-md:flex-col max-md:gap-5 md:grid md:grid-cols-7 md:gap-8 w-full",
	8:  "max-md:flex max-md:flex-col max-md:gap-5 md:grid md:grid-cols-8 md:gap-8 w-full",
	9:  "max-md:flex max-md:flex-col max-md:gap-5 md:grid md:grid-cols-9 md:gap-8 w-full",
	10: "max-md:flex max-md:flex-col max-md:gap-5 md:grid md:grid-cols-10 md:gap-8 w-full",
	11: "max-md:flex max-md:flex-col max-md:gap-5 md:grid md:grid-cols-11 md:gap-8 w-full",
	12: "max-md:flex max-md:flex-col max-md:gap-5 md:grid md:grid-cols-12 md:gap-8 w-full",
}

// GridClass returns the CSS class string for a given column count.
// On mobile (<md) falls back to vertical stack. On md+ uses CSS grid.
func GridClass(columns int) string {
	if columns < 1 || columns >= len(gridClasses) {
		return ""
	}
	return gridClasses[columns]
}
