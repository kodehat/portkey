package utils

import (
	"testing"

	"github.com/kodehat/portkey/internal/models"
)

func TestGroupPortalsEmpty(t *testing.T) {
	got := GroupPortals([]models.Portal{})
	if len(got) != 0 {
		t.Fatalf("expected 0 groups, got %d", len(got))
	}
}

func TestGroupPortalsNoGroups(t *testing.T) {
	portals := []models.Portal{
		{Title: "A", Link: "/a"},
		{Title: "B", Link: "/b"},
	}
	got := GroupPortals(portals)
	if len(got) != 1 {
		t.Fatalf("expected 1 group, got %d", len(got))
	}
	if got[0].Name != "" {
		t.Fatalf("expected unnamed group, got %q", got[0].Name)
	}
	if len(got[0].Portals) != 2 {
		t.Fatalf("expected 2 portals in group, got %d", len(got[0].Portals))
	}
}

func TestGroupPortalsWithGroups(t *testing.T) {
	portals := []models.Portal{
		{Title: "A", Link: "/a", Group: "dev"},
		{Title: "B", Link: "/b", Group: "social"},
		{Title: "C", Link: "/c"},
	}
	got := GroupPortals(portals)
	if len(got) != 3 {
		t.Fatalf("expected 3 groups, got %d", len(got))
	}
	if got[0].Name != "dev" {
		t.Fatalf("first group expected 'dev', got %q", got[0].Name)
	}
	if got[1].Name != "social" {
		t.Fatalf("second group expected 'social', got %q", got[1].Name)
	}
	if got[2].Name != "" {
		t.Fatalf("last group expected unnamed, got %q", got[2].Name)
	}
}

func TestGroupPortalsOrderPreserved(t *testing.T) {
	portals := []models.Portal{
		{Title: "A", Link: "/a", Group: "C"},
		{Title: "B", Link: "/b", Group: "A"},
		{Title: "D", Link: "/d"},
		{Title: "C", Link: "/c", Group: "B"},
	}
	got := GroupPortals(portals)
	if len(got) != 4 {
		t.Fatalf("expected 4 groups, got %d", len(got))
	}
	expectedOrder := []string{"C", "A", "B", ""}
	for i, name := range expectedOrder {
		if got[i].Name != name {
			t.Fatalf("group[%d] expected %q, got %q", i, name, got[i].Name)
		}
	}
}

func TestGroupPortalsMultiplePerGroup(t *testing.T) {
	portals := []models.Portal{
		{Title: "A1", Link: "/a1", Group: "dev"},
		{Title: "A2", Link: "/a2", Group: "dev"},
		{Title: "B1", Link: "/b1", Group: "social"},
		{Title: "B2", Link: "/b2", Group: "social"},
	}
	got := GroupPortals(portals)
	if len(got) != 2 {
		t.Fatalf("expected 2 groups, got %d", len(got))
	}
	if len(got[0].Portals) != 2 {
		t.Fatalf("expected 2 portals in dev group, got %d", len(got[0].Portals))
	}
	if len(got[1].Portals) != 2 {
		t.Fatalf("expected 2 portals in social group, got %d", len(got[1].Portals))
	}
}
