package service

import (
	"carwash_backend/internal/domain/settings/models"
	"carwash_backend/internal/domain/settings/repository"
)

// Service интерфейс для бизнес-логики настроек
type Service interface {
	GetAvailableRentalTimes(req *models.GetAvailableRentalTimesRequest) (*models.GetAvailableRentalTimesResponse, error)
	UpdateAvailableRentalTimes(req *models.UpdateAvailableRentalTimesRequest) (*models.UpdateAvailableRentalTimesResponse, error)
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

// GetAvailableRentalTimes получает доступное время аренды для определенного типа услуги
func (s *ServiceImpl) GetAvailableRentalTimes(req *models.GetAvailableRentalTimesRequest) (*models.GetAvailableRentalTimesResponse, error) {
	times, err := s.repo.GetAvailableRentalTimes(req.ServiceType)
	if err != nil {
		return nil, err
	}

	return &models.GetAvailableRentalTimesResponse{
		AvailableTimes: times,
	}, nil
}

// UpdateAvailableRentalTimes обновляет доступное время аренды для определенного типа услуги
func (s *ServiceImpl) UpdateAvailableRentalTimes(req *models.UpdateAvailableRentalTimesRequest) (*models.UpdateAvailableRentalTimesResponse, error) {
	err := s.repo.UpdateAvailableRentalTimes(req.ServiceType, req.AvailableTimes)
	if err != nil {
		return nil, err
	}

	return &models.UpdateAvailableRentalTimesResponse{
		Success: true,
	}, nil
}
