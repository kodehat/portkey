package types

type Portal struct {
	Link     string `json:"link"`
	Title    string `json:"title"`
	Emoji    string `json:"emoji"`
	External bool   `json:"external"`
}

type Page struct {
	Heading string `json:"heading"`
	Path    string `json:"path"`
	Content string `json:"content"`
}

type Config struct {
	Host               string
	Port               int
	Title              string
	FooterText         string
	SortAlphabetically bool
	Portals            []Portal
	Pages              []Page
}

type Flags struct {
	ConfigPath string
}
