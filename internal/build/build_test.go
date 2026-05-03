package build

import "testing"

func TestLoadBuildDetails(t *testing.T) {
	LoadBuildDetails("testhash")

	if B.CssHash != "testhash" {
		t.Fatalf("expected CssHash 'testhash', got %q", B.CssHash)
	}
	if B.Version != "dev" {
		t.Fatalf("expected Version 'dev', got %q", B.Version)
	}
	if B.BuildTime != "unknown" {
		t.Fatalf("expected BuildTime 'unknown', got %q", B.BuildTime)
	}
	if B.GoVersion != "unknown" {
		t.Fatalf("expected GoVersion 'unknown', got %q", B.GoVersion)
	}
}

func TestBuildDefaults(t *testing.T) {
	if Version == "" {
		t.Fatal("expected Version to have a default value")
	}
	if BuildTime == "" {
		t.Fatal("expected BuildTime to have a default value")
	}
	if GoVersion == "" {
		t.Fatal("expected GoVersion to have a default value")
	}
}
