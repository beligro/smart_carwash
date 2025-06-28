package service

import (
	"carwash_backend/internal/domain/washbox/models"
	"carwash_backend/internal/domain/washbox/repository"
	"errors"

	"github.com/google/uuid"
)

// Service интерфейс для бизнес-логики боксов мойки
type Service interface {
	GetWashBoxByID(id uuid.UUID) (*models.WashBox, error)
	UpdateWashBoxStatus(id uuid.UUID, status string) error
	GetFreeWashBoxes() ([]models.WashBox, error)
	GetFreeWashBoxesByServiceType(serviceType string) ([]models.WashBox, error)
	GetWashBoxesByServiceType(serviceType string) ([]models.WashBox, error)
	GetAllWashBoxes() ([]models.WashBox, error)

	// Административные методы
	AdminCreateWashBox(req *models.AdminCreateWashBoxRequest) (*models.AdminCreateWashBoxResponse, error)
	AdminUpdateWashBox(req *models.AdminUpdateWashBoxRequest) (*models.AdminUpdateWashBoxResponse, error)
	AdminDeleteWashBox(req *models.AdminDeleteWashBoxRequest) (*models.AdminDeleteWashBoxResponse, error)
	AdminGetWashBox(req *models.AdminGetWashBoxRequest) (*models.AdminGetWashBoxResponse, error)
	AdminListWashBoxes(req *models.AdminListWashBoxesRequest) (*models.AdminListWashBoxesResponse, error)
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

// GetFreeWashBoxesByServiceType получает все свободные боксы мойки определенного типа
func (s *ServiceImpl) GetFreeWashBoxesByServiceType(serviceType string) ([]models.WashBox, error) {
	return s.repo.GetFreeWashBoxesByServiceType(serviceType)
}

// GetWashBoxesByServiceType получает все боксы мойки определенного типа
func (s *ServiceImpl) GetWashBoxesByServiceType(serviceType string) ([]models.WashBox, error) {
	return s.repo.GetWashBoxesByServiceType(serviceType)
}

// GetAllWashBoxes получает все боксы мойки
func (s *ServiceImpl) GetAllWashBoxes() ([]models.WashBox, error) {
	return s.repo.GetAllWashBoxes()
}

// AdminCreateWashBox создает новый бокс мойки
func (s *ServiceImpl) AdminCreateWashBox(req *models.AdminCreateWashBoxRequest) (*models.AdminCreateWashBoxResponse, error) {
	// Проверяем, не существует ли уже бокс с таким номером
	existingBox, err := s.repo.GetWashBoxByNumber(req.Number)
	if err == nil && existingBox != nil {
		return nil, errors.New("бокс с таким номером уже существует")
	}

	// Создаем новый бокс
	washBox := &models.WashBox{
		Number:      req.Number,
		Status:      req.Status,
		ServiceType: req.ServiceType,
	}

	createdBox, err := s.repo.CreateWashBox(washBox)
	if err != nil {
		return nil, err
	}

	return &models.AdminCreateWashBoxResponse{
		WashBox: *createdBox,
	}, nil
}

// AdminUpdateWashBox обновляет бокс мойки
func (s *ServiceImpl) AdminUpdateWashBox(req *models.AdminUpdateWashBoxRequest) (*models.AdminUpdateWashBoxResponse, error) {
	// Получаем существующий бокс
	existingBox, err := s.repo.GetWashBoxByID(req.ID)
	if err != nil {
		return nil, errors.New("бокс не найден")
	}

	// Обновляем поля, если они переданы
	if req.Number != nil {
		// Проверяем, не существует ли уже бокс с таким номером
		if *req.Number != existingBox.Number {
			boxWithNumber, err := s.repo.GetWashBoxByNumber(*req.Number)
			if err == nil && boxWithNumber != nil {
				return nil, errors.New("бокс с таким номером уже существует")
			}
		}
		existingBox.Number = *req.Number
	}

	if req.Status != nil {
		existingBox.Status = *req.Status
	}

	if req.ServiceType != nil {
		existingBox.ServiceType = *req.ServiceType
	}

	// Сохраняем изменения
	updatedBox, err := s.repo.UpdateWashBox(existingBox)
	if err != nil {
		return nil, err
	}

	return &models.AdminUpdateWashBoxResponse{
		WashBox: *updatedBox,
	}, nil
}

// AdminDeleteWashBox удаляет бокс мойки
func (s *ServiceImpl) AdminDeleteWashBox(req *models.AdminDeleteWashBoxRequest) (*models.AdminDeleteWashBoxResponse, error) {
	// Проверяем, существует ли бокс
	existingBox, err := s.repo.GetWashBoxByID(req.ID)
	if err != nil {
		return nil, errors.New("бокс не найден")
	}

	// Проверяем, не занят ли бокс
	if existingBox.Status == models.StatusBusy || existingBox.Status == models.StatusReserved {
		return nil, errors.New("нельзя удалить занятый или зарезервированный бокс")
	}

	// Удаляем бокс
	err = s.repo.DeleteWashBox(req.ID)
	if err != nil {
		return nil, err
	}

	return &models.AdminDeleteWashBoxResponse{
		Message: "Бокс успешно удален",
	}, nil
}

// AdminGetWashBox получает бокс мойки по ID
func (s *ServiceImpl) AdminGetWashBox(req *models.AdminGetWashBoxRequest) (*models.AdminGetWashBoxResponse, error) {
	washBox, err := s.repo.GetWashBoxByID(req.ID)
	if err != nil {
		return nil, errors.New("бокс не найден")
	}

	return &models.AdminGetWashBoxResponse{
		WashBox: *washBox,
	}, nil
}

// AdminListWashBoxes получает список боксов мойки с фильтрацией
func (s *ServiceImpl) AdminListWashBoxes(req *models.AdminListWashBoxesRequest) (*models.AdminListWashBoxesResponse, error) {
	// Устанавливаем значения по умолчанию
	limit := 50
	offset := 0

	if req.Limit != nil {
		limit = *req.Limit
	}
	if req.Offset != nil {
		offset = *req.Offset
	}

	// Получаем боксы с фильтрацией
	washBoxes, total, err := s.repo.GetWashBoxesWithFilters(req.Status, req.ServiceType, limit, offset)
	if err != nil {
		return nil, err
	}

	return &models.AdminListWashBoxesResponse{
		WashBoxes: washBoxes,
		Total:     total,
		Limit:     limit,
		Offset:    offset,
	}, nil
}
