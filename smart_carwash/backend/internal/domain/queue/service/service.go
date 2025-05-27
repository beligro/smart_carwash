package service

import (
	"carwash_backend/internal/domain/queue/models"
	sessionModels "carwash_backend/internal/domain/session/models"
	sessionService "carwash_backend/internal/domain/session/service"
	washboxModels "carwash_backend/internal/domain/washbox/models"
	washboxService "carwash_backend/internal/domain/washbox/service"
)

// Service интерфейс для бизнес-логики очереди
type Service interface {
	GetQueueStatus() (*models.QueueStatus, error)
}

// ServiceImpl реализация Service
type ServiceImpl struct {
	sessionService sessionService.Service
	washboxService washboxService.Service
}

// NewService создает новый экземпляр Service
func NewService(sessionService sessionService.Service, washboxService washboxService.Service) *ServiceImpl {
	return &ServiceImpl{
		sessionService: sessionService,
		washboxService: washboxService,
	}
}

// GetQueueStatus получает статус очереди и боксов
func (s *ServiceImpl) GetQueueStatus() (*models.QueueStatus, error) {
	// Получаем все боксы мойки
	boxes, err := s.washboxService.GetAllWashBoxes()
	if err != nil {
		return nil, err
	}

	// Получаем количество сессий в очереди
	queueSize, err := s.sessionService.CountSessionsByStatus(sessionModels.SessionStatusCreated)
	if err != nil {
		return nil, err
	}

	// Получаем количество свободных боксов
	freeBoxes, err := s.washboxService.GetFreeWashBoxes()
	if err != nil {
		return nil, err
	}

	// Определяем, есть ли очередь
	hasQueue := queueSize > len(freeBoxes)

	// Формируем ответ
	return &models.QueueStatus{
		Boxes:     boxes,
		QueueSize: queueSize,
		HasQueue:  hasQueue,
	}, nil
}

// GetAllWashBoxes получает все боксы мойки
func (s *ServiceImpl) GetAllWashBoxes() ([]washboxModels.WashBox, error) {
	return s.washboxService.GetAllWashBoxes()
}
