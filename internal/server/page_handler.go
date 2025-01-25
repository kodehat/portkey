package server

import (
	"net/http"

	"github.com/a-h/templ"
	"github.com/kodehat/portkey/internal/build"
	"github.com/kodehat/portkey/internal/components"
	"github.com/kodehat/portkey/internal/config"
	"github.com/kodehat/portkey/internal/metrics"
	"github.com/kodehat/portkey/internal/models"
)

type pageHandlerInfo struct {
	pagePath    string
	handlerFunc func(w http.ResponseWriter, r *http.Request)
}

func pageHandler() []pageHandlerInfo {
	var pageHandlerInfos = make([]pageHandlerInfo, len(config.C.Pages))
	for i, page := range config.C.Pages {
		if config.C.EnableMetrics {
			// Required to initialize all portal metrics to "0".
			metrics.M.PageHitCounter.WithLabelValues(page.Path).Add(0)
		}
		pageHandlerInfos[i] = pageHandlerInfo{
			pagePath:    page.Path,
			handlerFunc: pageMetricsHandler(page, templ.Handler(components.ContentLayout(page.Heading, page.Subtitle, config.C, build.B, components.ContentPage(page.Content)))).ServeHTTP,
		}
	}
	return pageHandlerInfos
}

func pageMetricsHandler(p models.Page, h http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if config.C.EnableMetrics {
			metrics.M.PageHitCounter.WithLabelValues(p.Path).Inc()
		}
		h.ServeHTTP(w, r)
	})
}
