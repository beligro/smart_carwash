package handlers

import (
	"fmt"
	"net/http"
	"time"

	paymentModels "carwash_backend/internal/domain/payment/models"
	paymentService "carwash_backend/internal/domain/payment/service"
	sessionModels "carwash_backend/internal/domain/session/models"
	sessionService "carwash_backend/internal/domain/session/service"
	"carwash_backend/internal/logger"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

// GuestUserID — служебный пользователь для всех гостевых сессий.
// Создаётся миграцией 000045. Конкретные гости различаются по guest_token,
// а не по user_id (он общий, нужен только для FK sessions.user_id -> users.id).
var GuestUserID = uuid.MustParse("11111111-1111-1111-1111-111111111111")

// Handler обработчики гостевого API (/guest/*), без JWT
type Handler struct {
	sessionSvc sessionService.Service
	paymentSvc paymentService.Service
}

// NewHandler создает обработчик гостевого API
func NewHandler(sessionSvc sessionService.Service, paymentSvc paymentService.Service) *Handler {
	return &Handler{
		sessionSvc: sessionSvc,
		paymentSvc: paymentSvc,
	}
}

// RegisterRoutes регистрирует маршруты /guest/*
func (h *Handler) RegisterRoutes(router *gin.RouterGroup) {
	router.POST("/sessions/with-payment", h.createGuestSession)
	router.GET("/sessions/:token", h.getGuestSession)
	router.POST("/sessions/:token/extend", h.extendGuestSession)
	router.POST("/sessions/:token/start", h.startGuestSession)
	router.POST("/sessions/:token/cancel", h.cancelGuestSession)
	router.POST("/sessions/:token/enable-chemistry", h.enableGuestChemistry)
	router.POST("/sessions/:token/complete", h.completeGuestSession)
	router.GET("/sessions/:token/payments", h.getGuestSessionPayments)
}

// completeGuestSession — POST /api/guest/sessions/:token/complete
func (h *Handler) completeGuestSession(c *gin.Context) {
	session, ok := h.resolveGuestSession(c)
	if !ok {
		return
	}
	resp, err := h.sessionSvc.CompleteSession(c.Request.Context(), &sessionModels.CompleteSessionRequest{
		SessionID:        session.ID,
		CompletionSource: "client",
	})
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, resp)
}

// getGuestSessionPayments — GET /api/guest/sessions/:token/payments
func (h *Handler) getGuestSessionPayments(c *gin.Context) {
	session, ok := h.resolveGuestSession(c)
	if !ok {
		return
	}
	resp, err := h.sessionSvc.GetSessionPayments(c.Request.Context(), &sessionModels.GetSessionPaymentsRequest{
		SessionID: session.ID,
	})
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, resp)
}

// startGuestSession — POST /api/guest/sessions/:token/start
func (h *Handler) startGuestSession(c *gin.Context) {
	session, ok := h.resolveGuestSession(c)
	if !ok {
		return
	}
	s, err := h.sessionSvc.StartSession(c.Request.Context(), &sessionModels.StartSessionRequest{SessionID: session.ID})
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, sessionModels.StartSessionResponse{Session: s})
}

// cancelGuestSession — POST /api/guest/sessions/:token/cancel (с возвратом средств)
func (h *Handler) cancelGuestSession(c *gin.Context) {
	session, ok := h.resolveGuestSession(c)
	if !ok {
		return
	}
	resp, err := h.sessionSvc.CancelSession(c.Request.Context(), &sessionModels.CancelSessionRequest{
		SessionID:  session.ID,
		UserID:     session.UserID, // служебный гостевой пользователь
		SkipRefund: false,          // гостю возвращаем средства
	})
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, resp)
}

// enableGuestChemistry — POST /api/guest/sessions/:token/enable-chemistry
func (h *Handler) enableGuestChemistry(c *gin.Context) {
	session, ok := h.resolveGuestSession(c)
	if !ok {
		return
	}
	resp, err := h.sessionSvc.EnableChemistry(c.Request.Context(), &sessionModels.EnableChemistryRequest{SessionID: session.ID})
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, resp)
}

// resolveGuestSession находит сессию по :token из URL. Токен из cookie/URL и есть авторизация гостя.
func (h *Handler) resolveGuestSession(c *gin.Context) (*sessionModels.Session, bool) {
	token := c.Param("token")
	if token == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "не указан токен"})
		return nil, false
	}
	session, err := h.sessionSvc.GetSessionByGuestToken(c.Request.Context(), token)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "сессия не найдена"})
		return nil, false
	}
	return session, true
}

// createGuestSession — POST /api/guest/sessions/with-payment
// Создаёт сессию без авторизации, генерирует guest_token для дальнейшего доступа.
func (h *Handler) createGuestSession(c *gin.Context) {
	var body struct {
		ServiceType          string `json:"service_type" binding:"required"`
		WithChemistry        bool   `json:"with_chemistry"`
		ChemistryTimeMinutes int    `json:"chemistry_time_minutes"`
		CarNumber            string `json:"car_number" binding:"required"`
		CarNumberCountry     string `json:"car_number_country"`
		RentalTimeMinutes    int    `json:"rental_time_minutes" binding:"required"`
		IdempotencyKey       string `json:"idempotency_key" binding:"required"`
	}
	if err := c.ShouldBindJSON(&body); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Генерируем guest_token — уникальный токен для доступа к сессии без аккаунта
	guestToken := generateGuestToken()

	// Создаём сессию от имени служебного гостевого пользователя.
	// Идентификация конкретного гостя — через guest_token (cookie).
	req := &sessionModels.CreateSessionWithPaymentRequest{
		UserID:               GuestUserID,
		ServiceType:          body.ServiceType,
		WithChemistry:        body.WithChemistry,
		ChemistryTimeMinutes: body.ChemistryTimeMinutes,
		CarNumber:            body.CarNumber,
		CarNumberCountry:     body.CarNumberCountry,
		RentalTimeMinutes:    body.RentalTimeMinutes,
		IdempotencyKey:       body.IdempotencyKey,
		Source:               "guest",
		GuestToken:           guestToken, // Попадает в URL возврата Tinkoff (параметр gt) для восстановления cookie
	}

	resp, err := h.sessionSvc.CreateSessionWithPayment(c.Request.Context(), req)
	if err != nil {
		logger.Printf("Guest - createGuestSession: ошибка создания сессии: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": fmt.Sprintf("ошибка создания сессии: %v", err)})
		return
	}

	// Сохраняем guest_token в сессии
	if err := h.sessionSvc.UpdateSessionGuestToken(c.Request.Context(), resp.Session.ID, guestToken); err != nil {
		logger.Printf("Guest - createGuestSession: ошибка сохранения guest_token: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "ошибка сохранения токена"})
		return
	}

	resp.Session.GuestToken = &guestToken

	c.JSON(http.StatusOK, gin.H{
		"session":     resp.Session,
		"payment":     resp.Payment,
		"guest_token": guestToken,
	})
}

// getGuestSession — GET /api/guest/sessions/:token
// Возвращает статус сессии по guest_token.
func (h *Handler) getGuestSession(c *gin.Context) {
	token := c.Param("token")
	if token == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "не указан токен"})
		return
	}

	session, err := h.sessionSvc.GetSessionByGuestToken(c.Request.Context(), token)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "сессия не найдена"})
		return
	}

	// Получаем основной платёж для сессии
	paymentsResp, _ := h.sessionSvc.GetSessionPayments(c.Request.Context(), &sessionModels.GetSessionPaymentsRequest{
		SessionID: session.ID,
	})

	var mainPayment *sessionModels.Payment
	if paymentsResp != nil {
		mainPayment = paymentsResp.MainPayment
	}

	c.JSON(http.StatusOK, gin.H{
		"session": session,
		"payment": mainPayment,
	})
}

// extendGuestSession — POST /api/guest/sessions/:token/extend
// Создаёт платёж на продление сессии по guest_token.
func (h *Handler) extendGuestSession(c *gin.Context) {
	token := c.Param("token")
	if token == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "не указан токен"})
		return
	}

	var body struct {
		ExtensionTimeMinutes          int `json:"extension_time_minutes"`
		ExtensionChemistryTimeMinutes int `json:"extension_chemistry_time_minutes"`
	}
	if err := c.ShouldBindJSON(&body); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	session, err := h.sessionSvc.GetSessionByGuestToken(c.Request.Context(), token)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "сессия не найдена"})
		return
	}

	resp, err := h.sessionSvc.ExtendSessionWithPayment(c.Request.Context(), &sessionModels.ExtendSessionWithPaymentRequest{
		SessionID:                     session.ID,
		ExtensionTimeMinutes:          body.ExtensionTimeMinutes,
		ExtensionChemistryTimeMinutes: body.ExtensionChemistryTimeMinutes,
	})
	if err != nil {
		logger.Printf("Guest - extendGuestSession: ошибка продления сессии %s: %v", session.ID, err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": fmt.Sprintf("ошибка продления: %v", err)})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"session": resp.Session,
		"payment": resp.Payment,
	})
}

// generateGuestToken генерирует уникальный токен для гостевой сессии
func generateGuestToken() string {
	id := uuid.New()
	ts := time.Now().UnixNano() / 1e6 // миллисекунды
	return fmt.Sprintf("g_%d_%s", ts, id.String()[:8])
}

// getGuestPaymentStatus — вспомогательный для проверки статуса платежа по session_id и guest_token
func (h *Handler) getGuestPaymentStatus(c *gin.Context) {
	token := c.Param("token")
	paymentIDStr := c.Query("payment_id")

	if token == "" || paymentIDStr == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "не указаны обязательные параметры"})
		return
	}

	// Проверяем что токен принадлежит сессии
	_, err := h.sessionSvc.GetSessionByGuestToken(c.Request.Context(), token)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "сессия не найдена"})
		return
	}

	paymentID, err := uuid.Parse(paymentIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "неверный формат payment_id"})
		return
	}

	resp, err := h.paymentSvc.GetPaymentStatus(c.Request.Context(), &paymentModels.GetPaymentStatusRequest{
		PaymentID: paymentID,
	})
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "платёж не найден"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"payment": resp.Payment})
}
