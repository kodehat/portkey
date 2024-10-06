package metrics

import (
	"github.com/kodehat/portkey/internal/build"
	"github.com/prometheus/client_golang/prometheus"
)

const NAMESPACE = "portkey"

type Metrics struct {
	PortalHitCounter         *prometheus.CounterVec
	PageHitCounter           *prometheus.CounterVec
	SearchWithResultsCounter prometheus.Counter
	SearchNoResultsCounter   prometheus.Counter
	buildInfo                prometheus.Gauge
}

var M Metrics

func Load() {
	M = Metrics{
		PortalHitCounter: prometheus.NewCounterVec(prometheus.CounterOpts{
			Namespace: NAMESPACE,
			Name:      "portal_handler_requests_total",
			Help:      "Total number of HTTP requests by portal.",
		}, []string{"portal"}),
		PageHitCounter: prometheus.NewCounterVec(prometheus.CounterOpts{
			Namespace: NAMESPACE,
			Name:      "page_handler_requests_total",
			Help:      "Total number of HTTP requests by page.",
		}, []string{"path"}),
		SearchWithResultsCounter: prometheus.NewCounter(prometheus.CounterOpts{
			Namespace: NAMESPACE,
			Name:      "search_requests_with_results_total",
			Help:      "Total number of HTTP requests for search with at least one result.",
		}),
		SearchNoResultsCounter: prometheus.NewCounter(prometheus.CounterOpts{
			Namespace: NAMESPACE,
			Name:      "search_requests_no_results_total",
			Help:      "Total number of HTTP requests for search with no results.",
		}),
		buildInfo: prometheus.NewGauge(prometheus.GaugeOpts{
			Namespace: NAMESPACE,
			Name:      "version_info",
			Help:      "Version information about portkey.",
			ConstLabels: prometheus.Labels{
				"buildTime":  build.B.BuildTime,
				"commitHash": build.B.CommitHash,
				"version":    build.B.Version,
				"goVersion":  build.B.GoVersion,
			},
		}),
	}

	// Set build info gauge to "1" so it can be used in queries.
	M.buildInfo.Set(1.0)

	prometheus.MustRegister(M.PortalHitCounter)
	prometheus.MustRegister(M.PageHitCounter)
	prometheus.MustRegister(M.SearchWithResultsCounter)
	prometheus.MustRegister(M.SearchNoResultsCounter)
	prometheus.MustRegister(M.buildInfo)
}
