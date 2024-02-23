package routes

import (
	"net/http"
	"sync"

	"github.com/a-h/templ"
	"github.com/kodehat/portkey/internal/components"
	"github.com/kodehat/portkey/internal/config"
	"github.com/kodehat/portkey/internal/utils"
)

func homeHandler() http.HandlerFunc {
	var (
		init             sync.Once
		allFooterPortals []templ.Component
	)
	init.Do(func() {
		allFooterPortals = getAllFooterPortals()
	})
	home := components.HomePage()
	return func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/" {
			w.WriteHeader(http.StatusNotFound)
			templ.Handler(components.ContentLayout(utils.PageTitle("404 Not Found", config.C.Title), "404 Not Found", components.NotFound(), allFooterPortals, config.C.FooterText)).ServeHTTP(w, r)
			return
		}
		templ.Handler(components.HomeLayout(config.C.Title, home)).ServeHTTP(w, r)
	}
}
