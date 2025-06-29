package models

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// Статусы сессии
const (
	SessionStatusCreated  = "created"  // Создана
	SessionStatusAssigned = "assigned" // Назначена на бокс
	SessionStatusActive   = "active"   // Активна (клиент приступил к мойке)
	SessionStatusComplete = "complete" // Завершена
	SessionStatusCanceled = "canceled" // Отменена
	SessionStatusExpired  = "expired"  // Истек срок резервирования
)

// Session представляет сессию мойки
type Session struct {
	ID                           uuid.UUID      `json:"id" gorm:"primaryKey;type:uuid;default:gen_random_uuid()"`
	UserID                       uuid.UUID      `json:"user_id" gorm:"index;type:uuid"`
	BoxID                        *uuid.UUID     `json:"box_id,omitempty" gorm:"index;type:uuid"`
	BoxNumber                    *int           `json:"box_number,omitempty" gorm:"-"` // Виртуальное поле, не хранится в БД
	Status                       string         `json:"status" gorm:"default:created;index"`
	ServiceType                  string         `json:"service_type,omitempty" gorm:"default:null"`
	WithChemistry                bool           `json:"with_chemistry" gorm:"default:false"`
	RentalTimeMinutes            int            `json:"rental_time_minutes" gorm:"default:5"`    // Время аренды в минутах
	ExtensionTimeMinutes         int            `json:"extension_time_minutes" gorm:"default:0"` // Время продления в минутах
	IdempotencyKey               string         `json:"idempotency_key,omitempty" gorm:"index"`
	IsExpiringNotificationSent   bool           `json:"is_expiring_notification_sent" gorm:"default:false"`
	IsCompletingNotificationSent bool           `json:"is_completing_notification_sent" gorm:"default:false"`
	CreatedAt                    time.Time      `json:"created_at"`
	UpdatedAt                    time.Time      `json:"updated_at"`
	StatusUpdatedAt              time.Time      `json:"status_updated_at"` // Время последнего обновления статуса
	DeletedAt                    gorm.DeletedAt `json:"-" gorm:"index"`
}

// CreateSessionRequest представляет запрос на создание сессии
type CreateSessionRequest struct {
	UserID            uuid.UUID `json:"user_id" binding:"required"`
	ServiceType       string    `json:"service_type" binding:"required"`
	WithChemistry     bool      `json:"with_chemistry"`
	RentalTimeMinutes int       `json:"rental_time_minutes" binding:"required"`
	IdempotencyKey    string    `json:"idempotency_key" binding:"required"`
}

// CreateSessionResponse представляет ответ на создание сессии
type CreateSessionResponse struct {
	Session Session `json:"session"`
}

// GetUserSessionRequest представляет запрос на получение сессии пользователя
type GetUserSessionRequest struct {
	UserID uuid.UUID `json:"user_id" binding:"required"`
}

// GetUserSessionResponse представляет ответ на получение сессии пользователя
type GetUserSessionResponse struct {
	Session *Session `json:"session"`
}

// GetSessionRequest представляет запрос на получение сессии по ID
type GetSessionRequest struct {
	SessionID uuid.UUID `json:"session_id" binding:"required"`
}

// GetSessionResponse представляет ответ на получение сессии
type GetSessionResponse struct {
	Session *Session `json:"session"`
}

// StartSessionRequest представляет запрос на запуск сессии
type StartSessionRequest struct {
	SessionID uuid.UUID `json:"session_id" binding:"required"`
}

// StartSessionResponse представляет ответ на запуск сессии
type StartSessionResponse struct {
	Session *Session `json:"session"`
}

// CompleteSessionRequest представляет запрос на завершение сессии
type CompleteSessionRequest struct {
	SessionID uuid.UUID `json:"session_id" binding:"required"`
}

// CompleteSessionResponse представляет ответ на завершение сессии
type CompleteSessionResponse struct {
	Session *Session `json:"session"`
}

// GetUserSessionHistoryRequest представляет запрос на получение истории сессий пользователя
type GetUserSessionHistoryRequest struct {
	UserID uuid.UUID `json:"user_id" binding:"required"`
	Limit  int       `json:"limit"`
	Offset int       `json:"offset"`
}

// GetUserSessionHistoryResponse представляет ответ на получение истории сессий пользователя
type GetUserSessionHistoryResponse struct {
	Sessions []Session `json:"sessions"`
}

// ExtendSessionRequest представляет запрос на продление сессии
type ExtendSessionRequest struct {
	SessionID            uuid.UUID `json:"session_id" binding:"required"`
	ExtensionTimeMinutes int       `json:"extension_time_minutes" binding:"required"`
}

// ExtendSessionResponse представляет ответ на продление сессии
type ExtendSessionResponse struct {
	Session *Session `json:"session"`
}

// Административные модели

// AdminListSessionsRequest запрос на получение списка сессий с фильтрацией
type AdminListSessionsRequest struct {
	UserID      *uuid.UUID `json:"user_id"`
	BoxID       *uuid.UUID `json:"box_id"`
	BoxNumber   *int       `json:"box_number"`
	Status      *string    `json:"status" binding:"omitempty,oneof=created assigned active complete canceled expired"`
	ServiceType *string    `json:"service_type" binding:"omitempty,oneof=wash air_dry vacuum"`
	DateFrom    *time.Time `json:"date_from"`
	DateTo      *time.Time `json:"date_to"`
	Limit       *int       `json:"limit"`
	Offset      *int       `json:"offset"`
}

// AdminGetSessionRequest запрос на получение сессии по ID
type AdminGetSessionRequest struct {
	ID uuid.UUID `json:"id" binding:"required"`
}

// AdminListSessionsResponse ответ на получение списка сессий
type AdminListSessionsResponse struct {
	Sessions []Session `json:"sessions"`
	Total    int       `json:"total"`
	Limit    int       `json:"limit"`
	Offset   int       `json:"offset"`
}

// AdminGetSessionResponse ответ на получение сессии
type AdminGetSessionResponse struct {
	Session Session `json:"session"`
}
