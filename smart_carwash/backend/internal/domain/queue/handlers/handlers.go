package handlers

import (
	"net/http"

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
