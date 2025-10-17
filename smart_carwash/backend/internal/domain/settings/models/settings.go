package models

import (
	"encoding/json"
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// ServiceSetting представляет настройку для определенного типа услуги
type ServiceSetting struct {
	ID           uuid.UUID       `json:"id" gorm:"primaryKey;type:uuid;default:gen_random_uuid()"`
	ServiceType  string          `json:"service_type" gorm:"index;not null"`
	SettingKey   string          `json:"setting_key" gorm:"index;not null"`
	SettingValue json.RawMessage `json:"setting_value" gorm:"type:jsonb;not null"`
	CreatedAt    time.Time       `json:"created_at"`
	UpdatedAt    time.Time       `json:"updated_at"`
	DeletedAt    gorm.DeletedAt  `json:"-" gorm:"index"`
}

// GetAvailableRentalTimesRequest представляет запрос на получение доступного времени аренды
type GetAvailableRentalTimesRequest struct {
	ServiceType string `json:"service_type" binding:"required"`
}

// GetAvailableRentalTimesResponse представляет ответ на получение доступного времени аренды
type GetAvailableRentalTimesResponse struct {
	AvailableTimes []int `json:"available_times"`
}

// UpdateAvailableRentalTimesRequest представляет запрос на обновление доступного времени аренды
type UpdateAvailableRentalTimesRequest struct {
	ServiceType    string `json:"service_type" binding:"required"`
	AvailableTimes []int  `json:"available_times" binding:"required"`
}

// UpdateAvailableRentalTimesResponse представляет ответ на обновление времени аренды
type UpdateAvailableRentalTimesResponse struct {
	Success bool `json:"success"`
}

// AdminGetSettingsRequest запрос на получение всех настроек сервиса (админка)
type AdminGetSettingsRequest struct {
	ServiceType string `json:"service_type" binding:"required"`
}

// AdminGetSettingsResponse ответ на получение всех настроек сервиса (админка)
type AdminGetSettingsResponse struct {
	ServiceType             string `json:"service_type"`
	PricePerMinute          int    `json:"price_per_minute"`
	ChemistryPricePerMinute int    `json:"chemistry_price_per_minute"`
	AvailableRentalTimes    []int  `json:"available_rental_times"`
}

// AdminUpdatePricesRequest запрос на обновление цен (админка)
type AdminUpdatePricesRequest struct {
	ServiceType             string `json:"service_type" binding:"required"`
	PricePerMinute          int    `json:"price_per_minute" binding:"required"`
	ChemistryPricePerMinute *int   `json:"chemistry_price_per_minute"`
}

// AdminUpdatePricesResponse ответ на обновление цен (админка)
type AdminUpdatePricesResponse struct {
	Success bool   `json:"success"`
	Message string `json:"message"`
}

// AdminUpdateRentalTimesRequest запрос на обновление времени аренды (админка)
type AdminUpdateRentalTimesRequest struct {
	ServiceType          string `json:"service_type" binding:"required"`
	AvailableRentalTimes []int  `json:"available_rental_times" binding:"required"`
}

// AdminUpdateRentalTimesResponse ответ на обновление времени аренды (админка)
type AdminUpdateRentalTimesResponse struct {
	Success bool   `json:"success"`
	Message string `json:"message"`
}

// AdminGetChemistryTimeoutRequest запрос на получение времени доступности кнопки химии
type AdminGetChemistryTimeoutRequest struct {
	ServiceType string `json:"service_type" binding:"required"`
}

// AdminGetChemistryTimeoutResponse ответ на получение времени доступности кнопки химии
type AdminGetChemistryTimeoutResponse struct {
	ServiceType                   string `json:"service_type"`
	ChemistryEnableTimeoutMinutes int    `json:"chemistry_enable_timeout_minutes"`
}

// AdminUpdateChemistryTimeoutRequest запрос на обновление времени доступности кнопки химии
type AdminUpdateChemistryTimeoutRequest struct {
	ServiceType                   string `json:"service_type" binding:"required"`
	ChemistryEnableTimeoutMinutes int    `json:"chemistry_enable_timeout_minutes" binding:"required"`
}

// AdminUpdateChemistryTimeoutResponse ответ на обновление времени доступности кнопки химии
type AdminUpdateChemistryTimeoutResponse struct {
	Success bool   `json:"success"`
	Message string `json:"message"`
}
