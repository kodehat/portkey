package main

import (
	"net/http"
)

type NotFoundRespWr struct {
	http.ResponseWriter
	status int
}

func (w *NotFoundRespWr) WriteHeader(status int) {
	w.status = status
	if status != http.StatusNotFound {
		w.ResponseWriter.WriteHeader(status)
	}
}

func (w *NotFoundRespWr) Write(p []byte) (int, error) {
	if w.status != http.StatusNotFound {
		return w.ResponseWriter.Write(p)
	}
	return len(p), nil // Lie that we successfully written it
}

// staticHandler Redirect to 404 page if static file not found https://stackoverflow.com/a/47286697
func staticHandler(h http.Handler) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		nfrw := &NotFoundRespWr{ResponseWriter: w}
		h.ServeHTTP(nfrw, r)
		if nfrw.status == http.StatusNotFound {
			http.Redirect(w, r, "/404", http.StatusFound)
		}
	}
}
