package handlers

import (
	"bytes"
	"carwash_backend/internal/logger"
	"encoding/json"
	"io"
	"net/http"
	"strconv"
	"strings"
	"time"

	authService "carwash_backend/internal/domain/auth/service"
	paymentService "carwash_backend/internal/domain/payment/service"
	"carwash_backend/internal/domain/session/middleware"
	"carwash_backend/internal/domain/session/models"
	"carwash_backend/internal/domain/session/service"

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
		sessionRoutes.POST("/with-payment", h.createSessionWithPayment)        // создание сессии с платежом
		sessionRoutes.GET("", h.getUserSession)                                // user_id в query параметре
		sessionRoutes.GET("/for-payment", h.getUserSessionForPayment)          // user_id в query параметре, включает payment_failed
		sessionRoutes.GET("/check-active", h.checkActiveSession)               // user_id в query параметре, проверка активной сессии
		sessionRoutes.GET("/by-id", h.getSessionByID)                          // session_id в query параметре
		sessionRoutes.POST("/start", h.startSession)                           // session_id в теле запроса
		sessionRoutes.POST("/complete", h.completeSession)                     // session_id в теле запроса
		sessionRoutes.POST("/extend-with-payment", h.extendSessionWithPayment) // session_id и extension_time_minutes в теле запроса
		sessionRoutes.GET("/payments", h.getSessionPayments)                   // session_id в query параметре
		sessionRoutes.GET("/history", h.getUserSessionHistory)                 // user_id в query параметре
		sessionRoutes.POST("/cancel", h.cancelSession)                         // session_id и user_id в теле запроса
		sessionRoutes.POST("/enable-chemistry", h.enableChemistry)             // session_id и user_id в теле запроса
	}

	// Административные маршруты
	adminRoutes := router.Group("/admin/sessions")
	{
		adminRoutes.GET("", h.adminListSessions)
		adminRoutes.GET("/by-id", h.adminGetSession)
		adminRoutes.POST("/reassign", h.adminReassignSession) // переназначение сессии администратором
	}

	// Маршруты для кассира
	cashierRoutes := router.Group("/cashier/sessions", h.cashierMiddleware())
	{
		cashierRoutes.GET("/active", h.cashierGetActiveSessions)
		cashierRoutes.POST("/start", h.cashierStartSession)
		cashierRoutes.POST("/complete", h.cashierCompleteSession)
		cashierRoutes.POST("/cancel", h.cashierCancelSession)
		cashierRoutes.POST("/enable-chemistry", h.cashierEnableChemistry) // включение химии кассиром
		cashierRoutes.POST("/reassign", h.cashierReassignSession)         // переназначение сессии кассиром
	}

	// 1C webhook маршруты
	oneCRoutes := router.Group("/1c")
	{
		oneCRoutes.POST("/payment-callback", middleware.Auth1CMiddleware(h.apiKey1C), h.handle1CPaymentCallback)
	}
}

// createSessionWithPayment обработчик для создания сессии с платежом
func (h *Handler) createSessionWithPayment(c *gin.Context) {
	var req models.CreateSessionWithPaymentRequest

	// Парсим JSON из тела запроса
	if err := c.ShouldBindJSON(&req); err != nil {
		logger.WithContext(c).Errorf("API Error - createSessionWithPayment: ошибка парсинга JSON, error: %v", err)
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Создаем сессию с платежом
	response, err := h.service.CreateSessionWithPayment(c.Request.Context(), &req)
	if err != nil {
		logger.WithContext(c).Errorf("API Error - createSessionWithPayment: ошибка создания сессии с платежом, user_id: %s, service_type: %s, error: %v", req.UserID.String(), req.ServiceType, err)
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
		logger.WithContext(c).Errorf("API Error - getUserSession: не указан ID пользователя")
		c.JSON(http.StatusBadRequest, gin.H{"error": "Не указан ID пользователя"})
		return
	}

	userID, err := uuid.Parse(userIDStr)
	if err != nil {
		logger.WithContext(c).Errorf("API Error - getUserSession: некорректный ID пользователя '%s', error: %v", userIDStr, err)
		c.JSON(http.StatusBadRequest, gin.H{"error": "Некорректный ID пользователя"})
		return
	}

	// Получаем сессию пользователя
	response, err := h.service.GetUserSession(c.Request.Context(), &models.GetUserSessionRequest{
		UserID: userID,
	})
	if err != nil {
		logger.WithContext(c).Errorf("API Error - getUserSession: сессия не найдена для user_id: %s, error: %v", userID.String(), err)
		c.JSON(http.StatusNotFound, gin.H{"error": "Сессия не найдена"})
		return
	}

	// Возвращаем сессию пользователя
	c.JSON(http.StatusOK, response)
}

// checkActiveSession обработчик для проверки активной сессии пользователя
func (h *Handler) checkActiveSession(c *gin.Context) {
	// Получаем ID пользователя из query параметра
	userIDStr := c.Query("user_id")
	if userIDStr == "" {
		logger.WithContext(c).Errorf("API Error - checkActiveSession: не указан ID пользователя")
		c.JSON(http.StatusBadRequest, gin.H{"error": "Не указан ID пользователя"})
		return
	}

	userID, err := uuid.Parse(userIDStr)
	if err != nil {
		logger.WithContext(c).Errorf("API Error - checkActiveSession: некорректный ID пользователя '%s', error: %v", userIDStr, err)
		c.JSON(http.StatusBadRequest, gin.H{"error": "Некорректный ID пользователя"})
		return
	}

	// Проверяем активную сессию пользователя
	response, err := h.service.CheckActiveSession(c.Request.Context(), &models.CheckActiveSessionRequest{
		UserID: userID,
	})
	if err != nil {
		logger.WithContext(c).Errorf("API Error - checkActiveSession: ошибка проверки активной сессии для user_id: %s, error: %v", userID.String(), err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Ошибка проверки активной сессии"})
		return
	}

	// Возвращаем результат проверки
	c.JSON(http.StatusOK, response)
}

// getUserSessionForPayment обработчик для получения сессии пользователя для PaymentPage (включая payment_failed)
func (h *Handler) getUserSessionForPayment(c *gin.Context) {
	// Получаем ID пользователя из query параметра
	userIDStr := c.Query("user_id")
	if userIDStr == "" {
		logger.WithContext(c).Errorf("API Error - getUserSessionForPayment: не указан ID пользователя")
		c.JSON(http.StatusBadRequest, gin.H{"error": "Не указан ID пользователя"})
		return
	}

	userID, err := uuid.Parse(userIDStr)
	if err != nil {
		logger.WithContext(c).Errorf("API Error - getUserSessionForPayment: некорректный ID пользователя '%s', error: %v", userIDStr, err)
		c.JSON(http.StatusBadRequest, gin.H{"error": "Некорректный ID пользователя"})
		return
	}

	// Получаем сессию пользователя для PaymentPage
	response, err := h.service.GetUserSessionForPayment(c.Request.Context(), &models.GetUserSessionRequest{
		UserID: userID,
	})
	if err != nil {
		logger.WithContext(c).Errorf("API Error - getUserSessionForPayment: сессия не найдена для user_id: %s, error: %v", userID.String(), err)
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
	response, err := h.service.GetSession(c.Request.Context(), &models.GetSessionRequest{
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
	session, err := h.service.StartSession(c.Request.Context(), &req)
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
	response, err := h.service.CompleteSession(c.Request.Context(), &req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	// Возвращаем обновленную сессию с информацией о платеже
	c.JSON(http.StatusOK, response)
}

// extendSessionWithPayment обработчик для продления сессии с оплатой
func (h *Handler) extendSessionWithPayment(c *gin.Context) {
	var req models.ExtendSessionWithPaymentRequest

	// Парсим JSON из тела запроса
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Валидация: хотя бы одно из полей должно быть больше 0
	if req.ExtensionTimeMinutes <= 0 && req.ExtensionChemistryTimeMinutes <= 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "должно быть указано время продления или время химии"})
		return
	}

	// Логируем мета-параметр для поиска
	logger.WithContext(c).Infof("Запрос на продление сессии с оплатой: SessionID=%s, ExtensionTime=%d, ExtensionChemistryTime=%d", req.SessionID, req.ExtensionTimeMinutes, req.ExtensionChemistryTimeMinutes)

	// Продлеваем сессию с оплатой
	response, err := h.service.ExtendSessionWithPayment(c.Request.Context(), &req)
	if err != nil {
		logger.WithContext(c).Errorf("Ошибка продления сессии с оплатой: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	logger.WithContext(c).Infof("Успешно создан платеж продления: SessionID=%s, ExtensionTime=%d", req.SessionID, req.ExtensionTimeMinutes)
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
	logger.WithContext(c).Infof("Запрос на получение платежей сессии: SessionID=%s", sessionID)

	// Получаем платежи сессии
	response, err := h.service.GetSessionPayments(c.Request.Context(), &models.GetSessionPaymentsRequest{
		SessionID: sessionID,
	})
	if err != nil {
		logger.WithContext(c).Errorf("Ошибка получения платежей сессии: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	logger.WithContext(c).Infof("Успешно получены платежи сессии: SessionID=%s", sessionID)
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
	logger.WithContext(c).Infof("Запрос на отмену сессии: SessionID=%s, UserID=%s", req.SessionID, req.UserID)

	// Отменяем сессию
	response, err := h.service.CancelSession(c.Request.Context(), &req)
	if err != nil {
		logger.WithContext(c).Errorf("Ошибка отмены сессии: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	logger.WithContext(c).Infof("Успешно отменена сессия: SessionID=%s, UserID=%s", req.SessionID, req.UserID)
	c.JSON(http.StatusOK, response)
}

// getUserSessionHistory обработчик для получения истории сессий пользователя
func (h *Handler) getUserSessionHistory(c *gin.Context) {
	// Получаем user_id из query параметра
	userIDStr := c.Query("user_id")
	if userIDStr == "" {
		logger.WithContext(c).Errorf("API Error - getUserSessionHistory: не указан User ID")
		c.JSON(http.StatusBadRequest, gin.H{"error": "Не указан User ID"})
		return
	}

	// Преобразуем строку в uuid.UUID
	userID, err := uuid.Parse(userIDStr)
	if err != nil {
		logger.WithContext(c).Errorf("API Error - getUserSessionHistory: некорректный User ID '%s', error: %v", userIDStr, err)
		c.JSON(http.StatusBadRequest, gin.H{"error": "Некорректный User ID"})
		return
	}

	// Получаем параметры пагинации
	limit := 5
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
	sessions, err := h.service.GetUserSessionHistory(c.Request.Context(), &req)
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

	// Дата от (поддерживаем ISO 8601 с timezone)
	if dateFromStr := c.Query("date_from"); dateFromStr != "" {
		dateFrom, err := time.Parse(time.RFC3339, dateFromStr)
		if err != nil {
			// Пробуем парсить как простую дату для обратной совместимости
			dateFrom, err = time.Parse("2006-01-02", dateFromStr)
		}
		if err == nil {
			req.DateFrom = &dateFrom
		}
	}

	// Дата до (поддерживаем ISO 8601 с timezone)
	if dateToStr := c.Query("date_to"); dateToStr != "" {
		dateTo, err := time.Parse(time.RFC3339, dateToStr)
		if err != nil {
			// Пробуем парсить как простую дату для обратной совместимости
			dateTo, err = time.Parse("2006-01-02", dateToStr)
		}
		if err == nil {
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
	resp, err := h.service.AdminListSessions(c.Request.Context(), &req)
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
	resp, err := h.service.AdminGetSession(c.Request.Context(), &req)
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
		logger.WithContext(c).Infof("Error reading request body: %v", err)
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request body"})
		return
	}

	// Логируем тело запроса
	logger.WithContext(c).Infof("Raw request body: %s", string(bodyBytes))

	bodyBytes = bytes.TrimPrefix(bodyBytes, []byte("\xef\xbb\xbf"))

	// ВАЖНО: Восстанавливаем body для дальнейшего использования
	c.Request.Body = io.NopCloser(bytes.NewBuffer(bodyBytes))

	var req models.CashierPaymentRequest

	// Парсим JSON из тела запроса
	if err := json.Unmarshal(bodyBytes, &req); err != nil {
		logger.WithContext(c).Infof("Error parsing 1C payment request: %v", err)
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Логируем входящий запрос
	logger.WithContext(c).Infof("Received 1C payment callback: ServiceType=%s, WithChemistry=%t, Amount=%d, RentalTimeMinutes=%d, CarNumber=%s, PaymentTime=%s",
		req.ServiceType, req.WithChemistry, req.Amount, req.RentalTimeMinutes, req.CarNumber, req.PaymentTime.Format(time.RFC3339))

	// Создаем сессию через кассира
	session, err := h.service.CreateFromCashier(c.Request.Context(), &req)
	if err != nil {
		logger.WithContext(c).Errorf("Error creating session from cashier: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	// Логируем результат создания сессии (новая или существующая)
	if req.CarNumber != "" {
		logger.WithContext(c).Infof("Session processed for car '%s': SessionID=%s, Status=%s, CreatedAt=%s",
			req.CarNumber, session.ID.String(), session.Status, session.CreatedAt.Format(time.RFC3339))
	}

	// Создаем платеж для кассира
	payment, err := h.paymentService.CreateForCashier(c.Request.Context(), session.ID, req.Amount)
	if err != nil {
		logger.WithContext(c).Infof("Error creating payment for cashier: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	// Логируем успешное создание
	logger.WithContext(c).Infof("Successfully processed 1C payment: SessionID=%s, PaymentID=%s, Amount=%d",
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
	response, err := h.service.CashierGetActiveSessions(c.Request.Context(), req)
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
	logger.WithContext(c).Infof("Запрос на запуск сессии кассиром: SessionID=%s", req.SessionID)

	// Запускаем сессию
	session, err := h.service.CashierStartSession(c.Request.Context(), &req)
	if err != nil {
		logger.WithContext(c).Errorf("Ошибка запуска сессии кассиром: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	logger.WithContext(c).Infof("Успешно запущена сессия кассиром: SessionID=%s", req.SessionID)
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
	logger.WithContext(c).Infof("Запрос на завершение сессии кассиром: SessionID=%s", req.SessionID)

	// Завершаем сессию
	session, err := h.service.CashierCompleteSession(c.Request.Context(), &req)
	if err != nil {
		logger.WithContext(c).Errorf("Ошибка завершения сессии кассиром: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	logger.WithContext(c).Infof("Успешно завершена сессия кассиром: SessionID=%s", req.SessionID)
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
	logger.WithContext(c).Infof("Запрос на отмену сессии кассиром: SessionID=%s", req.SessionID)

	// Отменяем сессию
	session, err := h.service.CashierCancelSession(c.Request.Context(), &req)
	if err != nil {
		logger.WithContext(c).Errorf("Ошибка отмены сессии кассиром: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	logger.WithContext(c).Infof("Успешно отменена сессия кассиром: SessionID=%s", req.SessionID)
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
		claims, err := h.authService.ValidateToken(c.Request.Context(), token)
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
	logger.WithContext(c).Infof("Запрос на включение химии: SessionID=%s", req.SessionID)

	// Включаем химию
	response, err := h.service.EnableChemistry(c.Request.Context(), &req)
	if err != nil {
		logger.WithContext(c).Errorf("Ошибка включения химии: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	logger.WithContext(c).Infof("Успешно включена химия: SessionID=%s", req.SessionID)
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
	logger.WithContext(c).Infof("Запрос на включение химии кассиром: SessionID=%s", req.SessionID)

	// Включаем химию
	response, err := h.service.EnableChemistry(c.Request.Context(), &req)
	if err != nil {
		logger.WithContext(c).Errorf("Ошибка включения химии кассиром: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	logger.WithContext(c).Infof("Успешно включена химия кассиром: SessionID=%s", req.SessionID)
	c.JSON(http.StatusOK, response)
}

// adminReassignSession обработчик для переназначения сессии администратором
func (h *Handler) adminReassignSession(c *gin.Context) {
	var req models.ReassignSessionRequest

	// Парсим JSON из тела запроса
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Логируем мета-параметр для поиска
	logger.WithContext(c).Infof("Запрос на переназначение сессии администратором: SessionID=%s", req.SessionID)

	// Переназначаем сессию
	response, err := h.service.ReassignSession(c.Request.Context(), &req)
	if err != nil {
		logger.WithContext(c).Errorf("Ошибка переназначения сессии администратором: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	logger.WithContext(c).Infof("Успешно переназначена сессия администратором: SessionID=%s", req.SessionID)
	c.JSON(http.StatusOK, response)
}

// cashierReassignSession обработчик для переназначения сессии кассиром
func (h *Handler) cashierReassignSession(c *gin.Context) {
	var req models.ReassignSessionRequest

	// Парсим JSON из тела запроса
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Логируем мета-параметр для поиска
	logger.WithContext(c).Infof("Запрос на переназначение сессии кассиром: SessionID=%s", req.SessionID)

	// Переназначаем сессию
	response, err := h.service.ReassignSession(c.Request.Context(), &req)
	if err != nil {
		logger.WithContext(c).Errorf("Ошибка переназначения сессии кассиром: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	logger.WithContext(c).Infof("Успешно переназначена сессия кассиром: SessionID=%s", req.SessionID)
	c.JSON(http.StatusOK, response)
}
