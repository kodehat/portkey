package server

import (
	"encoding/json"
	"net/http"

	"github.com/kodehat/portkey/internal/config"
)

func portalsRestHandler() http.HandlerFunc {
	encoded, _ := json.Marshal(config.C.Portals)
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.Write(encoded)
	}
}

func pagesRestHandler() http.HandlerFunc {
	encoded, _ := json.Marshal(config.C.Pages)
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.Write(encoded)
	}
}
