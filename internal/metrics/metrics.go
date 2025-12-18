package metrics

import (
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
)

// Metrics holds all Prometheus metrics for the application
type Metrics struct {
	// HTTP Metrics
	HTTPRequestsTotal   *prometheus.CounterVec
	HTTPRequestDuration *prometheus.HistogramVec
	HTTPRequestSize     *prometheus.HistogramVec
	HTTPResponseSize    *prometheus.HistogramVec

	// Datastore Metrics
	DatastoreQueriesTotal    *prometheus.CounterVec
	DatastoreQueryDuration   *prometheus.HistogramVec
	DatastoreCacheHits       *prometheus.CounterVec
	DatastoreConnectionsOpen prometheus.Gauge

	// Application Metrics
	IPLookupsTotal    *prometheus.CounterVec
	IPLookupsNotFound prometheus.Counter
	IPLookupsErrors   *prometheus.CounterVec
}

// New creates and registers all Prometheus metrics
func New() *Metrics {
	return &Metrics{
		// HTTP Metrics
		HTTPRequestsTotal: promauto.NewCounterVec(
			prometheus.CounterOpts{
				Name: "http_requests_total",
				Help: "Total number of HTTP requests",
			},
			[]string{"method", "endpoint", "status"},
		),

		HTTPRequestDuration: promauto.NewHistogramVec(
			prometheus.HistogramOpts{
				Name:    "http_request_duration_seconds",
				Help:    "HTTP request latency in seconds",
				Buckets: prometheus.DefBuckets,
			},
			[]string{"method", "endpoint", "status"},
		),

		HTTPRequestSize: promauto.NewHistogramVec(
			prometheus.HistogramOpts{
				Name:    "http_request_size_bytes",
				Help:    "HTTP request size in bytes",
				Buckets: prometheus.ExponentialBuckets(100, 10, 7),
			},
			[]string{"method", "endpoint"},
		),

		HTTPResponseSize: promauto.NewHistogramVec(
			prometheus.HistogramOpts{
				Name:    "http_response_size_bytes",
				Help:    "HTTP response size in bytes",
				Buckets: prometheus.ExponentialBuckets(100, 10, 7),
			},
			[]string{"method", "endpoint", "status"},
		),

		// Datastore Metrics
		DatastoreQueriesTotal: promauto.NewCounterVec(
			prometheus.CounterOpts{
				Name: "datastore_queries_total",
				Help: "Total number of datastore queries",
			},
			[]string{"datastore", "operation", "status"},
		),

		DatastoreQueryDuration: promauto.NewHistogramVec(
			prometheus.HistogramOpts{
				Name:    "datastore_query_duration_seconds",
				Help:    "Datastore query latency in seconds",
				Buckets: prometheus.DefBuckets,
			},
			[]string{"datastore", "operation"},
		),

		DatastoreCacheHits: promauto.NewCounterVec(
			prometheus.CounterOpts{
				Name: "datastore_cache_hits_total",
				Help: "Total number of cache hits vs misses",
			},
			[]string{"datastore", "result"},
		),

		DatastoreConnectionsOpen: promauto.NewGauge(
			prometheus.GaugeOpts{
				Name: "datastore_connections_open",
				Help: "Number of open datastore connections",
			},
		),

		// Application Metrics
		IPLookupsTotal: promauto.NewCounterVec(
			prometheus.CounterOpts{
				Name: "ip_lookups_total",
				Help: "Total number of IP lookups",
			},
			[]string{"result"},
		),

		IPLookupsNotFound: promauto.NewCounter(
			prometheus.CounterOpts{
				Name: "ip_lookups_not_found_total",
				Help: "Total number of IP lookups that returned not found",
			},
		),

		IPLookupsErrors: promauto.NewCounterVec(
			prometheus.CounterOpts{
				Name: "ip_lookups_errors_total",
				Help: "Total number of IP lookup errors",
			},
			[]string{"error_type"},
		),
	}
}
