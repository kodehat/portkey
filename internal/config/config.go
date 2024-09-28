package config

import (
	"flag"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"github.com/kodehat/portkey/internal/models"
	"github.com/spf13/viper"
)

type Config struct {
	Host                       string
	Port                       string
	ContextPath                string
	Title                      string
	HideTitle                  bool
	Footer                     string
	ShowTopIcon                bool
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

var C Config
var F Flags

func Load() {
	LoadFlags()
	configPath, err := filepath.Abs(F.ConfigPath)
	if err != nil {
		panic(fmt.Errorf("fatal error config file: %w", err))
	}
	log.Printf("Looking for config.y[a]ml in: %s\n", configPath)
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
	viper.SetDefault("host", "localhost")
	viper.SetDefault("port", "1414")
	viper.SetDefault("contextPath", "")
	viper.SetDefault("title", "Your Portal")
	viper.SetDefault("footerText", "Works like a portal.")
	viper.SetDefault("minimumStringSimilarity", 0.75)
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
			if !C.Portals[i].External {
				C.Portals[i].Link = C.ContextPath + C.Portals[i].Link
			}
		}

		for i := range C.Pages {
			C.Pages[i].Path = C.ContextPath + C.Pages[i].Path
		}
	}
}
