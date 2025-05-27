package service

import (
	"carwash_backend/internal/domain/washbox/models"
	"carwash_backend/internal/domain/washbox/repository"

	"github.com/google/uuid"
)

// Service интерфейс для бизнес-логики боксов мойки
type Service interface {
	GetWashBoxByID(id uuid.UUID) (*models.WashBox, error)
	UpdateWashBoxStatus(id uuid.UUID, status string) error
	GetFreeWashBoxes() ([]models.WashBox, error)
	GetAllWashBoxes() ([]models.WashBox, error)
}

// ServiceImpl реализация Service
type ServiceImpl struct {
	repo repository.Repository
}

// NewService создает новый экземпляр Service
func NewService(repo repository.Repository) *ServiceImpl {
	return &ServiceImpl{
		repo: repo,
	}
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
