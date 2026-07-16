package middleware

import (
	"regexp"
	"sync/atomic"
	"time"

	"carwash_backend/internal/logger"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/sirupsen/logrus"
)

var (
	// Счётчик запросов
	requestCounter int64
	// Счётчик активных запросов
	activeRequests int64

	// Гостевой токен в URL (/guest/sessions/g_<ts>_<hex>/...) — это фактически пароль
	// доступа к сессии. В логи пишем усечённую форму, чтобы полный токен не оседал
	// в логах/мониторинге. Функционал не меняется — маскируется только логирование.
	guestTokenRe = regexp.MustCompile(`g_\d+_[0-9a-f]+`)
)

// MaskGuestToken усекает гостевой токен в пути для логов: g_1784100101692_3fb3c584 -> g_...c584
func MaskGuestToken(path string) string {
	return guestTokenRe.ReplaceAllStringFunc(path, func(tok string) string {
		if len(tok) <= 6 {
			return "g_..."
		}
		return "g_..." + tok[len(tok)-4:]
	})
}

// LoggingMiddleware создает middleware для логирования HTTP запросов
func LoggingMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		// Генерируем trace ID для запроса
		traceID := uuid.New().String()
		c.Set("trace_id", traceID)

		// Увеличиваем счётчики
		reqID := atomic.AddInt64(&requestCounter, 1)
		active := atomic.AddInt64(&activeRequests, 1)
		defer atomic.AddInt64(&activeRequests, -1)

		// Начало запроса
		start := time.Now()
		path := MaskGuestToken(c.Request.URL.Path)
		method := c.Request.Method
		clientIP := c.ClientIP()
		userAgent := c.Request.UserAgent()

		// Логируем начало запроса
		logger.WithFields(logrus.Fields{
			"trace_id":     traceID,
			"request_id":   reqID,
			"method":       method,
			"path":         path,
			"ip":           clientIP,
			"user_agent":   userAgent,
			"handler":      c.HandlerName(),
			"active_count": active,
		}).Info("→ HTTP request started")

		// Обрабатываем запрос
		c.Next()

		// Конец запроса
		latency := time.Since(start)
		status := c.Writer.Status()
		bodySize := c.Writer.Size()
		activeAfter := atomic.LoadInt64(&activeRequests)

		// Логируем завершение запроса
		logger.WithFields(logrus.Fields{
			"trace_id":     traceID,
			"request_id":   reqID,
			"method":       method,
			"path":         path,
			"status_code":  status,
			"duration_ms":  latency.Milliseconds(),
			"duration":     latency.String(),
			"body_size":    bodySize,
			"ip":           clientIP,
			"user_agent":   userAgent,
			"handler":      c.HandlerName(),
			"active_count": activeAfter,
		}).Info("← HTTP request completed")

		// Алерты для медленных запросов (исключаем таймауты - они логируются отдельно)
		if latency > 3*time.Second && status != 408 {
			logger.WithFields(logrus.Fields{
				"trace_id":    traceID,
				"request_id":  reqID,
				"method":      method,
				"path":        path,
				"duration":    latency.String(),
				"status_code": status,
			}).Warn("🐌 SLOW REQUEST detected")
		}

		// Алерты для таймаутов (более критично чем медленные запросы)
		if status == 408 {
			logger.WithFields(logrus.Fields{
				"trace_id":   traceID,
				"request_id": reqID,
				"method":     method,
				"path":       path,
				"duration":   latency.String(),
			}).Error("⏱️  REQUEST TIMEOUT: Request exceeded timeout limit")
		}

		// Алерты для высокой нагрузки (увеличили порог - 50 вместо 20)
		if activeAfter > 50 {
			logger.WithFields(logrus.Fields{
				"trace_id":     traceID,
				"request_id":   reqID,
				"active_count": activeAfter,
			}).Warn("⚠️  HIGH LOAD: Many active requests")
		}

		// Алерты для ошибок сервера
		if status >= 500 {
			logger.WithFields(logrus.Fields{
				"trace_id":    traceID,
				"request_id":  reqID,
				"method":      method,
				"path":        path,
				"status_code": status,
				"duration":    latency.String(),
			}).Error("🚨 SERVER ERROR detected")
		}
	}
}

// GetRequestStats возвращает статистику запросов
func GetRequestStats() (totalRequests int64, activeRequests int64) {
	return atomic.LoadInt64(&requestCounter), atomic.LoadInt64(&activeRequests)
}
