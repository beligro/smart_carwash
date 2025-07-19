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
		settingsGroup.GET("/service-setting", h.GetServiceSetting)
		settingsGroup.PUT("/service-setting", h.UpdateServiceSetting)
		settingsGroup.GET("/all", h.GetAllSettings)
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

// GetServiceSetting получает настройку для определенного типа услуги и ключа
// @Summary Получить настройку услуги
// @Description Получает значение настройки для определенного типа услуги и ключа
// @Tags settings
// @Accept json
// @Produce json
// @Param service_type query string true "Тип услуги"
// @Param setting_key query string true "Ключ настройки"
// @Success 200 {object} map[string]interface{}
// @Failure 400 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Router /settings/service-setting [get]
func (h *Handler) GetServiceSetting(c *gin.Context) {
	serviceType := c.Query("service_type")
	settingKey := c.Query("setting_key")
	
	if serviceType == "" || settingKey == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "service_type and setting_key are required"})
		return
	}

	value, err := h.service.GetServicePrice(serviceType, settingKey)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"value": value})
}

// UpdateServiceSetting обновляет настройку для определенного типа услуги и ключа
// @Summary Обновить настройку услуги
// @Description Обновляет значение настройки для определенного типа услуги и ключа
// @Tags settings
// @Accept json
// @Produce json
// @Param request body map[string]interface{} true "Запрос на обновление настройки"
// @Success 200 {object} map[string]bool
// @Failure 400 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Router /settings/service-setting [put]
func (h *Handler) UpdateServiceSetting(c *gin.Context) {
	var req struct {
		ServiceType  string `json:"service_type" binding:"required"`
		SettingKey   string `json:"setting_key" binding:"required"`
		SettingValue int    `json:"setting_value" binding:"required"`
	}
	
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	err := h.service.UpdateServiceSetting(req.ServiceType, req.SettingKey, req.SettingValue)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"success": true})
}

// GetAllSettings получает все настройки системы
// @Summary Получить все настройки
// @Description Получает все настройки системы: цены, время аренды, системные параметры
// @Tags settings
// @Accept json
// @Produce json
// @Success 200 {object} map[string]interface{}
// @Failure 500 {object} map[string]string
// @Router /settings/all [get]
func (h *Handler) GetAllSettings(c *gin.Context) {
	// Получаем цены для всех типов услуг
	prices := make(map[string]map[string]int)
	serviceTypes := []string{"wash", "air_dry", "vacuum"}
	priceKeys := []string{"price_per_minute", "chemistry_price_per_minute"}
	
	for _, serviceType := range serviceTypes {
		prices[serviceType] = make(map[string]int)
		for _, priceKey := range priceKeys {
			price, err := h.service.GetServicePrice(serviceType, priceKey)
			if err != nil {
				c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
				return
			}
			prices[serviceType][priceKey] = price
		}
	}
	
	// Получаем время аренды для всех типов услуг
	rentalTimes := make(map[string][]int)
	for _, serviceType := range serviceTypes {
		times, err := h.service.GetAvailableRentalTimes(&models.GetAvailableRentalTimesRequest{
			ServiceType: serviceType,
		})
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		rentalTimes[serviceType] = times.AvailableTimes
	}
	
	// Получаем системные настройки
	systemSettings := make(map[string]int)
	systemKeys := []string{"max_queue_size", "session_timeout_minutes", "reservation_timeout_minutes", "notification_enabled"}
	
	for _, key := range systemKeys {
		value, err := h.service.GetServicePrice("system", key)
		if err != nil {
			// Для системных настроек используем дефолтные значения
			switch key {
			case "max_queue_size":
				value = 10
			case "session_timeout_minutes":
				value = 5
			case "reservation_timeout_minutes":
				value = 3
			case "notification_enabled":
				value = 1
			}
		}
		systemSettings[key] = value
	}
	
	// Формируем ответ
	response := gin.H{
		"prices": prices,
		"rental_times": rentalTimes,
		"system_settings": systemSettings,
	}
	
	c.JSON(http.StatusOK, response)
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
