package components

import (
	"context"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/kodehat/portkey/internal/config"
)

func TestHomePage_SearchBar(t *testing.T) {
	config.C = config.Config{Server: config.ServerConfig{ContextPath: ""}, UI: config.UIConfig{ShowSearchBar: true}}
	rec := httptest.NewRecorder()
	HomePage().Render(context.Background(), rec)

	body := rec.Body.String()
	if !strings.Contains(body, `<input`) {
		t.Fatal("expected search input in output")
	}
	if !strings.Contains(body, "Search portals...") {
		t.Fatal("expected search placeholder in output")
	}
}

func TestHomePage_Hidden(t *testing.T) {
	config.C = config.Config{Server: config.ServerConfig{ContextPath: ""}, UI: config.UIConfig{ShowSearchBar: false}}
	rec := httptest.NewRecorder()
	HomePage().Render(context.Background(), rec)

	body := rec.Body.String()
	if !strings.Contains(body, "hidden") {
		t.Fatal("expected hidden class in output")
	}
	if !strings.Contains(body, `hx-get`) {
		t.Fatal("expected hx-get attribute in output")
	}
}

func TestLoadingBar_LargeMargin(t *testing.T) {
	rec := httptest.NewRecorder()
	loadingBar(true).Render(context.Background(), rec)

	body := rec.Body.String()
	if !strings.Contains(body, "htmx-indicator") {
		t.Fatal("expected htmx-indicator class in output")
	}
	if !strings.Contains(body, "mt-3") {
		t.Fatal("expected mt-3 class for large margin")
	}
}

func TestLoadingBar_SmallMargin(t *testing.T) {
	rec := httptest.NewRecorder()
	loadingBar(true).Render(context.Background(), rec)

	body := rec.Body.String()
	if !strings.Contains(body, "htmx-indicator") {
		t.Fatal("expected htmx-indicator class in output")
	}
	if !strings.Contains(body, "mt-3") {
		t.Fatal("expected mt-3 class when largeTopMargin is true")
	}
}

func TestContentPage(t *testing.T) {
	rec := httptest.NewRecorder()
	ContentPage("<p>hello world</p>").Render(context.Background(), rec)

	body := rec.Body.String()
	if !strings.Contains(body, "hello world") {
		t.Fatal("expected content in output")
	}
}

func TestHomePage(t *testing.T) {
	config.C = config.Config{

		Server: config.ServerConfig{

			ContextPath: "",
		},

		UI: config.UIConfig{

			Title: "portkey",
		},
	}
	rec := httptest.NewRecorder()
	HomePage().Render(context.Background(), rec)

	body := rec.Body.String()
	if !strings.Contains(body, "hx-get") {
		t.Fatal("expected hx-get attribute in output")
	}
}

func TestLoadingBar_NoMargin(t *testing.T) {
	rec := httptest.NewRecorder()
	loadingBar(false).Render(context.Background(), rec)

	body := rec.Body.String()
	if !strings.Contains(body, "htmx-indicator") {
		t.Fatal("expected htmx-indicator class in output")
	}
	if strings.Contains(body, "mt-3") {
		t.Fatal("did not expect mt-3 class when largeTopMargin is false")
	}
}

func TestHomePage_WithContextPath(t *testing.T) {
	config.C = config.Config{

		Server: config.ServerConfig{

			ContextPath: "/app",
		},

		UI: config.UIConfig{

			Title: "portkey",
		}}
	rec := httptest.NewRecorder()
	HomePage().Render(context.Background(), rec)

	body := rec.Body.String()
	if !strings.Contains(body, "/app") {
		t.Fatal("expected context path in output")
	}
}

func TestResultsContainer_Columns(t *testing.T) {
	config.C = config.Config{UI: config.UIConfig{LayoutColumns: 3}}
	rec := httptest.NewRecorder()
	ResultsContainer().Render(context.Background(), rec)

	body := rec.Body.String()
	if !strings.Contains(body, "max-w-7xl") {
		t.Fatal("expected wide column container class")
	}
}

func TestResultsContainer_NoColumns(t *testing.T) {
	config.C = config.Config{UI: config.UIConfig{LayoutColumns: 0}}
	rec := httptest.NewRecorder()
	ResultsContainer().Render(context.Background(), rec)

	body := rec.Body.String()
	if !strings.Contains(body, "search-results") {
		t.Fatal("expected search-results container")
	}
	if strings.Contains(body, "max-w-7xl") {
		t.Fatal("did not expect wide column container")
	}
}

func TestHomePage_WithColumns(t *testing.T) {
	config.C = config.Config{

		Server: config.ServerConfig{

			ContextPath: "",
		},

		UI: config.UIConfig{

			ShowSearchBar: true,

			LayoutColumns: 3,
		}}
	rec := httptest.NewRecorder()
	HomePage().Render(context.Background(), rec)

	body := rec.Body.String()
	if !strings.Contains(body, "max-w-7xl") {
		t.Fatal("expected wide column container in home page")
	}
	if !strings.Contains(body, `Search portals`) {
		t.Fatal("expected search bar in home page")
	}
}
