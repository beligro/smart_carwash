package handlers

import (
	"net/http"

	"carwash_backend/internal/models"
	"carwash_backend/internal/service"

	"github.com/gin-gonic/gin"
)

// Handler структура для обработчиков HTTP запросов
type Handler struct {
	service service.Service
}

// NewHandler создает новый экземпляр Handler
func NewHandler(service service.Service) *Handler {
	return &Handler{service: service}
}

// InitRoutes инициализирует маршруты
func (h *Handler) InitRoutes(router *gin.Engine) {
	api := router.Group("/")
	{
		// Информация о мойке
		api.GET("/wash-info", h.getWashInfo)

		// Пользователи
		api.POST("/users", h.createUser)
	}
}

// getWashInfo обработчик для получения информации о мойке
func (h *Handler) getWashInfo(c *gin.Context) {
	// Получаем информацию о мойке
	washInfo, err := h.service.GetWashInfo()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, washInfo)
}

// createUser обработчик для создания пользователя
func (h *Handler) createUser(c *gin.Context) {
	var req models.CreateUserRequest

	// Парсим JSON из тела запроса
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Создаем пользователя
	user, err := h.service.CreateUser(&req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	// Возвращаем созданного пользователя
	c.JSON(http.StatusOK, models.CreateUserResponse{User: *user})
}
