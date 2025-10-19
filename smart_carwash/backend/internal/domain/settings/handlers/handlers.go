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
		settingsGroup.GET("/available-chemistry-times", h.GetAvailableChemistryTimes)
	}

	// Административные маршруты
	adminSettingsGroup := router.Group("/admin/settings")
	{
		adminSettingsGroup.GET("", h.AdminGetSettings)
		adminSettingsGroup.PUT("/prices", h.AdminUpdatePrices)
		adminSettingsGroup.PUT("/rental-times", h.AdminUpdateRentalTimes)
		adminSettingsGroup.GET("/available-chemistry-times", h.AdminGetAvailableChemistryTimes)
		adminSettingsGroup.PUT("/available-chemistry-times", h.AdminUpdateAvailableChemistryTimes)
		adminSettingsGroup.GET("/cleaning-timeout", h.AdminGetCleaningTimeout)
		adminSettingsGroup.PUT("/cleaning-timeout", h.AdminUpdateCleaningTimeout)
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

// GetAvailableChemistryTimes получает доступное время химии для определенного типа услуги (публичный)
// @Summary Получить доступное время химии
// @Description Получает список доступного времени химии для определенного типа услуги
// @Tags settings
// @Accept json
// @Produce json
// @Param service_type query string true "Тип услуги"
// @Success 200 {object} models.GetAvailableChemistryTimesResponse
// @Failure 400 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Router /settings/available-chemistry-times [get]
func (h *Handler) GetAvailableChemistryTimes(c *gin.Context) {
	serviceType := c.Query("service_type")
	if serviceType == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "service_type is required"})
		return
	}

	req := &models.GetAvailableChemistryTimesRequest{
		ServiceType: serviceType,
	}

	resp, err := h.service.GetAvailableChemistryTimes(req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, resp)
}

// AdminGetAvailableChemistryTimes получает доступное время химии для определенного типа услуги (админка)
func (h *Handler) AdminGetAvailableChemistryTimes(c *gin.Context) {
	serviceType := c.Query("service_type")
	if serviceType == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "service_type is required"})
		return
	}

	req := &models.AdminGetAvailableChemistryTimesRequest{
		ServiceType: serviceType,
	}

	resp, err := h.service.AdminGetAvailableChemistryTimes(req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, resp)
}

// AdminUpdateAvailableChemistryTimes обновляет доступное время химии (админка)
func (h *Handler) AdminUpdateAvailableChemistryTimes(c *gin.Context) {
	var req models.AdminUpdateAvailableChemistryTimesRequest

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	resp, err := h.service.AdminUpdateAvailableChemistryTimes(&req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, resp)
}

// AdminGetCleaningTimeout получает время уборки (админка)
func (h *Handler) AdminGetCleaningTimeout(c *gin.Context) {
	timeout, err := h.service.GetCleaningTimeout()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	resp := &models.AdminGetCleaningTimeoutResponse{
		TimeoutMinutes: timeout,
	}

	c.JSON(http.StatusOK, resp)
}

// AdminUpdateCleaningTimeout обновляет время уборки (админка)
func (h *Handler) AdminUpdateCleaningTimeout(c *gin.Context) {
	var req models.AdminUpdateCleaningTimeoutRequest

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	err := h.service.UpdateCleaningTimeout(req.TimeoutMinutes)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	resp := &models.AdminUpdateCleaningTimeoutResponse{
		Success: true,
	}

	c.JSON(http.StatusOK, resp)
}
