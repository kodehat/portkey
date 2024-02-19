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
	"github.com/kodehat/thisismy.cloud/components"
	"github.com/spf13/viper"
)

//go:embed static
var static embed.FS

// build flags
var (
	BuildTime  string = "N/A"
	CommitHash string
	GoVersion  string = "N/A"
)

var C config
var F flags

func main() {
	loadFlags()
	configPath, err := filepath.Abs(F.ConfigPath)
	if err != nil {
		panic(fmt.Errorf("fatal error config file: %w", err))
	}
	log.Printf("Looking for config.y[a]ml in: %s\n", configPath)
	loadConfig(F.ConfigPath)

	if C.SortAlphabetically {
		sort.Slice(C.Portals, func(i, j int) bool {
			return strings.ToLower(C.Portals[i].Title) < strings.ToLower(C.Portals[j].Title)
		})
	}
	var allHomePortals = make([]templ.Component, len(C.Portals))
	var allFooterPortals = make([]templ.Component, len(C.Portals))
	for i, configPortal := range C.Portals {
		allHomePortals[i] = components.HomePortal(configPortal.Link, configPortal.Emoji, configPortal.Title, configPortal.External)
		allFooterPortals[i] = components.FooterPortal(configPortal.Link, configPortal.Emoji, configPortal.Title, configPortal.External)
	}
	home := components.HomePage(allHomePortals)

	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/" {
			w.WriteHeader(http.StatusNotFound)
			templ.Handler(components.ContentLayout(fmt.Sprintf("%s - %s", "404 Not Found", C.Title), "404 Not Found", components.NotFound(), allFooterPortals, C.FooterText)).ServeHTTP(w, r)
			return
		}
		templ.Handler(components.HomeLayout(C.Title, home)).ServeHTTP(w, r)
	})

	for _, page := range C.Pages {
		http.Handle(page.Path, templ.Handler(components.ContentLayout(fmt.Sprintf("%s - %s", page.Heading, C.Title), page.Heading, components.ContentPage(page.Content), allFooterPortals, C.FooterText)))
	}
	http.Handle("/version", templ.Handler(components.ContentLayout(fmt.Sprintf("%s - %s", "Version", C.Title), "Version", components.Version(BuildTime, CommitHash, GoVersion), allFooterPortals, C.FooterText)))
	http.Handle("/static/", staticHandler(http.FileServer(http.FS(static))))

	http.Handle("/api/v1/portals", http.HandlerFunc(returnPortalsAsJson))
	http.Handle("/api/v1/pages", http.HandlerFunc(returnPagessAsJson))

	log.Printf("Listening on %s:%d\n", C.Host, C.Port)
	http.ListenAndServe(fmt.Sprintf("%s:%d", C.Host, C.Port), nil)
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

type portal struct {
	Link     string `json:"link"`
	Title    string `json:"title"`
	Emoji    string `json:"emoji"`
	External bool   `json:"external"`
}

type page struct {
	Heading string `json:"heading"`
	Path    string `json:"path"`
	Content string `json:"content"`
}

type config struct {
	Host               string
	Port               int
	Title              string
	FooterText         string
	SortAlphabetically bool
	Portals            []portal
	Pages              []page
}

func loadConfig(configPath string) {
	viper.SetConfigName("config")
	viper.SetConfigType("yaml")
	viper.AddConfigPath(configPath)
	viper.SetDefault("host", "localhost")
	viper.SetDefault("port", "1414")
	viper.SetDefault("title", "Your Portal")
	viper.SetDefault("footerText", "Works like a portal.")
	err := viper.ReadInConfig()
	if err != nil {
		panic(fmt.Errorf("fatal error config file: %w", err))
	}
	viper.Unmarshal(&C)
}

type flags struct {
	ConfigPath string
}

func loadFlags() {
	var configPath string
	flag.StringVar(&configPath, "config-path", ".", "path where config.yml can be found")
	flag.Parse()
	F = flags{
		ConfigPath: configPath,
	}
}
