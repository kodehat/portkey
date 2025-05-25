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

	"github.com/kodehat/portkey/internal/models"
	"github.com/spf13/viper"
)

type Config struct {
	DevMode                    bool
	LogLevel                   string
	LogJson                    bool
	Host                       string
	Port                       string
	ContextPath                string
	EnableMetrics              bool
	MetricsHost                string
	MetricsPort                string
	Title                      string
	HideTitle                  bool
	Subtitle                   string
	Footer                     string
	ShowTopIcon                bool
	ShowKeywordsAsTooltips     bool
	SortAlphabetically         bool
	SearchWithStringSimilarity bool
	MinimumStringSimilarity    float64
	Portals                    []models.Portal
	Pages                      []models.Page
	HeaderAddition             string
	HideSearchBar              bool
}

type Flags struct {
	ConfigPath string
}

var C Config
var F Flags

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
	var envPrefix string
	workDir, err := os.Getwd()
	if err != nil {
		workDir = "."
	}
	flag.StringVar(&configPath, "config-path", workDir, "path where config.yml can be found")
	flag.StringVar(&envPrefix, "env-prefix", "", "prefix for environment variables")
	flag.Parse()
	F = Flags{
		ConfigPath: configPath,
	}
}

func loadConfig(configPath string) {
	viper.SetConfigName("config")
	viper.SetConfigType("yml")
	viper.AddConfigPath(configPath)
	viper.SetDefault("devMode", false)
	viper.SetDefault("logLevel", "INFO")
	viper.SetDefault("host", "localhost")
	viper.SetDefault("port", "1414")
	viper.SetDefault("contextPath", "")
	viper.SetDefault("metricsHost", "localhost")
	viper.SetDefault("metricsPort", "1515")
	viper.SetDefault("title", "Your Portal")
	viper.SetDefault("footerText", "Works like a portal.")
	viper.SetDefault("minimumStringSimilarity", 0.75)
	viper.SetDefault("headerAddition", "")
	viper.SetEnvPrefix("portkey")
	viper.AutomaticEnv()
	err := viper.ReadInConfig()
	if err != nil {
		panic(fmt.Errorf("fatal error config file: %w", err))
	}
	viper.Unmarshal(&C)

	postConfigHook()
}

// portConfigHook is used to make dynamic changes to already loaded config values.
func postConfigHook() {
	if C.SortAlphabetically {
		sort.Slice(C.Portals, func(i, j int) bool {
			return strings.ToLower(C.Portals[i].Title) < strings.ToLower(C.Portals[j].Title)
		})
	}

	if C.ContextPath != "" {
		for i := range C.Portals {
			if !C.Portals[i].IsExternal() {
				C.Portals[i].Link = C.ContextPath + C.Portals[i].Link
			}
		}

		for i := range C.Pages {
			C.Pages[i].Path = C.ContextPath + C.Pages[i].Path
		}
	}
}

func (c Config) GetLogLevel() (slog.Level, error) {
	var level slog.Level
	err := level.UnmarshalText([]byte(c.LogLevel))
	return level, err
}

func (c Config) GetLogHandler(w io.Writer) slog.Handler {
	logLevel, err := c.GetLogLevel()
	if err != nil {
		panic(fmt.Errorf("unable to unmarshal log level: %w", err))
	}
	logHandlerOptions := &slog.HandlerOptions{Level: logLevel}
	if c.LogJson {
		return slog.NewJSONHandler(w, logHandlerOptions)
	}
	return slog.NewTextHandler(w, logHandlerOptions)
}
