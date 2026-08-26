package auth

import (
	"net/http"
	"strings"

	"github.com/kodehat/portkey/internal/config"
)

const cookieName = "pk_sess"

// Require enforces authentication on all routes except /auth/* and static assets.
func Require(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		cp := config.C.Server.ContextPath
		p := r.URL.Path
		if strings.HasPrefix(p, cp+"/auth") ||
			strings.HasPrefix(p, cp+"/static/") ||
			p == cp+"/healthz" ||
			p == cp+"/version" {
			next.ServeHTTP(w, r)
			return
		}
		c, err := r.Cookie(cookieName)
		if err != nil || !ValidSession(c.Value) {
			http.Redirect(w, r, cp+"/auth", http.StatusFound)
			return
		}
		next.ServeHTTP(w, r)
	})
}

// SetSessionCookie writes a 30-day signed session cookie.
func SetSessionCookie(w http.ResponseWriter, token string) {
	http.SetCookie(w, &http.Cookie{
		Name:     cookieName,
		Value:    token,
		Path:     "/",
		MaxAge:   30 * 24 * 60 * 60,
		HttpOnly: true,
		Secure:   true,
		SameSite: http.SameSiteStrictMode,
	})
}

// ClearSessionCookie expires the session cookie.
func ClearSessionCookie(w http.ResponseWriter) {
	http.SetCookie(w, &http.Cookie{
		Name:   cookieName,
		Value:  "",
		Path:   "/",
		MaxAge: -1,
	})
}
