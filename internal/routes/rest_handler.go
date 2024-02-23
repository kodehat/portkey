package routes

import (
	"encoding/json"
	"net/http"

	"github.com/kodehat/portkey/internal/config"
)

func portalsRestHandler() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(config.C.Portals)
	}
}

func pagesRestHandler() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(config.C.Pages)
	}
}
