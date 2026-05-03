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

func TestBase_WithHeaderAddition(t *testing.T) {
	cfg := config.Config{
		Title:          "portkey",
		ContextPath:    "",
		HeaderAddition: `<meta name="custom" content="test"/>`,
	}
	details := build.BuildDetails{CssHash: "test"}
	build.LoadBuildDetails("test")
	rec := httptest.NewRecorder()
	Base("Test", "", cfg, details).Render(context.Background(), rec)

	body := rec.Body.String()
	if !strings.Contains(body, `content="test"`) {
		t.Fatal("expected header addition in output")
	}
}

func TestBase_DevMode(t *testing.T) {
	cfg := config.Config{
		Title:       "portkey",
		ContextPath: "",
		DevMode:     true,
	}
	details := build.BuildDetails{CssHash: "test"}
	rec := httptest.NewRecorder()
	Base("Test", "", cfg, details).Render(context.Background(), rec)

	body := rec.Body.String()
	if !strings.Contains(body, "<script") {
		t.Fatal("expected dev mode script in output")
	}
}

func TestBase_WithCommitHash(t *testing.T) {
	cfg := config.Config{
		Title:       "portkey",
		ContextPath: "",
	}
	details := build.BuildDetails{CssHash: "test", CommitHash: "abc123def"}
	build.LoadBuildDetails("test")
	build.CommitHash = "abc123def"
	build.BuildTime = "2024-01-01"
	build.Version = "1.0.0"
	build.GoVersion = "go1.21"
	rec := httptest.NewRecorder()
	Base("Test", "", cfg, details).Render(context.Background(), rec)

	body := rec.Body.String()
	if !strings.Contains(body, "abc123def") {
		t.Fatalf("expected commit hash in output, got body length %d", len(body))
	}
	if !strings.Contains(body, "1.0.0") {
		t.Fatal("expected version in output")
	}
	if !strings.Contains(body, "go1.21") {
		t.Fatal("expected go version in output")
	}
}

func TestContentLayout_WithPortals(t *testing.T) {
	cfg := config.Config{
		Title:       "portkey",
		ContextPath: "",
		Footer:      "test footer",
		Portals: []models.Portal{
			{Title: "GitHub", Link: "https://github.com"},
		},
	}
	details := build.BuildDetails{CssHash: "test"}
	rec := httptest.NewRecorder()
	ContentLayout("About", "", cfg, details, ContentPage("content")).Render(context.Background(), rec)

	body := rec.Body.String()
	if !strings.Contains(body, "GitHub") {
		t.Fatal("expected portal in footer")
	}
}

func TestHomeLayout_HideTitle(t *testing.T) {
	cfg := config.Config{
		Title:       "portkey",
		HideTitle:   true,
		ContextPath: "",
	}
	details := build.BuildDetails{CssHash: "test"}
	rec := httptest.NewRecorder()
	HomeLayout("Home", cfg, details, ContentPage("home content")).Render(context.Background(), rec)

	body := rec.Body.String()
	if !strings.Contains(body, "hidden") {
		t.Fatal("expected hidden class when HideTitle is true")
	}
}

func TestHomeLayout_NoTopIcon(t *testing.T) {
	cfg := config.Config{
		Title:       "portkey",
		ShowTopIcon: false,
		ContextPath: "",
	}
	details := build.BuildDetails{CssHash: "test"}
	rec := httptest.NewRecorder()
	HomeLayout("Home", cfg, details, ContentPage("home content")).Render(context.Background(), rec)

	body := rec.Body.String()
	if !strings.Contains(body, "hidden") {
		t.Fatal("expected hidden class when ShowTopIcon is false")
	}
}

func TestContentLayout_HideTitle(t *testing.T) {
	cfg := config.Config{
		Title:       "portkey",
		HideTitle:   true,
		ContextPath: "",
		Footer:      "footer",
		Portals:     []models.Portal{},
	}
	details := build.BuildDetails{CssHash: "test"}
	rec := httptest.NewRecorder()
	ContentLayout("About", "subtitle", cfg, details, ContentPage("content")).Render(context.Background(), rec)

	body := rec.Body.String()
	if !strings.Contains(body, "hidden") {
		t.Fatal("expected hidden class when HideTitle is true")
	}
}
