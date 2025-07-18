package handlers

import (
	"net/http"
	"strconv"
	"time"

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

	// Административные маршруты
	adminRoutes := router.Group("/admin/sessions")
	{
		adminRoutes.GET("", h.adminListSessions)
		adminRoutes.GET("/by-id", h.adminGetSession)
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
	// Получаем user_id из query параметра
	userIDStr := c.Query("user_id")
	if userIDStr == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Не указан User ID"})
		return
	}

	// Преобразуем строку в uuid.UUID
	userID, err := uuid.Parse(userIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Некорректный User ID"})
		return
	}

	// Получаем параметры пагинации
	limit := 10
	offset := 0

	if limitStr := c.Query("limit"); limitStr != "" {
		if l, err := strconv.Atoi(limitStr); err == nil {
			limit = l
		}
	}

	if offsetStr := c.Query("offset"); offsetStr != "" {
		if o, err := strconv.Atoi(offsetStr); err == nil {
			offset = o
		}
	}

	// Создаем запрос
	req := models.GetUserSessionHistoryRequest{
		UserID: userID,
		Limit:  limit,
		Offset: offset,
	}

	// Получаем историю сессий
	sessions, err := h.service.GetUserSessionHistory(&req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	// Возвращаем историю сессий
	c.JSON(http.StatusOK, models.GetUserSessionHistoryResponse{Sessions: sessions})
}

// adminListSessions обработчик для получения списка сессий для администратора
func (h *Handler) adminListSessions(c *gin.Context) {
	// Получаем параметры фильтрации из query
	var req models.AdminListSessionsRequest

	// User ID
	if userIDStr := c.Query("user_id"); userIDStr != "" {
		if userID, err := uuid.Parse(userIDStr); err == nil {
			req.UserID = &userID
		}
	}

	// Box ID
	if boxIDStr := c.Query("box_id"); boxIDStr != "" {
		if boxID, err := uuid.Parse(boxIDStr); err == nil {
			req.BoxID = &boxID
		}
	}

	// Box Number
	if boxNumberStr := c.Query("box_number"); boxNumberStr != "" {
		if boxNumber, err := strconv.Atoi(boxNumberStr); err == nil {
			req.BoxNumber = &boxNumber
		}
	}

	// Статус
	if status := c.Query("status"); status != "" {
		req.Status = &status
	}

	// Тип услуги
	if serviceType := c.Query("service_type"); serviceType != "" {
		req.ServiceType = &serviceType
	}

	// Дата от
	if dateFromStr := c.Query("date_from"); dateFromStr != "" {
		if dateFrom, err := time.Parse("2006-01-02", dateFromStr); err == nil {
			req.DateFrom = &dateFrom
		}
	}

	// Дата до
	if dateToStr := c.Query("date_to"); dateToStr != "" {
		if dateTo, err := time.Parse("2006-01-02", dateToStr); err == nil {
			req.DateTo = &dateTo
		}
	}

	// Лимит
	if limitStr := c.Query("limit"); limitStr != "" {
		if limit, err := strconv.Atoi(limitStr); err == nil {
			req.Limit = &limit
		}
	}

	// Смещение
	if offsetStr := c.Query("offset"); offsetStr != "" {
		if offset, err := strconv.Atoi(offsetStr); err == nil {
			req.Offset = &offset
		}
	}

	// Получаем список сессий
	resp, err := h.service.AdminListSessions(&req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	// Логируем мета-параметры
	c.Set("meta", gin.H{
		"filters": gin.H{
			"user_id":      req.UserID,
			"box_id":       req.BoxID,
			"box_number":   req.BoxNumber,
			"status":       req.Status,
			"service_type": req.ServiceType,
			"date_from":    req.DateFrom,
			"date_to":      req.DateTo,
			"limit":        req.Limit,
			"offset":       req.Offset,
		},
		"total": resp.Total,
	})

	c.JSON(http.StatusOK, resp)
}

// adminGetSession обработчик для получения сессии по ID для администратора
func (h *Handler) adminGetSession(c *gin.Context) {
	var req models.AdminGetSessionRequest

	// Получаем ID из query параметра
	idStr := c.Query("id")
	if idStr == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "ID сессии обязателен"})
		return
	}

	id, err := uuid.Parse(idStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Некорректный ID сессии"})
		return
	}

	req.ID = id

	// Получаем сессию
	resp, err := h.service.AdminGetSession(&req)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
		return
	}

	// Логируем мета-параметры
	c.Set("meta", gin.H{
		"session_id": req.ID,
		"user_id":    resp.Session.UserID,
		"status":     resp.Session.Status,
	})

	c.JSON(http.StatusOK, resp)
}
