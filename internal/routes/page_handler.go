package routes

import (
	"net/http"
	"sync"

	"github.com/a-h/templ"
	"github.com/kodehat/portkey/internal/components"
	"github.com/kodehat/portkey/internal/config"
	"github.com/kodehat/portkey/internal/utils"
)

type pageHandlerInfo struct {
	pagePath    string
	handlerFunc func(w http.ResponseWriter, r *http.Request)
}

func pageHandler() []pageHandlerInfo {
	var (
		init             sync.Once
		allFooterPortals []templ.Component
	)
	init.Do(func() {
		allFooterPortals = getAllFooterPortals()
	})

	var pageHandlerInfos = make([]pageHandlerInfo, len(config.C.Pages))
	for i, page := range config.C.Pages {
		pageHandlerInfos[i] = pageHandlerInfo{
			pagePath:    page.Path,
			handlerFunc: templ.Handler(components.ContentLayout(utils.PageTitle(page.Heading, config.C.Title), page.Heading, components.ContentPage(page.Content), allFooterPortals, config.C.FooterText)).ServeHTTP,
		}
	}
	return pageHandlerInfos
}
