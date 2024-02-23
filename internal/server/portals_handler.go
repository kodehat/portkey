package server

import (
	"net/http"
	"strings"

	"github.com/kodehat/portkey/internal/components"
	"github.com/kodehat/portkey/internal/config"
	"github.com/kodehat/portkey/internal/models"
	"github.com/kodehat/portkey/internal/utils"
)

func portalsHandler() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		query := r.URL.Query().Get("search")
		var homePortals = make([]models.Portal, 0)
		for _, configPortal := range config.C.Portals {
			if query != "" {
				if strings.Contains(configPortal.Title, query) || utils.ArrSubStr(configPortal.Keywords, query) {
					homePortals = append(homePortals, configPortal)
				}
			} else {
				homePortals = append(homePortals, configPortal)
			}
		}
		components.PortalPartial(homePortals).Render(r.Context(), w)
	}
}
