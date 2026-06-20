package server

import (
	"embed"
	"log/slog"
	"net/http"
	"strings"
	"time"

	"github.com/kodehat/portkey/internal/config"
	"github.com/kodehat/portkey/internal/metrics"
)

func NewServer(
	logger *slog.Logger,
	static embed.FS,
) http.Handler {
	mux := http.NewServeMux()
	addRoutes(mux, logger, static)
	return durationMiddleware(mux)
}

// durationMiddleware records HTTP request duration as a Prometheus histogram.
func durationMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()
		next.ServeHTTP(w, r)
		// Exclude static files and dev reload to reduce cardinality.
		cp := config.C.ContextPath
		if !strings.HasPrefix(r.URL.Path, cp+"/static/") && !strings.HasPrefix(r.URL.Path, cp+"/reload") {
			metrics.M.HTTPRequestDuration.WithLabelValues(r.URL.Path).Observe(time.Since(start).Seconds())
		}
	})
}

func NewMetricsServer(
	logger *slog.Logger,
) http.Handler {
	mux := http.NewServeMux()
	addMetricRoutes(mux)
	return mux
}
