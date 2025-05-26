package models

import (
	sessionModels "carwash_backend/internal/domain/session/models"
	washboxModels "carwash_backend/internal/domain/washbox/models"
)

// QueueStatus представляет статус очереди и боксов
type QueueStatus struct {
	Boxes     []washboxModels.WashBox `json:"boxes"`
	QueueSize int                     `json:"queue_size"`
	HasQueue  bool                    `json:"has_queue"`
}

// WashInfo представляет информацию о мойке для пользователя
type WashInfo struct {
	Boxes       []washboxModels.WashBox `json:"boxes"`
	QueueSize   int                     `json:"queue_size"`
	HasQueue    bool                    `json:"has_queue"`
	UserSession *sessionModels.Session  `json:"user_session,omitempty"`
}
