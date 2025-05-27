package handlers

import (
	"net/http"
	"strconv"

	"carwash_backend/internal/domain/user/models"
	"carwash_backend/internal/domain/user/service"

	"github.com/gin-gonic/gin"
)

// Handler структура для обработчиков HTTP запросов пользователей
type Handler struct {
	service service.Service
}

// NewHandler создает новый экземпляр Handler
func NewHandler(service service.Service) *Handler {
	return &Handler{
		service: service,
	}
}

// RegisterRoutes регистрирует маршруты для пользователей
func (h *Handler) RegisterRoutes(router *gin.RouterGroup) {
	userRoutes := router.Group("/users")
	{
		userRoutes.POST("", h.createUser)
		userRoutes.GET("/by-telegram-id", h.getUserByTelegramID) // telegram_id в query параметре
	}
}

// getUserByTelegramID обработчик для получения пользователя по telegram_id
func (h *Handler) getUserByTelegramID(c *gin.Context) {
	// Получаем telegram_id из query параметра
	telegramIDStr := c.Query("telegram_id")
	if telegramIDStr == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Не указан Telegram ID"})
		return
	}

	// Преобразуем строку в int64
	telegramID, err := strconv.ParseInt(telegramIDStr, 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Некорректный Telegram ID"})
		return
	}

	// Получаем пользователя по telegram_id
	user, err := h.service.GetUserByTelegramID(telegramID)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Пользователь не найден. Возможно, вы не зарегистрированы в боте. Пожалуйста, нажмите /start в боте."})
		return
	}

	// Возвращаем пользователя
	c.JSON(http.StatusOK, models.CreateUserResponse{User: *user})
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
