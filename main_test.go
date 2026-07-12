package main

import (
	"context"
	"encoding/json"
	"io"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/kodehat/portkey/internal/build"
	"github.com/kodehat/portkey/internal/config"
	"github.com/kodehat/portkey/internal/favicon"
	"github.com/kodehat/portkey/internal/metrics"
	"github.com/kodehat/portkey/internal/models"
	"github.com/kodehat/portkey/internal/server"
)

var globalsOnce sync.Once

func initGlobals(cfg config.Config) {
	config.C = cfg
	globalsOnce.Do(func() {
		build.LoadBuildDetails("testhash")
		// metrics.Load() may have already been called by main() in TestAAAMainDirectly.
		// Only load if the first metric hasn't been initialized yet.
		if metrics.M.PortalHitCounter == nil {
			metrics.Load()
		}
	})
}

func TestAAAMainDirectly(t *testing.T) {
	// This test must run before any test that calls initGlobals(),
	// because main() calls metrics.Load() which panics if already loaded.

	// Send interrupt after a brief delay so main() starts servers then shuts down.
	go func() {
		time.Sleep(300 * time.Millisecond)
		p, _ := os.FindProcess(os.Getpid())
		p.Signal(os.Interrupt)
	}()

	main()
}

func TestHealthz(t *testing.T) {
	initGlobals(config.Config{
		LogLevel: "INFO",
		Host:     "localhost",
		Port:     "3000",
		Portals:  []models.Portal{},
		Pages:    []models.Page{},
	})
	srv := server.NewServer(testLogger(), static)
	svr := httptest.NewServer(srv)
	defer svr.Close()

	res, err := http.Get(svr.URL + "/healthz")
	if err != nil {
		t.Fatal(err)
	}
	defer res.Body.Close()
	if res.StatusCode != http.StatusOK {
		t.Fatalf("expected 200, got %d", res.StatusCode)
	}
	body, _ := io.ReadAll(res.Body)
	if string(body) != "ok" {
		t.Fatalf("expected body 'ok', got %q", string(body))
	}
}

func TestHomePage(t *testing.T) {
	initGlobals(config.Config{
		LogLevel: "INFO",
		Host:     "localhost",
		Port:     "3000",
		Title:    "test-portkey",
		Portals:  []models.Portal{},
		Pages:    []models.Page{},
	})
	srv := server.NewServer(testLogger(), static)
	svr := httptest.NewServer(srv)
	defer svr.Close()

	res, err := http.Get(svr.URL + "/")
	if err != nil {
		t.Fatal(err)
	}
	defer res.Body.Close()
	if res.StatusCode != http.StatusOK {
		t.Fatalf("expected 200, got %d", res.StatusCode)
	}
}

func TestVersionPage(t *testing.T) {
	initGlobals(config.Config{
		LogLevel: "INFO",
		Host:     "localhost",
		Port:     "3000",
		Portals:  []models.Portal{},
		Pages:    []models.Page{},
	})
	srv := server.NewServer(testLogger(), static)
	svr := httptest.NewServer(srv)
	defer svr.Close()

	res, err := http.Get(svr.URL + "/version")
	if err != nil {
		t.Fatal(err)
	}
	defer res.Body.Close()
	if res.StatusCode != http.StatusOK {
		t.Fatalf("expected 200, got %d", res.StatusCode)
	}
}

func TestNotFound(t *testing.T) {
	initGlobals(config.Config{
		LogLevel: "INFO",
		Host:     "localhost",
		Port:     "3000",
		Portals:  []models.Portal{},
		Pages:    []models.Page{},
	})
	srv := server.NewServer(testLogger(), static)
	svr := httptest.NewServer(srv)
	defer svr.Close()

	res, err := http.Get(svr.URL + "/nonexistent")
	if err != nil {
		t.Fatal(err)
	}
	defer res.Body.Close()
	if res.StatusCode != http.StatusNotFound {
		t.Fatalf("expected 404, got %d", res.StatusCode)
	}
}

func TestAPIEndpoints(t *testing.T) { // NOSONAR
	initGlobals(config.Config{
		LogLevel: "INFO",
		Host:     "localhost",
		Port:     "3000",
		Portals: []models.Portal{
			{Title: "GitHub", Link: "https://github.com", Icon: "💻", Keywords: []string{"code"}},
		},
		Pages: []models.Page{
			{Heading: "About", Path: "/about", Content: "<p>hello</p>"},
		},
	})
	srv := server.NewServer(testLogger(), static)
	svr := httptest.NewServer(srv)
	defer svr.Close()

	t.Run("api/portals returns JSON", func(t *testing.T) {
		res, err := http.Get(svr.URL + "/api/portals")
		if err != nil {
			t.Fatal(err)
		}
		defer res.Body.Close()
		if res.StatusCode != http.StatusOK {
			t.Fatalf("expected 200, got %d", res.StatusCode)
		}
		if ct := res.Header.Get("Content-Type"); ct != "application/json" {
			t.Fatalf("expected Content-Type application/json, got %q", ct)
		}
		var portals []models.Portal
		if err := json.NewDecoder(res.Body).Decode(&portals); err != nil {
			t.Fatal(err)
		}
		if len(portals) != 1 || portals[0].Title != "GitHub" {
			t.Fatalf("unexpected portals: %+v", portals)
		}
	})

	t.Run("api/pages returns JSON", func(t *testing.T) {
		res, err := http.Get(svr.URL + "/api/pages")
		if err != nil {
			t.Fatal(err)
		}
		defer res.Body.Close()
		if res.StatusCode != http.StatusOK {
			t.Fatalf("expected 200, got %d", res.StatusCode)
		}
		if ct := res.Header.Get("Content-Type"); ct != "application/json" {
			t.Fatalf("expected Content-Type application/json, got %q", ct)
		}
		var pages []models.Page
		if err := json.NewDecoder(res.Body).Decode(&pages); err != nil {
			t.Fatal(err)
		}
		if len(pages) != 1 || pages[0].Heading != "About" {
			t.Fatalf("unexpected pages: %+v", pages)
		}
	})
}

func TestPortalRedirect(t *testing.T) {
	initGlobals(config.Config{
		LogLevel: "INFO",
		Host:     "localhost",
		Port:     "3000",
		Portals: []models.Portal{
			{Title: "GitHub", Link: "https://github.com", Icon: "💻", Keywords: []string{"code"}},
		},
		Pages: []models.Page{},
	})
	srv := server.NewServer(testLogger(), static)
	svr := httptest.NewServer(srv)
	defer svr.Close()

	client := &http.Client{
		CheckRedirect: func(req *http.Request, via []*http.Request) error {
			return http.ErrUseLastResponse
		},
	}
	res, err := client.Get(svr.URL + "/GitHub")
	if err != nil {
		t.Fatal(err)
	}
	defer res.Body.Close()
	if res.StatusCode != http.StatusTemporaryRedirect {
		t.Fatalf("expected 307, got %d", res.StatusCode)
	}
	loc := res.Header.Get("Location")
	if loc != "https://github.com" {
		t.Fatalf("expected Location https://github.com, got %q", loc)
	}
}

func TestPageServing(t *testing.T) {
	initGlobals(config.Config{
		LogLevel: "INFO",
		Host:     "localhost",
		Port:     "3000",
		Portals:  []models.Portal{},
		Pages: []models.Page{
			{Heading: "About", Path: "/about", Content: "<p>hello</p>"},
		},
	})
	srv := server.NewServer(testLogger(), static)
	svr := httptest.NewServer(srv)
	defer svr.Close()

	res, err := http.Get(svr.URL + "/about")
	if err != nil {
		t.Fatal(err)
	}
	defer res.Body.Close()
	if res.StatusCode != http.StatusOK {
		t.Fatalf("expected 200, got %d", res.StatusCode)
	}
	body, _ := io.ReadAll(res.Body)
	if !strings.Contains(string(body), "About") {
		t.Fatalf("expected page body to contain 'About', got %q", string(body))
	}
}

func TestSearchEndpoint(t *testing.T) {
	initGlobals(config.Config{
		LogLevel: "INFO",
		Host:     "localhost",
		Port:     "3000",
		Portals: []models.Portal{
			{Title: "GitHub", Link: "https://github.com", Icon: "💻", Keywords: []string{"code"}},
		},
		Pages: []models.Page{},
	})
	srv := server.NewServer(testLogger(), static)
	svr := httptest.NewServer(srv)
	defer svr.Close()

	t.Run("no query returns portals", func(t *testing.T) {
		res, err := http.Get(svr.URL + "/_/portals")
		if err != nil {
			t.Fatal(err)
		}
		defer res.Body.Close()
		if res.StatusCode != http.StatusOK {
			t.Fatalf("expected 200, got %d", res.StatusCode)
		}
	})

	t.Run("search query returns results", func(t *testing.T) {
		res, err := http.Get(svr.URL + "/_/portals?search=GitHub")
		if err != nil {
			t.Fatal(err)
		}
		defer res.Body.Close()
		if res.StatusCode != http.StatusOK {
			t.Fatalf("expected 200, got %d", res.StatusCode)
		}
	})

	t.Run("search with no match returns empty", func(t *testing.T) {
		res, err := http.Get(svr.URL + "/_/portals?search=zzzznotfound")
		if err != nil {
			t.Fatal(err)
		}
		defer res.Body.Close()
		if res.StatusCode != http.StatusOK {
			t.Fatalf("expected 200, got %d", res.StatusCode)
		}
	})
}

func TestStaticServing(t *testing.T) {
	initGlobals(config.Config{
		LogLevel: "INFO",
		Host:     "localhost",
		Port:     "3000",
		Portals:  []models.Portal{},
		Pages:    []models.Page{},
	})
	srv := server.NewServer(testLogger(), static)
	svr := httptest.NewServer(srv)
	defer svr.Close()

	res, err := http.Get(svr.URL + "/static/css/main.css")
	if err != nil {
		t.Fatal(err)
	}
	defer res.Body.Close()
	if res.StatusCode != http.StatusOK {
		t.Fatalf("expected 200 for static/css/main.css, got %d", res.StatusCode)
	}
}

func testLogger() *slog.Logger {
	return slog.New(slog.NewTextHandler(io.Discard, nil))
}

func TestGetCssResourceHash(t *testing.T) {
	hash := getCssResourceHash()
	if len(hash) != 8 {
		t.Fatalf("expected 8-char hash, got %d chars: %q", len(hash), hash)
	}
	for _, c := range hash {
		if !((c >= '0' && c <= '9') || (c >= 'a' && c <= 'f')) {
			t.Fatalf("expected hex char, got %q", c)
		}
	}
}

func TestHomePage_WithPortals(t *testing.T) {
	initGlobals(config.Config{
		LogLevel: "INFO",
		Host:     "localhost",
		Port:     "3000",
		Title:    "test",
		Portals: []models.Portal{
			{Title: "GitHub", Link: "https://github.com"},
			{Title: "GitLab", Link: "https://gitlab.com"},
		},
		Pages: []models.Page{},
	})
	srv := server.NewServer(testLogger(), static)
	svr := httptest.NewServer(srv)
	defer svr.Close()

	res, err := http.Get(svr.URL + "/")
	if err != nil {
		t.Fatal(err)
	}
	defer res.Body.Close()
	if res.StatusCode != http.StatusOK {
		t.Fatalf("expected 200, got %d", res.StatusCode)
	}

	res2, err := http.Get(svr.URL + "/_/portals")
	if err != nil {
		t.Fatal(err)
	}
	defer res2.Body.Close()
	body, _ := io.ReadAll(res2.Body)
	if !strings.Contains(string(body), "GitHub") {
		t.Fatal("expected GitHub in search results")
	}
}

func TestHomePage_WithGroupedPortals(t *testing.T) {
	initGlobals(config.Config{
		LogLevel: "INFO",
		Host:     "localhost",
		Port:     "3000",
		Title:    "test",
		Portals: []models.Portal{
			{Title: "GitHub", Link: "https://github.com", Group: "Dev"},
			{Title: "GitLab", Link: "https://gitlab.com", Group: "Dev"},
		},
		Pages: []models.Page{},
	})
	config.R.WithGroups = true
	srv := server.NewServer(testLogger(), static)
	svr := httptest.NewServer(srv)
	defer svr.Close()

	res, err := http.Get(svr.URL + "/_/portals")
	if err != nil {
		t.Fatal(err)
	}
	defer res.Body.Close()
	if res.StatusCode != http.StatusOK {
		t.Fatalf("expected 200, got %d", res.StatusCode)
	}
	body, _ := io.ReadAll(res.Body)
	if !strings.Contains(string(body), "Dev") {
		t.Fatalf("expected 'Dev' group in search results, got %q", string(body))
	}
}

func TestMetricsServer(t *testing.T) {
	initGlobals(config.Config{
		LogLevel: "INFO",
		Host:     "localhost",
		Port:     "3000",
		Portals:  []models.Portal{},
		Pages:    []models.Page{},
	})
	srv := server.NewMetricsServer(testLogger())
	svr := httptest.NewServer(srv)
	defer svr.Close()

	res, err := http.Get(svr.URL + "/metrics")
	if err != nil {
		t.Fatal(err)
	}
	defer res.Body.Close()
	if res.StatusCode != http.StatusOK {
		t.Fatalf("expected 200, got %d", res.StatusCode)
	}
}

func TestRun_CancelImmediately(t *testing.T) {
	cfg := config.Config{
		LogLevel: "INFO",
		Host:     "127.0.0.1",
		Port:     "0",
		Portals:  []models.Portal{},
		Pages:    []models.Page{},
	}
	initGlobals(cfg)
	favicon.Init(t.TempDir(), nil)

	ctx, cancel := context.WithCancel(context.Background())
	cancel() // pre-cancel so run shuts down immediately

	if err := run(ctx, cfg, strings.NewReader(""), io.Discard, io.Discard); err != nil {
		t.Fatalf("expected nil error, got %v", err)
	}
}

func TestRun_WithMetrics(t *testing.T) {
	cfg := config.Config{
		LogLevel:      "INFO",
		Host:          "127.0.0.1",
		Port:          "0",
		MetricsHost:   "127.0.0.1",
		MetricsPort:   "0",
		EnableMetrics: true,
		Portals:       []models.Portal{},
		Pages:         []models.Page{},
	}
	initGlobals(cfg)
	favicon.Init(t.TempDir(), nil)

	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	if err := run(ctx, cfg, strings.NewReader(""), io.Discard, io.Discard); err != nil {
		t.Fatalf("expected nil error with metrics enabled, got %v", err)
	}
}

func TestRun_ServerShutdown(t *testing.T) {
	cfg := config.Config{
		LogLevel:      "INFO",
		Host:          "127.0.0.1",
		Port:          "0",
		MetricsHost:   "127.0.0.1",
		MetricsPort:   "0",
		EnableMetrics: true,
		Portals:       []models.Portal{},
		Pages:         []models.Page{},
	}
	initGlobals(cfg)
	favicon.Init(t.TempDir(), nil)

	ctx, cancel := context.WithCancel(context.Background())

	// Run server in background and cancel after a brief moment
	errCh := make(chan error, 1)
	go func() {
		errCh <- run(ctx, cfg, strings.NewReader(""), io.Discard, io.Discard)
	}()

	// Give server time to start
	time.Sleep(50 * time.Millisecond)
	cancel()

	select {
	case err := <-errCh:
		if err != nil {
			t.Fatalf("expected nil error from run, got %v", err)
		}
	case <-time.After(5 * time.Second):
		t.Fatal("timed out waiting for server shutdown")
	}
}

func TestRun_ServerShutdownNoMetrics(t *testing.T) {
	cfg := config.Config{
		LogLevel: "INFO",
		Host:     "127.0.0.1",
		Port:     "0",
		Portals:  []models.Portal{},
		Pages:    []models.Page{},
	}
	initGlobals(cfg)
	favicon.Init(t.TempDir(), nil)

	ctx, cancel := context.WithCancel(context.Background())

	errCh := make(chan error, 1)
	go func() {
		errCh <- run(ctx, cfg, strings.NewReader(""), io.Discard, io.Discard)
	}()

	time.Sleep(50 * time.Millisecond)
	cancel()

	select {
	case err := <-errCh:
		if err != nil {
			t.Fatalf("expected nil error from run, got %v", err)
		}
	case <-time.After(5 * time.Second):
		t.Fatal("timed out waiting for server shutdown")
	}
}

func TestServer_CustomIconsDir(t *testing.T) {
	iconsDir := t.TempDir()
	// Create a test icon file
	if err := os.WriteFile(filepath.Join(iconsDir, "test-icon.svg"), []byte("<svg/>"), 0644); err != nil {
		t.Fatal(err)
	}

	initGlobals(config.Config{
		LogLevel:       "INFO",
		Host:           "localhost",
		Port:           "3000",
		CustomIconsDir: iconsDir,
		Portals:        []models.Portal{},
		Pages:          []models.Page{},
	})
	srv := server.NewServer(testLogger(), static)
	svr := httptest.NewServer(srv)
	defer svr.Close()

	res, err := http.Get(svr.URL + "/_/icons/test-icon.svg")
	if err != nil {
		t.Fatal(err)
	}
	defer res.Body.Close()
	if res.StatusCode != http.StatusOK {
		t.Fatalf("expected 200 for custom icon, got %d", res.StatusCode)
	}
}

func TestServer_DevMode(t *testing.T) {
	initGlobals(config.Config{
		LogLevel: "INFO",
		Host:     "localhost",
		Port:     "3000",
		DevMode:  true,
		Portals:  []models.Portal{},
		Pages:    []models.Page{},
	})
	// Creating the server with DevMode=true registers the /reload handler
	srv := server.NewServer(testLogger(), static)

	// Verify the /reload endpoint exists and responds
	req := httptest.NewRequest(http.MethodGet, "/reload", nil)
	rec := httptest.NewRecorder()
	srv.ServeHTTP(rec, req)

	// The reload handler expects a WebSocket upgrade; with a plain GET
	// it returns 426 Upgrade Required, which confirms the route is registered.
	if rec.Code != http.StatusUpgradeRequired {
		t.Fatalf("expected 426 for dev mode reload, got %d", rec.Code)
	}
}
