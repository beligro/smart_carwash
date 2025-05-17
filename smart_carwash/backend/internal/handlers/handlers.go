package handlers

import (
	"encoding/json"
	"net/http"

	"carwash_backend/internal/models"
	"carwash_backend/internal/service"
	"carwash_backend/internal/telegram"

	"github.com/gin-gonic/gin"
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"github.com/google/uuid"
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
		api.GET("/queue-status", h.getQueueStatus)
		api.GET("/wash-info/:user_id", h.getWashInfoForUser)

		// Пользователи
		api.POST("/users", h.createUser)

		// Сессии
		api.POST("/sessions", h.createSession)
		api.GET("/sessions/:user_id", h.getUserSession)
		api.GET("/sessions/by-id/:session_id", h.getSessionByID)

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

// getWashInfoForUser обработчик для получения информации о мойке для конкретного пользователя
func (h *Handler) getWashInfoForUser(c *gin.Context) {
	// Получаем ID пользователя из URL
	userIDStr := c.Param("user_id")
	userID, err := uuid.Parse(userIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Некорректный ID пользователя"})
		return
	}

	// Получаем информацию о мойке для пользователя
	washInfo, err := h.service.GetWashInfoForUser(userID)
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
	userID, err := uuid.Parse(userIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Некорректный ID пользователя"})
		return
	}

	// Получаем сессию пользователя
	session, err := h.service.GetUserSession(&models.GetUserSessionRequest{
		UserID: userID,
	})
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Сессия не найдена"})
		return
	}

	// Возвращаем сессию пользователя
	c.JSON(http.StatusOK, models.GetUserSessionResponse{Session: session})
}

// getSessionByID обработчик для получения сессии по ID
func (h *Handler) getSessionByID(c *gin.Context) {
	// Получаем ID сессии из URL
	sessionIDStr := c.Param("session_id")
	sessionID, err := uuid.Parse(sessionIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Некорректный ID сессии"})
		return
	}

	// Получаем сессию по ID
	session, err := h.service.GetSession(&models.GetSessionRequest{
		SessionID: sessionID,
	})
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Сессия не найдена"})
		return
	}

	// Возвращаем сессию
	c.JSON(http.StatusOK, models.GetSessionResponse{Session: session})
}
