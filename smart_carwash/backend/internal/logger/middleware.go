package logger

import (
	"bytes"
	"context"
	"io"
	"time"

	"github.com/gin-gonic/gin"
)

// bodyLogWriter - структура для перехвата тела ответа
type bodyLogWriter struct {
	gin.ResponseWriter
	body *bytes.Buffer
}

// Write перехватывает запись в ResponseWriter и дублирует ее в буфер
func (w bodyLogWriter) Write(b []byte) (int, error) {
	w.body.Write(b)
	return w.ResponseWriter.Write(b)
}

// LoggerMiddleware - middleware для логирования запросов и ответов
func LoggerMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		// Начало запроса
		start := time.Now()

		// Создаем контекст с trace_id
		ctx := WithTraceID(c.Request.Context())
		c.Request = c.Request.WithContext(ctx)

		// Читаем тело запроса
		var requestBody []byte
		if c.Request.Body != nil {
			requestBody, _ = io.ReadAll(c.Request.Body)
			// Восстанавливаем тело запроса
			c.Request.Body = io.NopCloser(bytes.NewBuffer(requestBody))
		}

		// Перехватываем тело ответа
		blw := &bodyLogWriter{
			ResponseWriter: c.Writer,
			body:           bytes.NewBufferString(""),
		}
		c.Writer = blw

		// Логируем запрос
		Info(ctx, "Incoming request",
			map[string]interface{}{
				"method":       c.Request.Method,
				"path":         c.Request.URL.Path,
				"query":        c.Request.URL.RawQuery,
				"ip":           c.ClientIP(),
				"user_agent":   c.Request.UserAgent(),
				"request_id":   c.GetHeader("X-Request-ID"),
				"request_body": string(requestBody),
			})

		// Продолжаем обработку запроса
		c.Next()

		// Логируем ответ
		latency := time.Since(start)
		Info(ctx, "Outgoing response",
			map[string]interface{}{
				"status":        c.Writer.Status(),
				"latency":       latency.String(),
				"latency_ms":    float64(latency.Nanoseconds()) / 1e6,
				"response_body": blw.body.String(),
				"errors":        c.Errors.Errors(),
			})
	}
}

// ExtractUserInfo - middleware для извлечения информации о пользователе
func ExtractUserInfo() gin.HandlerFunc {
	return func(c *gin.Context) {
		// Получаем контекст
		ctx := c.Request.Context()

		// Извлекаем user_id из заголовка или параметров запроса
		if userID := c.GetHeader("X-User-ID"); userID != "" {
			ctx = WithUserID(ctx, userID)
		}

		// Извлекаем telegram_id из заголовка или параметров запроса
		if telegramID := c.GetHeader("X-Telegram-ID"); telegramID != "" {
			ctx = WithTelegramID(ctx, telegramID)
		}

		// Извлекаем session_id из заголовка или параметров запроса
		if sessionID := c.GetHeader("X-Session-ID"); sessionID != "" {
			ctx = WithSessionID(ctx, sessionID)
		} else if sessionID := c.Query("session_id"); sessionID != "" {
			ctx = WithSessionID(ctx, sessionID)
		} else if sessionID := c.Param("session_id"); sessionID != "" {
			ctx = WithSessionID(ctx, sessionID)
		}

		// Извлекаем box_id из заголовка или параметров запроса
		if boxID := c.GetHeader("X-Box-ID"); boxID != "" {
			ctx = WithBoxID(ctx, boxID)
		} else if boxID := c.Query("box_id"); boxID != "" {
			ctx = WithBoxID(ctx, boxID)
		} else if boxID := c.Param("box_id"); boxID != "" {
			ctx = WithBoxID(ctx, boxID)
		}

		// Обновляем контекст запроса
		c.Request = c.Request.WithContext(ctx)
		c.Next()
	}
}

// ContextWithTraceID - функция для создания контекста с trace_id для периодических задач
func ContextWithTraceID() context.Context {
	return WithTraceID(context.Background())
}
