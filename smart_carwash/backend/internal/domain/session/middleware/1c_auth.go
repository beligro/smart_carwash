package middleware

import (
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
)

// Auth1CMiddleware создает middleware для аутентификации 1C webhook
func Auth1CMiddleware(apiKey string) gin.HandlerFunc {
	return func(c *gin.Context) {
		// Проверяем, что API ключ настроен
		if apiKey == "" {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "API_KEY_1C не настроен"})
			c.Abort()
			return
		}

		// Получаем заголовок Authorization
		authHeader := c.GetHeader("Authorization")
		if authHeader == "" {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Отсутствует заголовок Authorization"})
			c.Abort()
			return
		}

		// Проверяем формат Bearer token
		parts := strings.Split(authHeader, " ")
		if len(parts) != 2 || parts[0] != "Bearer" {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Неверный формат заголовка Authorization. Ожидается: Bearer <token>"})
			c.Abort()
			return
		}

		// Проверяем API ключ
		token := parts[1]
		if token != apiKey {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Неверный API ключ"})
			c.Abort()
			return
		}

		// Аутентификация успешна, продолжаем
		c.Next()
	}
} 