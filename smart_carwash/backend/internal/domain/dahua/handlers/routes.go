package handlers

import (
	"github.com/gin-gonic/gin"

	"carwash_backend/internal/domain/dahua/middleware"
)

// SetupRoutes настраивает маршруты для Dahua интеграции
func SetupRoutes(router *gin.RouterGroup, handler *Handler) {
	// Группа маршрутов с аутентификацией
	dahuaGroup := router.Group("/dahua")
	dahuaGroup.Use(middleware.DahuaAuthMiddleware())

	// Webhook для получения событий от камеры Dahua
	dahuaGroup.POST("/anpr-webhook", handler.ANPRWebhook)

	// Health check (без аутентификации для мониторинга)
	router.GET("/dahua/health", handler.HealthCheck)

	// Heartbeat endpoint для камер Dahua (без аутентификации)
	router.POST("/NotificationInfo/DeviceInfo", handler.DeviceInfo)
}
