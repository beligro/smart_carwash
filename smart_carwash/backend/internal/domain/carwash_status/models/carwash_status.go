package models

import (
	"time"

	"github.com/google/uuid"
)

// CarwashStatus представляет текущий статус мойки (singleton)
type CarwashStatus struct {
	ID           uuid.UUID  `json:"id" gorm:"primaryKey;type:uuid;default:gen_random_uuid()"`
	IsClosed     bool       `json:"is_closed" gorm:"default:false;not null"`
	ClosedReason *string    `json:"closed_reason,omitempty"`
	UpdatedAt    time.Time  `json:"updated_at" gorm:"not null"`
	UpdatedBy    *uuid.UUID `json:"updated_by,omitempty" gorm:"type:uuid"`
	CreatedAt    time.Time  `json:"created_at" gorm:"not null"`
}

// TableName указывает имя таблицы для GORM
func (CarwashStatus) TableName() string {
	return "carwash_status"
}

// CarwashStatusHistory представляет запись истории изменений статуса мойки
type CarwashStatusHistory struct {
	ID           uuid.UUID  `json:"id" gorm:"primaryKey;type:uuid;default:gen_random_uuid()"`
	IsClosed     bool       `json:"is_closed" gorm:"not null"`
	ClosedReason *string    `json:"closed_reason,omitempty"`
	CreatedAt    time.Time  `json:"created_at" gorm:"not null"`
	CreatedBy    *uuid.UUID `json:"created_by,omitempty" gorm:"type:uuid"`
}

// TableName указывает имя таблицы для GORM
func (CarwashStatusHistory) TableName() string {
	return "carwash_status_history"
}

// GetCurrentStatusRequest запрос на получение текущего статуса
type GetCurrentStatusRequest struct{}

// GetCurrentStatusResponse ответ на получение текущего статуса
type GetCurrentStatusResponse struct {
	IsClosed     bool      `json:"is_closed"`
	ClosedReason *string   `json:"closed_reason,omitempty"`
	UpdatedAt    time.Time `json:"updated_at"`
}

// CloseCarwashRequest запрос на закрытие мойки
type CloseCarwashRequest struct {
	Reason *string `json:"reason,omitempty"` // Опциональная причина закрытия
}

// CloseCarwashResponse ответ на закрытие мойки
type CloseCarwashResponse struct {
	Success           bool   `json:"success"`
	Message           string `json:"message"`
	CompletedSessions int    `json:"completed_sessions"` // Количество завершенных активных сессий
	CanceledSessions  int    `json:"canceled_sessions"`  // Количество отмененных сессий с возвратом
}

// OpenCarwashRequest запрос на открытие мойки
type OpenCarwashRequest struct{}

// OpenCarwashResponse ответ на открытие мойки
type OpenCarwashResponse struct {
	Success bool   `json:"success"`
	Message string `json:"message"`
}

// GetHistoryRequest запрос на получение истории изменений статуса
type GetHistoryRequest struct {
	Limit  int `json:"limit" form:"limit"`   // Лимит записей (по умолчанию 50)
	Offset int `json:"offset" form:"offset"` // Смещение (по умолчанию 0)
}

// GetHistoryResponse ответ на получение истории изменений статуса
type GetHistoryResponse struct {
	History []CarwashStatusHistory `json:"history"`
	Total   int                    `json:"total"`
}
