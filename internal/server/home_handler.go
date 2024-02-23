package server

import (
	"net/http"

	"github.com/a-h/templ"
	"github.com/kodehat/portkey/internal/components"
	"github.com/kodehat/portkey/internal/config"
)

func homeHandler() http.HandlerFunc {
	home := components.HomePage()
	return func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/" {
			w.WriteHeader(http.StatusNotFound)
			templ.Handler(components.ContentLayout("404 Not Found", config.C, components.NotFound())).ServeHTTP(w, r)
			return
		}
		templ.Handler(components.HomeLayout("Home", config.C, home)).ServeHTTP(w, r)
	}
}
