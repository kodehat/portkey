package routes

import (
	"net/http"
	"sync"

	"github.com/a-h/templ"
	"github.com/kodehat/portkey/internal/build"
	"github.com/kodehat/portkey/internal/components"
	"github.com/kodehat/portkey/internal/config"
	"github.com/kodehat/portkey/internal/utils"
)

func versionHandler() http.HandlerFunc {
	var (
		init             sync.Once
		allFooterPortals []templ.Component
	)
	init.Do(func() {
		allFooterPortals = getAllFooterPortals()
	})

	return templ.Handler(components.ContentLayout(utils.PageTitle("Version", config.C.Title), "Version", components.Version(build.B), allFooterPortals, config.C.FooterText)).ServeHTTP
}
