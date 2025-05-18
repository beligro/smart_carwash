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
		api.GET("/wash-info", h.getWashInfoForUser) // user_id в query параметре

		// Пользователи
		api.POST("/users", h.createUser)
		api.GET("/users/by-telegram-id", h.getUserByTelegramID) // telegram_id в query параметре

		// Сессии
		api.POST("/sessions", h.createSession)
		api.GET("/sessions", h.getUserSession)            // user_id в query параметре
		api.GET("/sessions/by-id", h.getSessionByID)      // session_id в query параметре
		api.POST("/sessions/start", h.startSession)       // session_id в теле запроса
		api.POST("/sessions/complete", h.completeSession) // session_id в теле запроса

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

// getWashInfoForUser обработчик для получения информации о сессии пользователя
func (h *Handler) getWashInfoForUser(c *gin.Context) {
	// Получаем ID пользователя из query параметра
	userIDStr := c.Query("user_id")
	if userIDStr == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Не указан ID пользователя"})
		return
	}

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
	if session == nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Сессия не найдена"})
		return
	}

	// Возвращаем только информацию о сессии пользователя
	c.JSON(http.StatusOK, models.GetUserSessionResponse{Session: session})
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
	// Получаем ID пользователя из query параметра
	userIDStr := c.Query("user_id")
	if userIDStr == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Не указан ID пользователя"})
		return
	}

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
	// Получаем ID сессии из query параметра
	sessionIDStr := c.Query("session_id")
	if sessionIDStr == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Не указан ID сессии"})
		return
	}

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

// startSession обработчик для запуска сессии (перевод в статус active)
func (h *Handler) startSession(c *gin.Context) {
	var req models.StartSessionRequest

	// Парсим JSON из тела запроса
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Запускаем сессию
	session, err := h.service.StartSession(&req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	// Возвращаем обновленную сессию
	c.JSON(http.StatusOK, models.StartSessionResponse{Session: session})
}

// completeSession обработчик для завершения сессии (перевод в статус complete)
func (h *Handler) completeSession(c *gin.Context) {
	var req models.CompleteSessionRequest

	// Парсим JSON из тела запроса
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Завершаем сессию
	session, err := h.service.CompleteSession(&req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	// Возвращаем обновленную сессию
	c.JSON(http.StatusOK, models.CompleteSessionResponse{Session: session})
}
