package handlers

import (
	"net/http"
	"strconv"

	"carwash_backend/internal/domain/session/models"
	"carwash_backend/internal/domain/session/service"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

// Handler структура для обработчиков HTTP запросов сессий
type Handler struct {
	service service.Service
}

// NewHandler создает новый экземпляр Handler
func NewHandler(service service.Service) *Handler {
	return &Handler{
		service: service,
	}
}

// RegisterRoutes регистрирует маршруты для сессий
func (h *Handler) RegisterRoutes(router *gin.RouterGroup) {
	sessionRoutes := router.Group("/sessions")
	{
		sessionRoutes.POST("", h.createSession)
		sessionRoutes.GET("", h.getUserSession)                // user_id в query параметре
		sessionRoutes.GET("/by-id", h.getSessionByID)          // session_id в query параметре
		sessionRoutes.POST("/start", h.startSession)           // session_id в теле запроса
		sessionRoutes.POST("/complete", h.completeSession)     // session_id в теле запроса
		sessionRoutes.POST("/extend", h.extendSession)         // session_id и extension_time_minutes в теле запроса
		sessionRoutes.GET("/history", h.getUserSessionHistory) // user_id в query параметре
	}
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

// extendSession обработчик для продления сессии
func (h *Handler) extendSession(c *gin.Context) {
	var req models.ExtendSessionRequest

	// Парсим JSON из тела запроса
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Продлеваем сессию
	session, err := h.service.ExtendSession(&req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	// Возвращаем обновленную сессию
	c.JSON(http.StatusOK, models.ExtendSessionResponse{Session: session})
}

// getUserSessionHistory обработчик для получения истории сессий пользователя
func (h *Handler) getUserSessionHistory(c *gin.Context) {
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

	// Получаем лимит и смещение из query параметров
	limit := 10 // По умолчанию 10 сессий
	offset := 0 // По умолчанию начинаем с первой сессии

	limitStr := c.Query("limit")
	if limitStr != "" {
		limitVal, err := strconv.Atoi(limitStr)
		if err == nil && limitVal > 0 {
			limit = limitVal
		}
	}

	offsetStr := c.Query("offset")
	if offsetStr != "" {
		offsetVal, err := strconv.Atoi(offsetStr)
		if err == nil && offsetVal >= 0 {
			offset = offsetVal
		}
	}

	// Получаем историю сессий пользователя
	sessions, err := h.service.GetUserSessionHistory(&models.GetUserSessionHistoryRequest{
		UserID: userID,
		Limit:  limit,
		Offset: offset,
	})
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Не удалось получить историю сессий"})
		return
	}

	// Возвращаем историю сессий пользователя
	c.JSON(http.StatusOK, models.GetUserSessionHistoryResponse{Sessions: sessions})
}
