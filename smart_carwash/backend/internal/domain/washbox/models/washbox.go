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

// WashBox представляет бокс автомойки
type WashBox struct {
	ID        uuid.UUID      `json:"id" gorm:"primaryKey;type:uuid;default:gen_random_uuid()"`
	Number    int            `json:"number" gorm:"uniqueIndex"`
	Status    string         `json:"status" gorm:"default:free"`
	CreatedAt time.Time      `json:"created_at"`
	UpdatedAt time.Time      `json:"updated_at"`
	DeletedAt gorm.DeletedAt `json:"-" gorm:"index"`
}

// GetQueueStatusResponse представляет ответ на получение статуса очереди и боксов
type GetQueueStatusResponse struct {
	Boxes     []WashBox `json:"boxes"`
	QueueSize int       `json:"queue_size"`
	HasQueue  bool      `json:"has_queue"`
}
