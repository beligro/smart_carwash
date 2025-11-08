package middleware

import (
	"context"
	"log"
	"net/http"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
)

func TimeoutMiddleware(timeout time.Duration) gin.HandlerFunc {
	return func(c *gin.Context) {
		// Создаем контекст с таймаутом
		ctx, cancel := context.WithTimeout(c.Request.Context(), timeout)
		defer cancel()

		// Заменяем контекст в запросе (DB запросы должны проверять контекст)
		c.Request = c.Request.WithContext(ctx)

		// Используем mutex для защиты записи в response
		var mu sync.Mutex
		responseWritten := false

		// Функция для безопасной записи ответа
		writeResponse := func(status int, body gin.H) bool {
			mu.Lock()
			defer mu.Unlock()
			if !responseWritten && !c.Writer.Written() {
				responseWritten = true
				c.JSON(status, body)
				c.Abort()
				return true
			}
			return false
		}

		// Используем канал для отслеживания завершения обработки
		done := make(chan struct{})
		var once sync.Once

		// Функция для завершения обработки
		finish := func() {
			once.Do(func() {
				close(done)
			})
		}

		// Запускаем обработку в горутине
		go func() {
			defer finish()
			c.Next()
		}()

		// Отслеживаем таймаут или завершение обработки
		select {
		case <-done:
			// Обработка завершилась
			mu.Lock()
			timedOut := ctx.Err() == context.DeadlineExceeded
			mu.Unlock()

			if timedOut && !responseWritten {
				log.Printf("⏱️  Request timeout: %s %s", c.Request.Method, c.Request.URL.Path)
				writeResponse(http.StatusRequestTimeout, gin.H{"error": "Request timeout"})
			}
		case <-ctx.Done():
			// Таймаут произошел до завершения обработки
			finish() // Сигнализируем, что горутина должна завершиться
			log.Printf("⏱️  Request timeout: %s %s", c.Request.Method, c.Request.URL.Path)

			// Отменяем контекст - это должно остановить DB запросы, если они проверяют контекст
			cancel()

			// Отправляем ответ таймаута
			writeResponse(http.StatusRequestTimeout, gin.H{"error": "Request timeout"})

			// Даем немного времени для завершения горутины
			// После этого она продолжит работать, но не сможет записать ответ (защищено mutex)
			select {
			case <-done:
				// Горутина завершилась
			case <-time.After(100 * time.Millisecond):
				// Горутина все еще работает, но это нормально - она не сможет записать ответ
			}
		}
	}
}
