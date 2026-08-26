package config

import (
	"flag"
	"fmt"
	"io"
	"log/slog"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"github.com/go-viper/mapstructure/v2"
	"github.com/kodehat/portkey/internal/models"
	"github.com/spf13/viper"
)

// Favicon loading modes. In "direct" mode portkey discovers favicons itself
// via the favifetch library (parsing the target site's HTML, manifest, and
// common fallback paths). In "proxied" mode portkey relays each request to a
// Vemetric-compatible favicon service (configured via faviconServiceURL),
// which performs all discovery; the relayed result is cached like a direct
// fetch.
const (
	FaviconModeDirect  = "direct"
	FaviconModeProxied = "proxied"
)

type Config struct {
	Server  ServerConfig
	Metrics MetricsConfig
	UI      UIConfig
	Search  SearchConfig
	Favicon FaviconConfig
	Auth    AuthConfig
	Portals []models.Portal
	Pages   []models.Page
}

// AuthConfig configures WebAuthn passkey authentication.
type AuthConfig struct {
	Enabled         bool
	RPId            string
	RPOrigin        string
	CredentialsFile string
}

// ServerConfig holds the HTTP server and logging settings.
type ServerConfig struct {
	LogLevel    string
	LogJson     bool
	Host        string
	Port        string
	ContextPath string
	DevMode     bool
}

// MetricsConfig configures the optional Prometheus metrics server.
type MetricsConfig struct {
	Enabled bool
	Host    string
	Port    string
}

// UIConfig holds front-page appearance settings.
type UIConfig struct {
	Title                  string
	ShowSearchBar          bool
	ShowTopIcon            bool
	ShowKeywordsAsTooltips bool
	SortAlphabetically     bool
	LayoutColumns          int
	HeaderAddition         string
	Footer                 string
}

// SearchConfig configures the portal search behavior.
type SearchConfig struct {
	StringSimilarity  bool
	MinimumSimilarity float64
}

// FaviconConfig configures favicon loading, caching, and custom icons.
type FaviconConfig struct {
	Mode           string
	ServiceURL     string
	CacheDir       string
	CacheEnabled   bool
	CustomIconsDir string
}

type Flags struct {
	ConfigPath string
}

type RuntimeConfig struct {
	WithGroups bool
}

var C Config
var F Flags
var R RuntimeConfig

func Load() {
	LoadFlags()
	_, err := filepath.Abs(F.ConfigPath)
	if err != nil {
		panic(fmt.Errorf("fatal error config file: %w", err))
	}
	loadConfig(F.ConfigPath)
}

func LoadFlags() {
	var configPath string
	workDir, err := os.Getwd()
	if err != nil {
		workDir = "."
	}
	flag.StringVar(&configPath, "config-path", workDir, "path where config.yml can be found")
	flag.Parse()
	F = Flags{
		ConfigPath: configPath,
	}
}

func loadConfig(configPath string) {
	viper.SetConfigName("config")
	viper.SetConfigType("yml")
	viper.AddConfigPath(configPath)
	viper.SetDefault("server.logLevel", "INFO")
	viper.SetDefault("server.logJson", false)
	viper.SetDefault("server.host", "localhost")
	viper.SetDefault("server.port", "3000")
	viper.SetDefault("server.contextPath", "")
	viper.SetDefault("server.devMode", false)
	viper.SetDefault("metrics.enabled", false)
	viper.SetDefault("metrics.host", "localhost")
	viper.SetDefault("metrics.port", "3030")
	viper.SetDefault("ui.title", "portkey")
	viper.SetDefault("ui.showSearchBar", true)
	viper.SetDefault("ui.showTopIcon", true)
	viper.SetDefault("ui.showKeywordsAsTooltips", false)
	viper.SetDefault("ui.sortAlphabetically", false)
	viper.SetDefault("ui.layoutColumns", 0)
	viper.SetDefault("ui.headerAddition", "")
	viper.SetDefault("ui.footer", "Works like a portal.")
	viper.SetDefault("search.stringSimilarity", false)
	viper.SetDefault("search.minimumSimilarity", 0.75)
	viper.SetDefault("favicon.mode", FaviconModeDirect)
	viper.SetDefault("favicon.serviceUrl", "")
	viper.SetDefault("favicon.cacheDir", "")
	viper.SetDefault("favicon.cacheEnabled", false)
	viper.SetDefault("favicon.customIconsDir", "")
	viper.SetDefault("auth.enabled", false)
	viper.SetDefault("auth.rpId", "")
	viper.SetDefault("auth.rpOrigin", "")
	viper.SetDefault("auth.credentialsFile", "/opt/portkey-credentials.json")
	viper.SetEnvPrefix("portkey")
	// Nested config keys (e.g. "server.port") map to PORTKEY_SERVER_PORT instead
	// of PORTKEY_SERVER.PORT when looking up environment variables.
	viper.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))
	viper.AutomaticEnv()
	err := viper.ReadInConfig()
	if err != nil {
		panic(fmt.Errorf("fatal error config file: %w", err))
	}
	// Unknown keys (e.g. old flat names like "faviconMode" or "hideTitle")
	// fail loudly so stale configs from previous major versions are caught.
	err = viper.Unmarshal(&C, viper.DecoderConfigOption(func(dc *mapstructure.DecoderConfig) {
		dc.ErrorUnused = true
	}))
	if err != nil {
		panic(fmt.Errorf("fatal error decoding config: %w", err))
	}

	postConfigHook()
}

// postConfigHook is used to make dynamic changes to already loaded config values.
func postConfigHook() {
	// Normalize the favicon loading mode. Empty or unrecognized values fall
	// back to "direct" so a typo never breaks startup; the favicon path is
	// non-critical and "direct" is always a safe default.
	switch strings.ToLower(strings.TrimSpace(C.Favicon.Mode)) {
	case FaviconModeProxied:
		C.Favicon.Mode = FaviconModeProxied
	default:
		C.Favicon.Mode = FaviconModeDirect
	}

	// The on-disk cache requires a directory. Fail fast on the misconfiguration
	// instead of silently running without caching.
	if C.Favicon.CacheEnabled && strings.TrimSpace(C.Favicon.CacheDir) == "" {
		panic(fmt.Errorf("favicon.cacheEnabled is true but favicon.cacheDir is empty: set favicon.cacheDir or disable the cache with favicon.cacheEnabled: false"))
	}

	if C.UI.SortAlphabetically {
		sort.Slice(C.Portals, func(i, j int) bool {
			return strings.ToLower(C.Portals[i].Title) < strings.ToLower(C.Portals[j].Title)
		})
	}

	if C.Server.ContextPath != "" {
		for i := range C.Portals {
			if !C.Portals[i].IsExternal() {
				C.Portals[i].Link = C.Server.ContextPath + C.Portals[i].Link
			}
		}

		for i := range C.Pages {
			C.Pages[i].Path = C.Server.ContextPath + C.Pages[i].Path
		}
	}

	for _, portal := range C.Portals {
		if portal.Group != "" {
			R.WithGroups = true
			break
		}
	}
}

func (c Config) GetLogLevel() (slog.Level, error) {
	var level slog.Level
	err := level.UnmarshalText([]byte(c.Server.LogLevel))
	return level, err
}

func (c Config) GetLogHandler(w io.Writer) slog.Handler {
	logLevel, err := c.GetLogLevel()
	if err != nil {
		panic(fmt.Errorf("unable to unmarshal log level: %w", err))
	}
	logHandlerOptions := &slog.HandlerOptions{Level: logLevel}
	if c.Server.LogJson {
		return slog.NewJSONHandler(w, logHandlerOptions)
	}
	return slog.NewTextHandler(w, logHandlerOptions)
}
