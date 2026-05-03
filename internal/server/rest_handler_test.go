package server

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/kodehat/portkey/internal/config"
	"github.com/kodehat/portkey/internal/models"
)

func TestPortalsRestHandler(t *testing.T) {
	setupServer()
	config.C.Portals = []models.Portal{
		{Title: "GitHub", Link: "https://github.com", Emoji: "💻", Keywords: []string{"code"}},
	}

	req := httptest.NewRequest(http.MethodGet, "/api/portals", nil)
	rec := httptest.NewRecorder()
	portalsRestHandler().ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rec.Code)
	}
	if ct := rec.Header().Get("Content-Type"); ct != "application/json" {
		t.Fatalf("expected Content-Type application/json, got %q", ct)
	}

	var portals []models.Portal
	if err := json.Unmarshal(rec.Body.Bytes(), &portals); err != nil {
		t.Fatal(err)
	}
	if len(portals) != 1 || portals[0].Title != "GitHub" {
		t.Fatalf("unexpected portals: %+v", portals)
	}
}

func TestPagesRestHandler(t *testing.T) {
	setupServer()
	config.C.Pages = []models.Page{
		{Heading: "About", Path: "/about", Content: "<p>info</p>"},
	}

	req := httptest.NewRequest(http.MethodGet, "/api/pages", nil)
	rec := httptest.NewRecorder()
	pagesRestHandler().ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rec.Code)
	}
	if ct := rec.Header().Get("Content-Type"); ct != "application/json" {
		t.Fatalf("expected Content-Type application/json, got %q", ct)
	}

	var pages []models.Page
	if err := json.Unmarshal(rec.Body.Bytes(), &pages); err != nil {
		t.Fatal(err)
	}
	if len(pages) != 1 || pages[0].Heading != "About" {
		t.Fatalf("unexpected pages: %+v", pages)
	}
}
