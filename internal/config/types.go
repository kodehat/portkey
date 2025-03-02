package config

import "github.com/kodehat/portkey/internal/models"

type EnvLevel int

const (
	Dev EnvLevel = iota
	Prod
)

type LogConfig struct {
	Level string
	Json  bool
}

type Config struct {
	Env                        EnvLevel
	Log                        LogConfig
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
}

type Flags struct {
	ConfigPath string
}
