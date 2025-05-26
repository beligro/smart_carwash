package service

import (
	"carwash_backend/internal/domain/washbox/models"
	"carwash_backend/internal/domain/washbox/repository"

	"github.com/google/uuid"
)

// SessionCounter интерфейс для подсчета сессий
type SessionCounter interface {
	CountSessionsByStatus(status string) (int, error)
}

// Service интерфейс для бизнес-логики боксов мойки
type Service interface {
	GetWashBoxByID(id uuid.UUID) (*models.WashBox, error)
	UpdateWashBoxStatus(id uuid.UUID, status string) error
	GetFreeWashBoxes() ([]models.WashBox, error)
	GetAllWashBoxes() ([]models.WashBox, error)
}

// ServiceImpl реализация Service
type ServiceImpl struct {
	repo           repository.Repository
	sessionCounter SessionCounter
}

// NewService создает новый экземпляр Service
func NewService(repo repository.Repository, sessionCounter SessionCounter) *ServiceImpl {
	return &ServiceImpl{
		repo:           repo,
		sessionCounter: sessionCounter,
	}
}

// GetQueueStatus получает статус очереди и боксов
func (s *ServiceImpl) GetQueueStatus() (*models.GetQueueStatusResponse, error) {
	// Получаем все боксы мойки
	boxes, err := s.repo.GetAllWashBoxes()
	if err != nil {
		return nil, err
	}

	// Получаем количество сессий в очереди
	queueSize, err := s.sessionCounter.CountSessionsByStatus("created")
	if err != nil {
		return nil, err
	}

	// Получаем количество свободных боксов
	freeBoxes, err := s.repo.GetFreeWashBoxes()
	if err != nil {
		return nil, err
	}

	// Определяем, есть ли очередь
	hasQueue := queueSize > len(freeBoxes)

	// Формируем ответ
	return &models.GetQueueStatusResponse{
		Boxes:     boxes,
		QueueSize: queueSize,
		HasQueue:  hasQueue,
	}, nil
}

// GetWashBoxByID получает бокс мойки по ID
func (s *ServiceImpl) GetWashBoxByID(id uuid.UUID) (*models.WashBox, error) {
	return s.repo.GetWashBoxByID(id)
}

// UpdateWashBoxStatus обновляет статус бокса мойки
func (s *ServiceImpl) UpdateWashBoxStatus(id uuid.UUID, status string) error {
	return s.repo.UpdateWashBoxStatus(id, status)
}

// GetFreeWashBoxes получает все свободные боксы мойки
func (s *ServiceImpl) GetFreeWashBoxes() ([]models.WashBox, error) {
	return s.repo.GetFreeWashBoxes()
}

// GetAllWashBoxes получает все боксы мойки
func (s *ServiceImpl) GetAllWashBoxes() ([]models.WashBox, error) {
	return s.repo.GetAllWashBoxes()
}
