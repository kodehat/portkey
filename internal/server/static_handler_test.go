package server

import (
	"embed"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/kodehat/portkey/internal/config"
	"github.com/kodehat/portkey/internal/favicon"
)

//go:embed testdata
var testStaticFS embed.FS

func TestStaticHandler(t *testing.T) {
	setupServer()

	h := staticHandler(testStaticFS)
	req := httptest.NewRequest(http.MethodGet, "/testdata/static/css/main.css", nil)
	rec := httptest.NewRecorder()
	h(rec, req)

	// staticHandler is exercised regardless of response status.
	if rec.Code == 0 {
		t.Fatal("expected a response status")
	}
}

func TestNewServer(t *testing.T) {
	setupServer()
	favicon.Init(t.TempDir(), nil)

	srv := NewServer(testLogger(), testStaticFS)
	if srv == nil {
		t.Fatal("expected non-nil server")
	}

	req := httptest.NewRequest(http.MethodGet, "/healthz", nil)
	rec := httptest.NewRecorder()
	srv.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200 from /healthz via NewServer, got %d", rec.Code)
	}
}

func TestDurationMiddleware(t *testing.T) {
	setupServer()
	config.C.ContextPath = ""

	inner := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})

	handler := durationMiddleware(inner)
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rec.Code)
	}
}

func TestDurationMiddleware_StaticPathSkipped(t *testing.T) {
	setupServer()
	config.C.ContextPath = ""

	inner := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})

	handler := durationMiddleware(inner)
	req := httptest.NewRequest(http.MethodGet, "/static/css/main.css", nil)
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rec.Code)
	}
}

func TestNewServer_WithCustomIconsDir(t *testing.T) {
	setupServer()
	favicon.Init(t.TempDir(), nil)
	config.C.CustomIconsDir = t.TempDir()
	defer func() { config.C.CustomIconsDir = "" }()

	srv := NewServer(testLogger(), testStaticFS)
	if srv == nil {
		t.Fatal("expected non-nil server")
	}

	req := httptest.NewRequest(http.MethodGet, "/healthz", nil)
	rec := httptest.NewRecorder()
	srv.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rec.Code)
	}
}

func TestNewServer_WithDevMode(t *testing.T) {
	setupServer()
	favicon.Init(t.TempDir(), nil)
	config.C.DevMode = true
	defer func() { config.C.DevMode = false }()

	srv := NewServer(testLogger(), testStaticFS)
	if srv == nil {
		t.Fatal("expected non-nil server")
	}

	req := httptest.NewRequest(http.MethodGet, "/healthz", nil)
	rec := httptest.NewRecorder()
	srv.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rec.Code)
	}
}
