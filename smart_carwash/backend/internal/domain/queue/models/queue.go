package models

import (
	washboxModels "carwash_backend/internal/domain/washbox/models"
)

// QueueStatus представляет статус очереди и боксов
type QueueStatus struct {
	Boxes     []washboxModels.WashBox `json:"boxes"`
	QueueSize int                     `json:"queue_size"`
	HasQueue  bool                    `json:"has_queue"`
}
