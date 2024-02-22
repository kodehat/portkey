package main

import (
	"embed"
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"net/http"
	"path/filepath"
	"sort"
	"strings"

	"github.com/a-h/templ"
	"github.com/kodehat/portkey/internal/components"
	"github.com/kodehat/portkey/internal/types"
	"github.com/kodehat/portkey/internal/utils"
	"github.com/spf13/viper"
)

//go:embed static
var static embed.FS

// Injected during build.
var (
	BuildTime  string = "N/A"
	CommitHash string
	GoVersion  string = "N/A"
)

var C types.Config
var F types.Flags
var B types.BuildDetails

func main() {
	loadBuildDetails()
	loadFlags()
	configPath, err := filepath.Abs(F.ConfigPath)
	if err != nil {
		panic(fmt.Errorf("fatal error config file: %w", err))
	}
	log.Printf("Looking for config.y[a]ml in: %s\n", configPath)
	loadConfig(configPath)

	if C.SortAlphabetically {
		sort.Slice(C.Portals, func(i, j int) bool {
			return strings.ToLower(C.Portals[i].Title) < strings.ToLower(C.Portals[j].Title)
		})
	}
	var allFooterPortals = make([]templ.Component, len(C.Portals))
	for i, configPortal := range C.Portals {
		allFooterPortals[i] = components.FooterPortal(configPortal.Link, configPortal.Emoji, configPortal.Title, configPortal.External)
	}
	home := components.HomePage()

	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/" {
			w.WriteHeader(http.StatusNotFound)
			templ.Handler(components.ContentLayout(utils.PageTitle("404 Not Found", C.Title), "404 Not Found", components.NotFound(), allFooterPortals, C.FooterText)).ServeHTTP(w, r)
			return
		}
		templ.Handler(components.HomeLayout(C.Title, home)).ServeHTTP(w, r)
	})

	for _, page := range C.Pages {
		http.Handle(page.Path, templ.Handler(components.ContentLayout(utils.PageTitle(page.Heading, C.Title), page.Heading, components.ContentPage(page.Content), allFooterPortals, C.FooterText)))
	}
	http.Handle("/version", templ.Handler(components.ContentLayout(utils.PageTitle("Version", C.Title), "Version", components.Version(B), allFooterPortals, C.FooterText)))
	http.Handle("/static/", staticHandler(http.FileServer(http.FS(static))))

	http.Handle("/_/portals", http.HandlerFunc(returnSearchedPortals))

	http.Handle("/api/v1/portals", http.HandlerFunc(returnPortalsAsJson))
	http.Handle("/api/v1/pages", http.HandlerFunc(returnPagessAsJson))

	log.Printf("Listening on %s:%d\n", C.Host, C.Port)
	http.ListenAndServe(fmt.Sprintf("%s:%d", C.Host, C.Port), nil)
}

func returnSearchedPortals(w http.ResponseWriter, r *http.Request) {
	query := r.URL.Query().Get("search")
	var allHomePortals = make([]templ.Component, 0)
	for _, configPortal := range C.Portals {
		portal := components.HomePortal(configPortal.Link, configPortal.Emoji, configPortal.Title, configPortal.External)
		if query != "" {
			if strings.Contains(configPortal.Title, query) || utils.ArrSubStr(configPortal.Keywords, query) {
				allHomePortals = append(allHomePortals, portal)
			}
		} else {
			allHomePortals = append(allHomePortals, portal)
		}
	}
	components.PortalPartial(allHomePortals).Render(r.Context(), w)
}

func returnPortalsAsJson(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(C.Portals)
}

func returnPagessAsJson(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(C.Pages)
}

func loadBuildDetails() {
	B = types.BuildDetails{
		BuildTime:  BuildTime,
		CommitHash: CommitHash,
		GoVersion:  GoVersion,
	}
}

func loadConfig(configPath string) {
	viper.SetConfigName("config")
	viper.SetConfigType("yml")
	viper.AddConfigPath(configPath)
	viper.SetDefault("host", "localhost")
	viper.SetDefault("port", "1414")
	viper.SetDefault("title", "Your Portal")
	viper.SetDefault("footerText", "Works like a portal.")
	viper.SafeWriteConfig()
	err := viper.ReadInConfig()
	if err != nil {
		panic(fmt.Errorf("fatal error config file: %w", err))
	}
	viper.Unmarshal(&C)
}

func loadFlags() {
	var configPath string
	flag.StringVar(&configPath, "config-path", ".", "path where config.yml can be found")
	flag.Parse()
	F = types.Flags{
		ConfigPath: configPath,
	}
}
