package main

import (
	"log"
	"net/http"
	"strconv"
	"time"

	"github.com/prometheus/client_golang/prometheus/promhttp"

	"github.com/prometheus/client_golang/prometheus"

	"gopkg.in/macaron.v1"
)

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

	prom_httpRequestDurationSeconds = prometheus.NewHistogramVec(
		prometheus.HistogramOpts{
			Name:    "http_request_duration_seconds",
			Help:    "HTTP request duration distribution",
			Buckets: []float64{1, 2, 5, 10, 20, 60},
		}, []string{"method"})
)

func job() {
	// your code here
}

func main() {
	// start server using Go-Macaron
	m := macaron.Classic()

	// register metrics
	prom := prometheus.NewRegistry()
	prom.MustRegister(prom_version)
	prom.MustRegister(prom_httpRequestTotal)
	prom.MustRegister(prom_httpRequestDurationSeconds)

	//Declare handlers
	m.Get("/", func(ctx *macaron.Context) {
		start := time.Now()

		job() // your code here

		duration := time.Since(start)
		prom_httpRequestDurationSeconds.With(prometheus.Labels{"method": "GET"}).Observe(duration.Seconds())
		prom_httpRequestTotal.With(prometheus.Labels{"method": "GET", "code": strconv.Itoa(ctx.Resp.Status())}).Inc()
	})

	m.Post("/", func(ctx *macaron.Context) {
		start := time.Now()

		job() // your code here

		duration := time.Since(start)
		prom_httpRequestDurationSeconds.With(prometheus.Labels{"method": "Post"}).Observe(duration.Seconds())
		prom_httpRequestTotal.With(prometheus.Labels{"method": "Post", "code": strconv.Itoa(ctx.Resp.Status())}).Inc()

	})

	m.Get("/metrics", promhttp.HandlerFor(prom, promhttp.HandlerOpts{}))

	log.Println("Server running... ...")
	log.Println(http.ListenAndServe(":8080", m))
}
