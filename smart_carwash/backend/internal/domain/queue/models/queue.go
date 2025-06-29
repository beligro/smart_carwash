package models

import (
	washboxModels "carwash_backend/internal/domain/washbox/models"
)

// ServiceQueueInfo представляет информацию об очереди для конкретного типа услуги
type ServiceQueueInfo struct {
	ServiceType  string                  `json:"service_type"`
	Boxes        []washboxModels.WashBox `json:"boxes"`
	QueueSize    int                     `json:"queue_size"`
	HasQueue     bool                    `json:"has_queue"`
	UsersInQueue []QueueUser             `json:"users_in_queue,omitempty"` // Список пользователей в очереди
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

// Административные модели

// AdminQueueStatusRequest запрос на получение статуса очереди для администратора
type AdminQueueStatusRequest struct {
	IncludeDetails bool `json:"include_details"`
}

// AdminQueueStatusResponse ответ на получение статуса очереди для администратора
type AdminQueueStatusResponse struct {
	QueueStatus QueueStatus   `json:"queue_status"`
	Details     *QueueDetails `json:"details,omitempty"`
}

// QueueDetails детальная информация об очереди
type QueueDetails struct {
	UsersInQueue []QueueUser `json:"users_in_queue"`
	QueueOrder   []string    `json:"queue_order"` // Список ID пользователей в порядке очереди
}

// QueueUser информация о пользователе в очереди
type QueueUser struct {
	UserID       string `json:"user_id"`
	Username     string `json:"username"`
	FirstName    string `json:"first_name"`
	LastName     string `json:"last_name"`
	ServiceType  string `json:"service_type"`
	Position     int    `json:"position"`
	WaitingSince string `json:"waiting_since"` // Время добавления в очередь
}
