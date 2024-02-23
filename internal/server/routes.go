package server

import (
	"embed"
	"log"
	"net/http"
)

func addRoutes(mux *http.ServeMux, logger *log.Logger, static embed.FS) {
	// Home
	mux.HandleFunc("/", homeHandler())

	// Custom pages
	for _, pageHandler := range pageHandler() {
		mux.HandleFunc(pageHandler.pagePath, pageHandler.handlerFunc)
	}

	// Fix pages
	mux.HandleFunc("/version", versionHandler())
	mux.HandleFunc("/healthz", healthHandler())

	// Static
	mux.HandleFunc("/static/", staticHandler(static))

	// htmx
	mux.HandleFunc("/_/portals", portalsHandler())

	// REST
	mux.HandleFunc("/api/portals", portalsRestHandler())
	mux.HandleFunc("/api/pages", pagesRestHandler())
}
