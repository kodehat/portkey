package server

import (
	"embed"
	"log"
	"net/http"

	"github.com/kodehat/portkey/internal/config"
)

func addRoutes(mux *http.ServeMux, logger *log.Logger, static embed.FS) {
	// Home
	mux.HandleFunc(config.C.ContextPath+"/", homeHandler())

	// Custom pages
	for _, pageHandler := range pageHandler() {
		mux.HandleFunc(pageHandler.pagePath, pageHandler.handlerFunc)
	}

	// Fix pages
	mux.HandleFunc(config.C.ContextPath+"/version", versionHandler())
	mux.HandleFunc(config.C.ContextPath+"/healthz", healthHandler())

	// Static
	mux.HandleFunc(config.C.ContextPath+"/static/", staticHandler(static))

	// htmx
	mux.HandleFunc(config.C.ContextPath+"/_/portals", portalsHandler{logger}.handle())

	// REST
	mux.HandleFunc(config.C.ContextPath+"/api/portals", portalsRestHandler())
	mux.HandleFunc(config.C.ContextPath+"/api/pages", pagesRestHandler())
}
