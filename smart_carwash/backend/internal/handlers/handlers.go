package handlers

import (
	"encoding/json"
	"net/http"
	"strconv"

	"carwash_backend/internal/models"
	"carwash_backend/internal/service"
	"carwash_backend/internal/telegram"

	"github.com/gin-gonic/gin"
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

// Handler структура для обработчиков HTTP запросов
type Handler struct {
	service service.Service
	bot     *telegram.Bot
}

// NewHandler создает новый экземпляр Handler
func NewHandler(service service.Service, bot *telegram.Bot) *Handler {
	return &Handler{
		service: service,
		bot:     bot,
	}
}

// InitRoutes инициализирует маршруты
func (h *Handler) InitRoutes(router *gin.Engine) {
	api := router.Group("/")
	{
		// Информация о мойке
		api.GET("/wash-info", h.getWashInfo)
		api.GET("/wash-info/:user_id", h.getWashInfoForUser)

		// Пользователи
		api.POST("/users", h.createUser)

		// Сессии
		api.POST("/sessions", h.createSession)
		api.GET("/sessions/:user_id", h.getUserSession)

		// Вебхук для Telegram бота
		api.POST("/webhook", h.handleWebhook)
	}
}

// handleWebhook обрабатывает вебхук от Telegram
func (h *Handler) handleWebhook(c *gin.Context) {
	// Читаем тело запроса
	body, err := c.GetRawData()
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Не удалось прочитать тело запроса"})
		return
	}

	// Парсим обновление
	var update tgbotapi.Update
	if err := json.Unmarshal(body, &update); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Не удалось распарсить обновление"})
		return
	}

	// Обрабатываем обновление
	h.bot.ProcessUpdate(update)

	c.JSON(http.StatusOK, gin.H{"status": "ok"})
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

// getWashInfoForUser обработчик для получения информации о мойке для конкретного пользователя
func (h *Handler) getWashInfoForUser(c *gin.Context) {
	// Получаем ID пользователя из URL
	userIDStr := c.Param("user_id")
	userID, err := strconv.ParseUint(userIDStr, 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Некорректный ID пользователя"})
		return
	}

	// Получаем информацию о мойке для пользователя
	washInfo, err := h.service.GetWashInfoForUser(uint(userID))
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

// createSession обработчик для создания сессии
func (h *Handler) createSession(c *gin.Context) {
	var req models.CreateSessionRequest

	// Парсим JSON из тела запроса
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Создаем сессию
	session, err := h.service.CreateSession(&req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	// Возвращаем созданную сессию
	c.JSON(http.StatusOK, models.CreateSessionResponse{Session: *session})
}

// getUserSession обработчик для получения сессии пользователя
func (h *Handler) getUserSession(c *gin.Context) {
	// Получаем ID пользователя из URL
	userIDStr := c.Param("user_id")
	userID, err := strconv.ParseUint(userIDStr, 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Некорректный ID пользователя"})
		return
	}

	// Получаем сессию пользователя
	session, err := h.service.GetUserSession(&models.GetUserSessionRequest{
		UserID: uint(userID),
	})
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Сессия не найдена"})
		return
	}

	// Возвращаем сессию пользователя
	c.JSON(http.StatusOK, models.GetUserSessionResponse{Session: session})
}
