package server

import (
	"net/http"

	"github.com/kodehat/portkey/internal/config"
	"github.com/kodehat/portkey/internal/favicon"
)

// faviconHandler serves cached favicons via the global favicon cache,
// but only for domains that exist in the configured portals.
type faviconHandler struct{}

func (faviconHandler) handle() http.HandlerFunc {
	// Build a set of allowed hostnames from the configured portals.
	allowed := make(map[string]bool)
	for _, p := range config.C.Portals {
		if p.IsExternal() {
			allowed[favicon.NormalizeHostname(favicon.DomainFromURL(p.Link))] = true
		}
	}

	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
			return
		}

		domain := favicon.NormalizeHostname(r.URL.Query().Get("domain"))
		if domain == "" || !allowed[domain] {
			http.NotFound(w, r)
			return
		}

		w.Header().Set("Cache-Control", "public, max-age=86400")
		favicon.C.ServeHTTP(w, r)
	}
}
