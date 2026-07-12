package server

import (
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"

	"github.com/kodehat/portkey/internal/config"
	"github.com/kodehat/portkey/internal/favicon"
	"github.com/kodehat/portkey/internal/models"
)

func TestFaviconHandler_MethodNotAllowed(t *testing.T) {
	setupServer()
	config.C.Portals = []models.Portal{}
	favicon.Init(t.TempDir())

	h := faviconHandler{}.handle()
	req := httptest.NewRequest(http.MethodPost, "/_/favicon?domain=github.com", nil)
	rec := httptest.NewRecorder()
	h(rec, req)

	if rec.Code != http.StatusMethodNotAllowed {
		t.Fatalf("expected 405, got %d", rec.Code)
	}
}

func TestFaviconHandler_EmptyDomain(t *testing.T) {
	setupServer()
	config.C.Portals = []models.Portal{}
	favicon.Init(t.TempDir())

	h := faviconHandler{}.handle()
	req := httptest.NewRequest(http.MethodGet, "/_/favicon", nil)
	rec := httptest.NewRecorder()
	h(rec, req)

	if rec.Code != http.StatusNotFound {
		t.Fatalf("expected 404 for empty domain, got %d", rec.Code)
	}
}

func TestFaviconHandler_DomainNotAllowed(t *testing.T) {
	setupServer()
	config.C.Portals = []models.Portal{
		{Title: "GitHub", Link: "https://github.com"},
	}
	favicon.Init(t.TempDir())

	h := faviconHandler{}.handle()
	req := httptest.NewRequest(http.MethodGet, "/_/favicon?domain=evil.com", nil)
	rec := httptest.NewRecorder()
	h(rec, req)

	if rec.Code != http.StatusNotFound {
		t.Fatalf("expected 404 for disallowed domain, got %d", rec.Code)
	}
}

func TestFaviconHandler_AllowedDomain(t *testing.T) {
	setupServer()
	config.C.Portals = []models.Portal{
		{Title: "GitHub", Link: "https://github.com"},
	}

	cacheDir := t.TempDir()
	favicon.Init(cacheDir)

	// Pre-populate cache so no network request is needed.
	cachePath := filepath.Join(cacheDir, "github.com.png")
	if err := os.WriteFile(cachePath, []byte("fake-favicon"), 0644); err != nil {
		t.Fatal(err)
	}

	h := faviconHandler{}.handle()
	req := httptest.NewRequest(http.MethodGet, "/_/favicon?domain=github.com", nil)
	rec := httptest.NewRecorder()
	h(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200 for allowed domain, got %d", rec.Code)
	}
	if cc := rec.Header().Get("Cache-Control"); cc != "public, max-age=86400" {
		t.Fatalf("expected Cache-Control header, got %q", cc)
	}
}
