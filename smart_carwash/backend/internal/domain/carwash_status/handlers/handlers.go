package handlers

import (
	"carwash_backend/internal/domain/carwash_status/models"
	"carwash_backend/internal/domain/carwash_status/service"
	"carwash_backend/internal/logger"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

// Handler структура для обработчиков HTTP запросов статуса мойки
type Handler struct {
	service         service.Service
	adminMiddleware gin.HandlerFunc
}

// NewHandler создает новый экземпляр Handler
func NewHandler(service service.Service, adminMiddleware gin.HandlerFunc) *Handler {
	return &Handler{
		service:         service,
		adminMiddleware: adminMiddleware,
	}
}

// RegisterRoutes регистрирует маршруты для статуса мойки
func (h *Handler) RegisterRoutes(router *gin.RouterGroup) {
	// Публичный маршрут для получения статуса (для mini app)
	router.GET("/carwash/status", h.getCurrentStatus)

	// Административные маршруты
	adminRoutes := router.Group("/admin/carwash", h.adminMiddleware)
	{
		adminRoutes.GET("/status", h.adminGetCurrentStatus)
		adminRoutes.POST("/close", h.closeCarwash)
		adminRoutes.POST("/open", h.openCarwash)
		adminRoutes.GET("/history", h.getHistory)
	}
}

// getCurrentStatus получает текущий статус мойки (публичный)
func (h *Handler) getCurrentStatus(c *gin.Context) {
	req := &models.GetCurrentStatusRequest{}

	resp, err := h.service.GetCurrentStatus(c.Request.Context(), req)
	if err != nil {
		logger.WithContext(c).Errorf("getCurrentStatus: ошибка получения статуса: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, resp)
}

// adminGetCurrentStatus получает текущий статус мойки (админка)
func (h *Handler) adminGetCurrentStatus(c *gin.Context) {
	req := &models.GetCurrentStatusRequest{}

	resp, err := h.service.GetCurrentStatus(c.Request.Context(), req)
	if err != nil {
		logger.WithContext(c).Errorf("adminGetCurrentStatus: ошибка получения статуса: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, resp)
}

// closeCarwash закрывает мойку (админка)
func (h *Handler) closeCarwash(c *gin.Context) {
	var req models.CloseCarwashRequest

	if err := c.ShouldBindJSON(&req); err != nil {
		logger.WithContext(c).Errorf("closeCarwash: ошибка парсинга запроса: %v", err)
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Получаем ID администратора из контекста
	adminIDValue, exists := c.Get("user_id")
	if !exists {
		logger.WithContext(c).Errorf("closeCarwash: user_id не найден в контексте")
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Требуется авторизация"})
		return
	}

	adminID, ok := adminIDValue.(uuid.UUID)
	if !ok {
		logger.WithContext(c).Errorf("closeCarwash: некорректный тип user_id")
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Ошибка получения ID администратора"})
		return
	}

	// Логируем мета-параметры
	logger.WithContext(c).Infof("closeCarwash: запрос на закрытие мойки, admin_id: %s, reason: %v", adminID, req.Reason)
	c.Set("meta", gin.H{
		"admin_id": adminID,
		"reason":   req.Reason,
	})

	resp, err := h.service.CloseCarwash(c.Request.Context(), &req, adminID)
	if err != nil {
		logger.WithContext(c).Errorf("closeCarwash: ошибка закрытия мойки: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	// Обновляем мета-параметры с результатами
	c.Set("meta", gin.H{
		"admin_id":           adminID,
		"reason":             req.Reason,
		"completed_sessions": resp.CompletedSessions,
		"canceled_sessions":  resp.CanceledSessions,
	})

	c.JSON(http.StatusOK, resp)
}

// openCarwash открывает мойку (админка)
func (h *Handler) openCarwash(c *gin.Context) {
	var req models.OpenCarwashRequest

	if err := c.ShouldBindJSON(&req); err != nil {
		logger.WithContext(c).Errorf("openCarwash: ошибка парсинга запроса: %v", err)
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Получаем ID администратора из контекста
	adminIDValue, exists := c.Get("user_id")
	if !exists {
		logger.WithContext(c).Errorf("openCarwash: user_id не найден в контексте")
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Требуется авторизация"})
		return
	}

	adminID, ok := adminIDValue.(uuid.UUID)
	if !ok {
		logger.WithContext(c).Errorf("openCarwash: некорректный тип user_id")
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Ошибка получения ID администратора"})
		return
	}

	// Логируем мета-параметры
	logger.WithContext(c).Infof("openCarwash: запрос на открытие мойки, admin_id: %s", adminID)
	c.Set("meta", gin.H{
		"admin_id": adminID,
	})

	resp, err := h.service.OpenCarwash(c.Request.Context(), &req, adminID)
	if err != nil {
		logger.WithContext(c).Errorf("openCarwash: ошибка открытия мойки: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, resp)
}

// getHistory получает историю изменений статуса мойки (админка)
func (h *Handler) getHistory(c *gin.Context) {
	var req models.GetHistoryRequest

	// Получаем параметры из query
	if limitStr := c.Query("limit"); limitStr != "" {
		if limit, err := strconv.Atoi(limitStr); err == nil {
			req.Limit = limit
		}
	}

	if offsetStr := c.Query("offset"); offsetStr != "" {
		if offset, err := strconv.Atoi(offsetStr); err == nil {
			req.Offset = offset
		}
	}

	resp, err := h.service.GetHistory(c.Request.Context(), &req)
	if err != nil {
		logger.WithContext(c).Errorf("getHistory: ошибка получения истории: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, resp)
}
