package components

import (
	"context"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/kodehat/portkey/internal/build"
	"github.com/kodehat/portkey/internal/config"
	"github.com/kodehat/portkey/internal/models"
)

func TestBase_Render(t *testing.T) {
	cfg := config.Config{
		Title:       "portkey",
		ContextPath: "",
	}
	details := build.BuildDetails{
		Version:    "1.0.0",
		BuildTime:  "2024-01-01",
		GoVersion:  "go1.21",
		CssHash:    "abc123",
	}
	build.LoadBuildDetails("abc123")
	rec := httptest.NewRecorder()
	Base("Test", "subtitle", cfg, details).Render(context.Background(), rec)

	body := rec.Body.String()
	if !strings.Contains(body, "portkey") {
		t.Fatal("expected config title in output")
	}
	if !strings.Contains(body, "Test") {
		t.Fatal("expected page title in output")
	}
	if !strings.Contains(body, "<html") {
		t.Fatal("expected html tag in output")
	}
	if !strings.Contains(body, "abc123") {
		t.Fatal("expected CSS hash in output")
	}
	if !strings.Contains(body, "dev") {
		t.Fatal("expected version comment in output")
	}
}

func TestHomeLayout(t *testing.T) {
	cfg := config.Config{
		Title:       "portkey",
		ShowTopIcon: true,
		ContextPath: "",
		Footer:      "test footer",
	}
	details := build.BuildDetails{CssHash: "test"}
	rec := httptest.NewRecorder()
	HomeLayout("Home", cfg, details, ContentPage("home content")).Render(context.Background(), rec)

	body := rec.Body.String()
	if !strings.Contains(body, "portkey") {
		t.Fatal("expected config title in output")
	}
	if !strings.Contains(body, "android-chrome-512x512.png") {
		t.Fatal("expected logo image in output")
	}
	if !strings.Contains(body, "test footer") {
		t.Fatal("expected footer in output")
	}
	if !strings.Contains(body, "home content") {
		t.Fatal("expected content in output")
	}
}

func TestContentLayout(t *testing.T) {
	cfg := config.Config{
		Title:       "portkey",
		ContextPath: "",
		Footer:      "test footer",
		Portals:     []models.Portal{},
	}
	details := build.BuildDetails{CssHash: "test"}
	rec := httptest.NewRecorder()
	ContentLayout("About", "Info", cfg, details, ContentPage("about content")).Render(context.Background(), rec)

	body := rec.Body.String()
	if !strings.Contains(body, "About") {
		t.Fatal("expected page title in output")
	}
	if !strings.Contains(body, "Info") {
		t.Fatal("expected page subtitle in output")
	}
	if !strings.Contains(body, "about content") {
		t.Fatal("expected content in output")
	}
	if !strings.Contains(body, "test footer") {
		t.Fatal("expected footer in output")
	}
}
