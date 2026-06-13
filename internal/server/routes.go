package server

import (
	"embed"
	"log/slog"
	"net/http"

	"github.com/adrg/strutil/metrics"
	"github.com/kodehat/livereload"
	"github.com/kodehat/portkey/internal/config"
)

const (
	devModeReloadPath = "/reload"
)

func addRoutes(mux *http.ServeMux, logger *slog.Logger, static embed.FS) {
	// Dev Mode browser reload
	if config.C.DevMode {
		logger.Info("registering dev mode", "devMode", true)
		devModeParams := livereload.NewParams(
			livereload.WithContextPath(config.C.ContextPath),
			livereload.WithReloadPath(devModeReloadPath),
		)
		mux.HandleFunc(config.C.ContextPath+devModeReloadPath, livereload.Handler(devModeParams))
	}

	// Home
	mux.HandleFunc(config.C.ContextPath+"/", homeHandler())

	// Dynamic portals
	ph := portalHandler{logger}
	for _, info := range ph.handle() {
		mux.HandleFunc(config.C.ContextPath+info.portalPath, info.handlerFunc)
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

	// Custom icons directory (served from disk, Docker-mountable)
	if config.C.CustomIconsDir != "" {
		fs := http.FileServer(http.Dir(config.C.CustomIconsDir))
		mux.Handle(config.C.ContextPath+"/_/icons/", http.StripPrefix(config.C.ContextPath+"/_/icons/", fs))
	}

	// Favicon cache
	mux.HandleFunc(config.C.ContextPath+"/_/favicon", faviconHandler{}.handle())

	// htmx
	mux.HandleFunc(config.C.ContextPath+"/_/portals", searchHandler{logger: logger, levenshtein: metrics.NewLevenshtein()}.handle())

	// REST
	mux.HandleFunc(config.C.ContextPath+"/api/portals", portalsRestHandler())
	mux.HandleFunc(config.C.ContextPath+"/api/pages", pagesRestHandler())
}

func addMetricRoutes(mux *http.ServeMux) {
	mux.HandleFunc("/metrics", metricsHandler())
}
