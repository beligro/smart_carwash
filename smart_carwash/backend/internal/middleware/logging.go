package middleware

import (
	"time"

	"carwash_backend/internal/logger"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/sirupsen/logrus"
)

// LoggingMiddleware создает middleware для логирования HTTP запросов
func LoggingMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		// Генерируем trace ID для запроса
		traceID := uuid.New().String()
		c.Set("trace_id", traceID)

		// Начало запроса
		start := time.Now()
		path := c.Request.URL.Path
		method := c.Request.Method
		clientIP := c.ClientIP()
		userAgent := c.Request.UserAgent()

		// Логируем начало запроса
		logger.WithFields(logrus.Fields{
			"trace_id":   traceID,
			"method":     method,
			"path":       path,
			"ip":         clientIP,
			"user_agent": userAgent,
			"handler":    c.HandlerName(),
		}).Info("HTTP request started")

		// Обрабатываем запрос
		c.Next()

		// Конец запроса
		latency := time.Since(start)
		status := c.Writer.Status()
		bodySize := c.Writer.Size()

		// Логируем завершение запроса
		logger.WithFields(logrus.Fields{
			"trace_id":    traceID,
			"method":      method,
			"path":        path,
			"status_code": status,
			"duration":    latency.Milliseconds(),
			"body_size":   bodySize,
			"ip":          clientIP,
			"user_agent":  userAgent,
			"handler":     c.HandlerName(),
		}).Info("HTTP request completed")
	}
}
