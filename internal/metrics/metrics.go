package metrics

import (
	"strconv"
	"time"

	"github.com/labstack/echo/v4"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

type Registry struct {
	Service        string
	Native         *prometheus.Registry
	HTTPRequests   *prometheus.CounterVec
	HTTPDuration   *prometheus.HistogramVec
	HTTPInFlight   prometheus.Gauge
	DomainCounters map[string]prometheus.Counter
}

func NewRegistry(service string) *Registry {
	reg := prometheus.NewRegistry()
	r := &Registry{
		Service: service,
		Native:  reg,
		HTTPRequests: prometheus.NewCounterVec(prometheus.CounterOpts{
			Name: "http_requests_total",
			Help: "Total HTTP requests.",
		}, []string{"service", "method", "route", "status"}),
		HTTPDuration: prometheus.NewHistogramVec(prometheus.HistogramOpts{
			Name:    "http_request_duration_seconds",
			Help:    "HTTP request duration.",
			Buckets: prometheus.DefBuckets,
		}, []string{"service", "method", "route", "status"}),
		HTTPInFlight: prometheus.NewGauge(prometheus.GaugeOpts{
			Name: "http_requests_in_flight",
			Help: "HTTP requests currently in flight.",
		}),
		DomainCounters: map[string]prometheus.Counter{},
	}
	reg.MustRegister(r.HTTPRequests, r.HTTPDuration, r.HTTPInFlight)
	for _, name := range []string{
		"users_created_total", "subscriptions_assigned_total", "qr_passes_generated_total",
		"qr_scans_total", "qr_scans_allowed_total", "qr_scans_denied_total",
		"access_checkins_total", "access_checkouts_total", "payments_uploaded_total",
		"payments_approved_total", "telegram_notifications_sent_total", "telegram_notifications_failed_total",
		"kafka_messages_published_total", "kafka_messages_consumed_total", "minio_uploads_total",
	} {
		counter := prometheus.NewCounter(prometheus.CounterOpts{Name: name, Help: name})
		r.DomainCounters[name] = counter
		reg.MustRegister(counter)
	}
	return r
}

func (r *Registry) Inc(name string) {
	if counter, ok := r.DomainCounters[name]; ok {
		counter.Inc()
	}
}

func (r *Registry) Middleware() echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			start := time.Now()
			r.HTTPInFlight.Inc()
			err := next(c)
			r.HTTPInFlight.Dec()

			status := c.Response().Status
			if status == 0 {
				status = 200
			}
			route := c.Path()
			if route == "" {
				route = c.Request().URL.Path
			}
			labels := prometheus.Labels{
				"service": r.Service,
				"method":  c.Request().Method,
				"route":   route,
				"status":  strconv.Itoa(status),
			}
			r.HTTPRequests.With(labels).Inc()
			r.HTTPDuration.With(labels).Observe(time.Since(start).Seconds())
			return err
		}
	}
}

func (r *Registry) Handler() echo.HandlerFunc {
	return echo.WrapHandler(promhttp.HandlerFor(r.Native, promhttp.HandlerOpts{}))
}
