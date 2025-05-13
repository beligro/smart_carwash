package models

import (
	"time"

	"gorm.io/gorm"
)

// Статусы бокса мойки
const (
	StatusFree        = "free"        // Свободен
	StatusBusy        = "busy"        // Занят
	StatusMaintenance = "maintenance" // На обслуживании
)

// User представляет пользователя системы
type User struct {
	ID         uint           `json:"id" gorm:"primaryKey"`
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
	ID        uint           `json:"id" gorm:"primaryKey"`
	Number    int            `json:"number" gorm:"uniqueIndex"`
	Status    string         `json:"status" gorm:"default:free"`
	CreatedAt time.Time      `json:"created_at"`
	UpdatedAt time.Time      `json:"updated_at"`
	DeletedAt gorm.DeletedAt `json:"-" gorm:"index"`
}

// WashInfo представляет информацию о мойке для API
type WashInfo struct {
	Boxes []WashBox `json:"boxes"`
}

// CreateUserRequest представляет запрос на создание пользователя
type CreateUserRequest struct {
	TelegramID int64  `json:"telegram_id" binding:"required"`
	Username   string `json:"username"`
	FirstName  string `json:"first_name"`
	LastName   string `json:"last_name"`
}

// CreateUserResponse представляет ответ на создание пользователя
type CreateUserResponse struct {
	User User `json:"user"`
}
