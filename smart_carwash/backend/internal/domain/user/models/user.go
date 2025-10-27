package models

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// User представляет пользователя системы
type User struct {
	ID              uuid.UUID      `json:"id" gorm:"primaryKey;type:uuid;default:gen_random_uuid()"`
	TelegramID      int64          `json:"telegram_id" gorm:"uniqueIndex"`
	Username        string         `json:"username"`
	FirstName       string         `json:"first_name"`
	LastName        string         `json:"last_name"`
	CarNumber       string         `json:"car_number"`
	CarNumberCountry string        `json:"car_number_country" gorm:"default:'RUS'"`
	Email           string         `json:"email"`
	IsAdmin         bool           `json:"is_admin" gorm:"default:false"`
	CreatedAt       time.Time      `json:"created_at"`
	UpdatedAt       time.Time      `json:"updated_at"`
	DeletedAt       gorm.DeletedAt `json:"-" gorm:"index"`
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

// Административные модели

// AdminListUsersRequest запрос на получение списка пользователей
type AdminListUsersRequest struct {
	Limit  *int `json:"limit"`
	Offset *int `json:"offset"`
}

// AdminGetUserRequest запрос на получение пользователя по ID
type AdminGetUserRequest struct {
	ID uuid.UUID `json:"id" binding:"required"`
}

// AdminListUsersResponse ответ на получение списка пользователей
type AdminListUsersResponse struct {
	Users  []User `json:"users"`
	Total  int    `json:"total"`
	Limit  int    `json:"limit"`
	Offset int    `json:"offset"`
}

// AdminGetUserResponse ответ на получение пользователя
type AdminGetUserResponse struct {
	User User `json:"user"`
}

// UpdateCarNumberRequest запрос на обновление номера машины
type UpdateCarNumberRequest struct {
	UserID           uuid.UUID `json:"user_id" binding:"required"`
	CarNumber        string    `json:"car_number" binding:"required"`
	CarNumberCountry string    `json:"car_number_country"`
}

// UpdateCarNumberResponse ответ на обновление номера машины
type UpdateCarNumberResponse struct {
	Success bool `json:"success"`
	User    User `json:"user"`
}

// UpdateEmailRequest запрос на обновление email
type UpdateEmailRequest struct {
	UserID uuid.UUID `json:"user_id" binding:"required"`
	Email  string    `json:"email" binding:"required"`
}

// UpdateEmailResponse ответ на обновление email
type UpdateEmailResponse struct {
	Success bool `json:"success"`
	User    User `json:"user"`
}
