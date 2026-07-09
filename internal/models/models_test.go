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

func TestTitleForUrl_ChineseCharacters(t *testing.T) {
	p := Portal{Title: "中国国家地理"}
	got := p.TitleForUrl()
	if got == "" {
		t.Fatal("TitleForUrl() returned empty string for Chinese title")
	}
}

func TestTitleForUrl_ChineseWithColon(t *testing.T) {
	p := Portal{Title: "导航: 首页"}
	got := p.TitleForUrl()
	if got == "" {
		t.Fatal("TitleForUrl() returned empty string for Chinese title with colon")
	}
}

func TestTitleForUrl_EmojiOnly(t *testing.T) {
	p := Portal{Title: "🍳🍕"}
	got := p.TitleForUrl()
	if got == "" {
		t.Fatal("TitleForUrl() returned empty string for emoji-only title")
	}
}

func TestTitleForUrl_FallbackIsDeterministic(t *testing.T) {
	p1 := Portal{Title: "中国国家地理"}
	p2 := Portal{Title: "中国国家地理"}
	if p1.TitleForUrl() != p2.TitleForUrl() {
		t.Fatal("TitleForUrl() should be deterministic for the same input")
	}
}

func TestTitleForUrl_HashDiffersForDifferentInputs(t *testing.T) {
	p1 := Portal{Title: "中国国家地理"}
	p2 := Portal{Title: "故宫博物院"}
	if p1.TitleForUrl() == p2.TitleForUrl() {
		t.Fatal("TitleForUrl() should produce different output for different Chinese titles")
	}
}
