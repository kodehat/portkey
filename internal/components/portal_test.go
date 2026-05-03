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

func TestHomePortal_WithEmoji(t *testing.T) {
	portal := models.Portal{Title: "GitHub", Link: "https://github.com", Emoji: "💻"}
	rec := httptest.NewRecorder()
	HomePortal(portal).Render(context.Background(), rec)

	body := rec.Body.String()
	if !strings.Contains(body, "💻") {
		t.Fatal("expected emoji in output")
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
