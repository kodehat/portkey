package server

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/kodehat/portkey/internal/config"
	"github.com/kodehat/portkey/internal/models"
)

func TestHomeHandler(t *testing.T) {
	setupServer()
	config.C.Portals = []models.Portal{}
	config.C.Pages = []models.Page{}

	t.Run("root path returns 200", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/", nil)
		rec := httptest.NewRecorder()
		homeHandler().ServeHTTP(rec, req)

		if rec.Code != http.StatusOK {
			t.Fatalf("expected 200, got %d", rec.Code)
		}
	})

	t.Run("non-root path returns 404", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/nonexistent", nil)
		rec := httptest.NewRecorder()
		homeHandler().ServeHTTP(rec, req)

		if rec.Code != http.StatusNotFound {
			t.Fatalf("expected 404, got %d", rec.Code)
		}
	})
}
