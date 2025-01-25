package server

import (
	"embed"
	"log/slog"
	"net/http"

	"github.com/kodehat/portkey/internal/config"
)

func addRoutes(mux *http.ServeMux, logger *slog.Logger, static embed.FS) {
	// Dev Mode browser reload
	if config.C.DevMode {
		logger.Info("registering dev mode", "devMode", true)
		mux.HandleFunc(config.C.ContextPath+"/reload", devModeHandler{logger}.handle())
	}

	// Home
	mux.HandleFunc(config.C.ContextPath+"/", homeHandler())

	// Dynamic portals
	portalHandler := portalHandler{logger}
	for _, portalHandler := range portalHandler.handle() {
		mux.HandleFunc(config.C.ContextPath+portalHandler.portalPath, portalHandler.handlerFunc)
	}

	// Dynamic pages
	for _, pageHandler := range pageHandler() {
		mux.HandleFunc(config.C.ContextPath+pageHandler.pagePath, pageHandler.handlerFunc)
	}

	// Fix pages
	mux.HandleFunc(config.C.ContextPath+"/version", versionHandler())
	mux.HandleFunc(config.C.ContextPath+"/healthz", healthHandler())

	// Static
	mux.HandleFunc(config.C.ContextPath+"/static/", staticHandler(static))

	// htmx
	mux.HandleFunc(config.C.ContextPath+"/_/portals", searchHandler{logger}.handle())

	// REST
	mux.HandleFunc(config.C.ContextPath+"/api/portals", portalsRestHandler())
	mux.HandleFunc(config.C.ContextPath+"/api/pages", pagesRestHandler())
}

func addMetricRoutes(mux *http.ServeMux) {
	mux.HandleFunc("/metrics", metricsHandler())
}
