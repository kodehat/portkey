package server

import (
	"embed"
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"testing"

	"github.com/adrg/strutil/metrics"
	"github.com/kodehat/portkey/internal/config"
	"github.com/kodehat/portkey/internal/models"
)

//go:embed testdata
var testStatic embed.FS

func TestNewMetricsServer(t *testing.T) {
	setupServer()

	srv := NewMetricsServer(testLogger())
	if srv == nil {
		t.Fatal("expected non-nil server")
	}

	mux, ok := srv.(*http.ServeMux)
	if !ok {
		t.Fatal("expected *http.ServeMux")
	}
	if mux == nil {
		t.Fatal("expected non-nil mux")
	}

	req := httptest.NewRequest(http.MethodGet, "/metrics", nil)
	rec := httptest.NewRecorder()
	srv.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rec.Code)
	}
}

func TestMetricsHandler(t *testing.T) {
	setupServer()

	req := httptest.NewRequest(http.MethodGet, "/metrics", nil)
	rec := httptest.NewRecorder()
	metricsHandler().ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rec.Code)
	}
	if rec.Body.Len() == 0 {
		t.Fatal("expected non-empty body")
	}
}

func TestServerHandlers_Healthz(t *testing.T) {
	setupServer()
	config.C.Portals = []models.Portal{}
	config.C.Pages = []models.Page{}

	req := httptest.NewRequest(http.MethodGet, "/healthz", nil)
	rec := httptest.NewRecorder()
	healthHandler().ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rec.Code)
	}
	if rec.Body.String() != "ok" {
		t.Fatalf("expected 'ok', got %q", rec.Body.String())
	}
}

func TestServerHandlers_Version(t *testing.T) {
	setupServer()

	req := httptest.NewRequest(http.MethodGet, "/version", nil)
	rec := httptest.NewRecorder()
	versionHandler().ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rec.Code)
	}
}

func TestServerHandlers_APIPortals(t *testing.T) {
	setupServer()
	config.C.Portals = []models.Portal{
		{Title: "Test", Link: "https://test.com"},
	}

	req := httptest.NewRequest(http.MethodGet, "/api/portals", nil)
	rec := httptest.NewRecorder()
	portalsRestHandler().ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rec.Code)
	}
}

func TestServerHandlers_APIPages(t *testing.T) {
	setupServer()
	config.C.Pages = []models.Page{
		{Heading: "Test", Path: "/test", Content: "<p>test</p>"},
	}

	req := httptest.NewRequest(http.MethodGet, "/api/pages", nil)
	rec := httptest.NewRecorder()
	pagesRestHandler().ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rec.Code)
	}
}

func TestServerHandlers_PortalRedirect(t *testing.T) {
	setupServer()
	config.C.Portals = []models.Portal{
		{Title: "GitHub", Link: "https://github.com"},
	}

	ph := portalHandler{logger: testLogger()}
	infos := ph.handle()
	if len(infos) == 0 {
		t.Fatal("expected at least one portal handler")
	}

	req := httptest.NewRequest(http.MethodGet, infos[0].portalPath, nil)
	rec := httptest.NewRecorder()
	infos[0].handlerFunc(rec, req)

	if rec.Code != http.StatusTemporaryRedirect {
		t.Fatalf("expected 307, got %d", rec.Code)
	}
}

func TestServerHandlers_PageServing(t *testing.T) {
	setupServer()
	config.C.Pages = []models.Page{
		{Heading: "About", Path: "/about", Content: "<p>hello</p>"},
	}

	infos := pageHandler()
	if len(infos) == 0 {
		t.Fatal("expected at least one page handler")
	}

	req := httptest.NewRequest(http.MethodGet, "/about", nil)
	rec := httptest.NewRecorder()
	infos[0].handlerFunc(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rec.Code)
	}
}

func TestSearchHandler_GroupedView(t *testing.T) {
	setupServer()
	config.C.Portals = []models.Portal{
		{Title: "GitHub", Link: "https://github.com", Group: "Dev"},
		{Title: "GitLab", Link: "https://gitlab.com", Group: "Dev"},
		{Title: "Blog", Link: "/blog"},
	}
	config.R.WithGroups = true

	req := httptest.NewRequest(http.MethodGet, "/_/portals", nil)
	rec := httptest.NewRecorder()
	sh := searchHandler{logger: testLogger(), levenshtein: metrics.NewLevenshtein()}
	sh.handle().ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rec.Code)
	}
	body := rec.Body.String()
	if !strings.Contains(body, "Dev") {
		t.Fatal("expected group name 'Dev' in grouped view")
	}
}

func TestSearchHandler_SimilarityMatch(t *testing.T) {
	setupServer()
	config.C.Portals = []models.Portal{
		{Title: "GitHub", Link: "https://github.com"},
	}
	config.C.Search.StringSimilarity = true
	config.C.Search.MinimumSimilarity = 0.5
	config.R.WithGroups = false

	req := httptest.NewRequest(http.MethodGet, "/_/portals?search=Githu", nil)
	rec := httptest.NewRecorder()
	sh := searchHandler{logger: testLogger(), levenshtein: metrics.NewLevenshtein()}
	sh.handle().ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rec.Code)
	}
	body := rec.Body.String()
	if !strings.Contains(body, "GitHub") {
		t.Fatal("expected 'GitHub' in similarity match results")
	}
}

func TestVersionHandler_ViaServer(t *testing.T) {
	setupServer()

	req := httptest.NewRequest(http.MethodGet, "/version", nil)
	rec := httptest.NewRecorder()
	versionHandler().ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rec.Code)
	}
	body := rec.Body.String()
	if body == "" {
		t.Fatal("expected non-empty version page")
	}
}

func TestSearchHandler_IsSimilar(t *testing.T) {
	sh := searchHandler{logger: testLogger(), levenshtein: metrics.NewLevenshtein()}

	if !sh.isSimilar("github", "GitHub", sh.levenshtein, 0.5) {
		t.Fatal("expected 'github' similar to 'GitHub'")
	}

	if sh.isSimilar("xyz123", "GitHub", sh.levenshtein, 0.5) {
		t.Fatal("expected 'xyz123' not similar to 'GitHub'")
	}

	if !sh.isSimilar("GitHb", "GitHub", sh.levenshtein, 0.5) {
		t.Fatal("expected 'GitHb' similar to 'GitHub'")
	}

	// Exact match is always above threshold
	if !sh.isSimilar("git", "git", sh.levenshtein, 0.9) {
		t.Fatal("expected exact match similar at high threshold")
	}

	// Just at the threshold boundary — similarity > minimum, not >=
	if sh.isSimilar("a", "b", sh.levenshtein, 0.99) {
		t.Fatal("expected single char diff below high threshold")
	}
}

func TestSearchHandler_IsSearchResult_DirectMatch(t *testing.T) {
	setupServer()
	config.C.Search.StringSimilarity = false
	portal := models.Portal{Title: "GitHub", Link: "https://github.com", Keywords: []string{"code", "git"}}

	sh := searchHandler{logger: testLogger(), levenshtein: metrics.NewLevenshtein()}

	if !sh.isSearchResult("Git", portal) {
		t.Fatal("expected direct title match")
	}
	if !sh.isSearchResult("git", portal) {
		t.Fatal("expected direct keyword match")
	}
	if sh.isSearchResult("nonexistent", portal) {
		t.Fatal("expected no match for unrelated query")
	}
}

func TestSearchHandler_IsSearchResult_SimilarityMatch(t *testing.T) {
	setupServer()
	config.C.Search.StringSimilarity = true
	config.C.Search.MinimumSimilarity = 0.5
	portal := models.Portal{Title: "GitHub", Link: "https://github.com", Keywords: []string{"code"}}

	sh := searchHandler{logger: testLogger(), levenshtein: metrics.NewLevenshtein()}

	if !sh.isSearchResult("Githu", portal) {
		t.Fatal("expected similarity title match")
	}
	if !sh.isSearchResult("cod", portal) {
		t.Fatal("expected similarity keyword match")
	}
}

func TestSearchHandler_IsSearchResult_SimilarityDisabled(t *testing.T) {
	setupServer()
	config.C.Search.StringSimilarity = false
	portal := models.Portal{Title: "GitHub", Link: "https://github.com"}

	sh := searchHandler{logger: testLogger(), levenshtein: metrics.NewLevenshtein()}

	if sh.isSearchResult("xyz", portal) {
		t.Fatal("expected no match when query doesn't match and similarity disabled")
	}
}

func TestSearchHandler_QueryHomePortals_EmptyQuery(t *testing.T) {
	setupServer()
	config.C.Portals = []models.Portal{
		{Title: "GitHub", Link: "https://github.com"},
		{Title: "GitLab", Link: "https://gitlab.com"},
	}

	sh := searchHandler{logger: testLogger(), levenshtein: metrics.NewLevenshtein()}
	result := sh.queryHomePortals("")

	if len(result) != 2 {
		t.Fatalf("expected 2 portals, got %d", len(result))
	}
}

func TestSearchHandler_QueryHomePortals_Filtered(t *testing.T) {
	setupServer()
	config.C.Portals = []models.Portal{
		{Title: "GitHub", Link: "https://github.com"},
		{Title: "GitLab", Link: "https://gitlab.com"},
	}

	sh := searchHandler{logger: testLogger(), levenshtein: metrics.NewLevenshtein()}
	result := sh.queryHomePortals("Hub")

	if len(result) != 1 {
		t.Fatalf("expected 1 portal, got %d", len(result))
	}
	if result[0].Title != "GitHub" {
		t.Fatalf("expected GitHub, got %s", result[0].Title)
	}
}

func TestSearchHandler_Handle_NoQueryNoGroups(t *testing.T) {
	setupServer()
	config.C.Portals = []models.Portal{
		{Title: "GitHub", Link: "https://github.com"},
	}
	config.R.WithGroups = false

	req := httptest.NewRequest(http.MethodGet, "/_/portals", nil)
	rec := httptest.NewRecorder()
	sh := searchHandler{logger: testLogger(), levenshtein: metrics.NewLevenshtein()}
	sh.handle().ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rec.Code)
	}
	body := rec.Body.String()
	if !strings.Contains(body, "GitHub") {
		t.Fatal("expected 'GitHub' in ungrouped view")
	}
}

func TestSearchHandler_IncreaseMetrics(t *testing.T) {
	setupServer()
	config.C.Metrics.Enabled = true

	sh := searchHandler{logger: testLogger(), levenshtein: metrics.NewLevenshtein()}

	sh.increaseMetrics(true)
	sh.increaseMetrics(false)
}

func TestSearchHandler_IsSearchResult_KeywordDirectMatch(t *testing.T) {
	setupServer()
	config.C.Search.StringSimilarity = false
	portal := models.Portal{Title: "DevTools", Link: "/tools", Keywords: []string{"code", "git", "repo"}}

	sh := searchHandler{logger: testLogger(), levenshtein: metrics.NewLevenshtein()}

	if !sh.isSearchResult("repo", portal) {
		t.Fatal("expected keyword match")
	}
	if sh.isSearchResult("something", portal) {
		t.Fatal("expected no match for unrelated query")
	}
}

func TestIsSearchResult_SimilarityKeywordMatch(t *testing.T) {
	setupServer()
	config.C.Search.StringSimilarity = true
	config.C.Search.MinimumSimilarity = 0.3

	sh := searchHandler{logger: testLogger(), levenshtein: metrics.NewLevenshtein()}
	portal := models.Portal{Title: "ZZZZ", Keywords: []string{"programming", "development"}}

	// Title doesn't match directly, keyword doesn't match directly, but close enough for similarity
	if !sh.isSearchResult("programing", portal) {
		t.Fatal("expected similarity keyword match")
	}
}

func TestIsSearchResult_NoSimilarityWhenDisabled(t *testing.T) {
	setupServer()
	config.C.Search.StringSimilarity = false

	sh := searchHandler{logger: testLogger(), levenshtein: metrics.NewLevenshtein()}
	portal := models.Portal{Title: "ZZZZ", Keywords: []string{"something_else"}}

	// No direct match in title or keywords, and similarity is disabled
	if sh.isSearchResult("completely_different", portal) {
		t.Fatal("expected no match when similarity disabled and no direct match")
	}
}

func TestIsSearchResult_SimilarityKeyword_LastResort(t *testing.T) {
	setupServer()
	config.C.Search.StringSimilarity = true
	config.C.Search.MinimumSimilarity = 0.3

	sh := searchHandler{logger: testLogger(), levenshtein: metrics.NewLevenshtein()}
	// Title has low similarity, but second keyword is close
	portal := models.Portal{
		Title:    "ZZZZZZZ",
		Keywords: []string{"aaaaaa", "github"},
	}

	if !sh.isSearchResult("githib", portal) {
		t.Fatal("expected similarity match on second keyword")
	}
}

func TestPageHandler_WithMetrics(t *testing.T) {
	setupServer()
	config.C.Metrics.Enabled = true
	config.C.Pages = []models.Page{
		{Heading: "About", Path: "/about", Content: "<p>info</p>"},
	}

	infos := pageHandler()
	if len(infos) != 1 {
		t.Fatalf("expected 1 page handler, got %d", len(infos))
	}

	req := httptest.NewRequest(http.MethodGet, "/about", nil)
	rec := httptest.NewRecorder()
	infos[0].handlerFunc(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rec.Code)
	}
}

func TestPageMetricsHandler(t *testing.T) {
	setupServer()
	config.C.Metrics.Enabled = true

	innerHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})
	page := models.Page{Path: "/test"}

	h := pageMetricsHandler(page, innerHandler)
	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rec.Code)
	}
}

func TestPortalHandler_Handle_WithMetrics(t *testing.T) {
	setupServer()
	config.C.Metrics.Enabled = true
	config.C.Portals = []models.Portal{
		{Title: "GitHub", Link: "https://github.com"},
	}

	ph := portalHandler{logger: testLogger()}
	infos := ph.handle()
	if len(infos) != 1 {
		t.Fatalf("expected 1 portal handler, got %d", len(infos))
	}
}

func TestPortalHandler_Handle_InternalLink(t *testing.T) {
	setupServer()
	config.C.Portals = []models.Portal{
		{Title: "About", Link: "/about"},
	}

	ph := portalHandler{logger: testLogger()}
	infos := ph.handle()
	if len(infos) != 0 {
		t.Fatalf("expected 0 portal handlers for internal links, got %d", len(infos))
	}
}

func TestPortalHandler_TitleModifiedWithMetrics(t *testing.T) {
	setupServer()
	config.C.Metrics.Enabled = true
	config.C.Portals = []models.Portal{
		{Title: "My Portal!", Link: "https://example.com"},
	}

	ph := portalHandler{logger: testLogger()}
	infos := ph.handle()

	if len(infos) != 1 {
		t.Fatalf("expected 1 handler info, got %d", len(infos))
	}
	if infos[0].portalPath != "/MyPortal" {
		t.Fatalf("expected portalPath /MyPortal, got %q", infos[0].portalPath)
	}
}

func TestPageHandler_EmptyPages(t *testing.T) {
	setupServer()
	config.C.Pages = []models.Page{}

	infos := pageHandler()
	if len(infos) != 0 {
		t.Fatalf("expected 0 page handlers, got %d", len(infos))
	}
}

func TestSearchHandler_Handle_SearchWithResults(t *testing.T) {
	setupServer()
	config.C.Portals = []models.Portal{
		{Title: "GitHub", Link: "https://github.com"},
		{Title: "Google", Link: "https://google.com"},
	}
	config.R.WithGroups = false
	config.C.Metrics.Enabled = true

	req := httptest.NewRequest(http.MethodGet, "/_/portals?search=Hub", nil)
	rec := httptest.NewRecorder()
	sh := searchHandler{logger: testLogger(), levenshtein: metrics.NewLevenshtein()}
	sh.handle().ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rec.Code)
	}
	body := rec.Body.String()
	if !strings.Contains(body, "GitHub") {
		t.Fatal("expected 'GitHub' in search results")
	}
	if strings.Contains(body, "Google") {
		t.Fatal("did not expect 'Google' in search results")
	}
}

func TestSearchHandler_Handle_SearchNoResults(t *testing.T) {
	setupServer()
	config.C.Portals = []models.Portal{
		{Title: "GitHub", Link: "https://github.com"},
	}
	config.R.WithGroups = false

	req := httptest.NewRequest(http.MethodGet, "/_/portals?search=zzzzz", nil)
	rec := httptest.NewRecorder()
	sh := searchHandler{logger: testLogger(), levenshtein: metrics.NewLevenshtein()}
	sh.handle().ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rec.Code)
	}
}

func TestAddRoutes_WithPages(t *testing.T) {
	setupServer()
	config.C.Pages = []models.Page{
		{Heading: "About", Path: "/about", Content: "<p>info</p>"},
	}

	srv := NewServer(testLogger(), testStatic)
	if srv == nil {
		t.Fatal("expected non-nil server")
	}

	// Verify the page route is registered
	req := httptest.NewRequest(http.MethodGet, "/about", nil)
	rec := httptest.NewRecorder()
	srv.ServeHTTP(rec, req)

	// Should return 200 since page is configured
	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200 for page route, got %d", rec.Code)
	}
}

func TestAddRoutes_WithCustomIcons(t *testing.T) {
	dir := t.TempDir()
	os.WriteFile(dir+"/icon.svg", []byte("<svg/>"), 0644)

	setupServer()
	config.C.Favicon.CustomIconsDir = dir

	srv := NewServer(testLogger(), testStatic)
	if srv == nil {
		t.Fatal("expected non-nil server")
	}

	// Verify the icons route is registered
	req := httptest.NewRequest(http.MethodGet, "/_/icons/icon.svg", nil)
	rec := httptest.NewRecorder()
	srv.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200 for custom icon route, got %d", rec.Code)
	}
}
