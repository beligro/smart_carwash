package service

import (
	modbusAdapter "carwash_backend/internal/domain/modbus/adapter"
	modbusRepository "carwash_backend/internal/domain/modbus/repository"
	sessionRepository "carwash_backend/internal/domain/session/repository"
	"carwash_backend/internal/domain/settings/service"
	"carwash_backend/internal/domain/washbox/models"
	"carwash_backend/internal/domain/washbox/repository"
	"context"
	"errors"
	"math/rand"
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// Service интерфейс для бизнес-логики боксов мойки
type Service interface {
	GetWashBoxByID(ctx context.Context, id uuid.UUID) (*models.WashBox, error)
	UpdateWashBoxStatus(ctx context.Context, id uuid.UUID, status string) error
	GetFreeWashBoxes(ctx context.Context) ([]models.WashBox, error)
	GetFreeWashBoxesByServiceType(ctx context.Context, serviceType string) ([]models.WashBox, error)
	GetFreeWashBoxesWithChemistry(ctx context.Context, serviceType string) ([]models.WashBox, error)
	GetWashBoxesByServiceType(ctx context.Context, serviceType string) ([]models.WashBox, error)
	GetAllWashBoxes(ctx context.Context) ([]models.WashBox, error)

	// Административные методы
	AdminCreateWashBox(ctx context.Context, req *models.AdminCreateWashBoxRequest) (*models.AdminCreateWashBoxResponse, error)
	AdminUpdateWashBox(ctx context.Context, req *models.AdminUpdateWashBoxRequest) (*models.AdminUpdateWashBoxResponse, error)
	AdminDeleteWashBox(ctx context.Context, req *models.AdminDeleteWashBoxRequest) (*models.AdminDeleteWashBoxResponse, error)
	AdminGetWashBox(ctx context.Context, req *models.AdminGetWashBoxRequest) (*models.AdminGetWashBoxResponse, error)
	AdminListWashBoxes(ctx context.Context, req *models.AdminListWashBoxesRequest) (*models.AdminListWashBoxesResponse, error)
	RestoreWashBox(ctx context.Context, id uuid.UUID, status string, serviceType string) (*models.WashBox, error)

	// Методы для кассира
	CashierListWashBoxes(ctx context.Context, req *models.CashierListWashBoxesRequest) (*models.CashierListWashBoxesResponse, error)
	CashierSetMaintenance(ctx context.Context, req *models.CashierSetMaintenanceRequest) (*models.CashierSetMaintenanceResponse, error)

	// Методы для уборщиков
	CleanerListWashBoxes(ctx context.Context, req *models.CleanerListWashBoxesRequest) (*models.CleanerListWashBoxesResponse, error)
	CleanerReserveCleaning(ctx context.Context, req *models.CleanerReserveCleaningRequest, cleanerID uuid.UUID) (*models.CleanerReserveCleaningResponse, error)
	CleanerStartCleaning(ctx context.Context, req *models.CleanerStartCleaningRequest, cleanerID uuid.UUID) (*models.CleanerStartCleaningResponse, error)
	CleanerCancelCleaning(ctx context.Context, req *models.CleanerCancelCleaningRequest, cleanerID uuid.UUID) (*models.CleanerCancelCleaningResponse, error)
	CleanerCompleteCleaning(ctx context.Context, req *models.CleanerCompleteCleaningRequest, cleanerID uuid.UUID) (*models.CleanerCompleteCleaningResponse, error)
	GetCleaningBoxes(ctx context.Context) ([]models.WashBox, error)
	AutoCompleteExpiredCleanings(ctx context.Context) error
	UpdateCleaningStartedAt(ctx context.Context, washBoxID uuid.UUID, startedAt time.Time) error
	ClearCleaningReservation(ctx context.Context, washBoxID uuid.UUID) error

	// Методы для логов уборки (админка)
	AdminListCleaningLogs(ctx context.Context, req *models.AdminListCleaningLogsRequest) (*models.AdminListCleaningLogsResponse, error)

	// Методы для работы с cooldown
	SetCooldown(ctx context.Context, boxID uuid.UUID, userID uuid.UUID, cooldownUntil time.Time) error
	SetCooldownByCarNumber(ctx context.Context, boxID uuid.UUID, carNumber string, cooldownUntil time.Time) error
	ClearCooldown(ctx context.Context, boxID uuid.UUID) error
	CheckCooldownExpired(ctx context.Context) error
}

// ServiceImpl реализация Service
type ServiceImpl struct {
	repo            repository.Repository
	sessionRepo     sessionRepository.Repository
	settingsService service.Service
	modbusRepo      *modbusRepository.ModbusRepository
	modbusAdapter   *modbusAdapter.ModbusAdapter
	// SafeDB удалён; используем контексты вызова
}

// shuffleBoxesWithSamePriority перемешивает боксы с одинаковым приоритетом
func shuffleBoxesWithSamePriority(boxes []models.WashBox) []models.WashBox {
	if len(boxes) <= 1 {
		return boxes
	}

	// Группируем боксы по приоритету
	priorityGroups := make(map[string][]models.WashBox)
	for _, box := range boxes {
		priorityGroups[box.Priority] = append(priorityGroups[box.Priority], box)
	}

	// Перемешиваем боксы в каждой группе с одинаковым приоритетом
	result := make([]models.WashBox, 0, len(boxes))
	// Проходим по всем возможным приоритетам от A до Z
	for priority := 'A'; priority <= 'Z'; priority++ {
		priorityStr := string(priority)
		if group, exists := priorityGroups[priorityStr]; exists {
			// Перемешиваем группу
			rand.Shuffle(len(group), func(i, j int) {
				group[i], group[j] = group[j], group[i]
			})
			result = append(result, group...)
		}
	}

	return result
}

// NewService создает новый экземпляр Service
func NewService(repo repository.Repository, sessionRepo sessionRepository.Repository, settingsService service.Service, db *gorm.DB, modbusAdapterInst *modbusAdapter.ModbusAdapter) *ServiceImpl {
	return &ServiceImpl{
		repo:            repo,
		sessionRepo:     sessionRepo,
		settingsService: settingsService,
		modbusRepo:      modbusRepository.NewModbusRepository(db),
		modbusAdapter:   modbusAdapterInst,
	}
}

// GetWashBoxByID получает бокс мойки по ID
func (s *ServiceImpl) GetWashBoxByID(ctx context.Context, id uuid.UUID) (*models.WashBox, error) {
	return s.repo.GetWashBoxByID(ctx, id)
}

// UpdateWashBoxStatus обновляет статус бокса мойки
func (s *ServiceImpl) UpdateWashBoxStatus(ctx context.Context, id uuid.UUID, status string) error {
	return s.repo.UpdateWashBoxStatus(ctx, id, status)
}

// GetFreeWashBoxes получает все свободные боксы мойки, отсортированные по приоритету с рандомизацией
func (s *ServiceImpl) GetFreeWashBoxes(ctx context.Context) ([]models.WashBox, error) {
	boxes, err := s.repo.GetFreeWashBoxes(ctx)
	if err != nil {
		return nil, err
	}
	return shuffleBoxesWithSamePriority(boxes), nil
}

// GetFreeWashBoxesByServiceType получает все свободные боксы мойки определенного типа, отсортированные по приоритету с рандомизацией
func (s *ServiceImpl) GetFreeWashBoxesByServiceType(ctx context.Context, serviceType string) ([]models.WashBox, error) {
	boxes, err := s.repo.GetFreeWashBoxesByServiceType(ctx, serviceType)
	if err != nil {
		return nil, err
	}
	return shuffleBoxesWithSamePriority(boxes), nil
}

// GetFreeWashBoxesWithChemistry получает все свободные боксы мойки с химией определенного типа, отсортированные по приоритету с рандомизацией
func (s *ServiceImpl) GetFreeWashBoxesWithChemistry(ctx context.Context, serviceType string) ([]models.WashBox, error) {
	boxes, err := s.repo.GetFreeWashBoxesWithChemistry(ctx, serviceType)
	if err != nil {
		return nil, err
	}
	return shuffleBoxesWithSamePriority(boxes), nil
}

// GetWashBoxesByServiceType получает все боксы мойки определенного типа
func (s *ServiceImpl) GetWashBoxesByServiceType(ctx context.Context, serviceType string) ([]models.WashBox, error) {
	return s.repo.GetWashBoxesByServiceType(ctx, serviceType)
}

// GetAllWashBoxes получает все боксы мойки
func (s *ServiceImpl) GetAllWashBoxes(ctx context.Context) ([]models.WashBox, error) {
	return s.repo.GetAllWashBoxes(ctx)
}

// AdminCreateWashBox создает новый бокс мойки
func (s *ServiceImpl) AdminCreateWashBox(ctx context.Context, req *models.AdminCreateWashBoxRequest) (*models.AdminCreateWashBoxResponse, error) {
	// Проверяем, не существует ли уже активный бокс с таким номером
	existingBox, err := s.repo.GetWashBoxByNumber(ctx, req.Number)
	if err == nil && existingBox != nil {
		return nil, errors.New("бокс с таким номером уже существует")
	}

	// Проверяем, есть ли удаленный бокс с таким номером
	deletedBox, err := s.repo.GetWashBoxByNumberIncludingDeleted(ctx, req.Number)
	if err == nil && deletedBox != nil && deletedBox.DeletedAt.Valid {
		// Восстанавливаем удаленный бокс
		restoredBox, err := s.RestoreWashBox(ctx, deletedBox.ID, req.Status, req.ServiceType)
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
		Priority:    req.Priority,
	}

	// Устанавливаем химию по умолчанию в зависимости от типа услуги
	if req.ChemistryEnabled != nil {
		washBox.ChemistryEnabled = *req.ChemistryEnabled
	} else {
		// По умолчанию химия включена только для wash
		washBox.ChemistryEnabled = req.ServiceType == "wash"
	}

	createdBox, err := s.repo.CreateWashBox(ctx, washBox)
	if err != nil {
		return nil, err
	}

	return &models.AdminCreateWashBoxResponse{
		WashBox: *createdBox,
	}, nil
}

// AdminUpdateWashBox обновляет бокс мойки
func (s *ServiceImpl) AdminUpdateWashBox(ctx context.Context, req *models.AdminUpdateWashBoxRequest) (*models.AdminUpdateWashBoxResponse, error) {
	// Получаем существующий бокс
	existingBox, err := s.repo.GetWashBoxByID(ctx, req.ID)
	if err != nil {
		return nil, errors.New("бокс не найден")
	}

	// Обновляем поля, если они переданы
	if req.Number != nil {
		// Проверяем, не существует ли уже бокс с таким номером
		if *req.Number != existingBox.Number {
			boxWithNumber, err := s.repo.GetWashBoxByNumber(ctx, *req.Number)
			if err == nil && boxWithNumber != nil {
				return nil, errors.New("бокс с таким номером уже существует")
			}
		}
		existingBox.Number = *req.Number
	}

	if req.Status != nil {
		existingBox.Status = *req.Status
	}

	existingBox.Comment = req.Comment

	if req.ServiceType != nil {
		existingBox.ServiceType = *req.ServiceType
	}

	if req.ChemistryEnabled != nil {
		existingBox.ChemistryEnabled = *req.ChemistryEnabled
	}

	if req.Priority != nil {
		existingBox.Priority = *req.Priority
	}

	if req.LightCoilRegister != nil {
		existingBox.LightCoilRegister = req.LightCoilRegister
	}

	if req.ChemistryCoilRegister != nil {
		existingBox.ChemistryCoilRegister = req.ChemistryCoilRegister
	}

	// Сохраняем изменения
	updatedBox, err := s.repo.UpdateWashBox(ctx, existingBox)
	if err != nil {
		return nil, err
	}

	return &models.AdminUpdateWashBoxResponse{
		WashBox: *updatedBox,
	}, nil
}

// AdminDeleteWashBox удаляет бокс мойки
func (s *ServiceImpl) AdminDeleteWashBox(ctx context.Context, req *models.AdminDeleteWashBoxRequest) (*models.AdminDeleteWashBoxResponse, error) {
	// Проверяем, существует ли бокс
	existingBox, err := s.repo.GetWashBoxByID(ctx, req.ID)
	if err != nil {
		return nil, errors.New("бокс не найден")
	}

	// Проверяем, не занят ли бокс
	if existingBox.Status == models.StatusBusy || existingBox.Status == models.StatusReserved {
		return nil, errors.New("нельзя удалить занятый или зарезервированный бокс")
	}

	// Удаляем бокс
	err = s.repo.DeleteWashBox(ctx, req.ID)
	if err != nil {
		return nil, err
	}

	return &models.AdminDeleteWashBoxResponse{
		Message: "Бокс успешно удален",
	}, nil
}

// AdminGetWashBox получает бокс мойки по ID
func (s *ServiceImpl) AdminGetWashBox(ctx context.Context, req *models.AdminGetWashBoxRequest) (*models.AdminGetWashBoxResponse, error) {
	washBox, err := s.repo.GetWashBoxByID(ctx, req.ID)
	if err != nil {
		return nil, errors.New("бокс не найден")
	}

	return &models.AdminGetWashBoxResponse{
		WashBox: *washBox,
	}, nil
}

// AdminListWashBoxes получает список боксов мойки с фильтрацией
func (s *ServiceImpl) AdminListWashBoxes(ctx context.Context, req *models.AdminListWashBoxesRequest) (*models.AdminListWashBoxesResponse, error) {
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
	boxes, total, err := s.repo.GetWashBoxesWithFilters(ctx, req.Status, req.ServiceType, limit, offset)
	if err != nil {
		return nil, err
	}

	// Получаем статусы modbus для всех боксов
	modbusStatuses, err := s.modbusRepo.GetAllModbusConnectionStatuses(ctx)
	if err == nil && len(modbusStatuses) > 0 {
		// Создаем мапу статусов для быстрого доступа
		statusMap := make(map[uuid.UUID]*bool)
		chemistryMap := make(map[uuid.UUID]*bool)

		for i := range modbusStatuses {
			if modbusStatuses[i].LightStatus != nil {
				lightStatus := *modbusStatuses[i].LightStatus
				statusMap[modbusStatuses[i].BoxID] = &lightStatus
			}
			if modbusStatuses[i].ChemistryStatus != nil {
				chemistryStatus := *modbusStatuses[i].ChemistryStatus
				chemistryMap[modbusStatuses[i].BoxID] = &chemistryStatus
			}
		}

		// Мапим статусы на боксы
		for i := range boxes {
			if lightStatus, exists := statusMap[boxes[i].ID]; exists {
				boxes[i].LightStatus = lightStatus
			}
			if chemistryStatus, exists := chemistryMap[boxes[i].ID]; exists {
				boxes[i].ChemistryStatus = chemistryStatus
			}
		}
	}

	return &models.AdminListWashBoxesResponse{
		WashBoxes: boxes,
		Total:     total,
		Limit:     limit,
		Offset:    offset,
	}, nil
}

// RestoreWashBox восстанавливает удаленный бокс мойки
func (s *ServiceImpl) RestoreWashBox(ctx context.Context, id uuid.UUID, status string, serviceType string) (*models.WashBox, error) {
	// Восстанавливаем бокс через repository
	restoredBox, err := s.repo.RestoreWashBox(ctx, id, status, serviceType)
	if err != nil {
		return nil, err
	}

	return restoredBox, nil
}

// CashierListWashBoxes возвращает список боксов для кассира
func (s *ServiceImpl) CashierListWashBoxes(ctx context.Context, req *models.CashierListWashBoxesRequest) (*models.CashierListWashBoxesResponse, error) {
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
	boxes, total, err := s.repo.GetWashBoxesWithFilters(ctx, req.Status, req.ServiceType, limit, offset)
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
func (s *ServiceImpl) CashierSetMaintenance(ctx context.Context, req *models.CashierSetMaintenanceRequest) (*models.CashierSetMaintenanceResponse, error) {
	// Получаем бокс по ID
	washBox, err := s.repo.GetWashBoxByID(ctx, req.ID)
	if err != nil {
		return nil, errors.New("бокс не найден")
	}

	// Проверяем, что бокс свободен (только свободные боксы можно переводить на сервис)
	if washBox.Status != models.StatusFree {
		return nil, errors.New("в режим обслуживания можно переводить только свободные боксы")
	}

	// Обновляем статус бокса на maintenance
	washBox.Status = models.StatusMaintenance
	updatedBox, err := s.repo.UpdateWashBox(ctx, washBox)
	if err != nil {
		return nil, err
	}

	return &models.CashierSetMaintenanceResponse{
		WashBox: *updatedBox,
		Message: "Бокс переведен в режим обслуживания",
	}, nil
}

// CleanerListWashBoxes получает список боксов для уборщика
func (s *ServiceImpl) CleanerListWashBoxes(ctx context.Context, req *models.CleanerListWashBoxesRequest) (*models.CleanerListWashBoxesResponse, error) {
	limit := 100 // По умолчанию
	offset := 0

	if req.Limit != nil {
		limit = *req.Limit
	}
	if req.Offset != nil {
		offset = *req.Offset
	}

	boxes, total, err := s.repo.GetWashBoxesForCleaner(ctx, limit, offset)
	if err != nil {
		return nil, err
	}

	// Батч: определяем возможность уборки без N+1
	// 1) Собираем все boxIDs
	boxIDs := make([]uuid.UUID, 0, len(boxes))
	for i := range boxes {
		boxIDs = append(boxIDs, boxes[i].ID)
	}

	// 2) Получаем последние логи по всем боксам
	lastLogsMap, err := s.repo.GetLastCleaningLogsByBoxIDs(ctx, boxIDs)
	if err != nil {
		// В случае ошибки считаем, что убирать можно (поведение как в checkIfBoxCanBeCleaned при ошибке)
		for i := range boxes {
			can := true
			boxes[i].CanBeCleaned = &can
		}
	} else {
		// 3) Находим минимальную дату CompletedAt среди завершенных логов для ограничения выборки сессий
		var minSince *time.Time
		now := time.Now()
		for _, l := range lastLogsMap {
			if l != nil && l.Status == models.CleaningLogStatusCompleted && l.CompletedAt != nil {
				if minSince == nil || l.CompletedAt.Before(*minSince) {
					t := *l.CompletedAt
					minSince = &t
				}
			}
		}

		var completedRows []struct {
			BoxID     uuid.UUID
			CreatedAt time.Time
		}
		if minSince != nil {
			// 4) Получаем завершенные сессии с minSince для всех боксов одним запросом
			completedRows, _ = s.sessionRepo.GetCompletedSessionsCreatedAtSinceForBoxes(ctx, boxIDs, *minSince)
		}
		// Группируем по боксу с последующей фильтрацией по конкретному CompletedAt
		createdAtByBox := make(map[uuid.UUID][]time.Time)
		for _, row := range completedRows {
			createdAtByBox[row.BoxID] = append(createdAtByBox[row.BoxID], row.CreatedAt)
		}

		for i := range boxes {
			lastLog := lastLogsMap[boxes[i].ID]
			// Если нет логов или последний лог не завершён — можно убирать
			if lastLog == nil || lastLog.Status != models.CleaningLogStatusCompleted || lastLog.CompletedAt == nil {
				can := true
				boxes[i].CanBeCleaned = &can
				continue
			}
			// Иначе проверяем, были ли завершённые сессии после CompletedAt этого лога
			hasCompletedAfter := false
			if times, ok := createdAtByBox[boxes[i].ID]; ok {
				for _, t := range times {
					if t.After(*lastLog.CompletedAt) && t.Before(now.Add(1*time.Second)) { // небольшой запас по границе
						hasCompletedAfter = true
						break
					}
				}
			}
			can := hasCompletedAfter
			boxes[i].CanBeCleaned = &can
		}
	}

	return &models.CleanerListWashBoxesResponse{
		WashBoxes: boxes,
		Total:     total,
	}, nil
}

// checkIfBoxCanBeCleaned проверяет, можно ли убирать бокс (не было повторного назначения)
func (s *ServiceImpl) checkIfBoxCanBeCleaned(ctx context.Context, boxID uuid.UUID) bool {
	// Получаем последний лог уборки для бокса
	lastLog, err := s.repo.GetLastCleaningLogByBox(ctx, boxID)
	if err != nil || lastLog == nil {
		// Если нет логов уборки, бокс можно убирать
		return true
	}

	// Проверяем, что последняя уборка была завершена
	if lastLog.Status != models.CleaningLogStatusCompleted || lastLog.CompletedAt == nil {
		// Если уборка не завершена, бокс можно убирать
		return true
	}

	// Проверяем, что после последней уборки была хотя бы одна завершенная сессия
	completedSessionsCount, err := s.sessionRepo.GetCompletedSessionsBetween(ctx,
		boxID,
		*lastLog.CompletedAt,
		time.Now(),
	)
	if err != nil {
		// В случае ошибки разрешаем уборку
		return true
	}

	// Если есть завершенные сессии между уборками, бокс можно убирать
	return completedSessionsCount > 0
}

// CleanerReserveCleaning резервирует уборку для бокса
func (s *ServiceImpl) CleanerReserveCleaning(ctx context.Context, req *models.CleanerReserveCleaningRequest, cleanerID uuid.UUID) (*models.CleanerReserveCleaningResponse, error) {
	// Проверяем, что бокс существует
	washBox, err := s.repo.GetWashBoxByID(ctx, req.WashBoxID)
	if err != nil {
		return nil, err
	}

	// Проверяем, что бокс не в статусе cleaning
	if washBox.Status == models.StatusCleaning {
		return nil, errors.New("бокс уже на уборке")
	}

	// Проверяем, что уборщик не убирает другой бокс
	activeLog, err := s.repo.GetActiveCleaningLogByCleaner(ctx, cleanerID)
	if err == nil && activeLog != nil {
		return nil, errors.New("уборщик уже убирает другой бокс")
	}

	// Проверяем, что бокс не убирался в предыдущей сессии
	lastLog, err := s.repo.GetLastCleaningLogByBox(ctx, req.WashBoxID)
	if err == nil && lastLog != nil {
		// Проверяем, что последняя уборка была завершена
		if lastLog.Status == models.CleaningLogStatusCompleted && lastLog.CompletedAt != nil {
			// Проверяем, что после последней уборки была хотя бы одна завершенная сессия
			completedSessionsCount, err := s.sessionRepo.GetCompletedSessionsBetween(ctx,
				req.WashBoxID,
				*lastLog.CompletedAt,
				time.Now(),
			)
			if err != nil {
				return nil, err
			}

			// Если нет завершенных сессий между уборками, запрещаем повторную уборку
			if completedSessionsCount == 0 {
				return nil, errors.New("уборщик не может 2 раза подряд брать уборку на одном и том же боксе, если между его уборками не было завершенной сессии")
			}
		}
	}

	// Резервируем уборку
	err = s.repo.ReserveCleaning(ctx, req.WashBoxID, cleanerID)
	if err != nil {
		return nil, err
	}

	return &models.CleanerReserveCleaningResponse{
		Success: true,
	}, nil
}

// CleanerStartCleaning начинает уборку бокса
func (s *ServiceImpl) CleanerStartCleaning(ctx context.Context, req *models.CleanerStartCleaningRequest, cleanerID uuid.UUID) (*models.CleanerStartCleaningResponse, error) {
	// Проверяем, что бокс существует
	washBox, err := s.repo.GetWashBoxByID(ctx, req.WashBoxID)
	if err != nil {
		return nil, err
	}

	// Проверяем, что бокс свободен или зарезервирован этим уборщиком
	if washBox.Status != models.StatusFree &&
		(washBox.CleaningReservedBy == nil || *washBox.CleaningReservedBy != cleanerID) {
		return nil, errors.New("бокс недоступен для уборки")
	}

	// Проверяем, что уборщик не убирает другой бокс
	activeLog, err := s.repo.GetActiveCleaningLogByCleaner(ctx, cleanerID)
	if err == nil && activeLog != nil {
		return nil, errors.New("уборщик уже убирает другой бокс")
	}

	// Проверяем, что бокс не убирался в предыдущей сессии
	lastLog, err := s.repo.GetLastCleaningLogByBox(ctx, req.WashBoxID)
	if err == nil && lastLog != nil {
		// Проверяем, что последняя уборка была завершена
		if lastLog.Status == models.CleaningLogStatusCompleted && lastLog.CompletedAt != nil {
			// Проверяем, что после последней уборки была хотя бы одна завершенная сессия
			completedSessionsCount, err := s.sessionRepo.GetCompletedSessionsBetween(ctx,
				req.WashBoxID,
				*lastLog.CompletedAt,
				time.Now(),
			)
			if err != nil {
				return nil, err
			}

			// Если нет завершенных сессий между уборками, запрещаем повторную уборку
			if completedSessionsCount == 0 {
				return nil, errors.New("уборщик не может 2 раза подряд брать уборку на одном и том же боксе, если между его уборками не было завершенной сессии")
			}
		}
	}

	// Начинаем уборку
	startedAt := time.Now()
	err = s.repo.StartCleaning(ctx, req.WashBoxID)
	if err != nil {
		return nil, err
	}

	// Обновляем время начала уборки в боксе
	err = s.repo.UpdateCleaningStartedAt(ctx, req.WashBoxID, startedAt)
	if err != nil {
		return nil, err
	}

	// Создаем лог уборки
	cleaningLog := &models.CleaningLog{
		CleanerID: cleanerID,
		WashBoxID: req.WashBoxID,
		StartedAt: startedAt,
		Status:    models.CleaningLogStatusInProgress,
	}

	err = s.repo.CreateCleaningLog(ctx, cleaningLog)
	if err != nil {
		return nil, err
	}

	// Включаем свет в боксе, если задан регистр и доступен адаптер Modbus
	if s.modbusAdapter != nil && washBox.LightCoilRegister != nil && *washBox.LightCoilRegister != "" {
		_ = s.modbusAdapter.WriteLightCoil(ctx, req.WashBoxID, *washBox.LightCoilRegister, true)
	}

	return &models.CleanerStartCleaningResponse{
		Success: true,
	}, nil
}

// CleanerCancelCleaning отменяет резервирование уборки
func (s *ServiceImpl) CleanerCancelCleaning(ctx context.Context, req *models.CleanerCancelCleaningRequest, cleanerID uuid.UUID) (*models.CleanerCancelCleaningResponse, error) {
	// Проверяем, что бокс существует
	washBox, err := s.repo.GetWashBoxByID(ctx, req.WashBoxID)
	if err != nil {
		return nil, err
	}

	// Проверяем, что уборка зарезервирована этим уборщиком
	if washBox.CleaningReservedBy == nil || *washBox.CleaningReservedBy != cleanerID {
		return nil, errors.New("уборка не зарезервирована этим уборщиком")
	}

	// Отменяем резервирование
	err = s.repo.CancelCleaning(ctx, req.WashBoxID)
	if err != nil {
		return nil, err
	}

	return &models.CleanerCancelCleaningResponse{
		Success: true,
	}, nil
}

// CleanerCompleteCleaning завершает уборку бокса
func (s *ServiceImpl) CleanerCompleteCleaning(ctx context.Context, req *models.CleanerCompleteCleaningRequest, cleanerID uuid.UUID) (*models.CleanerCompleteCleaningResponse, error) {
	// Проверяем, что бокс существует и в статусе cleaning
	washBox, err := s.repo.GetWashBoxByID(ctx, req.WashBoxID)
	if err != nil {
		return nil, err
	}

	if washBox.Status != models.StatusCleaning {
		return nil, errors.New("бокс не на уборке")
	}

	// Находим активный лог уборки для этого уборщика и бокса
	activeLog, err := s.repo.GetActiveCleaningLogByCleaner(ctx, cleanerID)
	if err != nil || activeLog == nil || activeLog.WashBoxID != req.WashBoxID {
		return nil, errors.New("активная уборка не найдена")
	}

	// Завершаем уборку
	completedAt := time.Now()
	err = s.repo.CompleteCleaning(ctx, req.WashBoxID)
	if err != nil {
		return nil, err
	}

	// Обновляем лог уборки
	duration := int(completedAt.Sub(activeLog.StartedAt).Minutes())
	activeLog.CompletedAt = &completedAt
	activeLog.DurationMinutes = &duration
	activeLog.Status = models.CleaningLogStatusCompleted

	err = s.repo.UpdateCleaningLog(ctx, activeLog)
	if err != nil {
		return nil, err
	}

	// Выключаем свет в боксе, если задан регистр и доступен адаптер Modbus
	if s.modbusAdapter != nil && washBox.LightCoilRegister != nil && *washBox.LightCoilRegister != "" {
		_ = s.modbusAdapter.WriteLightCoil(ctx, req.WashBoxID, *washBox.LightCoilRegister, false)
	}

	return &models.CleanerCompleteCleaningResponse{
		Success: true,
	}, nil
}

// GetCleaningBoxes получает все боксы в статусе уборки
func (s *ServiceImpl) GetCleaningBoxes(ctx context.Context) ([]models.WashBox, error) {
	return s.repo.GetCleaningBoxes(ctx)
}

// AdminListCleaningLogs получает логи уборки с фильтрами (админка)
func (s *ServiceImpl) AdminListCleaningLogs(ctx context.Context, req *models.AdminListCleaningLogsRequest) (*models.AdminListCleaningLogsResponse, error) {

	// Конвертируем строковые параметры в нужные типы
	convertedReq := &models.AdminListCleaningLogsInternalRequest{
		Limit:  req.Limit,
		Offset: req.Offset,
		Status: req.Status,
	}

	// Конвертируем даты из строк в time.Time (поддерживаем ISO 8601 с timezone)
	if req.DateFrom != nil && *req.DateFrom != "" {
		dateFrom, err := time.Parse(time.RFC3339, *req.DateFrom)
		if err != nil {
			// Пробуем парсить как datetime-local для обратной совместимости
			dateFrom, err = time.Parse("2006-01-02T15:04", *req.DateFrom)
			if err != nil {
				return nil, errors.New("неверный формат даты начала")
			}
		}
		convertedReq.DateFrom = &dateFrom
	}

	if req.DateTo != nil && *req.DateTo != "" {
		dateTo, err := time.Parse(time.RFC3339, *req.DateTo)
		if err != nil {
			// Пробуем парсить как datetime-local для обратной совместимости
			dateTo, err = time.Parse("2006-01-02T15:04", *req.DateTo)
			if err != nil {
				return nil, errors.New("неверный формат даты окончания")
			}
		}
		convertedReq.DateTo = &dateTo
	}

	// Получаем логи
	logs, err := s.repo.GetCleaningLogs(ctx, convertedReq)
	if err != nil {
		return nil, err
	}

	// Получаем общее количество
	total, err := s.repo.GetCleaningLogsCount(ctx, convertedReq)
	if err != nil {
		return nil, err
	}

	// Устанавливаем значения по умолчанию для пагинации
	limit := 20
	offset := 0
	if req.Limit != nil {
		limit = *req.Limit
	}
	if req.Offset != nil {
		offset = *req.Offset
	}

	return &models.AdminListCleaningLogsResponse{
		CleaningLogs: logs,
		Total:        int(total),
		Limit:        limit,
		Offset:       offset,
	}, nil
}

// AutoCompleteExpiredCleanings автоматически завершает просроченные уборки
func (s *ServiceImpl) AutoCompleteExpiredCleanings(ctx context.Context) error {
	// Получаем настройку времени уборки
	timeoutMinutes := 3 // По умолчанию 3 минуты
	if s.settingsService != nil {
		timeout, err := s.settingsService.GetCleaningTimeout(ctx)
		if err == nil {
			timeoutMinutes = timeout
		}
	}

	// Получаем просроченные логи уборки
	expiredLogs, err := s.repo.GetExpiredCleaningLogs(ctx, timeoutMinutes)
	if err != nil {
		return err
	}

	// Завершаем каждую просроченную уборку
	for _, log := range expiredLogs {
		// Завершаем уборку бокса
		err := s.repo.CompleteCleaning(ctx, log.WashBoxID)
		if err != nil {
			continue // Продолжаем с другими, если одна не удалась
		}

		// Обновляем лог уборки
		completedAt := time.Now()
		duration := int(completedAt.Sub(log.StartedAt).Minutes())
		log.CompletedAt = &completedAt
		log.DurationMinutes = &duration
		log.Status = models.CleaningLogStatusCompleted

		err = s.repo.UpdateCleaningLog(ctx, &log)
		if err != nil {
			continue // Продолжаем с другими, если одна не удалась
		}

		// Пытаемся выключить свет, если известен регистр
		if s.modbusAdapter != nil {
			// Получаем бокс для доступа к регистру света
			if box, getErr := s.repo.GetWashBoxByID(ctx, log.WashBoxID); getErr == nil && box != nil && box.LightCoilRegister != nil && *box.LightCoilRegister != "" {
				_ = s.modbusAdapter.WriteLightCoil(ctx, log.WashBoxID, *box.LightCoilRegister, false)
			}
		}
	}

	return nil
}

// UpdateCleaningStartedAt обновляет время начала уборки
func (s *ServiceImpl) UpdateCleaningStartedAt(ctx context.Context, washBoxID uuid.UUID, startedAt time.Time) error {
	return s.repo.UpdateCleaningStartedAt(ctx, washBoxID, startedAt)
}

// ClearCleaningReservation очищает резерв уборки
func (s *ServiceImpl) ClearCleaningReservation(ctx context.Context, washBoxID uuid.UUID) error {
	return s.repo.CancelCleaning(ctx, washBoxID)
}

// SetCooldown устанавливает cooldown для бокса после завершения сессии
func (s *ServiceImpl) SetCooldown(ctx context.Context, boxID uuid.UUID, userID uuid.UUID, cooldownUntil time.Time) error {
	return s.repo.SetCooldown(ctx, boxID, userID, cooldownUntil)
}

// SetCooldownByCarNumber устанавливает cooldown для бокса по госномеру после завершения сессии
func (s *ServiceImpl) SetCooldownByCarNumber(ctx context.Context, boxID uuid.UUID, carNumber string, cooldownUntil time.Time) error {
	return s.repo.SetCooldownByCarNumber(ctx, boxID, carNumber, cooldownUntil)
}

// ClearCooldown очищает поля кулдауна у бокса
func (s *ServiceImpl) ClearCooldown(ctx context.Context, boxID uuid.UUID) error {
	return s.repo.ClearCooldown(ctx, boxID)
}

// CheckCooldownExpired очищает истекшие cooldown'ы
func (s *ServiceImpl) CheckCooldownExpired(ctx context.Context) error {
	return s.repo.CheckCooldownExpired(ctx)
}
