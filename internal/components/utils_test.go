package components

import (
	"context"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/kodehat/portkey/internal/build"
)

func TestVersionComponent(t *testing.T) {
	details := build.BuildDetails{
		Version:    "1.0.0",
		CommitHash: "abc123",
		BuildTime:  "2024-01-01",
		GoVersion:  "go1.21",
	}
	rec := httptest.NewRecorder()
	Version(details).Render(context.Background(), rec)

	body := rec.Body.String()
	if !strings.Contains(body, "1.0.0") {
		t.Fatal("expected version in output")
	}
	if !strings.Contains(body, "abc123") {
		t.Fatal("expected commit hash in output")
	}
	if !strings.Contains(body, "2024-01-01") {
		t.Fatal("expected build time in output")
	}
	if !strings.Contains(body, "go1.21") {
		t.Fatal("expected go version in output")
	}
}

func TestNotFoundComponent(t *testing.T) {
	rec := httptest.NewRecorder()
	NotFound().Render(context.Background(), rec)

	body := rec.Body.String()
	if !strings.Contains(body, "404") {
		t.Fatal("expected 404 in output")
	}
	if !strings.Contains(body, "Back to home") {
		t.Fatal("expected 'Back to home' link in output")
	}
}

func TestDevModeSnippet(t *testing.T) {
	rec := httptest.NewRecorder()
	DevModeSnippet("/").Render(context.Background(), rec)

	body := rec.Body.String()
	if !strings.Contains(body, "<script") {
		t.Fatal("expected script tag in dev mode snippet")
	}
}
