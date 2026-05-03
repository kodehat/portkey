package config

import (
	"io"
	"log/slog"
	"testing"

	"github.com/kodehat/portkey/internal/models"
)

func TestGetLogLevel(t *testing.T) {
	c := Config{LogLevel: "INFO"}
	level, err := c.GetLogLevel()
	if err != nil {
		t.Fatal(err)
	}
	if level != slog.LevelInfo {
		t.Fatalf("expected LevelInfo, got %d", level)
	}
}

func TestGetLogLevelDebug(t *testing.T) {
	c := Config{LogLevel: "DEBUG"}
	level, err := c.GetLogLevel()
	if err != nil {
		t.Fatal(err)
	}
	if level != slog.LevelDebug {
		t.Fatalf("expected LevelDebug, got %d", level)
	}
}

func TestGetLogLevelWarn(t *testing.T) {
	c := Config{LogLevel: "WARN"}
	level, err := c.GetLogLevel()
	if err != nil {
		t.Fatal(err)
	}
	if level != slog.LevelWarn {
		t.Fatalf("expected LevelWarn, got %d", level)
	}
}

func TestGetLogLevelError(t *testing.T) {
	c := Config{LogLevel: "ERROR"}
	level, err := c.GetLogLevel()
	if err != nil {
		t.Fatal(err)
	}
	if level != slog.LevelError {
		t.Fatalf("expected LevelError, got %d", level)
	}
}

func TestGetLogLevelInvalid(t *testing.T) {
	c := Config{LogLevel: "INVALID"}
	_, err := c.GetLogLevel()
	if err == nil {
		t.Fatal("expected error for invalid log level")
	}
}

func TestGetLogHandlerText(t *testing.T) {
	c := Config{LogLevel: "INFO", LogJson: false}
	handler := c.GetLogHandler(io.Discard)
	if handler == nil {
		t.Fatal("expected non-nil handler")
	}
}

func TestGetLogHandlerJSON(t *testing.T) {
	c := Config{LogLevel: "INFO", LogJson: true}
	handler := c.GetLogHandler(io.Discard)
	if handler == nil {
		t.Fatal("expected non-nil handler")
	}
}

func TestPostConfigHook_SortAlphabetically(t *testing.T) {
	c := Config{
		SortAlphabetically: true,
		Portals: []models.Portal{
			{Title: "Zebra", Link: "/z"},
			{Title: "Alpha", Link: "/a"},
			{Title: "beta", Link: "/b"},
		},
	}
	C = c
	postConfigHook()

	if C.Portals[0].Title != "Alpha" {
		t.Fatalf("expected first portal 'Alpha', got %q", C.Portals[0].Title)
	}
	if C.Portals[1].Title != "beta" {
		t.Fatalf("expected second portal 'beta', got %q", C.Portals[1].Title)
	}
	if C.Portals[2].Title != "Zebra" {
		t.Fatalf("expected third portal 'Zebra', got %q", C.Portals[2].Title)
	}
}

func TestPostConfigHook_WithContextPath(t *testing.T) {
	c := Config{
		ContextPath: "/app",
		Portals: []models.Portal{
			{Title: "Internal", Link: "/internal"},
			{Title: "External", Link: "https://external.com"},
		},
		Pages: []models.Page{
			{Heading: "About", Path: "/about"},
		},
	}
	C = c
	postConfigHook()

	if C.Portals[0].Link != "/app/internal" {
		t.Fatalf("expected internal link /app/internal, got %q", C.Portals[0].Link)
	}
	if C.Portals[1].Link != "https://external.com" {
		t.Fatalf("expected external link unchanged, got %q", C.Portals[1].Link)
	}
	if C.Pages[0].Path != "/app/about" {
		t.Fatalf("expected page path /app/about, got %q", C.Pages[0].Path)
	}
}

func TestPostConfigHook_WithGroups(t *testing.T) {
	c := Config{
		Portals: []models.Portal{
			{Title: "A", Link: "/a", Group: "dev"},
			{Title: "B", Link: "/b"},
		},
	}
	C = c
	R = RuntimeConfig{}
	postConfigHook()

	if !R.WithGroups {
		t.Fatal("expected WithGroups to be true")
	}
}

func TestPostConfigHook_NoGroups(t *testing.T) {
	c := Config{
		Portals: []models.Portal{
			{Title: "A", Link: "/a"},
			{Title: "B", Link: "/b"},
		},
	}
	C = c
	R = RuntimeConfig{}
	postConfigHook()

	if R.WithGroups {
		t.Fatal("expected WithGroups to be false")
	}
}
