package bookserver

import "github.com/prometheus/client_golang/prometheus"

// Declare the metrics here
var (
	prom_version = prometheus.NewGauge(prometheus.GaugeOpts{
		Name: "version",
		Help: "Version information about this binary",
		ConstLabels: map[string]string{
			"version": "v0.0.1",
		},
	})
	prom_httpRequestTotal = prometheus.NewCounterVec(prometheus.CounterOpts{
		Name: "http_requests_total",
		Help: "Count of all http requests",
	}, []string{"method", "code"})

	prom_notFoundTotal = prometheus.NewCounterVec(prometheus.CounterOpts{
		Name: "http_not_found_request_total",
		Help: "Count of all http not found method",
	}, []string{"method", "URL"})

	prom_httpRequestDurationSeconds = prometheus.NewHistogramVec(
		prometheus.HistogramOpts{
			Name:    "http_request_duration_seconds",
			Help:    "HTTP request duration distribution",
			Buckets: []float64{1, 2, 5, 10, 20, 60},
		}, []string{"method"})
)

// Register the metrics here
func RegPrometheusMetrics() *prometheus.Registry {
	prom := prometheus.NewRegistry()
	prom.MustRegister(prom_version)
	prom.MustRegister(prom_httpRequestTotal)
	prom.MustRegister(prom_notFoundTotal)
	prom.MustRegister(prom_httpRequestDurationSeconds)
	return prom
}
