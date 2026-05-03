package metrics

import (
	"testing"

	"github.com/prometheus/client_golang/prometheus"
)

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
	reg.MustRegister(m.buildInfo)

	m.PortalHitCounter.WithLabelValues("test").Inc()
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
}

func TestNamespace(t *testing.T) {
	if NAMESPACE != "portkey" {
		t.Fatalf("expected namespace 'portkey', got %q", NAMESPACE)
	}
}
