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
	if config.C.Server.DevMode {
		logger.Info("registering dev mode", "devMode", true)
		devModeParams := livereload.NewParams(
			livereload.WithContextPath(config.C.Server.ContextPath),
			livereload.WithReloadPath(devModeReloadPath),
		)
		mux.HandleFunc(config.C.Server.ContextPath+devModeReloadPath, livereload.Handler(devModeParams))
	}

	// Home
	mux.HandleFunc(config.C.Server.ContextPath+"/", homeHandler())

	// Dynamic portals
	ph := portalHandler{logger}
	for _, info := range ph.handle() {
		mux.HandleFunc(config.C.Server.ContextPath+info.portalPath, info.handlerFunc)
	}

	// Dynamic pages
	for _, pageHandler := range pageHandler() {
		mux.HandleFunc(config.C.Server.ContextPath+pageHandler.pagePath, pageHandler.handlerFunc)
	}

	// Fix pages
	mux.HandleFunc(config.C.Server.ContextPath+"/version", versionHandler())
	mux.HandleFunc(config.C.Server.ContextPath+"/healthz", healthHandler())

	// Static
	mux.HandleFunc(config.C.Server.ContextPath+"/static/", staticHandler(static))

	// Custom icons directory (served from disk, Docker-mountable)
	if config.C.Favicon.CustomIconsDir != "" {
		fs := http.FileServer(http.Dir(config.C.Favicon.CustomIconsDir))
		mux.Handle(config.C.Server.ContextPath+"/_/icons/", http.StripPrefix(config.C.Server.ContextPath+"/_/icons/", fs))
	}

	// Favicon cache
	mux.HandleFunc(config.C.Server.ContextPath+"/_/favicon", faviconHandler{}.handle())

	// htmx
	mux.HandleFunc(config.C.Server.ContextPath+"/_/portals", searchHandler{logger: logger, levenshtein: metrics.NewLevenshtein()}.handle())

	// REST
	mux.HandleFunc(config.C.Server.ContextPath+"/api/portals", portalsRestHandler())
	mux.HandleFunc(config.C.Server.ContextPath+"/api/pages", pagesRestHandler())
}

func addMetricRoutes(mux *http.ServeMux) {
	mux.HandleFunc("/metrics", metricsHandler())
}
