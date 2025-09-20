package handlers

import (
	"github.com/gin-gonic/gin"
)

// RegisterRoutes регистрирует маршруты для Modbus API
func (h *Handler) RegisterRoutes(router *gin.RouterGroup) {
	// Группа маршрутов для админских операций с Modbus
	adminModbus := router.Group("/admin/modbus")
	{
		// Тестирование и статус
		adminModbus.POST("/test-connection", h.TestConnection)
		adminModbus.POST("/test-coil", h.TestCoil)
		adminModbus.GET("/status", h.GetStatus)
		
		// Мониторинг и дашборд
		adminModbus.GET("/dashboard", h.GetDashboard)
		adminModbus.GET("/history", h.GetHistory)
	}
} 