package config

import (
	"io"
	"log/slog"
	"os"
	"path/filepath"
	"strings"
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

func TestLoadFlags_SetsConfigPath(t *testing.T) {
	LoadFlags()
	if F.ConfigPath == "" {
		t.Fatal("expected ConfigPath to be set")
	}
}

func TestLoadConfig_FromTempFile(t *testing.T) {
	dir := t.TempDir()
	yaml := `logLevel: DEBUG
host: 0.0.0.0
port: "8080"
title: "Test Portal"
portals: []
pages: []
`
	if err := os.WriteFile(filepath.Join(dir, "config.yml"), []byte(yaml), 0644); err != nil {
		t.Fatal(err)
	}

	C = Config{}
	loadConfig(dir)

	if strings.ToLower(C.LogLevel) != "debug" {
		t.Fatalf("expected LogLevel DEBUG (case-insensitive), got %q", C.LogLevel)
	}
	if C.Host != "0.0.0.0" {
		t.Fatalf("expected Host 0.0.0.0, got %q", C.Host)
	}
	if C.Port != "8080" {
		t.Fatalf("expected Port 8080, got %q", C.Port)
	}
	if C.Title != "Test Portal" {
		t.Fatalf("expected Title 'Test Portal', got %q", C.Title)
	}
}

func TestLoadConfig_WithPortalsAndPages(t *testing.T) {
	dir := t.TempDir()
	yaml := `title: "My Portal"
portals:
  - title: "GitHub"
    link: "https://github.com"
    keywords: ["code"]
    group: "Dev"
pages:
  - heading: "About"
    path: /about
    content: "<p>info</p>"
`
	if err := os.WriteFile(filepath.Join(dir, "config.yml"), []byte(yaml), 0644); err != nil {
		t.Fatal(err)
	}

	C = Config{}
	loadConfig(dir)

	if len(C.Portals) != 1 || C.Portals[0].Title != "GitHub" {
		t.Fatalf("expected 1 portal 'GitHub', got %+v", C.Portals)
	}
	if len(C.Pages) != 1 || C.Pages[0].Heading != "About" {
		t.Fatalf("expected 1 page 'About', got %+v", C.Pages)
	}
}

func TestLoadConfig_EnvOverride(t *testing.T) {
	os.Setenv("PORTKEY_PORT", "9999")
	defer os.Unsetenv("PORTKEY_PORT")

	dir := t.TempDir()
	yaml := `title: "Test"
portals: []
pages: []
`
	if err := os.WriteFile(filepath.Join(dir, "config.yml"), []byte(yaml), 0644); err != nil {
		t.Fatal(err)
	}

	C = Config{}
	loadConfig(dir)

	if C.Port != "9999" {
		t.Fatalf("expected Port 9999 from env, got %q", C.Port)
	}
}

func TestLoadConfig_MissingFile(t *testing.T) {
	defer func() {
		if recover() == nil {
			t.Fatal("expected panic for missing config file")
		}
	}()
	C = Config{}
	loadConfig("/nonexistent/path")
}

func TestLoadConfig_InvalidYAML(t *testing.T) {
	dir := t.TempDir()
	yaml := `:invalid yaml content {{`
	if err := os.WriteFile(filepath.Join(dir, "config.yml"), []byte(yaml), 0644); err != nil {
		t.Fatal(err)
	}

	defer func() {
		if recover() == nil {
			t.Fatal("expected panic for invalid YAML")
		}
	}()
	C = Config{}
	loadConfig(dir)
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

func TestGetLogHandler_InvalidLogLevelPanics(t *testing.T) {
	c := Config{LogLevel: "INVALID", LogJson: false}
	defer func() {
		if recover() == nil {
			t.Fatal("expected panic for invalid log level")
		}
	}()
	c.GetLogHandler(io.Discard)
}

func TestGetLogLevelEmpty(t *testing.T) {
	c := Config{LogLevel: ""}
	_, err := c.GetLogLevel()
	if err == nil {
		t.Fatal("expected error for empty log level")
	}
}

func TestPostConfigHook_EmptyConfig(t *testing.T) {
	c := Config{}
	C = c
	R = RuntimeConfig{}
	postConfigHook()

	if R.WithGroups {
		t.Fatal("expected WithGroups to be false with empty config")
	}
}

func TestPostConfigHook_WithContextPathEmptyPages(t *testing.T) {
	c := Config{
		ContextPath: "/app",
		Portals:     []models.Portal{},
		Pages:       []models.Page{},
	}
	C = c
	postConfigHook()

	if len(C.Portals) != 0 {
		t.Fatalf("expected no portals, got %d", len(C.Portals))
	}
}
