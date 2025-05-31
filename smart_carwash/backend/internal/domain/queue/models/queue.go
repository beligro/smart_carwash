package models

import (
	washboxModels "carwash_backend/internal/domain/washbox/models"
)

// ServiceQueueInfo представляет информацию об очереди для конкретного типа услуги
type ServiceQueueInfo struct {
	ServiceType string                  `json:"service_type"`
	Boxes       []washboxModels.WashBox `json:"boxes"`
	QueueSize   int                     `json:"queue_size"`
	HasQueue    bool                    `json:"has_queue"`
}

// QueueStatus представляет статус очереди и боксов
type QueueStatus struct {
	AllBoxes       []washboxModels.WashBox `json:"all_boxes"`
	WashQueue      ServiceQueueInfo        `json:"wash_queue"`
	AirDryQueue    ServiceQueueInfo        `json:"air_dry_queue"`
	VacuumQueue    ServiceQueueInfo        `json:"vacuum_queue"`
	TotalQueueSize int                     `json:"total_queue_size"`
	HasAnyQueue    bool                    `json:"has_any_queue"`
}
