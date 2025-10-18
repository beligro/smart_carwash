package handlers

import (
	"net/http"
	"strconv"

	"carwash_backend/internal/domain/washbox/models"
	"carwash_backend/internal/domain/washbox/service"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

// Handler структура для обработчиков HTTP запросов боксов мойки
type Handler struct {
	service service.Service
}

// NewHandler создает новый экземпляр Handler
func NewHandler(service service.Service) *Handler {
	return &Handler{
		service: service,
	}
}

// RegisterRoutes регистрирует маршруты для боксов мойки
func (h *Handler) RegisterRoutes(router *gin.RouterGroup, cleanerMiddleware gin.HandlerFunc) {
	// Маршруты для боксов мойки
	// Пока нет маршрутов, так как queue-status перенесен в домен queue

	// Административные маршруты
	adminRoutes := router.Group("/admin/washboxes")
	{
		adminRoutes.GET("", h.adminListWashBoxes)
		adminRoutes.POST("", h.adminCreateWashBox)
		adminRoutes.PUT("", h.adminUpdateWashBox)
		adminRoutes.DELETE("", h.adminDeleteWashBox)
		adminRoutes.GET("/by-id", h.adminGetWashBox)
	}

	// Административные маршруты для логов уборки
	adminCleaningRoutes := router.Group("/admin/cleaning-logs")
	{
		adminCleaningRoutes.POST("", h.AdminListCleaningLogs)
	}

	// Маршруты для кассира
	cashierRoutes := router.Group("/cashier/washboxes")
	{
		cashierRoutes.GET("", h.cashierListWashBoxes)
		cashierRoutes.POST("/maintenance", h.cashierSetMaintenance)
	}

	// Маршруты для уборщика
	cleanerRoutes := router.Group("/cleaner/washboxes", cleanerMiddleware)
	{
		cleanerRoutes.GET("", h.cleanerListWashBoxes)
		cleanerRoutes.POST("/reserve-cleaning", h.cleanerReserveCleaning)
		cleanerRoutes.POST("/start-cleaning", h.cleanerStartCleaning)
		cleanerRoutes.POST("/cancel-cleaning", h.cleanerCancelCleaning)
		cleanerRoutes.POST("/complete-cleaning", h.cleanerCompleteCleaning)
	}
}

// adminListWashBoxes обработчик для получения списка боксов мойки
func (h *Handler) adminListWashBoxes(c *gin.Context) {
	// Получаем параметры фильтрации из query
	var req models.AdminListWashBoxesRequest

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
	resp, err := h.service.AdminListWashBoxes(&req)
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

// AdminListCleaningLogs получает логи уборки с фильтрами (админка)
func (h *Handler) AdminListCleaningLogs(c *gin.Context) {
	var req models.AdminListCleaningLogsRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Устанавливаем значения по умолчанию для пагинации
	if req.Limit == nil {
		limit := 20
		req.Limit = &limit
	}
	if req.Offset == nil {
		offset := 0
		req.Offset = &offset
	}

	resp, err := h.service.AdminListCleaningLogs(&req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, resp)
}

// adminCreateWashBox обработчик для создания бокса мойки
func (h *Handler) adminCreateWashBox(c *gin.Context) {
	var req models.AdminCreateWashBoxRequest

	// Парсим JSON из тела запроса
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Создаем бокс
	resp, err := h.service.AdminCreateWashBox(&req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	// Логируем мета-параметры
	c.Set("meta", gin.H{
		"washbox_id": resp.WashBox.ID,
		"number":     resp.WashBox.Number,
	})

	c.JSON(http.StatusCreated, resp)
}

// adminUpdateWashBox обработчик для обновления бокса мойки
func (h *Handler) adminUpdateWashBox(c *gin.Context) {
	var req models.AdminUpdateWashBoxRequest

	// Парсим JSON из тела запроса
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Обновляем бокс
	resp, err := h.service.AdminUpdateWashBox(&req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	// Логируем мета-параметры
	c.Set("meta", gin.H{
		"washbox_id": req.ID,
		"number":     resp.WashBox.Number,
	})

	c.JSON(http.StatusOK, resp)
}

// adminDeleteWashBox обработчик для удаления бокса мойки
func (h *Handler) adminDeleteWashBox(c *gin.Context) {
	var req models.AdminDeleteWashBoxRequest

	// Парсим JSON из тела запроса
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Удаляем бокс
	resp, err := h.service.AdminDeleteWashBox(&req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	// Логируем мета-параметры
	c.Set("meta", gin.H{
		"washbox_id": req.ID,
	})

	c.JSON(http.StatusOK, resp)
}


// adminGetWashBox обработчик для получения бокса мойки по ID
func (h *Handler) adminGetWashBox(c *gin.Context) {
	var req models.AdminGetWashBoxRequest

	// Получаем ID из query параметра
	idStr := c.Query("id")
	if idStr == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "ID бокса обязателен"})
		return
	}

	id, err := uuid.Parse(idStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Некорректный ID бокса"})
		return
	}

	req.ID = id

	// Получаем бокс
	resp, err := h.service.AdminGetWashBox(&req)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
		return
	}

	// Логируем мета-параметры
	c.Set("meta", gin.H{
		"washbox_id": req.ID,
		"number":     resp.WashBox.Number,
	})

	c.JSON(http.StatusOK, resp)
}

// cleanerListWashBoxes обработчик для получения списка боксов для уборщика
func (h *Handler) cleanerListWashBoxes(c *gin.Context) {
	var req models.CleanerListWashBoxesRequest

	// Парсим параметры из query
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

	// Получаем список боксов
	resp, err := h.service.CleanerListWashBoxes(&req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, resp)
}


// cleanerReserveCleaning обработчик для резервирования уборки
func (h *Handler) cleanerReserveCleaning(c *gin.Context) {
	var req models.CleanerReserveCleaningRequest

	// Парсим JSON из тела запроса
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Получаем ID уборщика из контекста
	cleanerIDInterface, exists := c.Get("cleaner_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Необходима авторизация"})
		return
	}

	cleanerID, ok := cleanerIDInterface.(uuid.UUID)
	if !ok {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Ошибка получения ID уборщика"})
		return
	}

	// Резервируем уборку
	resp, err := h.service.CleanerReserveCleaning(&req, cleanerID)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Логируем мета-параметры
	c.Set("meta", gin.H{
		"washbox_id": req.WashBoxID,
		"cleaner_id": cleanerID,
	})

	c.JSON(http.StatusOK, resp)
}


// cleanerStartCleaning обработчик для начала уборки
func (h *Handler) cleanerStartCleaning(c *gin.Context) {
	var req models.CleanerStartCleaningRequest

	// Парсим JSON из тела запроса
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Получаем ID уборщика из контекста
	cleanerIDInterface, exists := c.Get("cleaner_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Необходима авторизация"})
		return
	}

	cleanerID, ok := cleanerIDInterface.(uuid.UUID)
	if !ok {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Ошибка получения ID уборщика"})
		return
	}

	// Начинаем уборку
	resp, err := h.service.CleanerStartCleaning(&req, cleanerID)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Логируем мета-параметры
	c.Set("meta", gin.H{
		"washbox_id": req.WashBoxID,
		"cleaner_id": cleanerID,
	})

	c.JSON(http.StatusOK, resp)
}


// cleanerCancelCleaning обработчик для отмены уборки
func (h *Handler) cleanerCancelCleaning(c *gin.Context) {
	var req models.CleanerCancelCleaningRequest

	// Парсим JSON из тела запроса
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Получаем ID уборщика из контекста
	cleanerIDInterface, exists := c.Get("cleaner_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Необходима авторизация"})
		return
	}

	cleanerID, ok := cleanerIDInterface.(uuid.UUID)
	if !ok {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Ошибка получения ID уборщика"})
		return
	}

	// Отменяем уборку
	resp, err := h.service.CleanerCancelCleaning(&req, cleanerID)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Логируем мета-параметры
	c.Set("meta", gin.H{
		"washbox_id": req.WashBoxID,
		"cleaner_id": cleanerID,
	})

	c.JSON(http.StatusOK, resp)
}


// cleanerCompleteCleaning обработчик для завершения уборки
func (h *Handler) cleanerCompleteCleaning(c *gin.Context) {
	var req models.CleanerCompleteCleaningRequest

	// Парсим JSON из тела запроса
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Получаем ID уборщика из контекста
	cleanerIDInterface, exists := c.Get("cleaner_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Необходима авторизация"})
		return
	}

	cleanerID, ok := cleanerIDInterface.(uuid.UUID)
	if !ok {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Ошибка получения ID уборщика"})
		return
	}

	// Завершаем уборку
	resp, err := h.service.CleanerCompleteCleaning(&req, cleanerID)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Логируем мета-параметры
	c.Set("meta", gin.H{
		"washbox_id": req.WashBoxID,
		"cleaner_id": cleanerID,
	})

	c.JSON(http.StatusOK, resp)
}

