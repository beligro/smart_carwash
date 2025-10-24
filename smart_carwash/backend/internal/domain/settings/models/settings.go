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

// GetAvailableRentalTimesRequest представляет запрос на получение доступного времени мойки
type GetAvailableRentalTimesRequest struct {
	ServiceType string `json:"service_type" binding:"required"`
}

// GetAvailableRentalTimesResponse представляет ответ на получение доступного времени мойки
type GetAvailableRentalTimesResponse struct {
	AvailableTimes []int `json:"available_times"`
}

// UpdateAvailableRentalTimesRequest представляет запрос на обновление доступного времени мойки
type UpdateAvailableRentalTimesRequest struct {
	ServiceType    string `json:"service_type" binding:"required"`
	AvailableTimes []int  `json:"available_times" binding:"required"`
}

// UpdateAvailableRentalTimesResponse представляет ответ на обновление времени мойки
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

// AdminUpdateRentalTimesRequest запрос на обновление времени мойки (админка)
type AdminUpdateRentalTimesRequest struct {
	ServiceType          string `json:"service_type" binding:"required"`
	AvailableRentalTimes []int  `json:"available_rental_times" binding:"required"`
}

// AdminUpdateRentalTimesResponse ответ на обновление времени мойки (админка)
type AdminUpdateRentalTimesResponse struct {
	Success bool   `json:"success"`
	Message string `json:"message"`
}

// GetAvailableChemistryTimesRequest запрос на получение доступного времени химии (публичный)
type GetAvailableChemistryTimesRequest struct {
	ServiceType string `json:"service_type" binding:"required"`
}

// GetAvailableChemistryTimesResponse ответ на получение доступного времени химии (публичный)
type GetAvailableChemistryTimesResponse struct {
	AvailableChemistryTimes []int `json:"available_chemistry_times"`
}

// AdminGetAvailableChemistryTimesRequest запрос на получение доступного времени химии (админка)
type AdminGetAvailableChemistryTimesRequest struct {
	ServiceType string `json:"service_type" binding:"required"`
}

// AdminGetAvailableChemistryTimesResponse ответ на получение доступного времени химии (админка)
type AdminGetAvailableChemistryTimesResponse struct {
	ServiceType             string `json:"service_type"`
	AvailableChemistryTimes []int  `json:"available_chemistry_times"`
}

// AdminUpdateAvailableChemistryTimesRequest запрос на обновление доступного времени химии (админка)
type AdminUpdateAvailableChemistryTimesRequest struct {
	ServiceType             string `json:"service_type" binding:"required"`
	AvailableChemistryTimes []int  `json:"available_chemistry_times" binding:"required"`
}

// AdminUpdateAvailableChemistryTimesResponse ответ на обновление доступного времени химии (админка)
type AdminUpdateAvailableChemistryTimesResponse struct {
	Success bool   `json:"success"`
	Message string `json:"message"`
}

// AdminGetCleaningTimeoutRequest запрос на получение времени уборки (админка)
type AdminGetCleaningTimeoutRequest struct{}

// AdminGetCleaningTimeoutResponse ответ на получение времени уборки (админка)
type AdminGetCleaningTimeoutResponse struct {
	TimeoutMinutes int `json:"timeout_minutes"`
}

// AdminUpdateCleaningTimeoutRequest запрос на обновление времени уборки (админка)
type AdminUpdateCleaningTimeoutRequest struct {
	TimeoutMinutes int `json:"timeout_minutes" binding:"required,min=1,max=60"`
}

// AdminUpdateCleaningTimeoutResponse ответ на обновление времени уборки (админка)
type AdminUpdateCleaningTimeoutResponse struct {
	Success bool `json:"success"`
}

// AdminGetSessionTimeoutRequest запрос на получение времени ожидания старта мойки (админка)
type AdminGetSessionTimeoutRequest struct{}

// AdminGetSessionTimeoutResponse ответ на получение времени ожидания старта мойки (админка)
type AdminGetSessionTimeoutResponse struct {
	TimeoutMinutes int `json:"timeout_minutes"`
}

// AdminUpdateSessionTimeoutRequest запрос на обновление времени ожидания старта мойки (админка)
type AdminUpdateSessionTimeoutRequest struct {
	TimeoutMinutes int `json:"timeout_minutes" binding:"required,min=1,max=60"`
}

// AdminUpdateSessionTimeoutResponse ответ на обновление времени ожидания старта мойки (админка)
type AdminUpdateSessionTimeoutResponse struct {
	Success bool `json:"success"`
}
