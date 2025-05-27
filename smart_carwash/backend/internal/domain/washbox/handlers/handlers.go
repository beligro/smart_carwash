package handlers

import (
	"carwash_backend/internal/domain/washbox/service"

	"github.com/gin-gonic/gin"
)

// Handler структура для обработчиков HTTP запросов боксов мойки
type Handler struct {
	service service.Service
}

// NewHandler создает новый экземпляр Handler
func NewHandler(service service.Service) *Handler {
	return &Handler{
		service: service,
	}
}

// RegisterRoutes регистрирует маршруты для боксов мойки
func (h *Handler) RegisterRoutes(router *gin.RouterGroup) {
	// Маршруты для боксов мойки
	// Пока нет маршрутов, так как queue-status перенесен в домен queue
}
