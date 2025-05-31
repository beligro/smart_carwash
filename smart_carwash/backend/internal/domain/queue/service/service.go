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

// getServiceQueueInfo получает информацию об очереди для конкретного типа услуги
func (s *ServiceImpl) getServiceQueueInfo(serviceType string) (*models.ServiceQueueInfo, error) {
	// Получаем боксы для данного типа услуги
	boxes, err := s.washboxService.GetWashBoxesByServiceType(serviceType)
	if err != nil {
		return nil, err
	}

	// Получаем сессии со статусом "created"
	createdSessions, err := s.sessionService.GetSessionsByStatus(sessionModels.SessionStatusCreated)
	if err != nil {
		return nil, err
	}

	// Подсчитываем количество сессий для данного типа услуги
	queueSize := 0
	for _, session := range createdSessions {
		if session.ServiceType == serviceType {
			queueSize++
		}
	}

	// Получаем свободные боксы для данного типа услуги
	freeBoxes, err := s.washboxService.GetFreeWashBoxesByServiceType(serviceType)
	if err != nil {
		return nil, err
	}

	// Определяем, есть ли очередь
	hasQueue := queueSize > len(freeBoxes)

	// Формируем ответ
	return &models.ServiceQueueInfo{
		ServiceType: serviceType,
		Boxes:       boxes,
		QueueSize:   queueSize,
		HasQueue:    hasQueue,
	}, nil
}

// GetQueueStatus получает статус очереди и боксов
func (s *ServiceImpl) GetQueueStatus() (*models.QueueStatus, error) {
	// Получаем все боксы мойки
	allBoxes, err := s.washboxService.GetAllWashBoxes()
	if err != nil {
		return nil, err
	}

	// Получаем информацию об очереди для каждого типа услуги
	washQueueInfo, err := s.getServiceQueueInfo(washboxModels.ServiceTypeWash)
	if err != nil {
		return nil, err
	}

	airDryQueueInfo, err := s.getServiceQueueInfo(washboxModels.ServiceTypeAirDry)
	if err != nil {
		return nil, err
	}

	vacuumQueueInfo, err := s.getServiceQueueInfo(washboxModels.ServiceTypeVacuum)
	if err != nil {
		return nil, err
	}

	// Подсчитываем общее количество сессий в очереди
	totalQueueSize := washQueueInfo.QueueSize + airDryQueueInfo.QueueSize + vacuumQueueInfo.QueueSize

	// Определяем, есть ли очередь хотя бы для одного типа услуги
	hasAnyQueue := washQueueInfo.HasQueue || airDryQueueInfo.HasQueue || vacuumQueueInfo.HasQueue

	// Формируем ответ
	return &models.QueueStatus{
		AllBoxes:       allBoxes,
		WashQueue:      *washQueueInfo,
		AirDryQueue:    *airDryQueueInfo,
		VacuumQueue:    *vacuumQueueInfo,
		TotalQueueSize: totalQueueSize,
		HasAnyQueue:    hasAnyQueue,
	}, nil
}
