package components

import (
	"context"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/kodehat/portkey/internal/config"
	"github.com/kodehat/portkey/internal/models"
)

func TestPortalPartial_Empty(t *testing.T) {
	config.C = config.Config{}
	rec := httptest.NewRecorder()
	PortalPartial([]models.Portal{}).Render(context.Background(), rec)

	body := rec.Body.String()
	if !strings.Contains(body, "no results") {
		t.Fatal("expected 'no results' in output")
	}
}

func TestPortalPartial_Populated(t *testing.T) {
	config.C = config.Config{}
	portals := []models.Portal{
		{Title: "GitHub", Link: "https://github.com"},
		{Title: "Google", Link: "https://google.com"},
	}
	rec := httptest.NewRecorder()
	PortalPartial(portals).Render(context.Background(), rec)

	body := rec.Body.String()
	if !strings.Contains(body, "GitHub") {
		t.Fatal("expected 'GitHub' in output")
	}
	if !strings.Contains(body, "Google") {
		t.Fatal("expected 'Google' in output")
	}
}

func TestGroupedPortalPartial_Empty(t *testing.T) {
	config.C = config.Config{}
	rec := httptest.NewRecorder()
	GroupedPortalPartial([]models.PortalGroup{}).Render(context.Background(), rec)

	body := rec.Body.String()
	if !strings.Contains(body, "no results") {
		t.Fatal("expected 'no results' in output")
	}
}

func TestGroupedPortalPartial_Grouped(t *testing.T) {
	config.C = config.Config{}
	groups := []models.PortalGroup{
		{
			Name: "Dev",
			Portals: []models.Portal{
				{Title: "GitHub", Link: "https://github.com"},
			},
		},
		{
			Name: "",
			Portals: []models.Portal{
				{Title: "Blog", Link: "/blog"},
			},
		},
	}
	rec := httptest.NewRecorder()
	GroupedPortalPartial(groups).Render(context.Background(), rec)

	body := rec.Body.String()
	if !strings.Contains(body, "Dev") {
		t.Fatal("expected group name 'Dev' in output")
	}
	if !strings.Contains(body, "GitHub") {
		t.Fatal("expected portal title 'GitHub' in output")
	}
	if !strings.Contains(body, "Blog") {
		t.Fatal("expected portal title 'Blog' in output")
	}
}

func TestPortalPartial_WithKeywordsTooltips(t *testing.T) {
	config.C = config.Config{ShowKeywordsAsTooltips: true}
	portals := []models.Portal{
		{Title: "GitHub", Link: "https://github.com", Keywords: []string{"code", "git"}},
	}
	rec := httptest.NewRecorder()
	PortalPartial(portals).Render(context.Background(), rec)

	body := rec.Body.String()
	if !strings.Contains(body, "code") {
		t.Fatal("expected keyword 'code' in tooltip output")
	}
}

func TestGroupedPortalPartial_SingleGroup(t *testing.T) {
	config.C = config.Config{}
	groups := []models.PortalGroup{
		{
			Name:    "Tools",
			Portals: []models.Portal{{Title: "GitHub", Link: "https://github.com"}},
		},
	}
	rec := httptest.NewRecorder()
	GroupedPortalPartial(groups).Render(context.Background(), rec)

	body := rec.Body.String()
	if !strings.Contains(body, "Tools") {
		t.Fatal("expected group name 'Tools' in output")
	}
	if !strings.Contains(body, "GitHub") {
		t.Fatal("expected portal title 'GitHub' in output")
	}
}

func TestPortalPartial_MultiplePortals(t *testing.T) {
	config.C = config.Config{}
	portals := []models.Portal{
		{Title: "A", Link: "/a"},
		{Title: "B", Link: "/b"},
		{Title: "C", Link: "/c"},
	}
	rec := httptest.NewRecorder()
	PortalPartial(portals).Render(context.Background(), rec)

	body := rec.Body.String()
	if !strings.Contains(body, "A") || !strings.Contains(body, "B") || !strings.Contains(body, "C") {
		t.Fatal("expected all portal titles in output")
	}
}
