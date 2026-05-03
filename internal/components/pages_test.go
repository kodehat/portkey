package components

import (
	"context"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/kodehat/portkey/internal/config"
)

func TestHomePage_SearchBar(t *testing.T) {
	config.C = config.Config{HideSearchBar: false, ContextPath: ""}
	rec := httptest.NewRecorder()
	HomePage().Render(context.Background(), rec)

	body := rec.Body.String()
	if !strings.Contains(body, `<input`) {
		t.Fatal("expected search input in output")
	}
	if !strings.Contains(body, "Search...") {
		t.Fatal("expected search placeholder in output")
	}
}

func TestHomePage_Hidden(t *testing.T) {
	config.C = config.Config{HideSearchBar: true, ContextPath: ""}
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
	if !strings.Contains(body, "mt-4") {
		t.Fatal("expected mt-4 class for large margin")
	}
}

func TestLoadingBar_SmallMargin(t *testing.T) {
	rec := httptest.NewRecorder()
	loadingBar(false).Render(context.Background(), rec)

	body := rec.Body.String()
	if !strings.Contains(body, "htmx-indicator") {
		t.Fatal("expected htmx-indicator class in output")
	}
	if !strings.Contains(body, "mt-2") {
		t.Fatal("expected mt-2 class for small margin")
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

func TestHomePage_WithSubtitle(t *testing.T) {
	config.C = config.Config{
		Title:     "portkey",
		Subtitle:  "Where do you want to go?",
		ContextPath: "",
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
}

func TestHomePage_WithContextPath(t *testing.T) {
	config.C = config.Config{
		Title:       "portkey",
		ContextPath: "/app",
	}
	rec := httptest.NewRecorder()
	HomePage().Render(context.Background(), rec)

	body := rec.Body.String()
	if !strings.Contains(body, "/app") {
		t.Fatal("expected context path in output")
	}
}
