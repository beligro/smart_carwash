package handlers

import (
	"net/http"
	"time"

	"github.com/gin-gonic/gin"

	"carwash_backend/internal/domain/dahua/models"
	"carwash_backend/internal/domain/dahua/service"
)

// Handler представляет HTTP обработчики для Dahua интеграции
type Handler struct {
	dahuaService service.Service
}

// NewHandler создает новый экземпляр обработчика
func NewHandler(dahuaService service.Service) *Handler {
	return &Handler{
		dahuaService: dahuaService,
	}
}

// ANPRWebhook обрабатывает webhook от камеры Dahua
// POST /api/v1/dahua/anpr-webhook
func (h *Handler) ANPRWebhook(c *gin.Context) {
	// Парсим входящий JSON от камеры Dahua
	var webhookReq models.DahuaWebhookRequest
	if err := c.ShouldBindJSON(&webhookReq); err != nil {
		c.JSON(http.StatusBadRequest, models.DahuaWebhookResponse{
			Success: false,
			Message: "Неверный формат JSON: " + err.Error(),
		})
		return
	}

	// Валидируем направление
	if !webhookReq.ValidateDirection() {
		c.JSON(http.StatusBadRequest, models.DahuaWebhookResponse{
			Success: false,
			Message: "Неверное направление движения. Допустимые значения: in, out",
		})
		return
	}

	// Логируем входящий webhook
	c.Set("dahua_license_plate", webhookReq.LicensePlate)
	c.Set("dahua_direction", webhookReq.Direction)
	c.Set("dahua_confidence", webhookReq.Confidence)

	// Преобразуем в формат для обработки
	processReq := &models.ProcessANPREventRequest{
		LicensePlate: webhookReq.LicensePlate,
		Direction:    webhookReq.Direction,
		Confidence:   webhookReq.Confidence,
		EventType:    webhookReq.EventType,
		CaptureTime:  webhookReq.CaptureTime,
		ImagePath:    webhookReq.ImagePath,
	}

	// Обрабатываем событие
	response, err := h.dahuaService.ProcessANPREvent(processReq)
	if err != nil {
		c.JSON(http.StatusInternalServerError, models.DahuaWebhookResponse{
			Success: false,
			Message: "Ошибка обработки события: " + err.Error(),
		})
		return
	}

	// Возвращаем ответ
	if response.Success {
		c.JSON(http.StatusOK, models.DahuaWebhookResponse{
			Success: true,
			Message: response.Message,
		})
	} else {
		c.JSON(http.StatusInternalServerError, models.DahuaWebhookResponse{
			Success: false,
			Message: response.Message,
		})
	}
}

// HealthCheck проверяет состояние Dahua интеграции
// GET /api/v1/dahua/health
func (h *Handler) HealthCheck(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "Dahua интеграция работает",
		"service": "dahua",
	})
}

// DeviceInfo heartbeat endpoint для камер Dahua
// POST /NotificationInfo/DeviceInfo
func (h *Handler) DeviceInfo(c *gin.Context) {
	// Логируем heartbeat от камеры
	c.Set("dahua_heartbeat", true)
	
	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "Device heartbeat received",
		"timestamp": time.Now().Format("2006-01-02T15:04:05"),
	})
}
