package handlers

import (
	"net/http"
	"strconv"
	"time"

	"carwash_backend/internal/domain/maintenance/models"
	"carwash_backend/internal/domain/maintenance/service"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

// Handler HTTP-обработчики сервисных нарядов.
type Handler struct {
	service service.Service
}

// NewHandler создаёт Handler.
func NewHandler(svc service.Service) *Handler {
	return &Handler{service: svc}
}

// RegisterRoutes регистрирует маршруты кассира и админа.
func (h *Handler) RegisterRoutes(router *gin.RouterGroup, cashierMiddleware gin.HandlerFunc, adminMiddleware gin.HandlerFunc) {
	// Кассир: справочник симптомов и открытие наряда.
	cashier := router.Group("/cashier")
	if cashierMiddleware != nil {
		cashier.Use(cashierMiddleware)
	}
	{
		cashier.GET("/service-symptoms", h.getSymptoms)
		cashier.POST("/service-tickets", h.openTicket)
		cashier.GET("/service-tickets/open", h.listOpenTicketsCashier)
	}

	// Админ: журнал, справочник работ, закрытие наряда, получатели уведомлений.
	admin := router.Group("/admin")
	if adminMiddleware != nil {
		admin.Use(adminMiddleware)
	}
	{
		admin.GET("/service-tickets", h.listTickets)
		admin.GET("/service-catalog", h.getCatalog)
		admin.POST("/service-tickets/:id/close", h.closeTicket)
		admin.GET("/notification-recipients", h.listRecipients)
		admin.POST("/notification-recipients", h.createRecipient)
		admin.PATCH("/notification-recipients/:id", h.updateRecipient)
	}
}

// getSymptoms GET /cashier/service-symptoms?box_number=N
func (h *Handler) getSymptoms(c *gin.Context) {
	boxNumber, err := strconv.Atoi(c.Query("box_number"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "некорректный box_number"})
		return
	}
	symptoms, err := h.service.GetSymptoms(c.Request.Context(), boxNumber)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"symptoms": symptoms})
}

// openTicket POST /cashier/service-tickets
func (h *Handler) openTicket(c *gin.Context) {
	var req models.OpenTicketRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	cashierID := cashierIDFromContext(c)

	ticket, err := h.service.OpenTicket(c.Request.Context(), cashierID, &req)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"ticket": ticket})
}

// listOpenTicketsCashier GET /cashier/service-tickets/open — открытые наряды (для показа причины на боксах кассиру)
func (h *Handler) listOpenTicketsCashier(c *gin.Context) {
	open := models.TicketStatusOpen
	tickets, err := h.service.ListTickets(c.Request.Context(), &open, nil, nil, 200)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"tickets": tickets})
}

// getCatalog GET /admin/service-catalog?box_number=N
func (h *Handler) getCatalog(c *gin.Context) {
	boxNumber, err := strconv.Atoi(c.Query("box_number"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "некорректный box_number"})
		return
	}
	catalog, err := h.service.GetCatalog(c.Request.Context(), boxNumber)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"catalog": catalog})
}

// listTickets GET /admin/service-tickets?status=&box=&since=&limit=
func (h *Handler) listTickets(c *gin.Context) {
	var status *string
	if s := c.Query("status"); s != "" {
		status = &s
	}
	var boxNumber *int
	if b := c.Query("box"); b != "" {
		if n, err := strconv.Atoi(b); err == nil {
			boxNumber = &n
		}
	}
	var since *time.Time
	if sc := c.Query("since"); sc != "" {
		if t, err := time.Parse(time.RFC3339, sc); err == nil {
			since = &t
		}
	}
	limit := 200
	if l := c.Query("limit"); l != "" {
		if n, err := strconv.Atoi(l); err == nil && n > 0 {
			limit = n
		}
	}

	tickets, err := h.service.ListTickets(c.Request.Context(), status, boxNumber, since, limit)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"tickets": tickets})
}

// closeTicket POST /admin/service-tickets/:id/close
func (h *Handler) closeTicket(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "некорректный id наряда"})
		return
	}
	var req models.CloseTicketRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	closedBy := "Админ"
	if v, ok := c.Get("username"); ok {
		if s, ok := v.(string); ok && s != "" {
			closedBy = s
		}
	}

	ticket, err := h.service.CloseTicket(c.Request.Context(), id, closedBy, &req)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"ticket": ticket})
}

// listRecipients GET /admin/notification-recipients
func (h *Handler) listRecipients(c *gin.Context) {
	recipients, err := h.service.ListRecipients(c.Request.Context())
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"recipients": recipients})
}

// createRecipient POST /admin/notification-recipients
func (h *Handler) createRecipient(c *gin.Context) {
	var req models.CreateRecipientRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	rec, err := h.service.CreateRecipient(c.Request.Context(), &req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"recipient": rec})
}

// updateRecipient PATCH /admin/notification-recipients/:id
func (h *Handler) updateRecipient(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "некорректный id"})
		return
	}
	var req models.UpdateRecipientRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	if req.IsActive == nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "не указано поле is_active"})
		return
	}
	if err := h.service.SetRecipientActive(c.Request.Context(), id, *req.IsActive); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"status": "ok"})
}

// cashierIDFromContext извлекает UUID кассира из контекста (устанавливается CashierMiddleware).
func cashierIDFromContext(c *gin.Context) *uuid.UUID {
	v, ok := c.Get("cashier_id")
	if !ok {
		return nil
	}
	switch id := v.(type) {
	case uuid.UUID:
		return &id
	case string:
		if parsed, err := uuid.Parse(id); err == nil {
			return &parsed
		}
	}
	return nil
}
