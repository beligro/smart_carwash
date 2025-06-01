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

// UpdateAvailableRentalTimesResponse представляет ответ на обновление доступного времени аренды
type UpdateAvailableRentalTimesResponse struct {
	Success bool `json:"success"`
}
