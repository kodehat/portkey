package server

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/kodehat/portkey/internal/config"
	"github.com/kodehat/portkey/internal/models"
)

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
