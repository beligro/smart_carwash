package handlers

import (
	"carwash_backend/internal/domain/washboxlog/service"
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
)

type Handler struct {
	svc service.Service
}

func NewHandler(svc service.Service) *Handler {
	return &Handler{svc: svc}
}

// RegisterRoutes регистрирует админские ручки просмотра истории изменений
func (h *Handler) RegisterRoutes(rg *gin.RouterGroup, adminMiddleware gin.HandlerFunc) {
	admin := rg.Group("/admin")
	admin.Use(adminMiddleware)
	{
		admin.GET("/washbox-change-logs", h.listLogs)
	}
}

// listLogs возвращает историю изменений с фильтрами
// GET /admin/washbox-change-logs?box_number=&actor_type=&action=&date_from=&date_to=&limit=&offset=
func (h *Handler) listLogs(c *gin.Context) {
	// Ограничиваем доступ для limited_admin
	if roleAny, ok := c.Get("role"); ok {
		if role, _ := roleAny.(string); role == "limited_admin" {
			c.JSON(http.StatusForbidden, gin.H{"error": "доступ запрещен"})
			return
		}
	}
	ctx := c.Request.Context()

	var (
		boxNumber *int
		actorType *string
		action    *string
		since     *time.Time
		until     *time.Time
		limit     = 50
		offset    = 0
	)

	if v := c.Query("box_number"); v != "" {
		if n, err := strconv.Atoi(v); err == nil {
			boxNumber = &n
		}
	}
	if v := c.Query("actor_type"); v != "" {
		actorType = &v
	}
	if v := c.Query("action"); v != "" {
		action = &v
	}
	if v := c.Query("date_from"); v != "" {
		if t, err := time.Parse(time.RFC3339, v); err == nil {
			since = &t
		}
	}
	if v := c.Query("date_to"); v != "" {
		if t, err := time.Parse(time.RFC3339, v); err == nil {
			until = &t
		}
	}
	if v := c.Query("limit"); v != "" {
		if n, err := strconv.Atoi(v); err == nil {
			limit = n
		}
	}
	if v := c.Query("offset"); v != "" {
		if n, err := strconv.Atoi(v); err == nil {
			offset = n
		}
	}

	rows, total, err := h.svc.List(ctx, boxNumber, actorType, action, since, until, limit, offset)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to list logs"})
		return
	}
	c.JSON(http.StatusOK, gin.H{
		"logs":   rows,
		"total":  total,
		"limit":  limit,
		"offset": offset,
	})
}


