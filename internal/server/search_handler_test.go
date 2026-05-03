package server

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/adrg/strutil/metrics"
	"github.com/kodehat/portkey/internal/config"
	"github.com/kodehat/portkey/internal/models"
)

func TestSearchHandlerNoQuery(t *testing.T) {
	setupServer()
	config.C.Portals = []models.Portal{
		{Title: "GitHub", Link: "https://github.com"},
		{Title: "Google", Link: "https://google.com"},
	}
	config.R.WithGroups = true

	sh := searchHandler{logger: testLogger(), levenshtein: metrics.NewLevenshtein()}
	req := httptest.NewRequest(http.MethodGet, "/_/portals", nil)
	rec := httptest.NewRecorder()
	sh.handle().ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rec.Code)
	}
}

func TestSearchHandlerWithQuery(t *testing.T) {
	setupServer()
	config.C.Portals = []models.Portal{
		{Title: "GitHub", Link: "https://github.com", Keywords: []string{"code", "git"}},
		{Title: "Google", Link: "https://google.com"},
	}

	sh := searchHandler{logger: testLogger(), levenshtein: metrics.NewLevenshtein()}
	req := httptest.NewRequest(http.MethodGet, "/_/portals?search=GitHub", nil)
	rec := httptest.NewRecorder()
	sh.handle().ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rec.Code)
	}
}

func TestSearchHandlerNoMatch(t *testing.T) {
	setupServer()
	config.C.Portals = []models.Portal{
		{Title: "GitHub", Link: "https://github.com"},
	}

	sh := searchHandler{logger: testLogger(), levenshtein: metrics.NewLevenshtein()}
	req := httptest.NewRequest(http.MethodGet, "/_/portals?search=zzzznotfound", nil)
	rec := httptest.NewRecorder()
	sh.handle().ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rec.Code)
	}
}

func TestSearchHandlerKeywordMatch(t *testing.T) {
	setupServer()
	config.C.Portals = []models.Portal{
		{Title: "GitHub", Link: "https://github.com", Keywords: []string{"code", "git"}},
	}

	sh := searchHandler{logger: testLogger(), levenshtein: metrics.NewLevenshtein()}
	req := httptest.NewRequest(http.MethodGet, "/_/portals?search=code", nil)
	rec := httptest.NewRecorder()
	sh.handle().ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rec.Code)
	}
}

func TestIsSearchResult_DirectTitleMatch(t *testing.T) {
	setupServer()
	config.C.SearchWithStringSimilarity = false

	sh := searchHandler{logger: testLogger(), levenshtein: metrics.NewLevenshtein()}
	portal := models.Portal{Title: "GitHub", Keywords: []string{}}

	if !sh.isSearchResult("GitHub", portal) {
		t.Fatal("expected direct title match")
	}
}

func TestIsSearchResult_CaseInsensitive(t *testing.T) {
	setupServer()
	config.C.SearchWithStringSimilarity = false

	sh := searchHandler{logger: testLogger(), levenshtein: metrics.NewLevenshtein()}
	portal := models.Portal{Title: "GitHub", Keywords: []string{}}

	if !sh.isSearchResult("github", portal) {
		t.Fatal("expected case-insensitive title match")
	}
}

func TestIsSearchResult_KeywordMatch(t *testing.T) {
	setupServer()
	config.C.SearchWithStringSimilarity = false

	sh := searchHandler{logger: testLogger(), levenshtein: metrics.NewLevenshtein()}
	portal := models.Portal{Title: "GitHub", Keywords: []string{"code", "git"}}

	if !sh.isSearchResult("code", portal) {
		t.Fatal("expected keyword match")
	}
}

func TestIsSearchResult_NoMatch(t *testing.T) {
	setupServer()
	config.C.SearchWithStringSimilarity = false

	sh := searchHandler{logger: testLogger(), levenshtein: metrics.NewLevenshtein()}
	portal := models.Portal{Title: "GitHub", Keywords: []string{"code"}}

	if sh.isSearchResult("xyznotfound", portal) {
		t.Fatal("expected no match")
	}
}

func TestIsSearchResult_SimilarityMatch(t *testing.T) {
	setupServer()
	config.C.SearchWithStringSimilarity = true
	config.C.MinimumStringSimilarity = 0.5

	sh := searchHandler{logger: testLogger(), levenshtein: metrics.NewLevenshtein()}
	portal := models.Portal{Title: "GitHub", Keywords: []string{}}

	if !sh.isSearchResult("Githu", portal) {
		t.Fatal("expected similarity match")
	}
}

func TestIsSearchResult_SimilarityDisabled(t *testing.T) {
	setupServer()
	config.C.SearchWithStringSimilarity = false

	sh := searchHandler{logger: testLogger(), levenshtein: metrics.NewLevenshtein()}
	portal := models.Portal{Title: "GitHub", Keywords: []string{}}

	if sh.isSearchResult("Gthub", portal) {
		t.Fatal("expected no match when similarity disabled and no substring match")
	}
}

func TestQueryHomePortals_Empty(t *testing.T) {
	setupServer()
	config.C.Portals = []models.Portal{
		{Title: "GitHub", Link: "https://github.com"},
		{Title: "Google", Link: "https://google.com"},
	}

	sh := searchHandler{logger: testLogger(), levenshtein: metrics.NewLevenshtein()}
	got := sh.queryHomePortals("")

	if len(got) != 2 {
		t.Fatalf("expected 2 portals, got %d", len(got))
	}
}
