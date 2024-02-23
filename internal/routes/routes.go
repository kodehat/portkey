package routes

import (
	"embed"
	"net/http"

	"github.com/a-h/templ"
	"github.com/kodehat/portkey/internal/components"
	"github.com/kodehat/portkey/internal/config"
)

func AddRoutes(mux *http.ServeMux, static embed.FS) {
	// Home
	mux.HandleFunc("/", homeHandler())

	// Custom pages
	for _, pageHandler := range pageHandler() {
		mux.HandleFunc(pageHandler.pagePath, pageHandler.handlerFunc)
	}

	// Fix pages
	mux.HandleFunc("/version", versionHandler())

	// Static
	mux.HandleFunc("/static/", staticHandler(static))

	// htmx
	mux.HandleFunc("/_/portals", portalsHandler())

	// REST
	mux.HandleFunc("/api/portals", portalsRestHandler())
	mux.HandleFunc("/api/pages", pagesRestHandler())
}

func getAllFooterPortals() []templ.Component {
	var allFooterPortals = make([]templ.Component, len(config.C.Portals))
	for i, configPortal := range config.C.Portals {
		allFooterPortals[i] = components.FooterPortal(configPortal.Link, configPortal.Emoji, configPortal.Title, configPortal.External)
	}
	return allFooterPortals
}
