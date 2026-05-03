package server

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/kodehat/portkey/internal/config"
	"github.com/kodehat/portkey/internal/models"
)

func TestPageHandlerReturnsPage(t *testing.T) {
	setupServer()
	config.C.Pages = []models.Page{
		{Heading: "About", Subtitle: "Info", Path: "/about", Content: "<p>hello</p>"},
	}

	infos := pageHandler()

	if len(infos) != 1 {
		t.Fatalf("expected 1 handler info, got %d", len(infos))
	}
	if infos[0].pagePath != "/about" {
		t.Fatalf("expected pagePath /about, got %q", infos[0].pagePath)
	}

	req := httptest.NewRequest(http.MethodGet, "/about", nil)
	rec := httptest.NewRecorder()
	infos[0].handlerFunc(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rec.Code)
	}
	body := rec.Body.String()
	if body == "" {
		t.Fatal("expected non-empty body")
	}
}

func TestPageHandlerEmpty(t *testing.T) {
	setupServer()
	config.C.Pages = []models.Page{}

	infos := pageHandler()

	if len(infos) != 0 {
		t.Fatalf("expected 0 handler infos, got %d", len(infos))
	}
}
