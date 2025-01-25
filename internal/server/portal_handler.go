package server

import (
	"log/slog"
	"net/http"

	"github.com/kodehat/portkey/internal/config"
	"github.com/kodehat/portkey/internal/metrics"
)

type portalHandler struct {
	logger *slog.Logger
}

type portalHandlerInfo struct {
	portalPath  string
	handlerFunc func(w http.ResponseWriter, r *http.Request)
}

func (p portalHandler) handle() []portalHandlerInfo {
	portalHandlerInfos := make([]portalHandlerInfo, 0)
	for _, portal := range config.C.Portals {
		if portal.IsExternal() {
			fixedTitleUrl := portal.TitleForUrl()
			if fixedTitleUrl != portal.Title {
				slog.Debug("fixed title", "old", portal.Title, "new", fixedTitleUrl)
			}
			if config.C.EnableMetrics {
				// Required to initialize all portal metrics to "0".
				metrics.M.PortalHitCounter.WithLabelValues(portal.Title).Add(0)
			}
			portalHandlerInfos = append(portalHandlerInfos, portalHandlerInfo{
				portalPath: "/" + fixedTitleUrl,
				handlerFunc: func(w http.ResponseWriter, r *http.Request) {
					if config.C.EnableMetrics {
						metrics.M.PortalHitCounter.WithLabelValues(portal.Title).Inc()
					}
					p.logger.Debug("opening portal", "link", portal.Link)
					http.Redirect(w, r, portal.Link, http.StatusTemporaryRedirect)
				},
			})
		}
	}
	return portalHandlerInfos
}
