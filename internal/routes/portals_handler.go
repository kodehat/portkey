package routes

import (
	"net/http"
	"strings"

	"github.com/a-h/templ"
	"github.com/kodehat/portkey/internal/components"
	"github.com/kodehat/portkey/internal/config"
	"github.com/kodehat/portkey/internal/utils"
)

func portalsHandler() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		query := r.URL.Query().Get("search")
		var allHomePortals = make([]templ.Component, 0)
		for _, configPortal := range config.C.Portals {
			portal := components.HomePortal(configPortal.Link, configPortal.Emoji, configPortal.Title, configPortal.External)
			if query != "" {
				if strings.Contains(configPortal.Title, query) || utils.ArrSubStr(configPortal.Keywords, query) {
					allHomePortals = append(allHomePortals, portal)
				}
			} else {
				allHomePortals = append(allHomePortals, portal)
			}
		}
		components.PortalPartial(allHomePortals).Render(r.Context(), w)
	}
}
