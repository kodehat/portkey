package main

import (
	"embed"
	"fmt"
	"net/http"
	"sort"
	"strings"

	"github.com/a-h/templ"
	"github.com/kodehat/thisismy.cloud/components"
	"github.com/spf13/viper"
)

//go:embed static
var static embed.FS

var C config

func main() {
	loadConfig()
	fmt.Printf("Config is: %+v\n", C)

	if C.SortAlphabetically {
		sort.Slice(C.Portals, func(i, j int) bool {
			return strings.ToLower(C.Portals[i].Title) < strings.ToLower(C.Portals[j].Title)
		})
	}
	var allHomePortals = make([]templ.Component, len(C.Portals))
	var allFooterPortals = make([]templ.Component, len(C.Portals))
	for i, configPortal := range C.Portals {
		allHomePortals[i] = components.HomePortal(configPortal.Link, configPortal.Title, configPortal.External)
		allFooterPortals[i] = components.FooterPortal(configPortal.Link, configPortal.Emoji, configPortal.Title, configPortal.External)
	}
	home := components.HomePage(allHomePortals)
	http.Handle("/", templ.Handler(components.HomeLayout(C.Title, home)))
	http.Handle("/about", templ.Handler(components.ContentLayout("About", components.ContentPage("This is a portal"), allFooterPortals)))
	http.Handle("/static/", http.FileServer(http.FS(static)))
	http.ListenAndServe(fmt.Sprintf("%s:%d", C.Host, C.Port), nil)
}

type portal struct {
	Link     string
	Title    string
	Emoji    string
	External bool
}

type config struct {
	Host               string
	Port               int
	Title              string
	SortAlphabetically bool
	Portals            []portal
}

func loadConfig() {
	viper.SetConfigName("config")
	viper.SetConfigType("yaml")
	viper.AddConfigPath(".")
	viper.SetDefault("host", "localhost")
	viper.SetDefault("port", "1414")
	viper.SetDefault("title", "Your Portal")
	err := viper.ReadInConfig()
	if err != nil {
		panic(fmt.Errorf("fatal error config file: %w", err))
	}
	viper.Unmarshal(&C)
}
