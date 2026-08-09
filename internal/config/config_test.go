package config

import (
	"flag"
	"io"
	"log/slog"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/kodehat/portkey/internal/models"
)

func TestGetLogLevel(t *testing.T) {
	c := Config{Server: ServerConfig{LogLevel: "INFO"}}
	level, err := c.GetLogLevel()
	if err != nil {
		t.Fatal(err)
	}
	if level != slog.LevelInfo {
		t.Fatalf("expected LevelInfo, got %d", level)
	}
}

func TestGetLogLevelDebug(t *testing.T) {
	c := Config{Server: ServerConfig{LogLevel: "DEBUG"}}
	level, err := c.GetLogLevel()
	if err != nil {
		t.Fatal(err)
	}
	if level != slog.LevelDebug {
		t.Fatalf("expected LevelDebug, got %d", level)
	}
}

func TestGetLogLevelWarn(t *testing.T) {
	c := Config{Server: ServerConfig{LogLevel: "WARN"}}
	level, err := c.GetLogLevel()
	if err != nil {
		t.Fatal(err)
	}
	if level != slog.LevelWarn {
		t.Fatalf("expected LevelWarn, got %d", level)
	}
}

func TestGetLogLevelError(t *testing.T) {
	c := Config{Server: ServerConfig{LogLevel: "ERROR"}}
	level, err := c.GetLogLevel()
	if err != nil {
		t.Fatal(err)
	}
	if level != slog.LevelError {
		t.Fatalf("expected LevelError, got %d", level)
	}
}

func TestGetLogLevelInvalid(t *testing.T) {
	c := Config{Server: ServerConfig{LogLevel: "INVALID"}}
	_, err := c.GetLogLevel()
	if err == nil {
		t.Fatal("expected error for invalid log level")
	}
}

func TestGetLogHandlerText(t *testing.T) {
	c := Config{Server: ServerConfig{LogLevel: "INFO", LogJson: false}}
	handler := c.GetLogHandler(io.Discard)
	if handler == nil {
		t.Fatal("expected non-nil handler")
	}
}

func TestGetLogHandlerJSON(t *testing.T) {
	c := Config{Server: ServerConfig{LogLevel: "INFO", LogJson: true}}
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
	yaml := `server:
  logLevel: DEBUG
  host: 0.0.0.0
  port: "8080"
ui:
  title: "Test Portal"
portals: []
pages: []
`
	if err := os.WriteFile(filepath.Join(dir, "config.yml"), []byte(yaml), 0644); err != nil {
		t.Fatal(err)
	}

	C = Config{}
	loadConfig(dir)

	if strings.ToLower(C.Server.LogLevel) != "debug" {
		t.Fatalf("expected LogLevel DEBUG (case-insensitive), got %q", C.Server.LogLevel)
	}
	if C.Server.Host != "0.0.0.0" {
		t.Fatalf("expected Host 0.0.0.0, got %q", C.Server.Host)
	}
	if C.Server.Port != "8080" {
		t.Fatalf("expected Port 8080, got %q", C.Server.Port)
	}
	if C.UI.Title != "Test Portal" {
		t.Fatalf("expected Title 'Test Portal', got %q", C.UI.Title)
	}
}

func TestLoadConfig_WithPortalsAndPages(t *testing.T) {
	dir := t.TempDir()
	yaml := `ui:
  title: "My Portal"
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
	os.Setenv("PORTKEY_SERVER_PORT", "9999")
	defer os.Unsetenv("PORTKEY_SERVER_PORT")

	dir := t.TempDir()
	yaml := `ui:
  title: "Test"
portals: []
pages: []
`
	if err := os.WriteFile(filepath.Join(dir, "config.yml"), []byte(yaml), 0644); err != nil {
		t.Fatal(err)
	}

	C = Config{}
	loadConfig(dir)

	if C.Server.Port != "9999" {
		t.Fatalf("expected Port 9999 from env, got %q", C.Server.Port)
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

		UI: UIConfig{

			SortAlphabetically: true,
		},

		Portals: []models.Portal{
			{Title: "Zebra", Link: "/z"},
			{Title: "Alpha", Link: "/a"},
			{Title: "beta", Link: "/b"},
		}}
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

		Server: ServerConfig{

			ContextPath: "/app",
		},

		Portals: []models.Portal{
			{Title: "Internal", Link: "/internal"},
			{Title: "External", Link: "https://external.com"},
		},

		Pages: []models.Page{
			{Heading: "About", Path: "/about"},
		}}
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
		}}
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
		}}
	C = c
	R = RuntimeConfig{}
	postConfigHook()

	if R.WithGroups {
		t.Fatal("expected WithGroups to be false")
	}
}

func TestGetLogHandler_InvalidLogLevelPanics(t *testing.T) {
	c := Config{Server: ServerConfig{LogLevel: "INVALID", LogJson: false}}
	defer func() {
		if recover() == nil {
			t.Fatal("expected panic for invalid log level")
		}
	}()
	c.GetLogHandler(io.Discard)
}

func TestGetLogLevelEmpty(t *testing.T) {
	c := Config{Server: ServerConfig{LogLevel: ""}}
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

		Server: ServerConfig{

			ContextPath: "/app",
		},

		Portals: []models.Portal{},

		Pages: []models.Page{}}
	C = c
	postConfigHook()

	if len(C.Portals) != 0 {
		t.Fatalf("expected no portals, got %d", len(C.Portals))
	}
}

func TestLoad_Success(t *testing.T) {
	dir := t.TempDir()
	yaml := `ui:
  title: "LoadTest"
server:
  host: localhost
  port: "9090"
portals: []
pages: []
`
	if err := os.WriteFile(filepath.Join(dir, "config.yml"), []byte(yaml), 0644); err != nil {
		t.Fatal(err)
	}

	// Save and restore the config-path flag state
	oldConfigPath := F.ConfigPath
	defer func() { F.ConfigPath = oldConfigPath }()

	F.ConfigPath = dir
	C = Config{}
	loadConfig(dir)

	if C.UI.Title != "LoadTest" {
		t.Fatalf("expected Title 'LoadTest', got %q", C.UI.Title)
	}
	if C.Server.Port != "9090" {
		t.Fatalf("expected Port 9090, got %q", C.Server.Port)
	}
}

// TestLoadCall exercises the full Load() function path by temporarily
// replacing flag.CommandLine to avoid flag re-registration conflicts.
func TestLoadCall(t *testing.T) {
	dir := t.TempDir()
	yaml := "ui:\n  title: T\nportals: []\npages: []\n"
	if err := os.WriteFile(filepath.Join(dir, "config.yml"), []byte(yaml), 0644); err != nil {
		t.Fatal(err)
	}

	// Save original flag set and args.
	oldFlag := flag.CommandLine
	oldArgs := os.Args
	defer func() {
		flag.CommandLine = oldFlag
		os.Args = oldArgs
	}()

	// Replace flag set so LoadFlags() can register "config-path" without panic.
	flag.CommandLine = flag.NewFlagSet(os.Args[0], flag.ExitOnError)
	os.Args = []string{os.Args[0], "-config-path=" + dir}

	Load()

	if C.UI.Title != "T" {
		t.Fatalf("expected Title T, got %q", C.UI.Title)
	}
}

func TestLoadConfig_DefaultFaviconMode(t *testing.T) {
	dir := t.TempDir()
	yaml := `ui:
  title: "Test"
portals: []
pages: []
`
	if err := os.WriteFile(filepath.Join(dir, "config.yml"), []byte(yaml), 0644); err != nil {
		t.Fatal(err)
	}
	C = Config{}
	loadConfig(dir)
	if C.Favicon.Mode != FaviconModeDirect {
		t.Fatalf("expected default faviconMode %q, got %q", FaviconModeDirect, C.Favicon.Mode)
	}
}

func TestLoadConfig_ProxiedFaviconMode(t *testing.T) {
	dir := t.TempDir()
	yaml := `ui:
  title: "Test"
favicon:
  mode: proxied
portals: []
pages: []
`
	if err := os.WriteFile(filepath.Join(dir, "config.yml"), []byte(yaml), 0644); err != nil {
		t.Fatal(err)
	}
	C = Config{}
	loadConfig(dir)
	if C.Favicon.Mode != FaviconModeProxied {
		t.Fatalf("expected faviconMode %q, got %q", FaviconModeProxied, C.Favicon.Mode)
	}
}

func TestPostConfigHook_NormalizesFaviconMode(t *testing.T) {
	tests := []struct {
		name  string
		input string
		want  string
	}{
		{"empty", "", FaviconModeDirect},
		{"direct", "direct", FaviconModeDirect},
		{"direct uppercase", "DIRECT", FaviconModeDirect},
		{"proxied", "proxied", FaviconModeProxied},
		{"proxied uppercase", "PROXIED", FaviconModeProxied},
		{"invalid", "invalid", FaviconModeDirect},
		{"proxied with spaces", "  proxied  ", FaviconModeProxied},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			C = Config{Favicon: FaviconConfig{Mode: tt.input}}
			postConfigHook()
			if C.Favicon.Mode != tt.want {
				t.Errorf("input %q: got %q, want %q", tt.input, C.Favicon.Mode, tt.want)
			}
		})
	}
}

func TestPostConfigHook_CacheEnabledRequiresCacheDir(t *testing.T) {
	t.Run("enabled without dir panics", func(t *testing.T) {
		C = Config{Favicon: FaviconConfig{CacheEnabled: true, CacheDir: ""}}
		defer func() {
			if recover() == nil {
				t.Fatal("expected panic when cacheEnabled is true but cacheDir is empty")
			}
		}()
		postConfigHook()
	})

	t.Run("enabled with dir is fine", func(t *testing.T) {
		C = Config{Favicon: FaviconConfig{CacheEnabled: true, CacheDir: "./favicon-cache"}}
		postConfigHook() // must not panic
	})

	t.Run("disabled without dir is fine", func(t *testing.T) {
		C = Config{Favicon: FaviconConfig{CacheEnabled: false, CacheDir: ""}}
		postConfigHook() // must not panic
	})
}

func TestLoadConfig_CacheDisabledByDefault(t *testing.T) {
	dir := t.TempDir()
	yaml := `ui:
  title: "Test"
portals: []
pages: []
`
	if err := os.WriteFile(filepath.Join(dir, "config.yml"), []byte(yaml), 0644); err != nil {
		t.Fatal(err)
	}
	C = Config{}
	loadConfig(dir)
	if C.Favicon.CacheEnabled {
		t.Fatal("expected favicon cache to be disabled by default")
	}
	if C.Favicon.CacheDir != "" {
		t.Fatalf("expected empty favicon.cacheDir by default, got %q", C.Favicon.CacheDir)
	}
}
