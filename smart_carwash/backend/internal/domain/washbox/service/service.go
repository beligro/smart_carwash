package service

import (
	"carwash_backend/internal/domain/settings/service"
	"carwash_backend/internal/domain/washbox/models"
	"carwash_backend/internal/domain/washbox/repository"
	"errors"
	"time"

	"github.com/google/uuid"
)

// Service интерфейс для бизнес-логики боксов мойки
type Service interface {
	GetWashBoxByID(id uuid.UUID) (*models.WashBox, error)
	UpdateWashBoxStatus(id uuid.UUID, status string) error
	GetFreeWashBoxes() ([]models.WashBox, error)
	GetFreeWashBoxesByServiceType(serviceType string) ([]models.WashBox, error)
	GetFreeWashBoxesWithChemistry(serviceType string) ([]models.WashBox, error)
	GetWashBoxesByServiceType(serviceType string) ([]models.WashBox, error)
	GetAllWashBoxes() ([]models.WashBox, error)

	// Административные методы
	AdminCreateWashBox(req *models.AdminCreateWashBoxRequest) (*models.AdminCreateWashBoxResponse, error)
	AdminUpdateWashBox(req *models.AdminUpdateWashBoxRequest) (*models.AdminUpdateWashBoxResponse, error)
	AdminDeleteWashBox(req *models.AdminDeleteWashBoxRequest) (*models.AdminDeleteWashBoxResponse, error)
	AdminGetWashBox(req *models.AdminGetWashBoxRequest) (*models.AdminGetWashBoxResponse, error)
	AdminListWashBoxes(req *models.AdminListWashBoxesRequest) (*models.AdminListWashBoxesResponse, error)
	RestoreWashBox(id uuid.UUID, status string, serviceType string) (*models.WashBox, error)

	// Методы для кассира
	CashierListWashBoxes(req *models.CashierListWashBoxesRequest) (*models.CashierListWashBoxesResponse, error)
	CashierSetMaintenance(req *models.CashierSetMaintenanceRequest) (*models.CashierSetMaintenanceResponse, error)

	// Методы для уборщиков
	CleanerListWashBoxes(req *models.CleanerListWashBoxesRequest) (*models.CleanerListWashBoxesResponse, error)
	CleanerReserveCleaning(req *models.CleanerReserveCleaningRequest, cleanerID uuid.UUID) (*models.CleanerReserveCleaningResponse, error)
	CleanerStartCleaning(req *models.CleanerStartCleaningRequest, cleanerID uuid.UUID) (*models.CleanerStartCleaningResponse, error)
	CleanerCancelCleaning(req *models.CleanerCancelCleaningRequest, cleanerID uuid.UUID) (*models.CleanerCancelCleaningResponse, error)
	CleanerCompleteCleaning(req *models.CleanerCompleteCleaningRequest, cleanerID uuid.UUID) (*models.CleanerCompleteCleaningResponse, error)
	GetCleaningBoxes() ([]models.WashBox, error)
	AutoCompleteExpiredCleanings() error
	UpdateCleaningStartedAt(washBoxID uuid.UUID, startedAt time.Time) error
	ClearCleaningReservation(washBoxID uuid.UUID) error
}

// ServiceImpl реализация Service
type ServiceImpl struct {
	repo            repository.Repository
	settingsService service.Service
}

// NewService создает новый экземпляр Service
func NewService(repo repository.Repository, settingsService service.Service) *ServiceImpl {
	return &ServiceImpl{
		repo:            repo,
		settingsService: settingsService,
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

// GetFreeWashBoxesWithChemistry получает все свободные боксы мойки с химией определенного типа
func (s *ServiceImpl) GetFreeWashBoxesWithChemistry(serviceType string) ([]models.WashBox, error) {
	return s.repo.GetFreeWashBoxesWithChemistry(serviceType)
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
	// Проверяем, не существует ли уже активный бокс с таким номером
	existingBox, err := s.repo.GetWashBoxByNumber(req.Number)
	if err == nil && existingBox != nil {
		return nil, errors.New("бокс с таким номером уже существует")
	}

	// Проверяем, есть ли удаленный бокс с таким номером
	deletedBox, err := s.repo.GetWashBoxByNumberIncludingDeleted(req.Number)
	if err == nil && deletedBox != nil && deletedBox.DeletedAt.Valid {
		// Восстанавливаем удаленный бокс
		restoredBox, err := s.RestoreWashBox(deletedBox.ID, req.Status, req.ServiceType)
		if err != nil {
			return nil, err
		}
		return &models.AdminCreateWashBoxResponse{
			WashBox: *restoredBox,
		}, nil
	}

	// Создаем новый бокс
	washBox := &models.WashBox{
		Number:      req.Number,
		Status:      req.Status,
		ServiceType: req.ServiceType,
	}

	// Устанавливаем химию по умолчанию в зависимости от типа услуги
	if req.ChemistryEnabled != nil {
		washBox.ChemistryEnabled = *req.ChemistryEnabled
	} else {
		// По умолчанию химия включена только для wash
		washBox.ChemistryEnabled = req.ServiceType == "wash"
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

	if req.ChemistryEnabled != nil {
		existingBox.ChemistryEnabled = *req.ChemistryEnabled
	}

	if req.LightCoilRegister != nil {
		existingBox.LightCoilRegister = req.LightCoilRegister
	}

	if req.ChemistryCoilRegister != nil {
		existingBox.ChemistryCoilRegister = req.ChemistryCoilRegister
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
	boxes, total, err := s.repo.GetWashBoxesWithFilters(req.Status, req.ServiceType, limit, offset)
	if err != nil {
		return nil, err
	}

	return &models.AdminListWashBoxesResponse{
		WashBoxes: boxes,
		Total:     total,
		Limit:     limit,
		Offset:    offset,
	}, nil
}

// RestoreWashBox восстанавливает удаленный бокс мойки
func (s *ServiceImpl) RestoreWashBox(id uuid.UUID, status string, serviceType string) (*models.WashBox, error) {
	// Восстанавливаем бокс через repository
	restoredBox, err := s.repo.RestoreWashBox(id, status, serviceType)
	if err != nil {
		return nil, err
	}

	return restoredBox, nil
}

// CashierListWashBoxes возвращает список боксов для кассира
func (s *ServiceImpl) CashierListWashBoxes(req *models.CashierListWashBoxesRequest) (*models.CashierListWashBoxesResponse, error) {
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
	boxes, total, err := s.repo.GetWashBoxesWithFilters(req.Status, req.ServiceType, limit, offset)
	if err != nil {
		return nil, err
	}

	return &models.CashierListWashBoxesResponse{
		WashBoxes: boxes,
		Total:     total,
		Limit:     limit,
		Offset:    offset,
	}, nil
}

// CashierSetMaintenance переводит бокс в режим обслуживания (только свободные боксы)
func (s *ServiceImpl) CashierSetMaintenance(req *models.CashierSetMaintenanceRequest) (*models.CashierSetMaintenanceResponse, error) {
	// Получаем бокс по ID
	washBox, err := s.repo.GetWashBoxByID(req.ID)
	if err != nil {
		return nil, errors.New("бокс не найден")
	}

	// Проверяем, что бокс свободен (только свободные боксы можно переводить в обслуживание)
	if washBox.Status != models.StatusFree {
		return nil, errors.New("в режим обслуживания можно переводить только свободные боксы")
	}

	// Обновляем статус бокса на maintenance
	washBox.Status = models.StatusMaintenance
	updatedBox, err := s.repo.UpdateWashBox(washBox)
	if err != nil {
		return nil, err
	}

	return &models.CashierSetMaintenanceResponse{
		WashBox: *updatedBox,
		Message: "Бокс переведен в режим обслуживания",
	}, nil
}

// CleanerListWashBoxes получает список боксов для уборщика
func (s *ServiceImpl) CleanerListWashBoxes(req *models.CleanerListWashBoxesRequest) (*models.CleanerListWashBoxesResponse, error) {
	limit := 100 // По умолчанию
	offset := 0

	if req.Limit != nil {
		limit = *req.Limit
	}
	if req.Offset != nil {
		offset = *req.Offset
	}

	boxes, total, err := s.repo.GetWashBoxesForCleaner(limit, offset)
	if err != nil {
		return nil, err
	}

	return &models.CleanerListWashBoxesResponse{
		WashBoxes: boxes,
		Total:     total,
	}, nil
}

// CleanerReserveCleaning резервирует уборку для бокса
func (s *ServiceImpl) CleanerReserveCleaning(req *models.CleanerReserveCleaningRequest, cleanerID uuid.UUID) (*models.CleanerReserveCleaningResponse, error) {
	// Проверяем, что бокс существует
	washBox, err := s.repo.GetWashBoxByID(req.WashBoxID)
	if err != nil {
		return nil, err
	}

	// Проверяем, что бокс не в статусе cleaning
	if washBox.Status == models.StatusCleaning {
		return nil, errors.New("бокс уже на уборке")
	}

	// Резервируем уборку
	err = s.repo.ReserveCleaning(req.WashBoxID, cleanerID)
	if err != nil {
		return nil, err
	}

	return &models.CleanerReserveCleaningResponse{
		Success: true,
	}, nil
}

// CleanerStartCleaning начинает уборку бокса
func (s *ServiceImpl) CleanerStartCleaning(req *models.CleanerStartCleaningRequest, cleanerID uuid.UUID) (*models.CleanerStartCleaningResponse, error) {
	// Проверяем, что бокс существует
	washBox, err := s.repo.GetWashBoxByID(req.WashBoxID)
	if err != nil {
		return nil, err
	}

	// Проверяем, что бокс свободен или зарезервирован этим уборщиком
	if washBox.Status != models.StatusFree && 
		(washBox.CleaningReservedBy == nil || *washBox.CleaningReservedBy != cleanerID) {
		return nil, errors.New("бокс недоступен для уборки")
	}

	// Начинаем уборку
	err = s.repo.StartCleaning(req.WashBoxID)
	if err != nil {
		return nil, err
	}

	return &models.CleanerStartCleaningResponse{
		Success: true,
	}, nil
}

// CleanerCancelCleaning отменяет резервирование уборки
func (s *ServiceImpl) CleanerCancelCleaning(req *models.CleanerCancelCleaningRequest, cleanerID uuid.UUID) (*models.CleanerCancelCleaningResponse, error) {
	// Проверяем, что бокс существует
	washBox, err := s.repo.GetWashBoxByID(req.WashBoxID)
	if err != nil {
		return nil, err
	}

	// Проверяем, что уборка зарезервирована этим уборщиком
	if washBox.CleaningReservedBy == nil || *washBox.CleaningReservedBy != cleanerID {
		return nil, errors.New("уборка не зарезервирована этим уборщиком")
	}

	// Отменяем резервирование
	err = s.repo.CancelCleaning(req.WashBoxID)
	if err != nil {
		return nil, err
	}

	return &models.CleanerCancelCleaningResponse{
		Success: true,
	}, nil
}

// CleanerCompleteCleaning завершает уборку бокса
func (s *ServiceImpl) CleanerCompleteCleaning(req *models.CleanerCompleteCleaningRequest, cleanerID uuid.UUID) (*models.CleanerCompleteCleaningResponse, error) {
	// Проверяем, что бокс существует и в статусе cleaning
	washBox, err := s.repo.GetWashBoxByID(req.WashBoxID)
	if err != nil {
		return nil, err
	}

	if washBox.Status != models.StatusCleaning {
		return nil, errors.New("бокс не на уборке")
	}

	// Завершаем уборку
	err = s.repo.CompleteCleaning(req.WashBoxID)
	if err != nil {
		return nil, err
	}

	return &models.CleanerCompleteCleaningResponse{
		Success: true,
	}, nil
}

// GetCleaningBoxes получает все боксы в статусе уборки
func (s *ServiceImpl) GetCleaningBoxes() ([]models.WashBox, error) {
	return s.repo.GetCleaningBoxes()
}

// AutoCompleteExpiredCleanings автоматически завершает истекшие уборки
func (s *ServiceImpl) AutoCompleteExpiredCleanings() error {
	// Получаем все боксы в статусе уборки
	cleaningBoxes, err := s.repo.GetCleaningBoxes()
	if err != nil {
		return err
	}

	// Получаем настройку времени уборки из settings service
	// Пока используем фиксированное значение 30 минут, так как настройка времени уборки не реализована в settings
	timeoutMinutes := 30

	for _, box := range cleaningBoxes {
		if box.CleaningStartedAt != nil {
			// Проверяем, истекло ли время уборки
			timeSinceStart := time.Since(*box.CleaningStartedAt)
			if timeSinceStart.Minutes() >= float64(timeoutMinutes) {
				// Завершаем уборку
				err = s.repo.CompleteCleaning(box.ID)
				if err != nil {
					// Логируем ошибку, но продолжаем обработку других боксов
					continue
				}
			}
		}
	}

	return nil
}

// UpdateCleaningStartedAt обновляет время начала уборки
func (s *ServiceImpl) UpdateCleaningStartedAt(washBoxID uuid.UUID, startedAt time.Time) error {
	return s.repo.UpdateCleaningStartedAt(washBoxID, startedAt)
}

// ClearCleaningReservation очищает резерв уборки
func (s *ServiceImpl) ClearCleaningReservation(washBoxID uuid.UUID) error {
	return s.repo.CancelCleaning(washBoxID)
}
