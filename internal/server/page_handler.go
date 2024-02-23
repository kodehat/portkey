package server

import (
	"net/http"

	"github.com/a-h/templ"
	"github.com/kodehat/portkey/internal/components"
	"github.com/kodehat/portkey/internal/config"
)

type pageHandlerInfo struct {
	pagePath    string
	handlerFunc func(w http.ResponseWriter, r *http.Request)
}

func pageHandler() []pageHandlerInfo {
	var pageHandlerInfos = make([]pageHandlerInfo, len(config.C.Pages))
	for i, page := range config.C.Pages {
		pageHandlerInfos[i] = pageHandlerInfo{
			pagePath:    page.Path,
			handlerFunc: templ.Handler(components.ContentLayout(page.Heading, config.C, components.ContentPage(page.Content))).ServeHTTP,
		}
	}
	return pageHandlerInfos
}
