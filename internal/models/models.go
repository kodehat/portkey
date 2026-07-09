package models

import (
	"fmt"
	"hash/fnv"
	"regexp"
	"strings"
)

// Portal struct containing information about a portal.
// This is used later as a link destination shown to the user.
type Portal struct {
	// Link destination link of a portal. Can be absolute or relative.
	Link string `json:"link"`

	// Title of a destination link.
	Title string `json:"title"`

	// Icon can be an emoji, a relative path (e.g., "/static/icon.svg"),
	// or an absolute URL to an icon image. If empty, the global favicon
	// is used for external links and a file icon for internal pages.
	Icon string `json:"icon"`

	// Keywords allows defining additional keywords used by the search.
	// This can make getting reasonable search results a lot easier.
	Keywords []string `json:"keywords"`

	// Group optionally assigns this portal to a named section on the home page.
	// Portals with the same Group value are rendered together under a shared heading.
	// Portals with an empty Group are shown without a heading.
	Group string `json:"group"`
}

// PortalGroup is a named collection of portals used for grouped rendering on the home page.
type PortalGroup struct {
	// Name of the group. Empty string means "ungrouped" (no heading rendered).
	Name    string
	Portals []Portal
}

// IsExternal decides if a destination link opens an external page or a custom page.
func (p Portal) IsExternal() bool {
	return strings.HasPrefix(p.Link, "http")
}

var /* const */ alphaNumDashOnlyRegex = regexp.MustCompile("[^a-zA-Z0-9-]")

// TitleForUrl returns the portal's title with alpha-numerical (and dash) characters only.
// If the cleaned result is empty (e.g., pure CJK characters), a deterministic
// hash-based fallback is returned to ensure a valid and unique URL segment.
func (portal Portal) TitleForUrl() string {
	cleaned := alphaNumDashOnlyRegex.ReplaceAllString(portal.Title, "")
	if cleaned != "" {
		return cleaned
	}
	// FNV-64a hash produces a deterministic short string for non-Latin titles.
	h := fnv.New64a()
	_, _ = h.Write([]byte(portal.Title))
	return fmt.Sprintf("p-%x", h.Sum64())
}

// Page struct defines a custom page that consists of a heading, content and a path,
// where the page will be available at.
type Page struct {
	// Heading of the custom page.
	Heading string `json:"heading"`

	// Subtitle of the custom page.
	Subtitle string `json:"subtitle"`

	// Content of the custom page interpreted as HTML.
	Content string `json:"content"`

	// Path of the custom page.
	Path string `json:"path"`
}
