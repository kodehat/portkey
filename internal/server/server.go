package server

import (
	"embed"
	"log/slog"
	"net/http"

	"github.com/kodehat/portkey/internal/config"
)

func NewServer(
	logger *slog.Logger,
	config *config.Config,
	static embed.FS,
) http.Handler {
	mux := http.NewServeMux()
	addRoutes(mux, logger, static)
	var handler http.Handler = mux
	return handler
}
