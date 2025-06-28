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

// Типы услуг
const (
	ServiceTypeWash   = "wash"    // Мойка
	ServiceTypeAirDry = "air_dry" // Обдув воздухом
	ServiceTypeVacuum = "vacuum"  // Пылесос
)

// WashBox представляет бокс автомойки
type WashBox struct {
	ID          uuid.UUID      `json:"id" gorm:"primaryKey;type:uuid;default:gen_random_uuid()"`
	Number      int            `json:"number" gorm:"uniqueIndex"`
	Status      string         `json:"status" gorm:"default:free"`
	ServiceType string         `json:"service_type" gorm:"default:wash"`
	CreatedAt   time.Time      `json:"created_at"`
	UpdatedAt   time.Time      `json:"updated_at"`
	DeletedAt   gorm.DeletedAt `json:"-" gorm:"index"`
}

// GetQueueStatusResponse представляет ответ на получение статуса очереди и боксов
type GetQueueStatusResponse struct {
	Boxes     []WashBox `json:"boxes"`
	QueueSize int       `json:"queue_size"`
	HasQueue  bool      `json:"has_queue"`
}

// AdminCreateWashBoxRequest запрос на создание бокса мойки
type AdminCreateWashBoxRequest struct {
	Number      int    `json:"number" binding:"required"`
	Status      string `json:"status" binding:"required,oneof=free reserved busy maintenance"`
	ServiceType string `json:"service_type" binding:"required,oneof=wash air_dry vacuum"`
}

// AdminUpdateWashBoxRequest запрос на обновление бокса мойки
type AdminUpdateWashBoxRequest struct {
	ID          uuid.UUID `json:"id" binding:"required"`
	Number      *int      `json:"number"`
	Status      *string   `json:"status" binding:"omitempty,oneof=free reserved busy maintenance"`
	ServiceType *string   `json:"service_type" binding:"omitempty,oneof=wash air_dry vacuum"`
}

// AdminDeleteWashBoxRequest запрос на удаление бокса мойки
type AdminDeleteWashBoxRequest struct {
	ID uuid.UUID `json:"id" binding:"required"`
}

// AdminGetWashBoxRequest запрос на получение бокса мойки
type AdminGetWashBoxRequest struct {
	ID uuid.UUID `json:"id" binding:"required"`
}

// AdminListWashBoxesRequest запрос на получение списка боксов мойки
type AdminListWashBoxesRequest struct {
	Status      *string `json:"status" binding:"omitempty,oneof=free reserved busy maintenance"`
	ServiceType *string `json:"service_type" binding:"omitempty,oneof=wash air_dry vacuum"`
	Limit       *int    `json:"limit"`
	Offset      *int    `json:"offset"`
}

// AdminCreateWashBoxResponse ответ на создание бокса мойки
type AdminCreateWashBoxResponse struct {
	WashBox WashBox `json:"wash_box"`
}

// AdminUpdateWashBoxResponse ответ на обновление бокса мойки
type AdminUpdateWashBoxResponse struct {
	WashBox WashBox `json:"wash_box"`
}

// AdminGetWashBoxResponse ответ на получение бокса мойки
type AdminGetWashBoxResponse struct {
	WashBox WashBox `json:"wash_box"`
}

// AdminListWashBoxesResponse ответ на получение списка боксов мойки
type AdminListWashBoxesResponse struct {
	WashBoxes []WashBox `json:"wash_boxes"`
	Total     int       `json:"total"`
	Limit     int       `json:"limit"`
	Offset    int       `json:"offset"`
}

// AdminDeleteWashBoxResponse ответ на удаление бокса мойки
type AdminDeleteWashBoxResponse struct {
	Message string `json:"message"`
}
