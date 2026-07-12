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
		{Title: "GitHub", Link: "https://github.com", Icon: "💻"},
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

func TestPortalHandler_TitleModified(t *testing.T) {
	setupServer()
	// Title with spaces — TitleForUrl strips them, producing a different string.
	config.C.Portals = []models.Portal{
		{Title: "My Portal!", Link: "https://example.com"},
	}

	ph := portalHandler{logger: testLogger()}
	infos := ph.handle()

	if len(infos) != 1 {
		t.Fatalf("expected 1 handler info, got %d", len(infos))
	}
	// TitleForUrl removes non-alphanumeric/dash chars: "My Portal!" → "MyPortal"
	if infos[0].portalPath != "/MyPortal" {
		t.Fatalf("expected portalPath /MyPortal, got %q", infos[0].portalPath)
	}
}

func TestPortalHandler_CJKTitle(t *testing.T) {
	setupServer()
	config.C.Portals = []models.Portal{
		{Title: "中国国家地理", Link: "https://example.com"},
	}

	ph := portalHandler{logger: testLogger()}
	infos := ph.handle()

	if len(infos) != 1 {
		t.Fatalf("expected 1 handler info, got %d", len(infos))
	}
	if infos[0].portalPath != "/中国国家地理" {
		t.Fatalf("expected portalPath /中国国家地理, got %q", infos[0].portalPath)
	}

	req := httptest.NewRequest(http.MethodGet, "/中国国家地理", nil)
	rec := httptest.NewRecorder()
	infos[0].handlerFunc(rec, req)

	if rec.Code != http.StatusTemporaryRedirect {
		t.Fatalf("expected 307, got %d", rec.Code)
	}
	loc := rec.Header().Get("Location")
	if loc != "https://example.com" {
		t.Fatalf("expected Location https://example.com, got %q", loc)
	}
}

func TestPortalHandler_EmojiOnlyTitleFallback(t *testing.T) {
	setupServer()
	config.C.Portals = []models.Portal{
		{Title: "💰", Link: "https://example.com"},
	}

	ph := portalHandler{logger: testLogger()}
	infos := ph.handle()

	if len(infos) != 1 {
		t.Fatalf("expected 1 handler info for emoji-only title, got %d", len(infos))
	}
	// TitleForUrl falls back to percent-encoding.
	if infos[0].portalPath != "/%F0%9F%92%B0" {
		t.Fatalf("expected portalPath /%%F0%%9F%%92%%B0, got %q", infos[0].portalPath)
	}

	req := httptest.NewRequest(http.MethodGet, "/%F0%9F%92%B0", nil)
	rec := httptest.NewRecorder()
	infos[0].handlerFunc(rec, req)

	if rec.Code != http.StatusTemporaryRedirect {
		t.Fatalf("expected 307, got %d", rec.Code)
	}
	loc := rec.Header().Get("Location")
	if loc != "https://example.com" {
		t.Fatalf("expected Location https://example.com, got %q", loc)
	}
}
