package models

import "testing"

func TestIsExternal_HTTPS(t *testing.T) {
	p := Portal{Link: "https://github.com"}
	if !p.IsExternal() {
		t.Fatal("expected IsExternal() == true for https:// link")
	}
}

func TestIsExternal_HTTP(t *testing.T) {
	p := Portal{Link: "http://example.com"}
	if !p.IsExternal() {
		t.Fatal("expected IsExternal() == true for http:// link")
	}
}

func TestIsInternal_Path(t *testing.T) {
	p := Portal{Link: "/about"}
	if p.IsExternal() {
		t.Fatal("expected IsExternal() == false for relative path")
	}
}

func TestIsInternal_Empty(t *testing.T) {
	p := Portal{Link: ""}
	if p.IsExternal() {
		t.Fatal("expected IsExternal() == false for empty link")
	}
}

func TestTitleForUrl_Passthrough(t *testing.T) {
	p := Portal{Title: "github"}
	got := p.TitleForUrl()
	if got != "github" {
		t.Fatalf("TitleForUrl() == %q, want %q", got, "github")
	}
}

func TestTitleForUrl_SpacesRemoved(t *testing.T) {
	p := Portal{Title: "My Site"}
	got := p.TitleForUrl()
	if got != "MySite" {
		t.Fatalf("TitleForUrl() == %q, want %q", got, "MySite")
	}
}

func TestTitleForUrl_SpecialCharsRemoved(t *testing.T) {
	p := Portal{Title: "GitHub!@#$"}
	got := p.TitleForUrl()
	if got != "GitHub" {
		t.Fatalf("TitleForUrl() == %q, want %q", got, "GitHub")
	}
}

func TestTitleForUrl_PreservesDashes(t *testing.T) {
	p := Portal{Title: "my-cool-site"}
	got := p.TitleForUrl()
	if got != "my-cool-site" {
		t.Fatalf("TitleForUrl() == %q, want %q", got, "my-cool-site")
	}
}

func TestTitleForUrl_CJKPure(t *testing.T) {
	p := Portal{Title: "中国国家地理"}
	got := p.TitleForUrl()
	if got != "中国国家地理" {
		t.Fatalf("TitleForUrl() == %q, want %q", got, "中国国家地理")
	}
}

func TestTitleForUrl_CJKMixed(t *testing.T) {
	p := Portal{Title: "東京 Tokyo!"}
	got := p.TitleForUrl()
	if got != "東京Tokyo" {
		t.Fatalf("TitleForUrl() == %q, want %q", got, "東京Tokyo")
	}
}

func TestTitleForUrl_OnlySymbols(t *testing.T) {
	p := Portal{Title: "★★★"}
	got := p.TitleForUrl()
	// All-symbol titles fall back to percent-encoding.
	want := "%E2%98%85%E2%98%85%E2%98%85"
	if got != want {
		t.Fatalf("TitleForUrl() == %q, want %q", got, want)
	}
}
