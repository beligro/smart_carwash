package metrics

import (
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

// Metrics структура для хранения метрик
type Metrics struct {
	HTTPRequestsTotal     *prometheus.CounterVec
	HTTPRequestDuration   *prometheus.HistogramVec
	HTTPRequestSize       *prometheus.HistogramVec
	HTTPResponseSize      *prometheus.HistogramVec
	ActiveConnections     prometheus.Gauge
	DatabaseConnections   *prometheus.GaugeVec
	SessionMetrics        *prometheus.CounterVec
	QueueMetrics          *prometheus.GaugeVec
	ErrorMetrics          *prometheus.CounterVec
	MultipleSessionMetrics *prometheus.CounterVec
}

// NewMetrics создает новый экземпляр метрик
func NewMetrics() *Metrics {
	return &Metrics{
		HTTPRequestsTotal: promauto.NewCounterVec(
			prometheus.CounterOpts{
				Name: "http_requests_total",
				Help: "Total number of HTTP requests",
			},
			[]string{"method", "path", "status", "service"},
		),
		HTTPRequestDuration: promauto.NewHistogramVec(
			prometheus.HistogramOpts{
				Name:    "http_request_duration_seconds",
				Help:    "HTTP request duration in seconds",
				Buckets: prometheus.DefBuckets,
			},
			[]string{"method", "path", "status", "service"},
		),
		HTTPRequestSize: promauto.NewHistogramVec(
			prometheus.HistogramOpts{
				Name:    "http_request_size_bytes",
				Help:    "HTTP request size in bytes",
				Buckets: prometheus.ExponentialBuckets(100, 10, 8),
			},
			[]string{"method", "path", "service"},
		),
		HTTPResponseSize: promauto.NewHistogramVec(
			prometheus.HistogramOpts{
				Name:    "http_response_size_bytes",
				Help:    "HTTP response size in bytes",
				Buckets: prometheus.ExponentialBuckets(100, 10, 8),
			},
			[]string{"method", "path", "status", "service"},
		),
		ActiveConnections: promauto.NewGauge(
			prometheus.GaugeOpts{
				Name: "active_connections",
				Help: "Number of active connections",
			},
		),
		DatabaseConnections: promauto.NewGaugeVec(
			prometheus.GaugeOpts{
				Name: "database_connections",
				Help: "Number of database connections",
			},
			[]string{"state"},
		),
		SessionMetrics: promauto.NewCounterVec(
			prometheus.CounterOpts{
				Name: "sessions_total",
				Help: "Total number of sessions",
			},
			[]string{"status", "service_type", "with_chemistry"},
		),
		QueueMetrics: promauto.NewGaugeVec(
			prometheus.GaugeOpts{
				Name: "queue_size",
				Help: "Current queue size",
			},
			[]string{"service_type"},
		),
		ErrorMetrics: promauto.NewCounterVec(
			prometheus.CounterOpts{
				Name: "errors_total",
				Help: "Total number of errors",
			},
			[]string{"type", "service", "error_code"},
		),
		MultipleSessionMetrics: promauto.NewCounterVec(
			prometheus.CounterOpts{
				Name: "multiple_sessions_total",
				Help: "Total number of multiple session attempts",
			},
			[]string{"type", "user_id", "time_diff"},
		),
	}
}

// PrometheusMiddleware middleware для сбора HTTP метрик
func (m *Metrics) PrometheusMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		start := time.Now()
		path := c.FullPath()
		method := c.Request.Method
		service := "carwash-backend"

		// Увеличиваем счетчик активных соединений
		m.ActiveConnections.Inc()
		defer m.ActiveConnections.Dec()

		// Записываем размер запроса
		requestSize := float64(c.Request.ContentLength)
		if requestSize > 0 {
			m.HTTPRequestSize.WithLabelValues(method, path, service).Observe(requestSize)
		}

		// Обрабатываем запрос
		c.Next()

		// Записываем метрики после обработки
		status := c.Writer.Status()
		statusStr := strconv.Itoa(status)
		duration := time.Since(start).Seconds()
		responseSize := float64(c.Writer.Size())

		m.HTTPRequestsTotal.WithLabelValues(method, path, statusStr, service).Inc()
		m.HTTPRequestDuration.WithLabelValues(method, path, statusStr, service).Observe(duration)
		m.HTTPResponseSize.WithLabelValues(method, path, statusStr, service).Observe(responseSize)
	}
}

// MetricsHandler возвращает handler для /metrics endpoint
func (m *Metrics) MetricsHandler() gin.HandlerFunc {
	h := promhttp.Handler()
	return func(c *gin.Context) {
		h.ServeHTTP(c.Writer, c.Request)
	}
}

// RecordSession создает метрику для сессии
func (m *Metrics) RecordSession(status, serviceType, withChemistry string) {
	m.SessionMetrics.WithLabelValues(status, serviceType, withChemistry).Inc()
}

// UpdateQueueSize обновляет размер очереди
func (m *Metrics) UpdateQueueSize(serviceType string, size float64) {
	m.QueueMetrics.WithLabelValues(serviceType).Set(size)
}

// RecordError записывает метрику ошибки
func (m *Metrics) RecordError(errorType, service, errorCode string) {
	m.ErrorMetrics.WithLabelValues(errorType, service, errorCode).Inc()
}

// RecordMultipleSession записывает метрику попытки создания множественных сессий
func (m *Metrics) RecordMultipleSession(sessionType, userID, timeDiff string) {
	m.MultipleSessionMetrics.WithLabelValues(sessionType, userID, timeDiff).Inc()
}

// UpdateDatabaseConnections обновляет метрики подключений к БД
func (m *Metrics) UpdateDatabaseConnections(state string, count float64) {
	m.DatabaseConnections.WithLabelValues(state).Set(count)
}
