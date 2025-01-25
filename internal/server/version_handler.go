package server

import (
	"net/http"

	"github.com/a-h/templ"
	"github.com/kodehat/portkey/internal/build"
	"github.com/kodehat/portkey/internal/components"
	"github.com/kodehat/portkey/internal/config"
)

const TITLE string = "Version"

func versionHandler() http.HandlerFunc {
	return templ.Handler(components.ContentLayout(TITLE, "", config.C, build.B, components.Version(build.B))).ServeHTTP
}
