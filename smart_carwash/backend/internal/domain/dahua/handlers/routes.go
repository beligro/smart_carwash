package handlers

import (
	"github.com/gin-gonic/gin"

	"carwash_backend/internal/domain/dahua/middleware"
)

// SetupRoutes настраивает маршруты для Dahua интеграции
func SetupRoutes(router *gin.RouterGroup, handler *Handler) {
	// ANPR Webhook (только IP whitelist, без Basic Auth)
	router.POST("/dahua/anpr-webhook", middleware.DahuaIPWhitelistMiddleware(), handler.ANPRWebhook)

	// Health check (без аутентификации для мониторинга)
	router.GET("/dahua/health", handler.HealthCheck)

	// Heartbeat endpoint для камер Dahua (без аутентификации)
	router.POST("/NotificationInfo/DeviceInfo", handler.DeviceInfo)
	
	// KeepAlive heartbeat endpoint для камер Dahua (без аутентификации)
	router.GET("/NotificationInfo/KeepAlive", handler.KeepAlive)
	router.POST("/NotificationInfo/KeepAlive", handler.KeepAlive)
}
