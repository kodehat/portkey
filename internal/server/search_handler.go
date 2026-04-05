package server

import (
	"log/slog"
	"net/http"
	"strings"

	"github.com/adrg/strutil"
	"github.com/kodehat/portkey/internal/components"
	"github.com/kodehat/portkey/internal/config"
	imetrics "github.com/kodehat/portkey/internal/metrics"
	"github.com/kodehat/portkey/internal/models"
	"github.com/kodehat/portkey/internal/utils"
)

const searchQueryParam = "search"

type searchHandler struct {
	logger      *slog.Logger
	levenshtein strutil.StringMetric
}

func (p searchHandler) handle() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		query := r.URL.Query().Get(searchQueryParam)
		if query == "" && config.R.WithGroups {
			// No search query — render portals grouped by their Group field.
			groups := utils.GroupPortals(config.C.Portals)
			components.GroupedPortalPartial(groups).Render(r.Context(), w)
			return
		}
		homePortals := p.queryHomePortals(query)
		if config.C.EnableMetrics {
			p.increaseMetrics(len(homePortals) > 0)
		}
		components.PortalPartial(homePortals).Render(r.Context(), w)
	}
}

func (p searchHandler) queryHomePortals(query string) []models.Portal {
	if query == "" {
		return config.C.Portals
	}
	homePortals := make([]models.Portal, 0)
	for _, portal := range config.C.Portals {
		if p.isSearchResult(query, portal) {
			homePortals = append(homePortals, portal)
		}
	}
	return homePortals
}

func (p searchHandler) increaseMetrics(hasSearchResults bool) {
	if hasSearchResults {
		imetrics.M.SearchWithResultsCounter.Inc()
	} else {
		imetrics.M.SearchNoResultsCounter.Inc()
	}
}

func (p searchHandler) isSearchResult(query string, portal models.Portal) bool {
	lowerQuery := strings.ToLower(query)
	if strings.Contains(strings.ToLower(portal.Title), lowerQuery) || utils.ArrSubStr(portal.Keywords, lowerQuery) {
		p.logger.Debug("direct match for search found", "query", query, "portal", portal.Title)
		return true
	}

	if !config.C.SearchWithStringSimilarity {
		return false
	}
	p.logger.Debug("searching with string similarity", "query", query)

	similar := p.isSimilar(query, portal.Title, p.levenshtein, config.C.MinimumStringSimilarity)
	if similar {
		return similar
	}
	for _, keyword := range portal.Keywords {
		similar = p.isSimilar(query, keyword, p.levenshtein, config.C.MinimumStringSimilarity)
		if similar {
			return similar
		}
	}
	return false
}

func (p searchHandler) isSimilar(str string, reference string, metric strutil.StringMetric, minimumSimilarity float64) bool {
	similarity := strutil.Similarity(str, reference, metric)
	p.logger.Debug("similarity check", "str", str, "reference", reference, "similarity", similarity)
	return similarity > minimumSimilarity
}
