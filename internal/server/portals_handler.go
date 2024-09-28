package server

import (
	"log/slog"
	"net/http"
	"strings"

	"github.com/adrg/strutil"
	"github.com/adrg/strutil/metrics"
	"github.com/kodehat/portkey/internal/components"
	"github.com/kodehat/portkey/internal/config"
	"github.com/kodehat/portkey/internal/models"
	"github.com/kodehat/portkey/internal/utils"
)

type portalsHandler struct {
	logger *slog.Logger
}

func (p portalsHandler) handle() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		query := r.URL.Query().Get("search")
		var homePortals = make([]models.Portal, 0)
		for _, configPortal := range config.C.Portals {
			if query != "" {
				if p.isSearchResult(query, configPortal) {
					homePortals = append(homePortals, configPortal)
				}
			} else {
				homePortals = append(homePortals, configPortal)
			}
		}
		components.PortalPartial(homePortals).Render(r.Context(), w)
	}
}

func (p portalsHandler) isSearchResult(query string, portal models.Portal) bool {
	if strings.Contains(portal.Title, query) || utils.ArrSubStr(portal.Keywords, query) {
		p.logger.Debug("direct match for search found", "query", query, "portal", portal.Title)
		return true
	}

	if !config.C.SearchWithStringSimilarity {
		return false
	}
	p.logger.Debug("searching with string similarity", "query", query)

	levenshteinMetric := metrics.NewLevenshtein()
	similar := p.isSimilar(query, portal.Title, levenshteinMetric, config.C.MinimumStringSimilarity)
	if similar {
		return similar
	}
	for _, keyword := range portal.Keywords {
		similar = p.isSimilar(query, keyword, levenshteinMetric, config.C.MinimumStringSimilarity)
		if similar {
			return similar
		}
	}
	return false
}

func (p portalsHandler) isSimilar(str string, reference string, metric strutil.StringMetric, minimumSimilarity float64) bool {
	similarity := strutil.Similarity(str, reference, metric)
	p.logger.Debug("similarity check", "str", str, "reference", reference, "similarity", similarity)
	return similarity > minimumSimilarity
}
