package handlers

import (
	"net/http"
	"strconv"

	"carwash_backend/internal/domain/washbox/models"

	"github.com/gin-gonic/gin"
)

// cashierListWashBoxes обработчик для получения списка боксов мойки для кассира
func (h *Handler) cashierListWashBoxes(c *gin.Context) {
	// Получаем параметры фильтрации из query
	var req models.CashierListWashBoxesRequest

	// Статус
	if status := c.Query("status"); status != "" {
		req.Status = &status
	}

	// Тип услуги
	if serviceType := c.Query("service_type"); serviceType != "" {
		req.ServiceType = &serviceType
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

	// Получаем список боксов
	resp, err := h.service.CashierListWashBoxes(c.Request.Context(), &req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	// Логируем мета-параметры
	c.Set("meta", gin.H{
		"filters": gin.H{
			"status":       req.Status,
			"service_type": req.ServiceType,
			"limit":        req.Limit,
			"offset":       req.Offset,
		},
		"total": resp.Total,
	})

	c.JSON(http.StatusOK, resp)
}

// cashierSetMaintenance обработчик для перевода бокса в режим обслуживания кассиром
func (h *Handler) cashierSetMaintenance(c *gin.Context) {
	var req models.CashierSetMaintenanceRequest

	// Парсим JSON из тела запроса
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Переводим бокс в режим обслуживания
	resp, err := h.service.CashierSetMaintenance(c.Request.Context(), &req)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Логируем мета-параметры
	c.Set("meta", gin.H{
		"washbox_id": req.ID,
		"number":     resp.WashBox.Number,
		"status":     resp.WashBox.Status,
	})

	c.JSON(http.StatusOK, resp)
}
