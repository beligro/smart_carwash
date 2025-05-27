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
	ID             uuid.UUID      `json:"id" gorm:"primaryKey;type:uuid;default:gen_random_uuid()"`
	UserID         uuid.UUID      `json:"user_id" gorm:"index;type:uuid"`
	BoxID          *uuid.UUID     `json:"box_id,omitempty" gorm:"index;type:uuid"`
	Status         string         `json:"status" gorm:"default:created;index"`
	IdempotencyKey string         `json:"idempotency_key,omitempty" gorm:"index"`
	CreatedAt      time.Time      `json:"created_at"`
	UpdatedAt      time.Time      `json:"updated_at"`
	DeletedAt      gorm.DeletedAt `json:"-" gorm:"index"`
}

// CreateSessionRequest представляет запрос на создание сессии
type CreateSessionRequest struct {
	UserID         uuid.UUID `json:"user_id" binding:"required"`
	IdempotencyKey string    `json:"idempotency_key" binding:"required"`
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
