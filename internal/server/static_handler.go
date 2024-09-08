package server

import (
	"embed"
	"net/http"

	"github.com/kodehat/portkey/internal/config"
	"github.com/kodehat/portkey/internal/utils"
)

func staticHandler(static embed.FS) http.HandlerFunc {
	return utils.StaticHandler(http.StripPrefix(config.C.ContextPath, http.FileServer(http.FS(static))))
}
