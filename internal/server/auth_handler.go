package server

import (
	"encoding/json"
	"net/http"

	"github.com/a-h/templ"
	"github.com/kodehat/portkey/internal/auth"
	"github.com/kodehat/portkey/internal/build"
	"github.com/kodehat/portkey/internal/components"
	"github.com/kodehat/portkey/internal/config"
)

func authPageHandler() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if c, err := r.Cookie("pk_sess"); err == nil && auth.ValidSession(c.Value) {
			http.Redirect(w, r, config.C.Server.ContextPath+"/", http.StatusFound)
			return
		}
		templ.Handler(components.LoginPage(config.C, build.B, !auth.HasCredentials())).ServeHTTP(w, r)
	}
}

func authBeginHandler() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
			return
		}
		options, err := auth.BeginLogin()
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(options)
	}
}

func authFinishHandler() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
			return
		}
		if err := auth.FinishLogin(r); err != nil {
			http.Error(w, "authentication failed", http.StatusUnauthorized)
			return
		}
		auth.SetSessionCookie(w, auth.NewSessionToken())
		w.Header().Set("Content-Type", "application/json")
		w.Write([]byte(`{"ok":true}`))
	}
}

func authRegisterPageHandler() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if auth.HasCredentials() {
			http.Redirect(w, r, config.C.Server.ContextPath+"/auth", http.StatusFound)
			return
		}
		templ.Handler(components.RegisterPage(config.C, build.B)).ServeHTTP(w, r)
	}
}

func authRegisterBeginHandler() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
			return
		}
		options, err := auth.BeginRegistration()
		if err != nil {
			http.Error(w, err.Error(), http.StatusForbidden)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(options)
	}
}

func authRegisterFinishHandler() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
			return
		}
		if err := auth.FinishRegistration(r); err != nil {
			http.Error(w, "registration failed: "+err.Error(), http.StatusBadRequest)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		w.Write([]byte(`{"ok":true}`))
	}
}

func authLogoutHandler() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		auth.ClearSessionCookie(w)
		http.Redirect(w, r, config.C.Server.ContextPath+"/auth", http.StatusFound)
	}
}
