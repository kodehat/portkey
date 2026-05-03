package server

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/kodehat/portkey/internal/config"
	"github.com/kodehat/portkey/internal/models"
)

func TestPortalHandlerReturnsRedirect(t *testing.T) {
	setupServer()
	config.C.Portals = []models.Portal{
		{Title: "GitHub", Link: "https://github.com", Emoji: "💻"},
	}

	ph := portalHandler{logger: testLogger()}
	infos := ph.handle()

	if len(infos) != 1 {
		t.Fatalf("expected 1 handler info, got %d", len(infos))
	}
	if infos[0].portalPath != "/GitHub" {
		t.Fatalf("expected portalPath /GitHub, got %q", infos[0].portalPath)
	}

	req := httptest.NewRequest(http.MethodGet, "/GitHub", nil)
	rec := httptest.NewRecorder()
	infos[0].handlerFunc(rec, req)

	if rec.Code != http.StatusTemporaryRedirect {
		t.Fatalf("expected 307, got %d", rec.Code)
	}
	loc := rec.Header().Get("Location")
	if loc != "https://github.com" {
		t.Fatalf("expected Location https://github.com, got %q", loc)
	}
}

func TestPortalHandlerInternalOnly(t *testing.T) {
	setupServer()
	config.C.Portals = []models.Portal{
		{Title: "About", Link: "/about"},
	}

	ph := portalHandler{logger: testLogger()}
	infos := ph.handle()

	if len(infos) != 0 {
		t.Fatalf("expected 0 handler infos for internal link, got %d", len(infos))
	}
}

func TestPortalHandlerEmpty(t *testing.T) {
	setupServer()
	config.C.Portals = []models.Portal{}

	ph := portalHandler{logger: testLogger()}
	infos := ph.handle()

	if len(infos) != 0 {
		t.Fatalf("expected 0 handler infos, got %d", len(infos))
	}
}
