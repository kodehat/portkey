package server

import (
	"net/http"

	"github.com/a-h/templ"
	"github.com/kodehat/portkey/internal/build"
	"github.com/kodehat/portkey/internal/components"
	"github.com/kodehat/portkey/internal/config"
)

func homeHandler() http.HandlerFunc {
	home := components.HomePage()
	return func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != config.C.ContextPath+"/" {
			templ.Handler(components.ContentLayout("Oh no", "", config.C, build.B, components.NotFound()), templ.WithStatus(http.StatusNotFound)).ServeHTTP(w, r)
			return
		}
		templ.Handler(components.HomeLayout("Home", config.C, build.B, home)).ServeHTTP(w, r)
	}
}
