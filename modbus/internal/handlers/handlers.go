package handlers

import (
	"net/http"

	"github.com/gin-gonic/gin"

	"modbus-server/internal/models"
	"modbus-server/internal/service"
)

// Handler предоставляет HTTP обработчики для Modbus API
type Handler struct {
	modbusService *service.ModbusService
}

// NewHandler создает новый экземпляр Handler
func NewHandler(modbusService *service.ModbusService) *Handler {
	return &Handler{
		modbusService: modbusService,
	}
}

// WriteCoil записывает значение в coil Modbus устройства
func (h *Handler) WriteCoil(c *gin.Context) {
	var req models.WriteCoilRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Неверный формат запроса"})
		return
	}

	response := h.modbusService.WriteCoil(&req)
	c.JSON(http.StatusOK, response)
}

// WriteLightCoil включает или выключает свет для бокса
func (h *Handler) WriteLightCoil(c *gin.Context) {
	var req models.WriteLightCoilRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Неверный формат запроса"})
		return
	}

	response := h.modbusService.WriteLightCoil(&req)
	c.JSON(http.StatusOK, response)
}

// WriteChemistryCoil включает или выключает химию для бокса
func (h *Handler) WriteChemistryCoil(c *gin.Context) {
	var req models.WriteChemistryCoilRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Неверный формат запроса"})
		return
	}

	response := h.modbusService.WriteChemistryCoil(&req)
	c.JSON(http.StatusOK, response)
}

// TestConnection тестирует соединение с Modbus устройством
func (h *Handler) TestConnection(c *gin.Context) {
	var req models.TestConnectionRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Неверный формат запроса"})
		return
	}

	response := h.modbusService.TestConnection(&req)
	c.JSON(http.StatusOK, response)
}

// TestCoil тестирует запись в конкретный регистр
func (h *Handler) TestCoil(c *gin.Context) {
	var req models.TestCoilRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Неверный формат запроса"})
		return
	}

	response := h.modbusService.TestCoil(&req)
	c.JSON(http.StatusOK, response)
}

// RegisterRoutes регистрирует маршруты для Modbus API
func (h *Handler) RegisterRoutes(router *gin.Engine) {
	// API v1 группа
	v1 := router.Group("/api/v1")
	{
		// Modbus операции
		modbus := v1.Group("/modbus")
		{
			modbus.POST("/coil", h.WriteCoil)
			modbus.POST("/light", h.WriteLightCoil)
			modbus.POST("/chemistry", h.WriteChemistryCoil)
			modbus.POST("/test-connection", h.TestConnection)
			modbus.POST("/test-coil", h.TestCoil)
		}
	}

	// Health check
	router.GET("/health", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{
			"status": "ok",
			"service": "modbus-server",
		})
	})
}
