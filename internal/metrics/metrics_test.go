package metrics

import (
	"testing"

	"github.com/kodehat/portkey/internal/build"
	"github.com/prometheus/client_golang/prometheus"
)

func TestMain(m *testing.M) {
	build.LoadBuildDetails("test")
	m.Run()
}

func TestMetricsCreateNew(t *testing.T) {
	reg := prometheus.NewRegistry()
	m := Metrics{
		PortalHitCounter: prometheus.NewCounterVec(prometheus.CounterOpts{
			Namespace: NAMESPACE,
			Name:      "test_portal_handler_requests_total",
			Help:      "Test portal counter.",
		}, []string{"portal"}),
		PageHitCounter: prometheus.NewCounterVec(prometheus.CounterOpts{
			Namespace: NAMESPACE,
			Name:      "test_page_handler_requests_total",
			Help:      "Test page counter.",
		}, []string{"path"}),
		SearchWithResultsCounter: prometheus.NewCounter(prometheus.CounterOpts{
			Namespace: NAMESPACE,
			Name:      "test_search_with_results_total",
			Help:      "Test search counter.",
		}),
		SearchNoResultsCounter: prometheus.NewCounter(prometheus.CounterOpts{
			Namespace: NAMESPACE,
			Name:      "test_search_no_results_total",
			Help:      "Test search no results counter.",
		}),
		SearchDuration: prometheus.NewHistogram(prometheus.HistogramOpts{
			Namespace: NAMESPACE,
			Name:      "test_search_duration_seconds",
			Help:      "Test search duration.",
			Buckets:   []float64{0.001, 0.005, 0.01, 0.025, 0.05, 0.1},
		}),
		HTTPRequestDuration: prometheus.NewHistogramVec(prometheus.HistogramOpts{
			Namespace: NAMESPACE,
			Name:      "test_http_request_duration_seconds",
			Help:      "Test HTTP duration.",
			Buckets:   prometheus.DefBuckets,
		}, []string{"handler"}),
		FaviconCacheHits: prometheus.NewCounter(prometheus.CounterOpts{
			Namespace: NAMESPACE,
			Name:      "test_favicon_cache_hits_total",
			Help:      "Test favicon cache hits.",
		}),
		FaviconCacheMisses: prometheus.NewCounter(prometheus.CounterOpts{
			Namespace: NAMESPACE,
			Name:      "test_favicon_cache_misses_total",
			Help:      "Test favicon cache misses.",
		}),
		FaviconFetchFailures: prometheus.NewCounter(prometheus.CounterOpts{
			Namespace: NAMESPACE,
			Name:      "test_favicon_fetch_failures_total",
			Help:      "Test favicon fetch failures.",
		}),
		FaviconCacheSize: prometheus.NewGauge(prometheus.GaugeOpts{
			Namespace: NAMESPACE,
			Name:      "test_favicon_cache_size",
			Help:      "Test favicon cache size.",
		}),
		PortalCount: prometheus.NewGauge(prometheus.GaugeOpts{
			Namespace: NAMESPACE,
			Name:      "test_portals_total",
			Help:      "Test portal count.",
		}),
		GroupCount: prometheus.NewGauge(prometheus.GaugeOpts{
			Namespace: NAMESPACE,
			Name:      "test_groups_total",
			Help:      "Test group count.",
		}),
		buildInfo: prometheus.NewGauge(prometheus.GaugeOpts{
			Namespace: NAMESPACE,
			Name:      "test_version_info",
			Help:      "Test version info.",
		}),
	}

	m.buildInfo.Set(1.0)

	reg.MustRegister(m.PortalHitCounter)
	reg.MustRegister(m.PageHitCounter)
	reg.MustRegister(m.SearchWithResultsCounter)
	reg.MustRegister(m.SearchNoResultsCounter)
	reg.MustRegister(m.SearchDuration)
	reg.MustRegister(m.HTTPRequestDuration)
	reg.MustRegister(m.FaviconCacheHits)
	reg.MustRegister(m.FaviconCacheMisses)
	reg.MustRegister(m.FaviconFetchFailures)
	reg.MustRegister(m.FaviconCacheSize)
	reg.MustRegister(m.PortalCount)
	reg.MustRegister(m.GroupCount)
	reg.MustRegister(m.buildInfo)

	// Verify increment operations work.
	m.PortalHitCounter.WithLabelValues("test").Inc()
	m.FaviconCacheHits.Inc()
	m.FaviconCacheMisses.Inc()
	m.FaviconFetchFailures.Inc()
	m.FaviconCacheSize.Inc()
	m.PortalCount.Inc()
	m.GroupCount.Inc()
	m.SearchDuration.Observe(0.01)
	m.HTTPRequestDuration.WithLabelValues("/test").Observe(0.01)

	if m.PortalHitCounter == nil {
		t.Fatal("expected PortalHitCounter to be non-nil")
	}
	if m.PageHitCounter == nil {
		t.Fatal("expected PageHitCounter to be non-nil")
	}
	if m.SearchWithResultsCounter == nil {
		t.Fatal("expected SearchWithResultsCounter to be non-nil")
	}
	if m.SearchNoResultsCounter == nil {
		t.Fatal("expected SearchNoResultsCounter to be non-nil")
	}
	if m.SearchDuration == nil {
		t.Fatal("expected SearchDuration to be non-nil")
	}
	if m.HTTPRequestDuration == nil {
		t.Fatal("expected HTTPRequestDuration to be non-nil")
	}
	if m.FaviconCacheHits == nil {
		t.Fatal("expected FaviconCacheHits to be non-nil")
	}
	if m.FaviconCacheMisses == nil {
		t.Fatal("expected FaviconCacheMisses to be non-nil")
	}
	if m.FaviconFetchFailures == nil {
		t.Fatal("expected FaviconFetchFailures to be non-nil")
	}
	if m.FaviconCacheSize == nil {
		t.Fatal("expected FaviconCacheSize to be non-nil")
	}
	if m.PortalCount == nil {
		t.Fatal("expected PortalCount to be non-nil")
	}
	if m.GroupCount == nil {
		t.Fatal("expected GroupCount to be non-nil")
	}
}

func TestNamespace(t *testing.T) {
	if NAMESPACE != "portkey" {
		t.Fatalf("expected namespace 'portkey', got %q", NAMESPACE)
	}
}
