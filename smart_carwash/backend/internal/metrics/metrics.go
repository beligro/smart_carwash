package metrics

import (
	"time"

	"github.com/gin-gonic/gin"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

var (
	// RequestsTotal - счетчик общего количества запросов
	RequestsTotal = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Name: "carwash_http_requests_total",
			Help: "Общее количество HTTP запросов",
		},
		[]string{"method", "path", "status"},
	)

	// RequestDuration - гистограмма длительности запросов
	RequestDuration = promauto.NewHistogramVec(
		prometheus.HistogramOpts{
			Name:    "carwash_http_request_duration_seconds",
			Help:    "Длительность HTTP запросов в секундах",
			Buckets: prometheus.DefBuckets,
		},
		[]string{"method", "path", "status"},
	)

	// QueueSize - метрика размера очереди
	QueueSize = promauto.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "carwash_queue_size",
			Help: "Размер очереди по типам услуг",
		},
		[]string{"service_type"},
	)

	// ActiveSessions - метрика активных сессий
	ActiveSessions = promauto.NewGauge(
		prometheus.GaugeOpts{
			Name: "carwash_active_sessions",
			Help: "Количество активных сессий",
		},
	)

	// BoxesStatus - метрика статуса боксов
	BoxesStatus = promauto.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "carwash_boxes_status",
			Help: "Статус боксов по типам",
		},
		[]string{"status", "service_type"},
	)
)

// MetricsMiddleware - middleware для сбора метрик запросов
func MetricsMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		start := time.Now()

		// Продолжаем обработку запроса
		c.Next()

		// Собираем метрики после обработки запроса
		duration := time.Since(start).Seconds()
		status := c.Writer.Status()
		path := c.FullPath()
		if path == "" {
			path = "unknown"
		}
		method := c.Request.Method

		RequestsTotal.WithLabelValues(method, path, string(rune(status))).Inc()
		RequestDuration.WithLabelValues(method, path, string(rune(status))).Observe(duration)
	}
}

// SetupMetricsEndpoint - настраивает эндпоинт для метрик Prometheus
func SetupMetricsEndpoint(router *gin.Engine) {
	router.GET("/metrics", gin.WrapH(promhttp.Handler()))
}

// UpdateQueueSize - обновляет метрику размера очереди
func UpdateQueueSize(serviceType string, size float64) {
	QueueSize.WithLabelValues(serviceType).Set(size)
}

// UpdateActiveSessions - обновляет метрику активных сессий
func UpdateActiveSessions(count float64) {
	ActiveSessions.Set(count)
}

// UpdateBoxesStatus - обновляет метрику статуса боксов
func UpdateBoxesStatus(status, serviceType string, count float64) {
	BoxesStatus.WithLabelValues(status, serviceType).Set(count)
}
