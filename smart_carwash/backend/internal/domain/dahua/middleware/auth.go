package middleware

import (
	"net"
	"os"
	"strings"

	"github.com/gin-gonic/gin"
)

// DahuaAuthMiddleware создает middleware для аутентификации Dahua webhook
// Проверяет Basic Auth (username/password) и IP whitelist
func DahuaAuthMiddleware() gin.HandlerFunc {
	return gin.HandlerFunc(func(c *gin.Context) {
		// 1. Проверка Basic Authentication
		username, password, hasAuth := c.Request.BasicAuth()
		if !hasAuth {
			c.JSON(401, gin.H{
				"success": false,
				"message": "Требуется аутентификация",
			})
			c.Abort()
			return
		}

		// Проверяем username и password из переменных окружения
		expectedUsername := os.Getenv("DAHUA_WEBHOOK_USERNAME")
		expectedPassword := os.Getenv("DAHUA_WEBHOOK_PASSWORD")

		if expectedUsername == "" || expectedPassword == "" {
			c.JSON(500, gin.H{
				"success": false,
				"message": "Конфигурация аутентификации не настроена",
			})
			c.Abort()
			return
		}

		if username != expectedUsername || password != expectedPassword {
			c.JSON(401, gin.H{
				"success": false,
				"message": "Неверные учетные данные",
			})
			c.Abort()
			return
		}

		// 2. Проверка IP whitelist (временно отключена)
		// clientIP := getClientIP(c)
		// allowedIPs := getAllowedIPs()

		// if !isIPAllowed(clientIP, allowedIPs) {
		// 	c.JSON(403, gin.H{
		// 		"success": false,
		// 		"message": "IP адрес не разрешен",
		// 	})
		// 	c.Abort()
		// 	return
		// }

		// Добавляем информацию об аутентификации в контекст
		c.Set("dahua_authenticated", true)
		// c.Set("dahua_client_ip", clientIP) // временно отключено
		c.Set("dahua_username", username)

		c.Next()
	})
}

// getClientIP получает реальный IP адрес клиента
func getClientIP(c *gin.Context) string {
	// Проверяем заголовки прокси
	if xForwardedFor := c.GetHeader("X-Forwarded-For"); xForwardedFor != "" {
		// X-Forwarded-For может содержать несколько IP, берем первый
		ips := strings.Split(xForwardedFor, ",")
		if len(ips) > 0 {
			return strings.TrimSpace(ips[0])
		}
	}

	if xRealIP := c.GetHeader("X-Real-IP"); xRealIP != "" {
		return xRealIP
	}

	// Возвращаем IP из RemoteAddr
	ip, _, err := net.SplitHostPort(c.Request.RemoteAddr)
	if err != nil {
		return c.Request.RemoteAddr
	}
	return ip
}

// getAllowedIPs получает список разрешенных IP из переменной окружения
func getAllowedIPs() []string {
	allowedIPsStr := os.Getenv("DAHUA_ALLOWED_IPS")
	if allowedIPsStr == "" {
		return []string{}
	}

	// Разделяем по запятой и очищаем от пробелов
	ips := strings.Split(allowedIPsStr, ",")
	var cleanIPs []string
	for _, ip := range ips {
		cleanIP := strings.TrimSpace(ip)
		if cleanIP != "" {
			cleanIPs = append(cleanIPs, cleanIP)
		}
	}

	return cleanIPs
}

// isIPAllowed проверяет, разрешен ли IP адрес
func isIPAllowed(clientIP string, allowedIPs []string) bool {
	// Если список разрешенных IP пуст, разрешаем все
	if len(allowedIPs) == 0 {
		return true
	}

	// Проверяем точное совпадение
	for _, allowedIP := range allowedIPs {
		if clientIP == allowedIP {
			return true
		}
	}

	// Проверяем подсети (базовая поддержка CIDR)
	for _, allowedIP := range allowedIPs {
		if strings.Contains(allowedIP, "/") {
			// Это CIDR нотация, проверяем принадлежность к подсети
			_, network, err := net.ParseCIDR(allowedIP)
			if err != nil {
				continue
			}
			ip := net.ParseIP(clientIP)
			if ip != nil && network.Contains(ip) {
				return true
			}
		}
	}

	return false
}

