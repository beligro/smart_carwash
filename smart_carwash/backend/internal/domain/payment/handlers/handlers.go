package handlers

import (
	"carwash_backend/internal/domain/payment/models"
	"carwash_backend/internal/domain/payment/service"
	"net/http"
	"strconv"
	"log"

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
		paymentRoutes.POST("/create", h.createPayment)
		paymentRoutes.GET("/status", h.getPaymentStatus)
		paymentRoutes.POST("/retry", h.retryPayment)
		paymentRoutes.POST("/webhook", h.handleWebhook)
	}

	// Административные маршруты
	adminRoutes := router.Group("/admin/payments")
	{
		adminRoutes.GET("", h.adminListPayments)
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
	if sessionIDStr := c.Query("session_id"); sessionIDStr != "" {
		if sessionID, err := uuid.Parse(sessionIDStr); err == nil {
			req.SessionID = &sessionID
		}
	}

	if status := c.Query("status"); status != "" {
		req.Status = &status
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

	response, err := h.service.ListPayments(&req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, response)
} 