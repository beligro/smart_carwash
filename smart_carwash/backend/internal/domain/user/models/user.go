package models

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
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

// GetUserByTelegramIDRequest представляет запрос на получение пользователя по telegram_id
type GetUserByTelegramIDRequest struct {
	TelegramID int64 `json:"telegram_id" binding:"required"`
}

// GetUserByTelegramIDResponse представляет ответ на получение пользователя по telegram_id
type GetUserByTelegramIDResponse struct {
	User User `json:"user"`
}
