package handlers

import (
	"net/http"
	"strconv"

	authModels "carwash_backend/internal/domain/auth/models"
	linktokenService "carwash_backend/internal/domain/linktoken/service"
	paymentModels "carwash_backend/internal/domain/payment/models"
	paymentService "carwash_backend/internal/domain/payment/service"
	sessionModels "carwash_backend/internal/domain/session/models"
	sessionService "carwash_backend/internal/domain/session/service"
	userService "carwash_backend/internal/domain/user/service"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

// Handler обработчики API для веб-клиента (/web/*), user_id из JWT
type Handler struct {
	userSvc      userService.Service
	sessionSvc   sessionService.Service
	linkTokenSvc linktokenService.Service
	paymentSvc   paymentService.Service
}

// NewHandler создает обработчик веб-API
func NewHandler(
	userSvc userService.Service,
	sessionSvc sessionService.Service,
	linkTokenSvc linktokenService.Service,
	paymentSvc paymentService.Service,
) *Handler {
	return &Handler{
		userSvc:      userSvc,
		sessionSvc:   sessionSvc,
		linkTokenSvc: linkTokenSvc,
		paymentSvc:   paymentSvc,
	}
}

// userIDFromContext возвращает user_id из контекста (установлен WebAuthMiddleware)
func userIDFromContext(c *gin.Context) (uuid.UUID, bool) {
	v, ok := c.Get("user_id")
	if !ok {
		return uuid.Nil, false
	}
	id, ok := v.(uuid.UUID)
	return id, ok
}

// RegisterRoutes регистрирует маршруты /web/*
func (h *Handler) RegisterRoutes(router *gin.RouterGroup) {
	// router уже имеет базовый путь /web и middleware
	router.GET("/me", h.me)
	router.GET("/sessions", h.getUserSession)
	router.GET("/sessions/for-payment", h.getUserSessionForPayment)
	router.GET("/sessions/check-active", h.checkActiveSession)
	router.GET("/sessions/by-id", h.getSessionByID)
	router.POST("/sessions/start", h.startSession)
	router.POST("/sessions/complete", h.completeSession)
	router.POST("/sessions/extend-with-payment", h.extendSessionWithPayment)
	router.GET("/sessions/payments", h.getSessionPayments)
	router.GET("/sessions/history", h.getUserSessionHistory)
	router.POST("/sessions/cancel", h.cancelSession)
	router.POST("/sessions/enable-chemistry", h.enableChemistry)
	router.POST("/sessions/with-payment", h.createSessionWithPayment)
	router.POST("/link-telegram/request", h.linkTelegramRequest)
	router.GET("/payments/status", h.getPaymentStatus)
}

func (h *Handler) me(c *gin.Context) {
	userID, ok := userIDFromContext(c)
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Не авторизован"})
		return
	}
	user, err := h.userSvc.GetUserByID(c.Request.Context(), userID)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Пользователь не найден"})
		return
	}
	c.JSON(http.StatusOK, authModels.WebUser{
		ID:               user.ID,
		Email:            user.Email,
		TelegramID:       user.TelegramID,
		CarNumber:        user.CarNumber,
		CarNumberCountry: user.CarNumberCountry,
	})
}

func (h *Handler) getUserSession(c *gin.Context) {
	userID, ok := userIDFromContext(c)
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Не авторизован"})
		return
	}
	resp, err := h.sessionSvc.GetUserSession(c.Request.Context(), &sessionModels.GetUserSessionRequest{UserID: userID})
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Сессия не найдена"})
		return
	}
	c.JSON(http.StatusOK, resp)
}

func (h *Handler) getUserSessionForPayment(c *gin.Context) {
	userID, ok := userIDFromContext(c)
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Не авторизован"})
		return
	}
	resp, err := h.sessionSvc.GetUserSessionForPayment(c.Request.Context(), &sessionModels.GetUserSessionRequest{UserID: userID})
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Сессия не найдена"})
		return
	}
	c.JSON(http.StatusOK, resp)
}

func (h *Handler) checkActiveSession(c *gin.Context) {
	userID, ok := userIDFromContext(c)
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Не авторизован"})
		return
	}
	resp, err := h.sessionSvc.CheckActiveSession(c.Request.Context(), &sessionModels.CheckActiveSessionRequest{UserID: userID})
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, resp)
}

func (h *Handler) getSessionByID(c *gin.Context) {
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
	userID, ok := userIDFromContext(c)
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Не авторизован"})
		return
	}
	resp, err := h.sessionSvc.GetSession(c.Request.Context(), &sessionModels.GetSessionRequest{SessionID: sessionID})
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Сессия не найдена"})
		return
	}
	if resp.Session.UserID != userID {
		c.JSON(http.StatusForbidden, gin.H{"error": "Нет доступа к этой сессии"})
		return
	}
	c.JSON(http.StatusOK, resp)
}

func (h *Handler) startSession(c *gin.Context) {
	var req sessionModels.StartSessionRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	userID, ok := userIDFromContext(c)
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Не авторизован"})
		return
	}
	session, err := h.sessionSvc.GetSession(c.Request.Context(), &sessionModels.GetSessionRequest{SessionID: req.SessionID})
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Сессия не найдена"})
		return
	}
	if session.Session.UserID != userID {
		c.JSON(http.StatusForbidden, gin.H{"error": "Нет доступа"})
		return
	}
	s, err := h.sessionSvc.StartSession(c.Request.Context(), &req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, sessionModels.StartSessionResponse{Session: s})
}

func (h *Handler) completeSession(c *gin.Context) {
	var req sessionModels.CompleteSessionRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	userID, ok := userIDFromContext(c)
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Не авторизован"})
		return
	}
	session, err := h.sessionSvc.GetSession(c.Request.Context(), &sessionModels.GetSessionRequest{SessionID: req.SessionID})
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Сессия не найдена"})
		return
	}
	if session.Session.UserID != userID {
		c.JSON(http.StatusForbidden, gin.H{"error": "Нет доступа"})
		return
	}
	req.CompletionSource = "client"
	resp, err := h.sessionSvc.CompleteSession(c.Request.Context(), &req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, resp)
}

func (h *Handler) extendSessionWithPayment(c *gin.Context) {
	var req sessionModels.ExtendSessionWithPaymentRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	userID, ok := userIDFromContext(c)
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Не авторизован"})
		return
	}
	session, err := h.sessionSvc.GetSession(c.Request.Context(), &sessionModels.GetSessionRequest{SessionID: req.SessionID})
	if err != nil || session == nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Сессия не найдена"})
		return
	}
	if session.Session.UserID != userID {
		c.JSON(http.StatusForbidden, gin.H{"error": "Нет доступа"})
		return
	}
	resp, err := h.sessionSvc.ExtendSessionWithPayment(c.Request.Context(), &req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, resp)
}

func (h *Handler) getSessionPayments(c *gin.Context) {
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
	userID, ok := userIDFromContext(c)
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Не авторизован"})
		return
	}
	session, err := h.sessionSvc.GetSession(c.Request.Context(), &sessionModels.GetSessionRequest{SessionID: sessionID})
	if err != nil || session == nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Сессия не найдена"})
		return
	}
	if session.Session.UserID != userID {
		c.JSON(http.StatusForbidden, gin.H{"error": "Нет доступа"})
		return
	}
	resp, err := h.sessionSvc.GetSessionPayments(c.Request.Context(), &sessionModels.GetSessionPaymentsRequest{SessionID: sessionID})
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, resp)
}

func (h *Handler) getUserSessionHistory(c *gin.Context) {
	userID, ok := userIDFromContext(c)
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Не авторизован"})
		return
	}
	limit := 5
	offset := 0
	if l := c.Query("limit"); l != "" {
		if n, err := strconv.Atoi(l); err == nil {
			limit = n
		}
	}
	if o := c.Query("offset"); o != "" {
		if n, err := strconv.Atoi(o); err == nil {
			offset = n
		}
	}
	sessions, err := h.sessionSvc.GetUserSessionHistory(c.Request.Context(), &sessionModels.GetUserSessionHistoryRequest{
		UserID: userID,
		Limit:  limit,
		Offset: offset,
	})
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, sessionModels.GetUserSessionHistoryResponse{Sessions: sessions})
}

func (h *Handler) cancelSession(c *gin.Context) {
	userID, ok := userIDFromContext(c)
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Не авторизован"})
		return
	}
	var body struct {
		SessionID  uuid.UUID `json:"session_id" binding:"required"`
		SkipRefund bool      `json:"skip_refund"`
	}
	if err := c.ShouldBindJSON(&body); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	req := &sessionModels.CancelSessionRequest{
		SessionID:  body.SessionID,
		UserID:     userID,
		SkipRefund: body.SkipRefund,
	}
	resp, err := h.sessionSvc.CancelSession(c.Request.Context(), req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, resp)
}

func (h *Handler) enableChemistry(c *gin.Context) {
	var req sessionModels.EnableChemistryRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	userID, ok := userIDFromContext(c)
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Не авторизован"})
		return
	}
	session, err := h.sessionSvc.GetSession(c.Request.Context(), &sessionModels.GetSessionRequest{SessionID: req.SessionID})
	if err != nil || session == nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Сессия не найдена"})
		return
	}
	if session.Session.UserID != userID {
		c.JSON(http.StatusForbidden, gin.H{"error": "Нет доступа"})
		return
	}
	resp, err := h.sessionSvc.EnableChemistry(c.Request.Context(), &req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, resp)
}

func (h *Handler) createSessionWithPayment(c *gin.Context) {
	userID, ok := userIDFromContext(c)
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Не авторизован"})
		return
	}
	var body struct {
		ServiceType          string    `json:"service_type" binding:"required"`
		WithChemistry        bool      `json:"with_chemistry"`
		ChemistryTimeMinutes int       `json:"chemistry_time_minutes"`
		CarNumber            string    `json:"car_number" binding:"required"`
		CarNumberCountry     string    `json:"car_number_country"`
		Email                string    `json:"email"`
		RentalTimeMinutes    int       `json:"rental_time_minutes" binding:"required"`
		IdempotencyKey       string    `json:"idempotency_key" binding:"required"`
	}
	if err := c.ShouldBindJSON(&body); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	req := &sessionModels.CreateSessionWithPaymentRequest{
		UserID:               userID,
		ServiceType:          body.ServiceType,
		WithChemistry:        body.WithChemistry,
		ChemistryTimeMinutes: body.ChemistryTimeMinutes,
		CarNumber:            body.CarNumber,
		CarNumberCountry:     body.CarNumberCountry,
		Email:                body.Email,
		RentalTimeMinutes:    body.RentalTimeMinutes,
		IdempotencyKey:       body.IdempotencyKey,
		Source:               "web",
	}
	resp, err := h.sessionSvc.CreateSessionWithPayment(c.Request.Context(), req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, resp)
}

func (h *Handler) linkTelegramRequest(c *gin.Context) {
	userID, ok := userIDFromContext(c)
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Не авторизован"})
		return
	}
	resp, err := h.linkTokenSvc.CreateToken(c.Request.Context(), userID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, resp)
}

func (h *Handler) getPaymentStatus(c *gin.Context) {
	userID, ok := userIDFromContext(c)
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Не авторизован"})
		return
	}
	paymentIDStr := c.Query("payment_id")
	if paymentIDStr == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Не указан payment_id"})
		return
	}
	paymentID, err := uuid.Parse(paymentIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Некорректный payment_id"})
		return
	}
	paymentResp, err := h.paymentSvc.GetPaymentStatus(c.Request.Context(), &paymentModels.GetPaymentStatusRequest{PaymentID: paymentID})
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Платеж не найден"})
		return
	}
	sessResp, err := h.sessionSvc.GetSession(c.Request.Context(), &sessionModels.GetSessionRequest{SessionID: paymentResp.Payment.SessionID})
	if err != nil || sessResp == nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Сессия не найдена"})
		return
	}
	if sessResp.Session.UserID != userID {
		c.JSON(http.StatusForbidden, gin.H{"error": "Нет доступа к этому платежу"})
		return
	}
	c.JSON(http.StatusOK, paymentResp)
}
