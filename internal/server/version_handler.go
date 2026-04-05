package server

import (
	"net/http"

	"github.com/a-h/templ"
	"github.com/kodehat/portkey/internal/build"
	"github.com/kodehat/portkey/internal/components"
	"github.com/kodehat/portkey/internal/config"
)

const title string = "Version"

func versionHandler() http.HandlerFunc {
	return templ.Handler(components.ContentLayout(title, "", config.C, build.B, components.Version(build.B))).ServeHTTP
}
