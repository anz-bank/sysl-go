package metrics

import (
	"context"
	"net/http"
	"strconv"
	"time"

	"github.com/prometheus/client_golang/prometheus/promhttp"

	"github.com/prometheus/client_golang/prometheus"
)

// Middleware is a handler which exposes prometheus metrics. Requests will log
// the total number of requests partitioned by code, status and path. Latency will
// log the total duration of requests partitioned by code, status and path.
// getPathPattern allows you to pass a function which will group paths by
// specified patterns. If not value is assigned, the requests and latency will
// be partitioned by the standalone path on the incoming request.
type Middleware struct {
	requests       *prometheus.CounterVec
	latency        *prometheus.SummaryVec
	getPathPattern func(ctx context.Context) string
}

// NewHTTPServerMetricsMiddleware returns a new Prometheus Middleware handler.
func NewHTTPServerMetricsMiddleware(registry *prometheus.Registry, serviceName string,
	getPathPattern func(ctx context.Context) string) func(next http.Handler) http.Handler {
	requestCounterVec := prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Name:        "http_server_requests_total",
			Help:        "HTTP requests processed, by status code, method and HTTP path",
			ConstLabels: prometheus.Labels{"service": serviceName},
		},
		[]string{"code", "method", "path"},
	)
	registry.MustRegister(requestCounterVec)

	latencySummary := prometheus.NewSummaryVec(
		prometheus.SummaryOpts{
			Name:        "http_server_request_duration_seconds",
			Help:        "Duration of the processed request, by status code, method and HTTP path",
			ConstLabels: prometheus.Labels{"service": serviceName},
			Objectives:  map[float64]float64{0.50: 0.01, 0.90: 0.001, 0.95: 0.001, 0.99: 0.0001},
		},
		[]string{"code", "method", "path"},
	)
	registry.MustRegister(latencySummary)

	m := Middleware{
		requests:       requestCounterVec,
		latency:        latencySummary,
		getPathPattern: getPathPattern,
	}

	return m.MonitorMetrics
}

// MonitorMetrics returns a http.Handler which will monitor the incoming
// request metrics and update the prometheus metric values.
func (m *Middleware) MonitorMetrics(next http.Handler) http.Handler {
	fn := func(w http.ResponseWriter, r *http.Request) {
		requestStart := time.Now()
		ww := NewStatusResponseWriter(w)
		next.ServeHTTP(ww, r)
		pathPattern := r.URL.Path
		if m.getPathPattern != nil {
			pathPattern = m.getPathPattern(r.Context())
		}
		m.updateMetrics(ww.Status(), r.Method, pathPattern, requestStart)
	}
	return http.HandlerFunc(fn)
}

func (m *Middleware) updateMetrics(status int, method, path string, requestStart time.Time) {
	statusString := strconv.Itoa(status)
	durationSecs := time.Since(requestStart).Seconds()
	m.requests.WithLabelValues(statusString, method, path).Inc()
	m.latency.WithLabelValues(statusString, method, path).Observe(durationSecs)
}

func Handler(registry *prometheus.Registry) http.Handler {
	registry.MustRegister(prometheus.NewProcessCollector(prometheus.ProcessCollectorOpts{}))
	registry.MustRegister(prometheus.NewGoCollector())
	return promhttp.InstrumentMetricHandler(
		registry, promhttp.HandlerFor(registry, promhttp.HandlerOpts{}),
	)
}
