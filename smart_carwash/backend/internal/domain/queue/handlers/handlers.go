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
	router.GET("/wash-info", h.getWashInfo)
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

// getWashInfo обработчик для получения информации о мойке для пользователя
func (h *Handler) getWashInfo(c *gin.Context) {
	// Получаем ID пользователя из запроса
	userID := c.Query("user_id")
	if userID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "user_id is required"})
		return
	}

	// Получаем информацию о мойке для пользователя
	washInfo, err := h.service.GetWashInfo(userID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, washInfo)
}
