package server

import (
	"net/http"

	"github.com/a-h/templ"
	"github.com/kodehat/portkey/internal/build"
	"github.com/kodehat/portkey/internal/components"
	"github.com/kodehat/portkey/internal/config"
)

func versionHandler() http.HandlerFunc {
	return templ.Handler(components.ContentLayout("Version", config.C, components.Version(build.B))).ServeHTTP
}
