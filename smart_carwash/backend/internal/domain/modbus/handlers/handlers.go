package handlers

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"

	"carwash_backend/internal/domain/modbus/models"
	"carwash_backend/internal/domain/modbus/service"
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

// TestConnection тестирует соединение с Modbus устройством
func (h *Handler) TestConnection(c *gin.Context) {
	var req models.TestModbusConnectionRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Неверный формат запроса"})
		return
	}

	response, err := h.modbusService.TestConnection(req.BoxID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, response)
}

// TestCoil тестирует запись в конкретный регистр
func (h *Handler) TestCoil(c *gin.Context) {
	var req models.TestModbusCoilRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Неверный формат запроса"})
		return
	}

	response, err := h.modbusService.TestCoil(req.BoxID, req.Register, req.Value)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, response)
}

// GetStatus получает статус Modbus устройства
func (h *Handler) GetStatus(c *gin.Context) {
	boxIDStr := c.Query("box_id")
	if boxIDStr == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "box_id обязателен"})
		return
	}

	boxID, err := uuid.Parse(boxIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Неверный формат box_id"})
		return
	}

	response, err := h.modbusService.GetStatus(boxID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, response)
}

// GetDashboard получает данные дашборда мониторинга Modbus
func (h *Handler) GetDashboard(c *gin.Context) {
	timeRange := c.DefaultQuery("time_range", "24h")

	response, err := h.modbusService.GetDashboard(timeRange)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, response)
}

// GetHistory получает историю операций Modbus
func (h *Handler) GetHistory(c *gin.Context) {
	var req models.GetModbusHistoryRequest

	// Парсим query параметры
	if boxIDStr := c.Query("box_id"); boxIDStr != "" {
		boxID, err := uuid.Parse(boxIDStr)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Неверный формат box_id"})
			return
		}
		req.BoxID = &boxID
	}

	if operation := c.Query("operation"); operation != "" {
		req.Operation = &operation
	}

	if successStr := c.Query("success"); successStr != "" {
		success := successStr == "true"
		req.Success = &success
	}

	// Парсим лимит и оффсет
	req.Limit = 50 // По умолчанию
	if limitStr := c.Query("limit"); limitStr != "" {
		if limit, err := strconv.Atoi(limitStr); err == nil && limit > 0 && limit <= 100 {
			req.Limit = limit
		}
	}

	req.Offset = 0 // По умолчанию
	if offsetStr := c.Query("offset"); offsetStr != "" {
		if offset, err := strconv.Atoi(offsetStr); err == nil && offset >= 0 {
			req.Offset = offset
		}
	}

	response, err := h.modbusService.GetHistory(&req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, response)
}
