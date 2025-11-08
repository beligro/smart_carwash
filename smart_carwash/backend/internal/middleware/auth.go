package middleware

import (
	"net/http"
	"strings"

	authService "carwash_backend/internal/domain/auth/service"

	"github.com/gin-gonic/gin"
)

// CashierMiddleware создает middleware для проверки авторизации кассира
func CashierMiddleware(authService authService.Service) gin.HandlerFunc {
	return func(c *gin.Context) {
		// Получаем токен из заголовка
		token := c.GetHeader("Authorization")
		token = strings.TrimPrefix(token, "Bearer ")

		if token == "" {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Токен не предоставлен"})
			c.Abort()
			return
		}

		// Проверяем токен через auth service (с кэшем невалидных токенов)
		claims, err := authService.ValidateToken(c.Request.Context(), token)
		if err != nil {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Недействительный токен"})
			c.Abort()
			return
		}

		// Проверяем, что это кассир, а не администратор
		if claims.IsAdmin {
			c.JSON(http.StatusForbidden, gin.H{"error": "Доступ запрещен"})
			c.Abort()
			return
		}

		// Устанавливаем ID кассира в контекст
		c.Set("cashier_id", claims.ID)
		c.Next()
	}
}













