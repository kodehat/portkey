package server

import (
	"embed"
	"log/slog"
	"net/http"
)

func NewServer(
	logger *slog.Logger,
	static embed.FS,
) http.Handler {
	mux := http.NewServeMux()
	addRoutes(mux, logger, static)
	return mux
}

func NewMetricsServer(
	logger *slog.Logger,
) http.Handler {
	mux := http.NewServeMux()
	addMetricRoutes(mux)
	return mux
}
