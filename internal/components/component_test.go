package components

import (
	"testing"

	"github.com/kodehat/portkey/internal/config"
)

func TestDomainFromURL(t *testing.T) {
	tests := []struct {
		url      string
		expected string
	}{
		{"https://github.com", "github.com"},
		{"https://www.github.com", "www.github.com"},
		{"http://example.com/path", "example.com"},
		{"https://sub.domain.com:8080/page", "sub.domain.com"},
		{"/about", ""},
		{"relative/path", ""},
		{"", ""},
	}

	for _, tt := range tests {
		got := domainFromURL(tt.url)
		if got != tt.expected {
			t.Errorf("domainFromURL(%q) = %q, want %q", tt.url, got, tt.expected)
		}
	}
}

func TestDomainFromURL_InvalidURL(t *testing.T) {
	got := domainFromURL("not a valid url://")
	if got != "" {
		t.Errorf("expected empty for invalid URL, got %q", got)
	}
}

func TestFaviconURL(t *testing.T) {
	// Set up a known context path.
	config.C = config.Config{ContextPath: ""}

	tests := []struct {
		url      string
		expected string
	}{
		{"https://github.com", "/_/favicon?domain=github.com"},
		{"https://www.example.com/page", "/_/favicon?domain=www.example.com"},
	}

	for _, tt := range tests {
		got := faviconURL(tt.url)
		if got != tt.expected {
			t.Errorf("faviconURL(%q) with empty context = %q, want %q", tt.url, got, tt.expected)
		}
	}
}

func TestFaviconURL_RelativeURL(t *testing.T) {
	config.C = config.Config{ContextPath: ""}
	got := faviconURL("/about")
	if got != "" {
		t.Errorf("expected empty for relative URL, got %q", got)
	}
}

func TestFaviconURL_WithContextPath(t *testing.T) {
	config.C = config.Config{ContextPath: "/portkey"}
	got := faviconURL("https://github.com")
	expected := "/portkey/_/favicon?domain=github.com"
	if got != expected {
		t.Errorf("faviconURL with context path = %q, want %q", got, expected)
	}
}

func TestGridClass_Valid(t *testing.T) {
	for n := 1; n <= 12; n++ {
		cls := GridClass(n)
		if cls == "" {
			t.Errorf("GridClass(%d) returned empty", n)
		}
		if len(cls) < 10 {
			t.Errorf("GridClass(%d) = %q, too short", n, cls)
		}
	}
}

func TestGridClass_Invalid(t *testing.T) {
	if cls := GridClass(0); cls != "" {
		t.Errorf("GridClass(0) = %q, want empty", cls)
	}
	if cls := GridClass(13); cls != "" {
		t.Errorf("GridClass(13) = %q, want empty", cls)
	}
	if cls := GridClass(-1); cls != "" {
		t.Errorf("GridClass(-1) = %q, want empty", cls)
	}
}

func TestGridClass_ContainsColumnCount(t *testing.T) {
	for n := 1; n <= 12; n++ {
		cls := GridClass(n)
		expectedCol := "grid-cols-"
		if !contains(cls, expectedCol) {
			t.Errorf("GridClass(%d) = %q, missing %q", n, cls, expectedCol)
		}
	}
}

func TestGridClass_MobileFallback(t *testing.T) {
	for n := 1; n <= 4; n++ {
		cls := GridClass(n)
		if !contains(cls, "max-md:flex") {
			t.Errorf("GridClass(%d) missing mobile fallback: %q", n, cls)
		}
	}
}

// contains is a small helper since strings.Contains isn't imported.
func contains(s, substr string) bool {
	return len(s) >= len(substr) && containsSub(s, substr)
}

func containsSub(s, sub string) bool {
	for i := 0; i <= len(s)-len(sub); i++ {
		if s[i:i+len(sub)] == sub {
			return true
		}
	}
	return false
}
