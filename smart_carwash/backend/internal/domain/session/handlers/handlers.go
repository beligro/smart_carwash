package handlers

import (
	"log"
	"io"
	"bytes"
	"encoding/json"
	"net/http"
	"strconv"
	"strings"
	"time"

	"carwash_backend/internal/domain/session/models"
	"carwash_backend/internal/domain/session/service"
	paymentService "carwash_backend/internal/domain/payment/service"
	authService "carwash_backend/internal/domain/auth/service"
	"carwash_backend/internal/domain/session/middleware"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

// Handler структура для обработчиков HTTP запросов сессий
type Handler struct {
	service        service.Service
	paymentService paymentService.Service
	authService    authService.Service
	apiKey1C       string
}

// NewHandler создает новый экземпляр Handler
func NewHandler(service service.Service, paymentService paymentService.Service, authService authService.Service, apiKey1C string) *Handler {
	return &Handler{
		service:        service,
		paymentService: paymentService,
		authService:    authService,
		apiKey1C:       apiKey1C,
	}
}

// RegisterRoutes регистрирует маршруты для сессий
func (h *Handler) RegisterRoutes(router *gin.RouterGroup) {
	sessionRoutes := router.Group("/sessions")
	{
		sessionRoutes.POST("", h.createSession)
		sessionRoutes.POST("/with-payment", h.createSessionWithPayment) // создание сессии с платежом
		sessionRoutes.GET("", h.getUserSession)                // user_id в query параметре
		sessionRoutes.GET("/for-payment", h.getUserSessionForPayment) // user_id в query параметре, включает payment_failed
		sessionRoutes.GET("/by-id", h.getSessionByID)          // session_id в query параметре
		sessionRoutes.POST("/start", h.startSession)           // session_id в теле запроса
		sessionRoutes.POST("/complete", h.completeSession)     // session_id в теле запроса
		sessionRoutes.POST("/extend", h.extendSession)         // session_id и extension_time_minutes в теле запроса
		sessionRoutes.POST("/extend-with-payment", h.extendSessionWithPayment) // session_id и extension_time_minutes в теле запроса
		sessionRoutes.GET("/payments", h.getSessionPayments)   // session_id в query параметре
		sessionRoutes.GET("/history", h.getUserSessionHistory) // user_id в query параметре
		sessionRoutes.POST("/cancel", h.cancelSession)         // session_id и user_id в теле запроса
		sessionRoutes.POST("/enable-chemistry", h.enableChemistry) // session_id и user_id в теле запроса
	}

	// Административные маршруты
	adminRoutes := router.Group("/admin/sessions")
	{
		adminRoutes.GET("", h.adminListSessions)
		adminRoutes.GET("/by-id", h.adminGetSession)
		adminRoutes.GET("/chemistry-stats", h.getChemistryStats) // статистика химии
	}

	// Маршруты для кассира
	cashierRoutes := router.Group("/cashier/sessions", h.cashierMiddleware())
	{
		cashierRoutes.GET("", h.cashierListSessions)
		cashierRoutes.GET("/active", h.cashierGetActiveSessions)
		cashierRoutes.POST("/start", h.cashierStartSession)
		cashierRoutes.POST("/complete", h.cashierCompleteSession)
		cashierRoutes.POST("/cancel", h.cashierCancelSession)
		cashierRoutes.POST("/enable-chemistry", h.cashierEnableChemistry) // включение химии кассиром
	}

	// 1C webhook маршруты
	oneCRoutes := router.Group("/1c")
	{
		oneCRoutes.POST("/payment-callback", middleware.Auth1CMiddleware(h.apiKey1C), h.handle1CPaymentCallback)
	}
}

// createSession обработчик для создания сессии
func (h *Handler) createSession(c *gin.Context) {
	var req models.CreateSessionRequest

	// Парсим JSON из тела запроса
	if err := c.ShouldBindJSON(&req); err != nil {
		log.Printf("API Error - createSession: ошибка парсинга JSON, error: %v", err)
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Создаем сессию
	session, err := h.service.CreateSession(&req)
	if err != nil {
		log.Printf("API Error - createSession: ошибка создания сессии, user_id: %s, service_type: %s, error: %v", req.UserID.String(), req.ServiceType, err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	// Возвращаем созданную сессию
	c.JSON(http.StatusOK, models.CreateSessionResponse{Session: *session})
}

// createSessionWithPayment обработчик для создания сессии с платежом
func (h *Handler) createSessionWithPayment(c *gin.Context) {
	var req models.CreateSessionWithPaymentRequest

	// Парсим JSON из тела запроса
	if err := c.ShouldBindJSON(&req); err != nil {
		log.Printf("API Error - createSessionWithPayment: ошибка парсинга JSON, error: %v", err)
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Создаем сессию с платежом
	response, err := h.service.CreateSessionWithPayment(&req)
	if err != nil {
		log.Printf("API Error - createSessionWithPayment: ошибка создания сессии с платежом, user_id: %s, service_type: %s, error: %v", req.UserID.String(), req.ServiceType, err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	// Возвращаем созданную сессию с платежом
	c.JSON(http.StatusOK, response)
}

// getUserSession обработчик для получения сессии пользователя
func (h *Handler) getUserSession(c *gin.Context) {
	// Получаем ID пользователя из query параметра
	userIDStr := c.Query("user_id")
	if userIDStr == "" {
		log.Printf("API Error - getUserSession: не указан ID пользователя")
		c.JSON(http.StatusBadRequest, gin.H{"error": "Не указан ID пользователя"})
		return
	}

	userID, err := uuid.Parse(userIDStr)
	if err != nil {
		log.Printf("API Error - getUserSession: некорректный ID пользователя '%s', error: %v", userIDStr, err)
		c.JSON(http.StatusBadRequest, gin.H{"error": "Некорректный ID пользователя"})
		return
	}

	// Получаем сессию пользователя
	response, err := h.service.GetUserSession(&models.GetUserSessionRequest{
		UserID: userID,
	})
	if err != nil {
		log.Printf("API Error - getUserSession: сессия не найдена для user_id: %s, error: %v", userID.String(), err)
		c.JSON(http.StatusNotFound, gin.H{"error": "Сессия не найдена"})
		return
	}

	// Возвращаем сессию пользователя
	c.JSON(http.StatusOK, response)
}

// getUserSessionForPayment обработчик для получения сессии пользователя для PaymentPage (включая payment_failed)
func (h *Handler) getUserSessionForPayment(c *gin.Context) {
	// Получаем ID пользователя из query параметра
	userIDStr := c.Query("user_id")
	if userIDStr == "" {
		log.Printf("API Error - getUserSessionForPayment: не указан ID пользователя")
		c.JSON(http.StatusBadRequest, gin.H{"error": "Не указан ID пользователя"})
		return
	}

	userID, err := uuid.Parse(userIDStr)
	if err != nil {
		log.Printf("API Error - getUserSessionForPayment: некорректный ID пользователя '%s', error: %v", userIDStr, err)
		c.JSON(http.StatusBadRequest, gin.H{"error": "Некорректный ID пользователя"})
		return
	}

	// Получаем сессию пользователя для PaymentPage
	response, err := h.service.GetUserSessionForPayment(&models.GetUserSessionRequest{
		UserID: userID,
	})
	if err != nil {
		log.Printf("API Error - getUserSessionForPayment: сессия не найдена для user_id: %s, error: %v", userID.String(), err)
		c.JSON(http.StatusNotFound, gin.H{"error": "Сессия не найдена"})
		return
	}

	// Возвращаем сессию пользователя
	c.JSON(http.StatusOK, response)
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
	response, err := h.service.GetSession(&models.GetSessionRequest{
		SessionID: sessionID,
	})
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Сессия не найдена"})
		return
	}

	// Возвращаем сессию
	c.JSON(http.StatusOK, response)
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
	response, err := h.service.CompleteSession(&req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	// Возвращаем обновленную сессию с информацией о платеже
	c.JSON(http.StatusOK, response)
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

// extendSessionWithPayment обработчик для продления сессии с оплатой
func (h *Handler) extendSessionWithPayment(c *gin.Context) {
	var req models.ExtendSessionWithPaymentRequest

	// Парсим JSON из тела запроса
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Логируем мета-параметр для поиска
	log.Printf("Запрос на продление сессии с оплатой: SessionID=%s, ExtensionTime=%d", req.SessionID, req.ExtensionTimeMinutes)

	// Продлеваем сессию с оплатой
	response, err := h.service.ExtendSessionWithPayment(&req)
	if err != nil {
		log.Printf("Ошибка продления сессии с оплатой: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	log.Printf("Успешно создан платеж продления: SessionID=%s, ExtensionTime=%d", req.SessionID, req.ExtensionTimeMinutes)
	c.JSON(http.StatusOK, response)
}

// getSessionPayments обработчик для получения платежей сессии
func (h *Handler) getSessionPayments(c *gin.Context) {
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

	// Логируем мета-параметр для поиска
	log.Printf("Запрос на получение платежей сессии: SessionID=%s", sessionID)

	// Получаем платежи сессии
	response, err := h.service.GetSessionPayments(&models.GetSessionPaymentsRequest{
		SessionID: sessionID,
	})
	if err != nil {
		log.Printf("Ошибка получения платежей сессии: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	log.Printf("Успешно получены платежи сессии: SessionID=%s", sessionID)
	c.JSON(http.StatusOK, response)
}

// cancelSession обработчик для отмены сессии
func (h *Handler) cancelSession(c *gin.Context) {
	var req models.CancelSessionRequest

	// Парсим JSON из тела запроса
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Логируем мета-параметр для поиска
	log.Printf("Запрос на отмену сессии: SessionID=%s, UserID=%s", req.SessionID, req.UserID)

	// Отменяем сессию
	response, err := h.service.CancelSession(&req)
	if err != nil {
		log.Printf("Ошибка отмены сессии: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	log.Printf("Успешно отменена сессия: SessionID=%s, UserID=%s", req.SessionID, req.UserID)
	c.JSON(http.StatusOK, response)
}

// getUserSessionHistory обработчик для получения истории сессий пользователя
func (h *Handler) getUserSessionHistory(c *gin.Context) {
	// Получаем user_id из query параметра
	userIDStr := c.Query("user_id")
	if userIDStr == "" {
		log.Printf("API Error - getUserSessionHistory: не указан User ID")
		c.JSON(http.StatusBadRequest, gin.H{"error": "Не указан User ID"})
		return
	}

	// Преобразуем строку в uuid.UUID
	userID, err := uuid.Parse(userIDStr)
	if err != nil {
		log.Printf("API Error - getUserSessionHistory: некорректный User ID '%s', error: %v", userIDStr, err)
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

// handle1CPaymentCallback обработчик для webhook от 1C для платежей через кассира
func (h *Handler) handle1CPaymentCallback(c *gin.Context) {
	bodyBytes, err := io.ReadAll(c.Request.Body)
    if err != nil {
        log.Printf("Error reading request body: %v", err)
        c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request body"})
        return
    }
    
    // Логируем тело запроса
    log.Printf("Raw request body: %s", string(bodyBytes))

	bodyBytes = bytes.TrimPrefix(bodyBytes, []byte("\xef\xbb\xbf"))
    
    // ВАЖНО: Восстанавливаем body для дальнейшего использования
    c.Request.Body = io.NopCloser(bytes.NewBuffer(bodyBytes))

	var req models.CashierPaymentRequest

	// Парсим JSON из тела запроса
	if err := json.Unmarshal(bodyBytes, &req); err != nil {
		log.Printf("Error parsing 1C payment request: %v", err)
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Логируем входящий запрос
	log.Printf("Received 1C payment callback: ServiceType=%s, WithChemistry=%t, Amount=%d, RentalTimeMinutes=%d, CarNumber=%s",
		req.ServiceType, req.WithChemistry, req.Amount, req.RentalTimeMinutes, req.CarNumber)

	// Создаем сессию через кассира
	session, err := h.service.CreateFromCashier(&req)
	if err != nil {
		log.Printf("Error creating session from cashier: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	// Создаем платеж для кассира
	payment, err := h.paymentService.CreateForCashier(session.ID, req.Amount)
	if err != nil {
		log.Printf("Error creating payment for cashier: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	// Логируем успешное создание
	log.Printf("Successfully processed 1C payment: SessionID=%s, PaymentID=%s, Amount=%d",
		session.ID, payment.ID, req.Amount)

	// Возвращаем успешный ответ
	response := models.CashierPaymentResponse{
		Success:   true,
		SessionID: session.ID.String(),
		PaymentID: payment.ID.String(),
		Message:   "Payment processed successfully",
	}

	c.JSON(http.StatusOK, response)
}

// cashierListSessions обработчик для получения списка сессий кассира
func (h *Handler) cashierListSessions(c *gin.Context) {
	// Получаем время начала смены из query параметра
	shiftStartedAtStr := c.Query("shift_started_at")
	if shiftStartedAtStr == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Не указано время начала смены"})
		return
	}

	shiftStartedAt, err := time.Parse(time.RFC3339, shiftStartedAtStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Некорректный формат времени"})
		return
	}

	// Получаем параметры пагинации
	limitStr := c.DefaultQuery("limit", "50")
	offsetStr := c.DefaultQuery("offset", "0")

	limit, err := strconv.Atoi(limitStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Некорректный параметр limit"})
		return
	}

	offset, err := strconv.Atoi(offsetStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Некорректный параметр offset"})
		return
	}

	// Создаем запрос
	req := &models.CashierSessionsRequest{
		ShiftStartedAt: shiftStartedAt,
		Limit:          limit,
		Offset:         offset,
	}

	// Получаем список сессий
	response, err := h.service.CashierListSessions(req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	// Логируем мета-параметры
	c.Set("meta", gin.H{
		"shift_started_at": shiftStartedAt,
		"limit":            limit,
		"offset":           offset,
		"total_sessions":   response.Total,
	})

	// Возвращаем результат
	c.JSON(http.StatusOK, response)
}

// cashierGetActiveSessions обработчик для получения активных сессий кассира
func (h *Handler) cashierGetActiveSessions(c *gin.Context) {
	// Получаем параметры пагинации
	limitStr := c.DefaultQuery("limit", "50")
	offsetStr := c.DefaultQuery("offset", "0")

	limit, err := strconv.Atoi(limitStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Некорректный параметр limit"})
		return
	}

	offset, err := strconv.Atoi(offsetStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Некорректный параметр offset"})
		return
	}

	// Создаем запрос
	req := &models.CashierActiveSessionsRequest{
		Limit:  limit,
		Offset: offset,
	}

	// Получаем активные сессии
	response, err := h.service.CashierGetActiveSessions(req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	// Логируем мета-параметры
	c.Set("meta", gin.H{
		"limit":          limit,
		"offset":         offset,
		"total_sessions": response.Total,
	})

	// Возвращаем результат
	c.JSON(http.StatusOK, response)
}

// cashierStartSession обработчик для запуска сессии кассиром
func (h *Handler) cashierStartSession(c *gin.Context) {
	var req models.CashierStartSessionRequest

	// Парсим JSON из тела запроса
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Логируем мета-параметр для поиска
	log.Printf("Запрос на запуск сессии кассиром: SessionID=%s", req.SessionID)

	// Запускаем сессию
	session, err := h.service.CashierStartSession(&req)
	if err != nil {
		log.Printf("Ошибка запуска сессии кассиром: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	log.Printf("Успешно запущена сессия кассиром: SessionID=%s", req.SessionID)
	c.JSON(http.StatusOK, models.CashierStartSessionResponse{Session: *session})
}

// cashierCompleteSession обработчик для завершения сессии кассиром
func (h *Handler) cashierCompleteSession(c *gin.Context) {
	var req models.CashierCompleteSessionRequest

	// Парсим JSON из тела запроса
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Логируем мета-параметр для поиска
	log.Printf("Запрос на завершение сессии кассиром: SessionID=%s", req.SessionID)

	// Завершаем сессию
	session, err := h.service.CashierCompleteSession(&req)
	if err != nil {
		log.Printf("Ошибка завершения сессии кассиром: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	log.Printf("Успешно завершена сессия кассиром: SessionID=%s", req.SessionID)
	c.JSON(http.StatusOK, models.CashierCompleteSessionResponse{Session: *session})
}

// cashierCancelSession обработчик для отмены сессии кассиром
func (h *Handler) cashierCancelSession(c *gin.Context) {
	var req models.CashierCancelSessionRequest

	// Парсим JSON из тела запроса
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Логируем мета-параметр для поиска
	log.Printf("Запрос на отмену сессии кассиром: SessionID=%s", req.SessionID)

	// Отменяем сессию
	session, err := h.service.CashierCancelSession(&req)
	if err != nil {
		log.Printf("Ошибка отмены сессии кассиром: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	log.Printf("Успешно отменена сессия кассиром: SessionID=%s", req.SessionID)
	c.JSON(http.StatusOK, models.CashierCancelSessionResponse{Session: *session})
}

// cashierMiddleware middleware для проверки авторизации кассира
func (h *Handler) cashierMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		// Получаем токен из заголовка
		token := c.GetHeader("Authorization")
		token = strings.TrimPrefix(token, "Bearer ")

		if token == "" {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Токен не предоставлен"})
			c.Abort()
			return
		}

		// Проверяем токен через auth service
		claims, err := h.authService.ValidateToken(token)
		if err != nil {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Недействительный токен"})
			c.Abort()
			return
		}

		// Проверяем, что это кассир, а не администратор
		if claims.IsAdmin {
			c.JSON(http.StatusForbidden, gin.H{"error": "Доступ запрещен"})
			c.Abort()
			return
		}

		// Устанавливаем ID кассира в контекст
		c.Set("cashier_id", claims.ID)
		c.Next()
	}
}

// enableChemistry обработчик для включения химии
func (h *Handler) enableChemistry(c *gin.Context) {
	var req models.EnableChemistryRequest

	// Парсим JSON из тела запроса
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Логируем мета-параметр для поиска
	log.Printf("Запрос на включение химии: SessionID=%s", req.SessionID)

	// Включаем химию
	response, err := h.service.EnableChemistry(&req)
	if err != nil {
		log.Printf("Ошибка включения химии: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	log.Printf("Успешно включена химия: SessionID=%s", req.SessionID)
	c.JSON(http.StatusOK, response)
}

// cashierEnableChemistry обработчик для включения химии кассиром
func (h *Handler) cashierEnableChemistry(c *gin.Context) {
	var req models.EnableChemistryRequest

	// Парсим JSON из тела запроса
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Логируем мета-параметр для поиска
	log.Printf("Запрос на включение химии кассиром: SessionID=%s", req.SessionID)

	// Включаем химию
	response, err := h.service.EnableChemistry(&req)
	if err != nil {
		log.Printf("Ошибка включения химии кассиром: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	log.Printf("Успешно включена химия кассиром: SessionID=%s", req.SessionID)
	c.JSON(http.StatusOK, response)
}

// getChemistryStats обработчик для получения статистики химии
func (h *Handler) getChemistryStats(c *gin.Context) {
	// Парсим параметры запроса
	dateFromStr := c.Query("date_from")
	dateToStr := c.Query("date_to")

	var dateFrom, dateTo *time.Time

	// Парсим дату начала периода
	if dateFromStr != "" {
		parsedDateFrom, err := time.Parse("2006-01-02", dateFromStr)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Неверный формат даты начала периода"})
			return
		}
		dateFrom = &parsedDateFrom
	}

	// Парсим дату окончания периода
	if dateToStr != "" {
		parsedDateTo, err := time.Parse("2006-01-02", dateToStr)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Неверный формат даты окончания периода"})
			return
		}
		// Устанавливаем время на конец дня
		parsedDateTo = parsedDateTo.Add(23*time.Hour + 59*time.Minute + 59*time.Second)
		dateTo = &parsedDateTo
	}

	req := &models.GetChemistryStatsRequest{
		DateFrom: dateFrom,
		DateTo:   dateTo,
	}

	// Логируем запрос
	log.Printf("Запрос статистики химии: DateFrom=%v, DateTo=%v", dateFrom, dateTo)

	// Получаем статистику
	response, err := h.service.GetChemistryStats(req)
	if err != nil {
		log.Printf("Ошибка получения статистики химии: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	log.Printf("Успешно получена статистика химии")
	c.JSON(http.StatusOK, response)
}
