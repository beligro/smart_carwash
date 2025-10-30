package middleware

import (
	"context"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
)

// TimeoutMiddleware создает middleware для таймаутов запросов
func TimeoutMiddleware(timeout time.Duration) gin.HandlerFunc {
	return func(c *gin.Context) {
		// Создаем контекст с таймаутом
		ctx, cancel := context.WithTimeout(c.Request.Context(), timeout)
		defer cancel()

		// Заменяем контекст запроса
		c.Request = c.Request.WithContext(ctx)

		// Создаем канал для завершения запроса
		done := make(chan bool, 1)

		// Запускаем обработку запроса в отдельной горутине
		go func() {
			c.Next()
			done <- true
		}()

		// Ждем завершения или таймаута
		select {
		case <-done:
			// Запрос завершился успешно
			return
		case <-ctx.Done():
			// Таймаут
			c.JSON(http.StatusRequestTimeout, gin.H{
				"error": "Request timeout",
			})
			c.Abort()
			return
		}
	}
}
