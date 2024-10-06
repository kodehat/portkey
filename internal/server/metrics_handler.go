package server

import (
	"net/http"

	"github.com/prometheus/client_golang/prometheus/promhttp"
)

func metricsHandler() http.HandlerFunc {
	return promhttp.Handler().ServeHTTP
}
