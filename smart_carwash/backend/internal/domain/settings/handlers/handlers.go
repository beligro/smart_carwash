package handlers

import (
	"carwash_backend/internal/domain/settings/models"
	"carwash_backend/internal/domain/settings/service"
	"net/http"

	"github.com/gin-gonic/gin"
)

// Handler обработчик запросов для настроек
type Handler struct {
	service service.Service
}

// NewHandler создает новый экземпляр Handler
func NewHandler(service service.Service) *Handler {
	return &Handler{
		service: service,
	}
}

// RegisterRoutes регистрирует маршруты для настроек
func (h *Handler) RegisterRoutes(router gin.IRouter) {
	settingsGroup := router.Group("/settings")
	{
		settingsGroup.GET("/rental-times", h.GetAvailableRentalTimes)
		settingsGroup.PUT("/rental-times", h.UpdateAvailableRentalTimes)
	}

	// Административные маршруты
	adminSettingsGroup := router.Group("/admin/settings")
	{
		adminSettingsGroup.GET("", h.AdminGetSettings)
		adminSettingsGroup.PUT("/prices", h.AdminUpdatePrices)
		adminSettingsGroup.PUT("/rental-times", h.AdminUpdateRentalTimes)
	}
}

// GetAvailableRentalTimes получает доступное время аренды для определенного типа услуги
// @Summary Получить доступное время аренды
// @Description Получает список доступного времени аренды для определенного типа услуги
// @Tags settings
// @Accept json
// @Produce json
// @Param service_type query string true "Тип услуги"
// @Success 200 {object} models.GetAvailableRentalTimesResponse
// @Failure 400 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Router /settings/rental-times [get]
func (h *Handler) GetAvailableRentalTimes(c *gin.Context) {
	serviceType := c.Query("service_type")
	if serviceType == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "service_type is required"})
		return
	}

	req := &models.GetAvailableRentalTimesRequest{
		ServiceType: serviceType,
	}

	resp, err := h.service.GetAvailableRentalTimes(req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, resp)
}

// UpdateAvailableRentalTimes обновляет доступное время аренды для определенного типа услуги
// @Summary Обновить доступное время аренды
// @Description Обновляет список доступного времени аренды для определенного типа услуги
// @Tags settings
// @Accept json
// @Produce json
// @Param request body models.UpdateAvailableRentalTimesRequest true "Запрос на обновление времени аренды"
// @Success 200 {object} models.UpdateAvailableRentalTimesResponse
// @Failure 400 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Router /settings/rental-times [put]
func (h *Handler) UpdateAvailableRentalTimes(c *gin.Context) {
	var req models.UpdateAvailableRentalTimesRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	resp, err := h.service.UpdateAvailableRentalTimes(&req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, resp)
}

// AdminGetSettings получает все настройки сервиса (админка)
func (h *Handler) AdminGetSettings(c *gin.Context) {
	serviceType := c.Query("service_type")
	if serviceType == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "service_type is required"})
		return
	}

	req := &models.AdminGetSettingsRequest{
		ServiceType: serviceType,
	}

	resp, err := h.service.GetSettings(req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, resp)
}

// AdminUpdatePrices обновляет цены сервиса (админка)
func (h *Handler) AdminUpdatePrices(c *gin.Context) {
	var req models.AdminUpdatePricesRequest

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	resp, err := h.service.UpdatePrices(&req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, resp)
}

// AdminUpdateRentalTimes обновляет доступное время аренды (админка)
func (h *Handler) AdminUpdateRentalTimes(c *gin.Context) {
	var req models.AdminUpdateRentalTimesRequest

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	resp, err := h.service.UpdateRentalTimes(&req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, resp)
}
