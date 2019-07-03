# Build An App That Exports Prometheus Metrics

In this tutorial, we are going to build a simple http api server that exports prometheus metrics at `/metrics` endpoint.


## Prometheus Metric Types

The Prometheus client libraries offer four core metric types. 

1. **Counter:**  A counter is a cumulative metric that represents a single monotonically increasing counter whose value can only increase or be reset to zero on restart. For example, you can use a counter to represent the number of requests served, tasks completed, or errors. Do not use a counter to expose a value that can decrease like number of currently running processes.

    ```go
    type Counter interface {
    Metric
    Collector

    // Inc increments the counter by 1. Use Add to increment it by arbitrary
    // non-negative values.
    Inc()
    // Add adds the given value to the counter. It panics if the value is < 0.
    Add(float64)
    }
    ```
2. **Gauge:** A gauge is a metric that represents a single numerical value that can arbitrarily go up and down. Gauges are typically used for measured values like temperatures or current memory usage, but also "counts" that can go up and down, like the number of concurrent requests.
    
    ```go
    type Gauge interface {
    Metric
    Collector

    // Set sets the Gauge to an arbitrary value.
    Set(float64)
    // Inc increments the Gauge by 1. Use Add to increment it by arbitrary
    // values.
    Inc()
    // Dec decrements the Gauge by 1. Use Sub to decrement it by arbitrary
    // values.
    Dec()
    // Add adds the given value to the Gauge. (The value can be negative,
    // resulting in a decrease of the Gauge.)
    Add(float64)
    // Sub subtracts the given value from the Gauge. (The value can be
    // negative, resulting in an increase of the Gauge.)
    Sub(float64)

    // SetToCurrentTime sets the Gauge to the current Unix time in seconds.
    SetToCurrentTime()
    } 
    ```
3. **Histogram:**  A histogram samples observations (usually things like request durations or response sizes) and counts them in configurable buckets. It also provides a sum of all observed values. A histogram with a base metric name of `<basename>` exposes multiple time series during a scrape. Let's consider `<basename>` be `http_request_duration_seconds`
    
    * `<basename>_bucket` (example: `http_request_duration_seconds_bucket`)
    * `<basename>_sum` (example: `http_request_duration_seconds_sum`)
    * `<basename>_count` (example: `http_request_duration_seconds_count`)
    
    ```go
    type Histogram interface {
    Metric
    Collector

    // Observe adds a single observation to the histogram.
    Observe(float64)
    }
    ```
4. **Summary:** Creating a Summary is somewhat similar to an Histogram, the difference is that prometheus.NewSummary() must specify which quantiles to calculate (instead of specifying buckets).
    
    ```go
    type Summary interface {
    Metric
    Collector

    // Observe adds a single observation to the summary.
    Observe(float64)
    }
    ```
    
## Making of HTTP API Server

Let's make a simple http web server using [Go Macaron](https://go-macaron.com/docs)

```go
func main() {
	// start server using Go-Macaron
	m := macaron.Classic()
	
	//Declare handlers
	m.Get("/", func(ctx *macaron.Context) {
		job() // your code here
	})
	
	m.Post("/", func(ctx *macaron.Context) {
		job()// your code here
	})

	log.Println("Server running... ...")
	log.Println(http.ListenAndServe(":8080", m))
}
```

## Integrate Prometheus Client Library with API Server

Go is one of the officially supported languages for Prometheus instrumentation. Not hugely surprising, since Prometheus is written in Go! 

```bash
go get git@github.com:prometheus/client_golang.git
```

The prometheus [Go-client](https://github.com/prometheus/client_golang) provides:

* Built-in Go metrics (memory usage, goroutines, GC, …)
* The ability to create custom metrics
* An HTTP handler for the `/metrics` endpoint

For better understanding we will divide the whole process into several steps:

1. **Declaration of metrics:** Metrics need to be declared before using them.
    
    ```go
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
    ```
2. **Register metrics:** Since metrics are declared in previous step. Now they needed to be registered.
    
    ```go 
    prom := prometheus.NewRegistry()
    prom.MustRegister(prom_version)
    prom.MustRegister(prom_httpRequestTotal)
    prom.MustRegister(prom_httpRequestDurationSeconds)
    ```
3. **Export metrics:** Once metrics are registered, they are ready to be exported. The client library we are using provides a http handler to for this job.

    ```go
    m.Get("/metrics", promhttp.HandlerFor(Prom,promhttp.HandlerOpts{}))
    ```
    
4. **Set metrics values:** When our server will get a `GET` request at `/metrics` endpoint, it will expose the metrics along with the value of it. Now we need to set the value of those metrics as our app encounter any changes. For example, we need to increment the value of `http_request_total` metric when a http request occurs.

    ```go
    m.Get("/", func(ctx *macaron.Context) {
	    start := time.Now()
		
	    job() // your code here
		
	    duration := time.Since(start)
	    prom_httpRequestDurationSeconds.With(prometheus.Labels{"method": "GET"}).Observe(duration.Seconds())
	    prom_httpRequestTotal.With(prometheus.Labels{"method": "GET", "code": strconv.Itoa(ctx.Resp.Status())}).Inc()
	})
    ```
    

## Summary
Let's summarize the whole steps in one singe code

```go
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

```

