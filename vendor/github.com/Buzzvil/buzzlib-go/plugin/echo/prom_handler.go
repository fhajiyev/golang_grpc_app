package echo

import (
	"strconv"
	"time"

	"github.com/labstack/echo"
	"github.com/labstack/echo/middleware"
	"github.com/prometheus/client_golang/prometheus"
)

// PrometheusConfig is config struct for prom.
type PrometheusConfig struct {
	Skipper   middleware.Skipper
	Namespace string
}

var (
	promInitialized = false
	// DefaultPrometheusConfig is default configuration.
	DefaultPrometheusConfig = PrometheusConfig{
		Skipper:   middleware.DefaultSkipper,
		Namespace: "echo",
	}
)

var (
	echoReqQPS       *prometheus.CounterVec
	echoReqQPSPerApp *prometheus.CounterVec
	echoReqDuration  *prometheus.SummaryVec
	echoOutBytes     prometheus.Summary
)

func initCollector(namespace string) {
	echoReqQPS = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Namespace: namespace,
			Name:      "http_request_total",
			Help:      "HTTP requests processed.",
		},
		[]string{"code", "method", "host", "url"},
	)
	echoReqQPSPerApp = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Namespace: namespace,
			Name:      "http_request_total_per_app",
			Help:      "HTTP requests processed.",
		},
		[]string{"url", "app_id"},
	)
	echoReqDuration = prometheus.NewSummaryVec(
		prometheus.SummaryOpts{
			Namespace: namespace,
			Name:      "http_request_duration_seconds",
			Help:      "HTTP request latencies in seconds.",
		},
		[]string{"method", "host", "url"},
	)
	echoOutBytes = prometheus.NewSummary(
		prometheus.SummaryOpts{
			Namespace: namespace,
			Name:      "http_response_size_bytes",
			Help:      "HTTP response bytes.",
		},
	)
	prometheus.MustRegister(echoReqQPS, echoReqQPSPerApp, echoReqDuration, echoOutBytes)
}

// NewMetric initializes prom.
func NewMetric() echo.MiddlewareFunc {
	return NewMetricWithConfig(DefaultPrometheusConfig)
}

// NewMetricWithConfig initializes prom with configuration.
func NewMetricWithConfig(config PrometheusConfig) echo.MiddlewareFunc {
	if !promInitialized {
		initCollector(config.Namespace)
		promInitialized = true
	}
	if config.Skipper == nil {
		config.Skipper = DefaultPrometheusConfig.Skipper
	}
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			if config.Skipper(c) {
				return next(c)
			}

			req := c.Request()
			res := c.Response()
			start := time.Now()

			err := next(c)

			uri := c.Path()
			status := strconv.Itoa(res.Status)
			elapsed := time.Since(start).Seconds()
			bytesOut := float64(res.Size)
			appID := req.Header.Get("Buzz-App-ID")
			echoReqQPS.WithLabelValues(status, req.Method, req.Host, uri).Inc()
			echoReqQPSPerApp.WithLabelValues(uri, appID).Inc()
			echoReqDuration.WithLabelValues(req.Method, req.Host, uri).Observe(elapsed)
			echoOutBytes.Observe(bytesOut)

			return err
		}
	}
}
