package components

import (
	"context"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/kodehat/portkey/internal/models"
)

func TestHomePortal_External(t *testing.T) {
	portal := models.Portal{Title: "GitHub", Link: "https://github.com"}
	rec := httptest.NewRecorder()
	HomePortal(portal).Render(context.Background(), rec)

	body := rec.Body.String()
	if !strings.Contains(body, "GitHub") {
		t.Fatal("expected title in output")
	}
	if !strings.Contains(body, "target=\"_blank\"") {
		t.Fatal("expected target=_blank for external link")
	}
	if !strings.Contains(body, "nofollow") {
		t.Fatal("expected rel=nofollow for external link")
	}
	if !strings.Contains(body, "onerror=") {
		t.Fatal("expected favicon error handler to reveal the inline fallback")
	}
}

func TestHomePortal_Internal(t *testing.T) {
	portal := models.Portal{Title: "About", Link: "/about"}
	rec := httptest.NewRecorder()
	HomePortal(portal).Render(context.Background(), rec)

	body := rec.Body.String()
	if !strings.Contains(body, "About") {
		t.Fatal("expected title in output")
	}
	if strings.Contains(body, "target=\"_blank\"") {
		t.Fatal("expected no target=_blank for internal link")
	}
}

func TestHomePortal_WithIcon(t *testing.T) {
	portal := models.Portal{Title: "GitHub", Link: "https://github.com", Icon: "💻"}
	rec := httptest.NewRecorder()
	HomePortal(portal).Render(context.Background(), rec)

	body := rec.Body.String()
	if !strings.Contains(body, "GitHub") {
		t.Fatal("expected portal title in output")
	}
	if !strings.Contains(body, "github.com") {
		t.Fatal("expected domain in output")
	}
}

func TestHomePortalWithToolTip(t *testing.T) {
	portal := models.Portal{Title: "GitHub", Link: "https://github.com", Keywords: []string{"code", "git"}}
	rec := httptest.NewRecorder()
	HomePortalWithToolTip(portal).Render(context.Background(), rec)

	body := rec.Body.String()
	if !strings.Contains(body, "GitHub") {
		t.Fatal("expected title in output")
	}
	if !strings.Contains(body, "code") {
		t.Fatal("expected keyword 'code' in tooltip")
	}
	if !strings.Contains(body, "git") {
		t.Fatal("expected keyword 'git' in tooltip")
	}
}

func TestTooltip(t *testing.T) {
	rec := httptest.NewRecorder()
	tooltip("code").Render(context.Background(), rec)

	body := rec.Body.String()
	if !strings.Contains(body, "code") {
		t.Fatal("expected keyword in tooltip output")
	}
}

func TestFooterPortal(t *testing.T) {
	portal := models.Portal{Title: "GitHub", Link: "https://github.com"}
	rec := httptest.NewRecorder()
	FooterPortal(portal).Render(context.Background(), rec)

	body := rec.Body.String()
	if !strings.Contains(body, "GitHub") {
		t.Fatal("expected title in output")
	}
}

func TestFooterPortal_Internal(t *testing.T) {
	portal := models.Portal{Title: "About", Link: "/about"}
	rec := httptest.NewRecorder()
	FooterPortal(portal).Render(context.Background(), rec)

	body := rec.Body.String()
	if !strings.Contains(body, "About") {
		t.Fatal("expected title in output")
	}
	if strings.Contains(body, "target=\"_blank\"") {
		t.Fatal("expected no target=_blank for internal link")
	}
}

func TestHomePortalWithToolTip_NoKeywords(t *testing.T) {
	portal := models.Portal{Title: "GitHub", Link: "https://github.com"}
	rec := httptest.NewRecorder()
	HomePortalWithToolTip(portal).Render(context.Background(), rec)

	body := rec.Body.String()
	if !strings.Contains(body, "GitHub") {
		t.Fatal("expected title in output")
	}
}

func TestHomePortal_WithSpecialCharacters(t *testing.T) {
	portal := models.Portal{Title: "My <cool> Site", Link: "https://example.com"}
	rec := httptest.NewRecorder()
	HomePortal(portal).Render(context.Background(), rec)

	body := rec.Body.String()
	if !strings.Contains(body, "My") {
		t.Fatal("expected title in output")
	}
}

func TestHomePortal_WithImageIcon(t *testing.T) {
	portal := models.Portal{Title: "MyApp", Link: "https://example.com", Icon: "/static/icon.svg"}
	rec := httptest.NewRecorder()
	HomePortal(portal).Render(context.Background(), rec)

	body := rec.Body.String()
	if !strings.Contains(body, "<img") {
		t.Fatal("expected <img> tag for image icon")
	}
	if !strings.Contains(body, "/static/icon.svg") {
		t.Fatal("expected icon src in output")
	}
}

func TestFooterPortal_WithImageIcon(t *testing.T) {
	portal := models.Portal{Title: "MyApp", Link: "https://example.com", Icon: "/icon.png"}
	rec := httptest.NewRecorder()
	FooterPortal(portal).Render(context.Background(), rec)

	body := rec.Body.String()
	if !strings.Contains(body, "<img") {
		t.Fatal("expected <img> tag for image icon in footer")
	}
}

func TestFooterPortal_WithEmojiIcon(t *testing.T) {
	portal := models.Portal{Title: "MyApp", Link: "https://example.com", Icon: "🚀"}
	rec := httptest.NewRecorder()
	FooterPortal(portal).Render(context.Background(), rec)

	body := rec.Body.String()
	if !strings.Contains(body, "🚀") {
		t.Fatal("expected emoji icon in footer output")
	}
}

func TestFooterPortal_NoIcon(t *testing.T) {
	portal := models.Portal{Title: "MyApp", Link: "https://example.com"}
	rec := httptest.NewRecorder()
	FooterPortal(portal).Render(context.Background(), rec)

	body := rec.Body.String()
	if !strings.Contains(body, "MyApp") {
		t.Fatal("expected portal title in footer output")
	}
	if strings.Contains(body, "<img") {
		t.Fatal("did not expect img tag when no icon")
	}
}

func TestFooterPortal_NoIconInternal(t *testing.T) {
	portal := models.Portal{Title: "About", Link: "/about"}
	rec := httptest.NewRecorder()
	FooterPortal(portal).Render(context.Background(), rec)

	body := rec.Body.String()
	if !strings.Contains(body, "About") {
		t.Fatal("expected portal title in footer output")
	}
}

func TestHomePortal_DataURIIcon(t *testing.T) {
	portal := models.Portal{
		Title: "SVG",
		Link:  "https://example.com",
		Icon:  "data:image/svg+xml,%3Csvg xmlns='http://www.w3.org/2000/svg' width='16' height='16'%3E%3C/svg%3E",
	}
	rec := httptest.NewRecorder()
	HomePortal(portal).Render(context.Background(), rec)

	body := rec.Body.String()
	if !strings.Contains(body, "<img") {
		t.Fatal("expected img tag for data URI icon")
	}
	if !strings.Contains(body, "SVG") {
		t.Fatal("expected portal title in output")
	}
}

func TestHomePortal_AbsoluteURLIcon(t *testing.T) {
	portal := models.Portal{
		Title: "ExternalIcon",
		Link:  "https://example.com",
		Icon:  "https://cdn.example.com/icon.png",
	}
	rec := httptest.NewRecorder()
	HomePortal(portal).Render(context.Background(), rec)

	body := rec.Body.String()
	if !strings.Contains(body, "<img") {
		t.Fatal("expected img tag for absolute URL icon")
	}
	if !strings.Contains(body, "cdn.example.com") {
		t.Fatal("expected icon URL in output")
	}
}

func TestHomePortalWithToolTip_WithImageIcon(t *testing.T) {
	portal := models.Portal{
		Title:    "MyApp",
		Link:     "https://example.com",
		Icon:     "/icon.svg",
		Keywords: []string{"tool", "app"},
	}
	rec := httptest.NewRecorder()
	HomePortalWithToolTip(portal).Render(context.Background(), rec)

	body := rec.Body.String()
	if !strings.Contains(body, "<img") {
		t.Fatal("expected img tag for icon")
	}
	if !strings.Contains(body, "tool") {
		t.Fatal("expected keyword in tooltip")
	}
	if !strings.Contains(body, "app") {
		t.Fatal("expected keyword in tooltip")
	}
}

func TestIsIconImage(t *testing.T) {
	tests := []struct {
		icon     string
		expected bool
	}{
		{"https://example.com/icon.png", true},
		{"http://cdn.com/icon.svg", true},
		{"/static/icon.png", true},
		{"/_/icons/github.svg", true},
		{"data:image/svg+xml,...", true},
		{"data:image/png;base64,...", true},
		{"🚀", false},
		{"💻", false},
		{"🔗", false},
		{"", false},
	}
	for _, tc := range tests {
		if got := isIconImage(tc.icon); got != tc.expected {
			t.Errorf("isIconImage(%q) = %v, want %v", tc.icon, got, tc.expected)
		}
	}
}
