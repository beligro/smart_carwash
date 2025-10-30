package handlers

import (
	authService "carwash_backend/internal/domain/auth/service"
	"carwash_backend/internal/domain/payment/models"
	"carwash_backend/internal/domain/payment/service"
	"carwash_backend/internal/logger"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

// Handler структура для обработчиков HTTP запросов платежей
type Handler struct {
	service     service.Service
	authService authService.Service
}

// NewHandler создает новый экземпляр Handler
func NewHandler(service service.Service, authService authService.Service) *Handler {
	return &Handler{
		service:     service,
		authService: authService,
	}
}

// RegisterRoutes регистрирует маршруты для платежей
func (h *Handler) RegisterRoutes(router *gin.RouterGroup) {
	paymentRoutes := router.Group("/payments")
	{
		paymentRoutes.POST("/calculate-price", h.calculatePrice)
		paymentRoutes.POST("/calculate-extension-price", h.calculateExtensionPrice)
		paymentRoutes.POST("/create", h.createPayment)
		paymentRoutes.GET("/status", h.getPaymentStatus)
		paymentRoutes.POST("/webhook", h.handleWebhook)
	}

	// Административные маршруты
	adminRoutes := router.Group("/admin/payments")
	{
		adminRoutes.GET("", h.adminListPayments)
		adminRoutes.GET("/statistics", h.adminGetPaymentStatistics)
		adminRoutes.POST("/refund", h.adminRefundPayment)
	}

	// Маршруты для кассира
	cashierRoutes := router.Group("/cashier/payments", h.cashierMiddleware())
	{
		cashierRoutes.GET("", h.cashierListPayments)
	}

	// Маршруты для статистики кассира
	cashierStatsRoutes := router.Group("/cashier/statistics", h.cashierMiddleware())
	{
		cashierStatsRoutes.GET("/last-shift", h.cashierGetLastShiftStatistics)
	}
}

// calculatePrice обработчик для расчета цены
func (h *Handler) calculatePrice(c *gin.Context) {
	var req models.CalculatePriceRequest

	if err := c.ShouldBindJSON(&req); err != nil {
		logger.WithContext(c).Errorf("API Error - calculatePrice: ошибка парсинга JSON, error: %v", err)
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	response, err := h.service.CalculatePrice(c.Request.Context(), &req)
	if err != nil {
		logger.WithContext(c).Errorf("API Error - calculatePrice: ошибка расчета цены, service_type: %s, rental_time: %d, with_chemistry: %t, error: %v", req.ServiceType, req.RentalTimeMinutes, req.WithChemistry, err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, response)
}

// calculateExtensionPrice обработчик для расчета цены продления
func (h *Handler) calculateExtensionPrice(c *gin.Context) {
	var req models.CalculateExtensionPriceRequest

	if err := c.ShouldBindJSON(&req); err != nil {
		logger.WithContext(c).Errorf("API Error - calculateExtensionPrice: ошибка парсинга JSON, error: %v", err)
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	response, err := h.service.CalculateExtensionPrice(c.Request.Context(), &req)
	if err != nil {
		logger.WithContext(c).Errorf("API Error - calculateExtensionPrice: ошибка расчета цены продления, service_type: %s, extension_minutes: %d, error: %v", req.ServiceType, req.ExtensionTimeMinutes, err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, response)
}

// createPayment обработчик для создания платежа
func (h *Handler) createPayment(c *gin.Context) {
	var req models.CreatePaymentRequest

	if err := c.ShouldBindJSON(&req); err != nil {
		logger.WithContext(c).Errorf("API Error - createPayment: ошибка парсинга JSON, error: %v", err)
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	response, err := h.service.CreatePayment(c.Request.Context(), &req)
	if err != nil {
		logger.WithContext(c).Errorf("API Error - createPayment: ошибка создания платежа, session_id: %s, amount: %d, error: %v", req.SessionID.String(), req.Amount, err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, response)
}

// getPaymentStatus обработчик для получения статуса платежа
func (h *Handler) getPaymentStatus(c *gin.Context) {
	paymentIDStr := c.Query("payment_id")
	if paymentIDStr == "" {
		logger.WithContext(c).Errorf("API Error - getPaymentStatus: не указан ID платежа")
		c.JSON(http.StatusBadRequest, gin.H{"error": "Не указан ID платежа"})
		return
	}

	paymentID, err := uuid.Parse(paymentIDStr)
	if err != nil {
		logger.WithContext(c).Errorf("API Error - getPaymentStatus: некорректный ID платежа '%s', error: %v", paymentIDStr, err)
		c.JSON(http.StatusBadRequest, gin.H{"error": "Некорректный ID платежа"})
		return
	}

	response, err := h.service.GetPaymentStatus(c.Request.Context(), &models.GetPaymentStatusRequest{
		PaymentID: paymentID,
	})
	if err != nil {
		logger.WithContext(c).Errorf("API Error - getPaymentStatus: платеж не найден, payment_id: %s, error: %v", paymentID.String(), err)
		c.JSON(http.StatusNotFound, gin.H{"error": "Платеж не найден"})
		return
	}

	c.JSON(http.StatusOK, response)
}

// handleWebhook обработчик для webhook от Tinkoff
func (h *Handler) handleWebhook(c *gin.Context) {
	// Парсим webhook
	var webhookReq models.WebhookRequest
	if err := c.ShouldBindJSON(&webhookReq); err != nil {
		logger.WithContext(c).Errorf("API Error - handleWebhook: ошибка привязки JSON, error: %v", err)
		c.JSON(http.StatusBadRequest, gin.H{"error": "Неверный формат webhook"})
		return
	}

	// Обрабатываем webhook
	if err := h.service.HandleWebhook(c.Request.Context(), &webhookReq); err != nil {
		logger.WithContext(c).Errorf("API Error - handleWebhook: ошибка обработки webhook, terminal_key: %s, order_id: %s, error: %v", webhookReq.TerminalKey, webhookReq.OrderId, err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	// Отвечаем успехом Tinkoff
	c.JSON(http.StatusOK, gin.H{"success": true})
}

// adminListPayments обработчик для получения списка платежей (админка)
func (h *Handler) adminListPayments(c *gin.Context) {
	var req models.AdminListPaymentsRequest

	// Парсим query параметры
	if paymentIDStr := c.Query("payment_id"); paymentIDStr != "" {
		if paymentID, err := uuid.Parse(paymentIDStr); err == nil {
			req.PaymentID = &paymentID
		}
	}

	if sessionIDStr := c.Query("session_id"); sessionIDStr != "" {
		if sessionID, err := uuid.Parse(sessionIDStr); err == nil {
			req.SessionID = &sessionID
		}
	}

	if userIDStr := c.Query("user_id"); userIDStr != "" {
		if userID, err := uuid.Parse(userIDStr); err == nil {
			req.UserID = &userID
		}
	}

	if status := c.Query("status"); status != "" {
		req.Status = &status
	}

	if paymentType := c.Query("payment_type"); paymentType != "" {
		req.PaymentType = &paymentType
	}

	if paymentMethod := c.Query("payment_method"); paymentMethod != "" {
		req.PaymentMethod = &paymentMethod
	}

	if dateFromStr := c.Query("date_from"); dateFromStr != "" {
		dateFrom, err := time.Parse(time.RFC3339, dateFromStr)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "invalid date_from format, expected ISO 8601"})
			return
		}
		req.DateFrom = &dateFrom
	}

	if dateToStr := c.Query("date_to"); dateToStr != "" {
		dateTo, err := time.Parse(time.RFC3339, dateToStr)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "invalid date_to format, expected ISO 8601"})
			return
		}
		req.DateTo = &dateTo
	}

	if limitStr := c.Query("limit"); limitStr != "" {
		if limit, err := strconv.Atoi(limitStr); err == nil {
			req.Limit = &limit
		}
	}

	if offsetStr := c.Query("offset"); offsetStr != "" {
		if offset, err := strconv.Atoi(offsetStr); err == nil {
			req.Offset = &offset
		}
	}

	// Логируем запрос с мета-параметрами
	logger.WithContext(c).Infof("Запрос списка платежей (админка): PaymentID=%v, SessionID=%v, UserID=%v, Status=%v, PaymentType=%v, PaymentMethod=%v, DateFrom=%v, DateTo=%v, Limit=%v, Offset=%v",
		req.PaymentID, req.SessionID, req.UserID, req.Status, req.PaymentType, req.PaymentMethod, req.DateFrom, req.DateTo, req.Limit, req.Offset)

	response, err := h.service.ListPayments(c.Request.Context(), &req)
	if err != nil {
		logger.WithContext(c).Errorf("Ошибка получения списка платежей: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	logger.WithContext(c).Infof("Получен список платежей: Total=%d, Limit=%d, Offset=%d",
		response.Total, response.Limit, response.Offset)

	c.JSON(http.StatusOK, response)
}

// adminRefundPayment обработчик для возврата платежа (админка)
func (h *Handler) adminRefundPayment(c *gin.Context) {
	var req models.AdminRefundPaymentRequest

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Логируем запрос с мета-параметрами
	logger.WithContext(c).Infof("Запрос на возврат платежа (админка): PaymentID=%s, Amount=%d",
		req.PaymentID, req.Amount)

	response, err := h.service.RefundPayment(c.Request.Context(), &models.RefundPaymentRequest{
		PaymentID: req.PaymentID,
		Amount:    req.Amount,
	})
	if err != nil {
		logger.WithContext(c).Errorf("Ошибка возврата платежа: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	adminResponse := &models.AdminRefundPaymentResponse{
		Payment: response.Payment,
		Refund:  response.Refund,
		Success: true,
		Message: "Возврат выполнен успешно",
	}

	logger.WithContext(c).Infof("Успешно выполнен возврат платежа: PaymentID=%s, Amount=%d",
		req.PaymentID, req.Amount)

	c.JSON(http.StatusOK, adminResponse)
}

// adminGetPaymentStatistics обработчик для получения статистики платежей
func (h *Handler) adminGetPaymentStatistics(c *gin.Context) {
	var req models.PaymentStatisticsRequest

	// Парсим query параметры
	if userIDStr := c.Query("user_id"); userIDStr != "" {
		userID, err := uuid.Parse(userIDStr)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "invalid user_id format"})
			return
		}
		req.UserID = &userID
	}

	if dateFromStr := c.Query("date_from"); dateFromStr != "" {
		dateFrom, err := time.Parse(time.RFC3339, dateFromStr)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "invalid date_from format, expected ISO 8601"})
			return
		}
		req.DateFrom = &dateFrom
	}

	if dateToStr := c.Query("date_to"); dateToStr != "" {
		dateTo, err := time.Parse(time.RFC3339, dateToStr)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "invalid date_to format, expected ISO 8601"})
			return
		}
		req.DateTo = &dateTo
	}

	if serviceType := c.Query("service_type"); serviceType != "" {
		req.ServiceType = &serviceType
	}

	logger.WithContext(c).Infof("Admin payment statistics request: user_id=%v, date_from=%v, date_to=%v, service_type=%v",
		req.UserID, req.DateFrom, req.DateTo, req.ServiceType)

	response, err := h.service.GetPaymentStatistics(c.Request.Context(), &req)
	if err != nil {
		logger.WithContext(c).Infof("Error getting payment statistics: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, response)
}

// cashierListPayments обработчик для получения списка платежей кассира
func (h *Handler) cashierListPayments(c *gin.Context) {
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
	req := &models.CashierPaymentsRequest{
		ShiftStartedAt: shiftStartedAt,
		Limit:          &limit,
		Offset:         &offset,
	}

	// Получаем список платежей
	response, err := h.service.CashierListPayments(c.Request.Context(), req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	// Логируем мета-параметры
	c.Set("meta", gin.H{
		"shift_started_at": shiftStartedAt,
		"limit":            limit,
		"offset":           offset,
		"total_payments":   response.Total,
	})

	// Возвращаем результат
	c.JSON(http.StatusOK, response)
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

// cashierGetLastShiftStatistics обработчик для получения статистики последней смены кассира
func (h *Handler) cashierGetLastShiftStatistics(c *gin.Context) {
	// Получаем ID кассира из контекста (установлен в middleware)
	cashierIDInterface, exists := c.Get("cashier_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Не авторизован"})
		return
	}

	cashierID, ok := cashierIDInterface.(uuid.UUID)
	if !ok {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Некорректный ID кассира в контексте"})
		return
	}

	// Создаем запрос
	req := &models.CashierLastShiftStatisticsRequest{
		CashierID: cashierID,
	}

	// Логируем запрос
	logger.WithContext(c).Infof("Cashier last shift statistics request: CashierID=%s", cashierID)

	// Получаем статистику
	response, err := h.service.GetCashierLastShiftStatistics(c.Request.Context(), req)
	if err != nil {
		logger.WithContext(c).Infof("Error getting cashier last shift statistics: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	// Логируем результат
	if response.HasShift {
		logger.WithContext(c).Infof("Successfully retrieved cashier last shift statistics: CashierID=%s, HasShift=%t, Message=%s",
			cashierID, response.HasShift, response.Message)
	} else {
		logger.WithContext(c).Infof("No completed shifts found for cashier: CashierID=%s, Message=%s",
			cashierID, response.Message)
	}

	// Возвращаем результат
	c.JSON(http.StatusOK, response)
}
