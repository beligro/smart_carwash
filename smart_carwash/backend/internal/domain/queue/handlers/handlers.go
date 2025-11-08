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
// Для мини-приложения не загружаем пользователей (includeUsers=false) для оптимизации
func (h *Handler) getQueueStatus(c *gin.Context) {
	// Получаем статус очереди и боксов без пользователей (для мини-приложения)
	queueStatus, err := h.service.GetQueueStatus(c.Request.Context(), false)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, queueStatus)
}

// adminGetQueueStatus обработчик для получения детального статуса очереди для администратора
func (h *Handler) adminGetQueueStatus(c *gin.Context) {
	// Получаем детальный статус очереди (всегда включаем детали)
	req := &models.AdminQueueStatusRequest{
		IncludeDetails: true,
	}

	// Получаем детальный статус очереди
	resp, err := h.service.AdminGetQueueStatus(c.Request.Context(), req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	// Логируем мета-параметры
	c.Set("meta", gin.H{
		"include_details":  true,
		"total_queue_size": resp.QueueStatus.TotalQueueSize,
		"has_any_queue":    resp.QueueStatus.HasAnyQueue,
	})

	c.JSON(http.StatusOK, resp)
}
