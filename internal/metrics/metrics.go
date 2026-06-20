package metrics

import (
	"github.com/kodehat/portkey/internal/build"
	"github.com/kodehat/portkey/internal/config"
	"github.com/prometheus/client_golang/prometheus"
)

const NAMESPACE = "portkey"

type Metrics struct {
	PortalHitCounter         *prometheus.CounterVec
	PageHitCounter           *prometheus.CounterVec
	SearchWithResultsCounter prometheus.Counter
	SearchNoResultsCounter   prometheus.Counter
	SearchDuration           prometheus.Histogram
	HTTPRequestDuration      *prometheus.HistogramVec
	FaviconCacheHits         prometheus.Counter
	FaviconCacheMisses       prometheus.Counter
	FaviconFetchFailures     prometheus.Counter
	FaviconCacheSize         prometheus.Gauge
	PortalCount              prometheus.Gauge
	GroupCount               prometheus.Gauge
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
		SearchDuration: prometheus.NewHistogram(prometheus.HistogramOpts{
			Namespace: NAMESPACE,
			Name:      "search_duration_seconds",
			Help:      "Search query duration in seconds.",
			Buckets:   []float64{0.001, 0.005, 0.01, 0.025, 0.05, 0.1},
		}),
		HTTPRequestDuration: prometheus.NewHistogramVec(prometheus.HistogramOpts{
			Namespace: NAMESPACE,
			Name:      "http_request_duration_seconds",
			Help:      "HTTP request duration by handler pattern.",
			Buckets:   prometheus.DefBuckets,
		}, []string{"handler"}),
		FaviconCacheHits: prometheus.NewCounter(prometheus.CounterOpts{
			Namespace: NAMESPACE,
			Name:      "favicon_cache_hits_total",
			Help:      "Total number of favicon cache hits.",
		}),
		FaviconCacheMisses: prometheus.NewCounter(prometheus.CounterOpts{
			Namespace: NAMESPACE,
			Name:      "favicon_cache_misses_total",
			Help:      "Total number of favicon cache misses.",
		}),
		FaviconFetchFailures: prometheus.NewCounter(prometheus.CounterOpts{
			Namespace: NAMESPACE,
			Name:      "favicon_fetch_failures_total",
			Help:      "Total number of failed favicon fetches.",
		}),
		FaviconCacheSize: prometheus.NewGauge(prometheus.GaugeOpts{
			Namespace: NAMESPACE,
			Name:      "favicon_cache_size",
			Help:      "Current number of favicons in the on-disk cache.",
		}),
		PortalCount: prometheus.NewGauge(prometheus.GaugeOpts{
			Namespace: NAMESPACE,
			Name:      "portals_total",
			Help:      "Total number of configured portals.",
		}),
		GroupCount: prometheus.NewGauge(prometheus.GaugeOpts{
			Namespace: NAMESPACE,
			Name:      "groups_total",
			Help:      "Total number of portal groups.",
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

	// Set static configuration gauges.
	M.PortalCount.Set(float64(len(config.C.Portals)))
	groups := make(map[string]bool)
	for _, p := range config.C.Portals {
		groups[p.Group] = true
	}
	M.GroupCount.Set(float64(len(groups)))

	prometheus.MustRegister(M.PortalHitCounter)
	prometheus.MustRegister(M.PageHitCounter)
	prometheus.MustRegister(M.SearchWithResultsCounter)
	prometheus.MustRegister(M.SearchNoResultsCounter)
	prometheus.MustRegister(M.SearchDuration)
	prometheus.MustRegister(M.HTTPRequestDuration)
	prometheus.MustRegister(M.FaviconCacheHits)
	prometheus.MustRegister(M.FaviconCacheMisses)
	prometheus.MustRegister(M.FaviconFetchFailures)
	prometheus.MustRegister(M.FaviconCacheSize)
	prometheus.MustRegister(M.PortalCount)
	prometheus.MustRegister(M.GroupCount)
	prometheus.MustRegister(M.buildInfo)
}
