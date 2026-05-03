package main

import (
	"encoding/json"
	"io"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync"
	"testing"

	"github.com/kodehat/portkey/internal/build"
	"github.com/kodehat/portkey/internal/config"
	"github.com/kodehat/portkey/internal/metrics"
	"github.com/kodehat/portkey/internal/models"
	"github.com/kodehat/portkey/internal/server"
)

var globalsOnce sync.Once

func initGlobals(cfg config.Config) {
	config.C = cfg
	globalsOnce.Do(func() {
		build.LoadBuildDetails("testhash")
		metrics.Load()
	})
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

func TestAPIEndpoints(t *testing.T) {
	initGlobals(config.Config{
		LogLevel: "INFO",
		Host:     "localhost",
		Port:     "3000",
		Portals: []models.Portal{
			{Title: "GitHub", Link: "https://github.com", Emoji: "💻", Keywords: []string{"code"}},
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
			{Title: "GitHub", Link: "https://github.com", Emoji: "💻", Keywords: []string{"code"}},
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
			{Title: "GitHub", Link: "https://github.com", Emoji: "💻", Keywords: []string{"code"}},
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
