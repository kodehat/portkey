package server

import (
	"net/http"

	"github.com/kodehat/portkey/internal/favicon"
)

// faviconHandler serves cached favicons via the global favicon cache.
type faviconHandler struct{}

func (faviconHandler) handle() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
			return
		}
		// Set cache headers so browsers cache the favicon response.
		w.Header().Set("Cache-Control", "public, max-age=86400")
		favicon.C.ServeHTTP(w, r)
	}
}
