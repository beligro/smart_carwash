package handlers

import (
	"net/http"

	"carwash_backend/internal/domain/queue/models"
	"carwash_backend/internal/domain/queue/service"

	"github.com/gin-gonic/gin"
)

// Handler структура для обработчиков HTTP запросов очереди
type Handler struct {
	service service.Service
}

// NewHandler создает новый экземпляр Handler
func NewHandler(service service.Service) *Handler {
	return &Handler{
		service: service,
	}
}

// RegisterRoutes регистрирует маршруты для очереди
func (h *Handler) RegisterRoutes(router *gin.RouterGroup) {
	router.GET("/queue-status", h.getQueueStatus)

	// Административные маршруты
	adminRoutes := router.Group("/admin/queue")
	{
		adminRoutes.GET("/status", h.adminGetQueueStatus)
	}
}

// getQueueStatus обработчик для получения статуса очереди и боксов
func (h *Handler) getQueueStatus(c *gin.Context) {
	// Получаем статус очереди и боксов
	queueStatus, err := h.service.GetQueueStatus()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, queueStatus)
}

// adminGetQueueStatus обработчик для получения детального статуса очереди для администратора
func (h *Handler) adminGetQueueStatus(c *gin.Context) {
	// Получаем параметры из query
	includeDetails := false
	if detailsStr := c.Query("include_details"); detailsStr == "true" {
		includeDetails = true
	}

	req := &models.AdminQueueStatusRequest{
		IncludeDetails: includeDetails,
	}

	// Получаем детальный статус очереди
	resp, err := h.service.AdminGetQueueStatus(req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	// Логируем мета-параметры
	c.Set("meta", gin.H{
		"include_details":  includeDetails,
		"total_queue_size": resp.QueueStatus.TotalQueueSize,
		"has_any_queue":    resp.QueueStatus.HasAnyQueue,
	})

	c.JSON(http.StatusOK, resp)
}
