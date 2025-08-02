package handlers

import (
	"carwash_backend/internal/domain/payment/models"
	"carwash_backend/internal/domain/payment/service"
	"net/http"
	"strconv"
	"log"
	"time"

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

// RegisterRoutes регистрирует маршруты для платежей
func (h *Handler) RegisterRoutes(router *gin.RouterGroup) {
	paymentRoutes := router.Group("/payments")
	{
		paymentRoutes.POST("/calculate-price", h.calculatePrice)
		paymentRoutes.POST("/calculate-extension-price", h.calculateExtensionPrice)
		paymentRoutes.POST("/create", h.createPayment)
		paymentRoutes.GET("/status", h.getPaymentStatus)
		paymentRoutes.POST("/retry", h.retryPayment)
		paymentRoutes.POST("/webhook", h.handleWebhook)
		paymentRoutes.POST("/calculate-session-refund", h.calculateSessionRefund)
	}

	// Административные маршруты
	adminRoutes := router.Group("/admin/payments")
	{
		adminRoutes.GET("", h.adminListPayments)
		adminRoutes.GET("/statistics", h.adminGetPaymentStatistics)
		adminRoutes.POST("/refund", h.adminRefundPayment)
	}
}

// calculatePrice обработчик для расчета цены
func (h *Handler) calculatePrice(c *gin.Context) {
	var req models.CalculatePriceRequest

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	response, err := h.service.CalculatePrice(&req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, response)
}

// calculateExtensionPrice обработчик для расчета цены продления
func (h *Handler) calculateExtensionPrice(c *gin.Context) {
	var req models.CalculateExtensionPriceRequest

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	response, err := h.service.CalculateExtensionPrice(&req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, response)
}

// createPayment обработчик для создания платежа
func (h *Handler) createPayment(c *gin.Context) {
	var req models.CreatePaymentRequest

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	response, err := h.service.CreatePayment(&req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, response)
}

// getPaymentStatus обработчик для получения статуса платежа
func (h *Handler) getPaymentStatus(c *gin.Context) {
	paymentIDStr := c.Query("payment_id")
	if paymentIDStr == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Не указан ID платежа"})
		return
	}

	paymentID, err := uuid.Parse(paymentIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Некорректный ID платежа"})
		return
	}

	response, err := h.service.GetPaymentStatus(&models.GetPaymentStatusRequest{
		PaymentID: paymentID,
	})
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Платеж не найден"})
		return
	}

	c.JSON(http.StatusOK, response)
}

// retryPayment обработчик для повторной попытки оплаты
func (h *Handler) retryPayment(c *gin.Context) {
	var req models.RetryPaymentRequest

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	response, err := h.service.RetryPayment(&req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, response)
}

// handleWebhook обработчик для webhook от Tinkoff
func (h *Handler) handleWebhook(c *gin.Context) {
	// Парсим webhook
	var webhookReq models.WebhookRequest
	if err := c.ShouldBindJSON(&webhookReq); err != nil {
		log.Printf("Ошибка привязки JSON: %v", err)
		c.JSON(http.StatusBadRequest, gin.H{"error": "Неверный формат webhook"})
		return
	}

	// Обрабатываем webhook
	if err := h.service.HandleWebhook(&webhookReq); err != nil {
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

	if dateFromStr := c.Query("date_from"); dateFromStr != "" {
		if dateFrom, err := time.Parse("2006-01-02", dateFromStr); err == nil {
			req.DateFrom = &dateFrom
		}
	}

	if dateToStr := c.Query("date_to"); dateToStr != "" {
		if dateTo, err := time.Parse("2006-01-02", dateToStr); err == nil {
			req.DateTo = &dateTo
		}
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
	log.Printf("Запрос списка платежей (админка): PaymentID=%v, SessionID=%v, UserID=%v, Status=%v, PaymentType=%v, Limit=%v, Offset=%v", 
		req.PaymentID, req.SessionID, req.UserID, req.Status, req.PaymentType, req.Limit, req.Offset)

	response, err := h.service.ListPayments(&req)
	if err != nil {
		log.Printf("Ошибка получения списка платежей: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	log.Printf("Получен список платежей: Total=%d, Limit=%d, Offset=%d", 
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
	log.Printf("Запрос на возврат платежа (админка): PaymentID=%s, Amount=%d", 
		req.PaymentID, req.Amount)

	response, err := h.service.RefundPayment(&models.RefundPaymentRequest{
		PaymentID: req.PaymentID,
		Amount:    req.Amount,
	})
	if err != nil {
		log.Printf("Ошибка возврата платежа: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	adminResponse := &models.AdminRefundPaymentResponse{
		Payment: response.Payment,
		Refund:  response.Refund,
		Success: true,
		Message: "Возврат выполнен успешно",
	}

	log.Printf("Успешно выполнен возврат платежа: PaymentID=%s, Amount=%d", 
		req.PaymentID, req.Amount)

	c.JSON(http.StatusOK, adminResponse)
}

// calculateSessionRefund обработчик для расчета возврата по всем платежам сессии
func (h *Handler) calculateSessionRefund(c *gin.Context) {
	var req models.CalculateSessionRefundRequest

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Логируем запрос с мета-параметрами
	log.Printf("Запрос на расчет возврата по сессии: SessionID=%s, ServiceType=%s, UsedTime=%ds", 
		req.SessionID, req.ServiceType, req.UsedTimeSeconds)

	response, err := h.service.CalculateSessionRefund(&req)
	if err != nil {
		log.Printf("Ошибка расчета возврата по сессии: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	log.Printf("Рассчитан возврат по сессии: SessionID=%s, TotalRefundAmount=%d", 
		req.SessionID, response.TotalRefundAmount)

	c.JSON(http.StatusOK, response)
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
		dateFrom, err := time.Parse("2006-01-02", dateFromStr)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "invalid date_from format, expected YYYY-MM-DD"})
			return
		}
		req.DateFrom = &dateFrom
	}

	if dateToStr := c.Query("date_to"); dateToStr != "" {
		dateTo, err := time.Parse("2006-01-02", dateToStr)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "invalid date_to format, expected YYYY-MM-DD"})
			return
		}
		req.DateTo = &dateTo
	}

	if serviceType := c.Query("service_type"); serviceType != "" {
		req.ServiceType = &serviceType
	}

	log.Printf("Admin payment statistics request: user_id=%v, date_from=%v, date_to=%v, service_type=%v",
		req.UserID, req.DateFrom, req.DateTo, req.ServiceType)

	response, err := h.service.GetPaymentStatistics(&req)
	if err != nil {
		log.Printf("Error getting payment statistics: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, response)
} 