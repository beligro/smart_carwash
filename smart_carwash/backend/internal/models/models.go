package models

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// Статусы бокса мойки
const (
	StatusFree        = "free"        // Свободен
	StatusReserved    = "reserved"    // Зарезервирован (назначен на пользователя, но сессия не запущена)
	StatusBusy        = "busy"        // Занят (сессия активна)
	StatusMaintenance = "maintenance" // На обслуживании
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

// User представляет пользователя системы
type User struct {
	ID         uuid.UUID      `json:"id" gorm:"primaryKey;type:uuid;default:gen_random_uuid()"`
	TelegramID int64          `json:"telegram_id" gorm:"uniqueIndex"`
	Username   string         `json:"username"`
	FirstName  string         `json:"first_name"`
	LastName   string         `json:"last_name"`
	IsAdmin    bool           `json:"is_admin" gorm:"default:false"`
	CreatedAt  time.Time      `json:"created_at"`
	UpdatedAt  time.Time      `json:"updated_at"`
	DeletedAt  gorm.DeletedAt `json:"-" gorm:"index"`
}

// WashBox представляет бокс автомойки
type WashBox struct {
	ID        uuid.UUID      `json:"id" gorm:"primaryKey;type:uuid;default:gen_random_uuid()"`
	Number    int            `json:"number" gorm:"uniqueIndex"`
	Status    string         `json:"status" gorm:"default:free"`
	CreatedAt time.Time      `json:"created_at"`
	UpdatedAt time.Time      `json:"updated_at"`
	DeletedAt gorm.DeletedAt `json:"-" gorm:"index"`
}

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

// WashInfo представляет информацию о мойке для API
type WashInfo struct {
	Boxes       []WashBox `json:"boxes"`
	QueueSize   int       `json:"queue_size"`
	HasQueue    bool      `json:"has_queue"`
	UserSession *Session  `json:"user_session,omitempty"`
}

// CreateUserRequest представляет запрос на создание пользователя
type CreateUserRequest struct {
	TelegramID     int64  `json:"telegram_id" binding:"required"`
	Username       string `json:"username"`
	FirstName      string `json:"first_name"`
	LastName       string `json:"last_name"`
	IdempotencyKey string `json:"idempotency_key" binding:"required"`
}

// CreateUserResponse представляет ответ на создание пользователя
type CreateUserResponse struct {
	User User `json:"user"`
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

// GetUserByTelegramIDRequest представляет запрос на получение пользователя по telegram_id
type GetUserByTelegramIDRequest struct {
	TelegramID int64 `json:"telegram_id" binding:"required"`
}

// GetUserByTelegramIDResponse представляет ответ на получение пользователя по telegram_id
type GetUserByTelegramIDResponse struct {
	User User `json:"user"`
}

// GetQueueStatusRequest представляет запрос на получение статуса очереди и боксов
type GetQueueStatusRequest struct {
}

// GetQueueStatusResponse представляет ответ на получение статуса очереди и боксов
type GetQueueStatusResponse struct {
	Boxes     []WashBox `json:"boxes"`
	QueueSize int       `json:"queue_size"`
	HasQueue  bool      `json:"has_queue"`
}
