package server

import (
	"io"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"sync"
	"testing"

	"github.com/kodehat/portkey/internal/build"
	"github.com/kodehat/portkey/internal/config"
	"github.com/kodehat/portkey/internal/metrics"
	"github.com/kodehat/portkey/internal/models"
)

var setupOnce sync.Once

func setupServer() {
	setupOnce.Do(func() {
		build.LoadBuildDetails("testhash")
		config.C = config.Config{
			LogLevel:     "INFO",
			Host:         "localhost",
			Port:         "3000",
			Title:        "portkey",
			Portals:      []models.Portal{},
			Pages:        []models.Page{},
			EnableMetrics: false,
		}
		metrics.Load()
	})
}

func testLogger() *slog.Logger {
	return slog.New(slog.NewTextHandler(io.Discard, nil))
}

func TestHealthHandler(t *testing.T) {
	setupServer()
	req := httptest.NewRequest(http.MethodGet, "/healthz", nil)
	rec := httptest.NewRecorder()
	healthHandler().ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rec.Code)
	}
	if rec.Body.String() != "ok" {
		t.Fatalf("expected body 'ok', got %q", rec.Body.String())
	}
}
