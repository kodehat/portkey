package models

import "regexp"

// Portal struct containing information about a portal.
// This is used later as a link destination shown to the user.
type Portal struct {

	// Link destination link of a portal.
	Link string `json:"link"`

	// Title of a destination link.
	Title string `json:"title"`

	// Emoji used as a prefix of the title.
	Emoji string `json:"emoji"`

	// External defines if a destination link opens an external page or a custom page.
	External bool `json:"external"`

	// Keywords allows defining additional keywords used by the search.
	// This can make getting reasonable search results a lot easier.
	Keywords []string `json:"keywords"`
}

var /* const */ alphaNumDashOnlyRegex = regexp.MustCompile("[^a-zA-Z0-9-]")

// TitleForUrl returns the portal's title with alpha-numerical (and dash) characters only.
func (portal Portal) TitleForUrl() string {
	return alphaNumDashOnlyRegex.ReplaceAllString(portal.Title, "")
}

// Page struct defines a custom page that consists of a heading, content and a path,
// where the page will be available at.
type Page struct {

	// Heading of the custom page.
	Heading string `json:"heading"`

	// Content of the custom page interpreted as HTML.
	Content string `json:"content"`

	// Path of the custom page.
	Path string `json:"path"`
}
