package handlers

import (
	"net/http"
	"strconv"
	"time"
	"log"

	"carwash_backend/internal/domain/payment/models"
	"carwash_backend/internal/domain/payment/service"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

// Handler структура для обработчиков HTTP запросов платежей
type Handler struct {
	service service.Service
}

// NewHandler создает новый экземпляр Handler
func NewHandler(service service.Service) *Handler {
	return &Handler{
		service: service,
	}
}

// RegisterRoutes регистрирует публичные и интеграционные маршруты для платежей
func (h *Handler) RegisterRoutes(router *gin.RouterGroup) {
	paymentRoutes := router.Group("/payments")
	{
		// Основные маршруты платежей
		paymentRoutes.POST("", h.createPayment)
		paymentRoutes.GET("/by-id", h.getPayment)
		paymentRoutes.GET("/status", h.getPaymentStatus)
		paymentRoutes.POST("/webhook", h.handleWebhook)

		// Маршруты возвратов
		paymentRoutes.POST("/refunds", h.createRefund)
		paymentRoutes.POST("/refunds/automatic", h.processAutomaticRefund)
		paymentRoutes.POST("/refunds/full", h.processFullRefund)

		// Маршруты для интеграции с другими доменами
		paymentRoutes.POST("/queue", h.createQueuePayment)
		paymentRoutes.POST("/session-extension", h.createSessionExtensionPayment)
		paymentRoutes.GET("/session", h.checkPaymentForSession)

		// Маршруты расчёта стоимости
		paymentRoutes.POST("/calculate-price", h.calculatePrice)
	}
}

// RegisterAdminRoutes регистрирует админские маршруты для платежей
func (h *Handler) RegisterAdminRoutes(router *gin.RouterGroup) {
	adminRoutes := router.Group("/payments", h.adminMiddleware())
	{
		adminRoutes.GET("", h.adminListPayments)
		adminRoutes.GET("/by-id", h.adminGetPayment)
		adminRoutes.POST("/reconcile", h.reconcilePayments)
	}
}

// createPayment обработчик для создания платежа
func (h *Handler) createPayment(c *gin.Context) {
	var req models.CreatePaymentRequest

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	resp, err := h.service.CreatePayment(&req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, resp)
}

// getPayment обработчик для получения платежа
func (h *Handler) getPayment(c *gin.Context) {
	paymentIDStr := c.Query("id")
	if paymentIDStr == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "id обязателен"})
		return
	}

	paymentID, err := uuid.Parse(paymentIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "неверный формат id"})
		return
	}

	req := &models.GetPaymentRequest{ID: paymentID}
	resp, err := h.service.GetPayment(req)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, resp)
}

// getPaymentStatus обработчик для получения статуса платежа
func (h *Handler) getPaymentStatus(c *gin.Context) {
	paymentIDStr := c.Query("id")
	if paymentIDStr == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "id обязателен"})
		return
	}

	paymentID, err := uuid.Parse(paymentIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "неверный формат id"})
		return
	}

	req := &models.GetPaymentStatusRequest{ID: paymentID}
	resp, err := h.service.GetPaymentStatus(req)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, resp)
}

// handleWebhook обработчик для webhook'ов от Tinkoff
func (h *Handler) handleWebhook(c *gin.Context) {
	var webhook models.TinkoffWebhook

	if err := c.ShouldBindJSON(&webhook); err != nil {
		log.Println("Error binding JSON:", err)
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Логируем webhook для отладки
	c.Set("meta", map[string]interface{}{
		"payment_id": webhook.PaymentId,
		"status":     webhook.Status,
		"amount":     webhook.Amount,
	})

	if err := h.service.HandleTinkoffWebhook(&webhook); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"status": "ok"})
}

// createRefund обработчик для создания возврата
func (h *Handler) createRefund(c *gin.Context) {
	var req models.CreateRefundRequest

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	resp, err := h.service.CreateRefund(&req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, resp)
}

// processAutomaticRefund обработчик для автоматического возврата
func (h *Handler) processAutomaticRefund(c *gin.Context) {
	sessionIDStr := c.Query("session_id")
	if sessionIDStr == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "session_id обязателен"})
		return
	}

	sessionID, err := uuid.Parse(sessionIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "неверный формат session_id"})
		return
	}

	// Логируем операцию
	c.Set("meta", map[string]interface{}{
		"session_id": sessionID,
		"operation":  "automatic_refund",
	})

	if err := h.service.ProcessAutomaticRefund(sessionID); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"status": "ok"})
}

// processFullRefund обработчик для полного возврата
func (h *Handler) processFullRefund(c *gin.Context) {
	paymentIDStr := c.Query("payment_id")
	if paymentIDStr == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "payment_id обязателен"})
		return
	}

	paymentID, err := uuid.Parse(paymentIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "неверный формат payment_id"})
		return
	}

	// Логируем операцию
	c.Set("meta", map[string]interface{}{
		"payment_id": paymentID,
		"operation":  "full_refund",
	})

	if err := h.service.ProcessFullRefund(paymentID); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"status": "ok"})
}

// createQueuePayment обработчик для создания платежа за очередь
func (h *Handler) createQueuePayment(c *gin.Context) {
	userIDStr := c.Query("user_id")
	serviceType := c.Query("service_type")
	rentalTimeMinutesStr := c.Query("rental_time_minutes")
	withChemistryStr := c.Query("with_chemistry")

	if userIDStr == "" || serviceType == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "user_id и service_type обязательны"})
		return
	}

	userID, err := uuid.Parse(userIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "неверный формат user_id"})
		return
	}

	// Парсим rental_time_minutes (обязательный параметр)
	rentalTimeMinutes := 5 // значение по умолчанию
	if rentalTimeMinutesStr != "" {
		rentalTimeMinutes, err = strconv.Atoi(rentalTimeMinutesStr)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "неверный формат rental_time_minutes"})
			return
		}
	}

	// Парсим with_chemistry (опциональный параметр)
	withChemistry := false
	if withChemistryStr != "" {
		withChemistry, err = strconv.ParseBool(withChemistryStr)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "неверный формат with_chemistry"})
			return
		}
	}

	// Логируем операцию
	c.Set("meta", map[string]interface{}{
		"user_id":             userID,
		"service_type":        serviceType,
		"rental_time_minutes": rentalTimeMinutes,
		"with_chemistry":      withChemistry,
		"operation":           "queue_payment",
	})

	payment, err := h.service.CreateQueuePayment(userID, serviceType, rentalTimeMinutes, withChemistry)
	if err != nil {
		log.Printf("Ошибка создания платежа за очередь: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	log.Printf("Платеж за очередь успешно создан: ID=%s, Amount=%d, URL=%s", 
		payment.ID, payment.AmountKopecks, payment.PaymentURL)

	c.JSON(http.StatusOK, gin.H{
		"payment":     payment,
		"payment_url": payment.PaymentURL,
	})
}

// createSessionExtensionPayment обработчик для создания платежа за продление сессии
func (h *Handler) createSessionExtensionPayment(c *gin.Context) {
	sessionIDStr := c.Query("session_id")
	extensionMinutesStr := c.Query("extension_minutes")

	if sessionIDStr == "" || extensionMinutesStr == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "session_id и extension_minutes обязательны"})
		return
	}

	sessionID, err := uuid.Parse(sessionIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "неверный формат session_id"})
		return
	}

	extensionMinutes, err := strconv.Atoi(extensionMinutesStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "неверный формат extension_minutes"})
		return
	}

	// Логируем операцию
	c.Set("meta", map[string]interface{}{
		"session_id":        sessionID,
		"extension_minutes": extensionMinutes,
		"operation":         "session_extension_payment",
	})

	payment, err := h.service.CreateSessionExtensionPayment(sessionID, extensionMinutes)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, gin.H{
		"payment":     payment,
		"payment_url": payment.PaymentURL,
	})
}

// checkPaymentForSession обработчик для проверки платежа сессии
func (h *Handler) checkPaymentForSession(c *gin.Context) {
	sessionIDStr := c.Query("session_id")
	if sessionIDStr == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "session_id обязателен"})
		return
	}

	sessionID, err := uuid.Parse(sessionIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "неверный формат session_id"})
		return
	}

	payment, err := h.service.CheckPaymentForSession(sessionID)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"payment": payment})
}

// calculatePrice обработчик для расчёта стоимости
func (h *Handler) calculatePrice(c *gin.Context) {
	var req models.CalculatePriceRequest

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Логируем операцию
	c.Set("meta", map[string]interface{}{
		"service_type":        req.ServiceType,
		"rental_time_minutes": req.RentalTimeMinutes,
		"with_chemistry":      req.WithChemistry,
		"operation":           "calculate_price",
	})

	resp, err := h.service.CalculatePrice(&req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, resp)
}

// adminListPayments обработчик для получения списка платежей (админ)
func (h *Handler) adminListPayments(c *gin.Context) {
	limitStr := c.Query("limit")
	offsetStr := c.Query("offset")
	userIDStr := c.Query("user_id")
	status := c.Query("status")
	paymentType := c.Query("type")
	dateFromStr := c.Query("date_from")
	dateToStr := c.Query("date_to")

	if limitStr == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "limit обязателен"})
		return
	}

	limit, err := strconv.Atoi(limitStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "неверный формат limit"})
		return
	}

	offset := 0
	if offsetStr != "" {
		offset, err = strconv.Atoi(offsetStr)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "неверный формат offset"})
			return
		}
	}

	var userID *uuid.UUID
	if userIDStr != "" {
		parsedUserID, err := uuid.Parse(userIDStr)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "неверный формат user_id"})
			return
		}
		userID = &parsedUserID
	}

	var dateFrom, dateTo *time.Time
	if dateFromStr != "" {
		parsedDateFrom, err := time.Parse("2006-01-02", dateFromStr)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "неверный формат date_from"})
			return
		}
		dateFrom = &parsedDateFrom
	}

	if dateToStr != "" {
		parsedDateTo, err := time.Parse("2006-01-02", dateToStr)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "неверный формат date_to"})
			return
		}
		dateTo = &parsedDateTo
	}

	req := &models.AdminListPaymentsRequest{
		UserID:    userID,
		Status:    &status,
		Type:      &paymentType,
		DateFrom:  dateFrom,
		DateTo:    dateTo,
		Limit:     limit,
		Offset:    offset,
	}

	resp, err := h.service.AdminListPayments(req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, resp)
}

// adminGetPayment обработчик для получения платежа (админ)
func (h *Handler) adminGetPayment(c *gin.Context) {
	paymentIDStr := c.Query("id")
	if paymentIDStr == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "id обязателен"})
		return
	}

	paymentID, err := uuid.Parse(paymentIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "неверный формат id"})
		return
	}

	req := &models.AdminGetPaymentRequest{ID: paymentID}
	resp, err := h.service.AdminGetPayment(req)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, resp)
}

// reconcilePayments обработчик для сверки платежей (админ)
func (h *Handler) reconcilePayments(c *gin.Context) {
	dateStr := c.Query("date")
	if dateStr == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "date обязателен"})
		return
	}

	date, err := time.Parse("2006-01-02", dateStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "неверный формат date"})
		return
	}

	// Сверяем за один день
	startDate := date
	endDate := date.Add(24 * time.Hour)

	report, err := h.service.ReconcilePayments(startDate, endDate)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, report)
}

// adminMiddleware middleware для проверки прав администратора
func (h *Handler) adminMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		// Здесь должна быть проверка JWT токена и прав администратора
		// Пока что просто пропускаем
		c.Next()
	}
} 