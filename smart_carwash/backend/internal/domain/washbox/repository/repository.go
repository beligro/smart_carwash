package repository

import (
	"carwash_backend/internal/domain/washbox/models"
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// Repository интерфейс для работы с боксами мойки в базе данных
type Repository interface {
	GetAllWashBoxes() ([]models.WashBox, error)
	GetWashBoxByID(id uuid.UUID) (*models.WashBox, error)
	UpdateWashBoxStatus(id uuid.UUID, status string) error
	CreateWashBox(box *models.WashBox) (*models.WashBox, error)
	GetFreeWashBoxes() ([]models.WashBox, error)
	GetFreeWashBoxesByServiceType(serviceType string) ([]models.WashBox, error)
	GetFreeWashBoxesWithChemistry(serviceType string) ([]models.WashBox, error)
	GetWashBoxesByServiceType(serviceType string) ([]models.WashBox, error)

	// Административные методы
	GetWashBoxByNumber(number int) (*models.WashBox, error)
	GetWashBoxByNumberIncludingDeleted(number int) (*models.WashBox, error)
	UpdateWashBox(box *models.WashBox) (*models.WashBox, error)
	DeleteWashBox(id uuid.UUID) error
	GetWashBoxesWithFilters(status *string, serviceType *string, limit int, offset int) ([]models.WashBox, int, error)
	RestoreWashBox(id uuid.UUID, status string, serviceType string) (*models.WashBox, error)

	// Методы для уборщиков
	GetWashBoxesForCleaner(limit int, offset int) ([]models.WashBox, int, error)
	ReserveCleaning(washBoxID uuid.UUID, cleanerID uuid.UUID) error
	StartCleaning(washBoxID uuid.UUID) error
	CancelCleaning(washBoxID uuid.UUID) error
	CompleteCleaning(washBoxID uuid.UUID) error
	GetCleaningBoxes() ([]models.WashBox, error)
	UpdateCleaningStartedAt(washBoxID uuid.UUID, startedAt time.Time) error

	// Методы для логов уборки
	CreateCleaningLog(log *models.CleaningLog) error
	UpdateCleaningLog(log *models.CleaningLog) error
	GetCleaningLogs(req *models.AdminListCleaningLogsInternalRequest) ([]models.CleaningLogWithDetails, error)
	GetCleaningLogsCount(req *models.AdminListCleaningLogsInternalRequest) (int64, error)
	GetActiveCleaningLogByCleaner(cleanerID uuid.UUID) (*models.CleaningLog, error)
	GetLastCleaningLogByBox(washBoxID uuid.UUID) (*models.CleaningLog, error)
	GetExpiredCleaningLogs(timeoutMinutes int) ([]models.CleaningLog, error)

	// Методы для работы с cooldown
	SetCooldown(boxID uuid.UUID, userID uuid.UUID, cooldownUntil time.Time) error
	SetCooldownByCarNumber(boxID uuid.UUID, carNumber string, cooldownUntil time.Time) error
	GetCooldownBoxesForUser(userID uuid.UUID) ([]models.WashBox, error)
	GetCooldownBoxesByCarNumber(carNumber string) ([]models.WashBox, error)
	CheckCooldownExpired() error
}

// PostgresRepository реализация Repository для PostgreSQL
type PostgresRepository struct {
	db *gorm.DB
}

// NewPostgresRepository создает новый экземпляр PostgresRepository
func NewPostgresRepository(db *gorm.DB) *PostgresRepository {
	return &PostgresRepository{db: db}
}

// GetAllWashBoxes получает все боксы мойки
func (r *PostgresRepository) GetAllWashBoxes() ([]models.WashBox, error) {
	var boxes []models.WashBox
	err := r.db.Find(&boxes).Error
	return boxes, err
}

// GetWashBoxByID получает бокс мойки по ID
func (r *PostgresRepository) GetWashBoxByID(id uuid.UUID) (*models.WashBox, error) {
	var box models.WashBox
	err := r.db.First(&box, id).Error
	if err != nil {
		return nil, err
	}
	return &box, nil
}

// UpdateWashBoxStatus обновляет статус бокса мойки
func (r *PostgresRepository) UpdateWashBoxStatus(id uuid.UUID, status string) error {
	return r.db.Model(&models.WashBox{}).Where("id = ?", id).Update("status", status).Error
}

// CreateWashBox создает новый бокс мойки
func (r *PostgresRepository) CreateWashBox(box *models.WashBox) (*models.WashBox, error) {
	err := r.db.Create(box).Error
	if err != nil {
		return nil, err
	}
	return box, nil
}

// GetFreeWashBoxes получает все свободные боксы мойки, отсортированные по приоритету
func (r *PostgresRepository) GetFreeWashBoxes() ([]models.WashBox, error) {
	var boxes []models.WashBox
	err := r.db.Where("status = ?", models.StatusFree).Order("priority ASC, number ASC").Find(&boxes).Error
	return boxes, err
}

// GetFreeWashBoxesByServiceType получает все свободные боксы мойки определенного типа, отсортированные по приоритету
func (r *PostgresRepository) GetFreeWashBoxesByServiceType(serviceType string) ([]models.WashBox, error) {
	var boxes []models.WashBox
	err := r.db.Where("status = ? AND service_type = ?", models.StatusFree, serviceType).Order("priority ASC, number ASC").Find(&boxes).Error
	return boxes, err
}

// GetFreeWashBoxesWithChemistry получает все свободные боксы мойки с химией определенного типа, отсортированные по приоритету
func (r *PostgresRepository) GetFreeWashBoxesWithChemistry(serviceType string) ([]models.WashBox, error) {
	var boxes []models.WashBox
	err := r.db.Where("status = ? AND service_type = ? AND chemistry_enabled = ?", models.StatusFree, serviceType, true).Order("priority ASC, number ASC").Find(&boxes).Error
	return boxes, err
}

// GetWashBoxesByServiceType получает все боксы мойки определенного типа
func (r *PostgresRepository) GetWashBoxesByServiceType(serviceType string) ([]models.WashBox, error) {
	var boxes []models.WashBox
	err := r.db.Where("service_type = ?", serviceType).Find(&boxes).Error
	return boxes, err
}

// GetWashBoxByNumber получает бокс мойки по номеру
func (r *PostgresRepository) GetWashBoxByNumber(number int) (*models.WashBox, error) {
	var box models.WashBox
	err := r.db.Where("number = ?", number).First(&box).Error
	if err != nil {
		return nil, err
	}
	return &box, nil
}

// GetWashBoxByNumberIncludingDeleted получает бокс мойки по номеру, включая удаленные
func (r *PostgresRepository) GetWashBoxByNumberIncludingDeleted(number int) (*models.WashBox, error) {
	var box models.WashBox
	err := r.db.Unscoped().Where("number = ?", number).First(&box).Error
	if err != nil {
		return nil, err
	}
	return &box, nil
}

// UpdateWashBox обновляет бокс мойки
func (r *PostgresRepository) UpdateWashBox(box *models.WashBox) (*models.WashBox, error) {
	err := r.db.Save(box).Error
	if err != nil {
		return nil, err
	}
	return box, nil
}

// DeleteWashBox удаляет бокс мойки
func (r *PostgresRepository) DeleteWashBox(id uuid.UUID) error {
	return r.db.Delete(&models.WashBox{}, id).Error
}

// GetWashBoxesWithFilters получает боксы мойки с фильтрацией
func (r *PostgresRepository) GetWashBoxesWithFilters(status *string, serviceType *string, limit int, offset int) ([]models.WashBox, int, error) {
	var boxes []models.WashBox
	var total int64

	query := r.db.Model(&models.WashBox{})

	// Применяем фильтры
	if status != nil {
		query = query.Where("status = ?", *status)
	}
	if serviceType != nil {
		query = query.Where("service_type = ?", *serviceType)
	}

	// Получаем общее количество
	err := query.Count(&total).Error
	if err != nil {
		return nil, 0, err
	}

	// Получаем данные с пагинацией и сортировкой по номеру
	err = query.Order("number ASC").Limit(limit).Offset(offset).Find(&boxes).Error
	if err != nil {
		return nil, 0, err
	}

	return boxes, int(total), nil
}

// RestoreWashBox восстанавливает удаленный бокс мойки
func (r *PostgresRepository) RestoreWashBox(id uuid.UUID, status string, serviceType string) (*models.WashBox, error) {
	// Получаем удаленный бокс
	var box models.WashBox
	err := r.db.Unscoped().First(&box, id).Error
	if err != nil {
		return nil, err
	}

	// Обновляем бокс: убираем deleted_at и обновляем статус и тип услуги
	box.Status = status
	box.ServiceType = serviceType
	box.DeletedAt = gorm.DeletedAt{}

	// Сохраняем обновленный бокс
	updatedBox, err := r.UpdateWashBox(&box)
	if err != nil {
		return nil, err
	}

	return updatedBox, nil
}

// GetWashBoxesForCleaner получает список боксов для уборщика
func (r *PostgresRepository) GetWashBoxesForCleaner(limit int, offset int) ([]models.WashBox, int, error) {
	var boxes []models.WashBox
	var total int64

	// Подсчитываем общее количество
	err := r.db.Model(&models.WashBox{}).Count(&total).Error
	if err != nil {
		return nil, 0, err
	}

	// Получаем боксы с пагинацией
	query := r.db.Order("number ASC")
	if limit > 0 {
		query = query.Limit(limit)
	}
	if offset > 0 {
		query = query.Offset(offset)
	}

	err = query.Find(&boxes).Error
	if err != nil {
		return nil, 0, err
	}

	return boxes, int(total), nil
}

// ReserveCleaning резервирует уборку для бокса
func (r *PostgresRepository) ReserveCleaning(washBoxID uuid.UUID, cleanerID uuid.UUID) error {
	return r.db.Model(&models.WashBox{}).
		Where("id = ?", washBoxID).
		Update("cleaning_reserved_by", cleanerID).Error
}

// StartCleaning начинает уборку бокса
func (r *PostgresRepository) StartCleaning(washBoxID uuid.UUID) error {
	now := time.Now()
	return r.db.Model(&models.WashBox{}).
		Where("id = ?", washBoxID).
		Updates(map[string]interface{}{
			"status":               models.StatusCleaning,
			"cleaning_started_at":  now,
			"cleaning_reserved_by": nil,
		}).Error
}

// CancelCleaning отменяет резервирование уборки
func (r *PostgresRepository) CancelCleaning(washBoxID uuid.UUID) error {
	return r.db.Model(&models.WashBox{}).
		Where("id = ?", washBoxID).
		Update("cleaning_reserved_by", nil).Error
}

// CompleteCleaning завершает уборку бокса
func (r *PostgresRepository) CompleteCleaning(washBoxID uuid.UUID) error {
	return r.db.Model(&models.WashBox{}).
		Where("id = ?", washBoxID).
		Updates(map[string]interface{}{
			"status":               models.StatusFree,
			"cleaning_started_at":  nil,
			"cleaning_reserved_by": nil,
		}).Error
}

// GetCleaningBoxes получает все боксы в статусе уборки
func (r *PostgresRepository) GetCleaningBoxes() ([]models.WashBox, error) {
	var boxes []models.WashBox
	err := r.db.Where("status = ?", models.StatusCleaning).Find(&boxes).Error
	return boxes, err
}

// UpdateCleaningStartedAt обновляет время начала уборки
func (r *PostgresRepository) UpdateCleaningStartedAt(washBoxID uuid.UUID, startedAt time.Time) error {
	return r.db.Model(&models.WashBox{}).
		Where("id = ?", washBoxID).
		Update("cleaning_started_at", startedAt).Error
}

// CreateCleaningLog создает новый лог уборки
func (r *PostgresRepository) CreateCleaningLog(log *models.CleaningLog) error {
	return r.db.Create(log).Error
}

// UpdateCleaningLog обновляет лог уборки
func (r *PostgresRepository) UpdateCleaningLog(log *models.CleaningLog) error {
	return r.db.Save(log).Error
}

// GetCleaningLogs получает логи уборки с фильтрами
func (r *PostgresRepository) GetCleaningLogs(req *models.AdminListCleaningLogsInternalRequest) ([]models.CleaningLogWithDetails, error) {
	var logs []models.CleaningLogWithDetails


	query := r.db.Table("cleaning_logs cl").
		Select(`
			cl.*,
			c.username as cleaner_username,
			wb.number as wash_box_number,
			wb.service_type as wash_box_type
		`).
		Joins("LEFT JOIN cleaners c ON cl.cleaner_id = c.id").
		Joins("LEFT JOIN wash_boxes wb ON cl.wash_box_id = wb.id")

	// Применяем фильтры
	if req.Status != nil {
		query = query.Where("cl.status = ?", *req.Status)
	}

	if req.DateFrom != nil {
		query = query.Where("cl.started_at >= ?", *req.DateFrom)
	}

	if req.DateTo != nil {
		query = query.Where("cl.started_at <= ?", *req.DateTo)
	}

	// Сортировка по времени начала (новые сначала)
	query = query.Order("cl.started_at DESC")

	// Пагинация
	if req.Limit != nil && *req.Limit > 0 {
		query = query.Limit(*req.Limit)
	}

	if req.Offset != nil && *req.Offset > 0 {
		query = query.Offset(*req.Offset)
	}

	err := query.Scan(&logs).Error
	return logs, err
}

// GetCleaningLogsCount получает общее количество логов уборки с фильтрами
func (r *PostgresRepository) GetCleaningLogsCount(req *models.AdminListCleaningLogsInternalRequest) (int64, error) {
	var count int64

	query := r.db.Model(&models.CleaningLog{})

	// Применяем те же фильтры
	if req.Status != nil {
		query = query.Where("status = ?", *req.Status)
	}

	if req.DateFrom != nil {
		query = query.Where("started_at >= ?", *req.DateFrom)
	}

	if req.DateTo != nil {
		query = query.Where("started_at <= ?", *req.DateTo)
	}

	err := query.Count(&count).Error
	return count, err
}

// GetActiveCleaningLogByCleaner получает активный лог уборки для уборщика
func (r *PostgresRepository) GetActiveCleaningLogByCleaner(cleanerID uuid.UUID) (*models.CleaningLog, error) {
	var log models.CleaningLog
	err := r.db.Where("cleaner_id = ? AND status = ?", cleanerID, models.CleaningLogStatusInProgress).
		First(&log).Error
	if err != nil {
		return nil, err
	}
	return &log, nil
}

// GetLastCleaningLogByBox получает последний лог уборки для бокса
func (r *PostgresRepository) GetLastCleaningLogByBox(washBoxID uuid.UUID) (*models.CleaningLog, error) {
	var log models.CleaningLog
	err := r.db.Where("wash_box_id = ?", washBoxID).
		Order("started_at DESC").
		First(&log).Error
	if err != nil {
		return nil, err
	}
	return &log, nil
}

// GetExpiredCleaningLogs получает логи уборки, которые нужно автоматически завершить
func (r *PostgresRepository) GetExpiredCleaningLogs(timeoutMinutes int) ([]models.CleaningLog, error) {
	var logs []models.CleaningLog
	timeout := time.Now().Add(-time.Duration(timeoutMinutes) * time.Minute)

	err := r.db.Where("status = ? AND started_at <= ?",
		models.CleaningLogStatusInProgress, timeout).
		Find(&logs).Error
	return logs, err
}

// SetCooldown устанавливает cooldown для бокса после завершения сессии
func (r *PostgresRepository) SetCooldown(boxID uuid.UUID, userID uuid.UUID, cooldownUntil time.Time) error {
	return r.db.Model(&models.WashBox{}).
		Where("id = ?", boxID).
		Updates(map[string]interface{}{
			"last_completed_session_user_id": userID,
			"last_completed_at":              time.Now(),
			"cooldown_until":                 cooldownUntil,
		}).Error
}

// GetCooldownBoxesForUser получает боксы в cooldown для конкретного пользователя
func (r *PostgresRepository) GetCooldownBoxesForUser(userID uuid.UUID) ([]models.WashBox, error) {
	var boxes []models.WashBox
	now := time.Now()

	err := r.db.Where("last_completed_session_user_id = ? AND cooldown_until > ?",
		userID, now).
		Find(&boxes).Error
	return boxes, err
}

// SetCooldownByCarNumber устанавливает cooldown для бокса по госномеру после завершения сессии
func (r *PostgresRepository) SetCooldownByCarNumber(boxID uuid.UUID, carNumber string, cooldownUntil time.Time) error {
	return r.db.Model(&models.WashBox{}).
		Where("id = ?", boxID).
		Updates(map[string]interface{}{
			"last_completed_session_car_number": carNumber,
			"last_completed_at":                 time.Now(),
			"cooldown_until":                    cooldownUntil,
		}).Error
}

// GetCooldownBoxesByCarNumber получает боксы в cooldown для конкретного госномера
func (r *PostgresRepository) GetCooldownBoxesByCarNumber(carNumber string) ([]models.WashBox, error) {
	var boxes []models.WashBox
	now := time.Now()

	err := r.db.Where("last_completed_session_car_number = ? AND cooldown_until > ?",
		carNumber, now).
		Find(&boxes).Error
	return boxes, err
}

// CheckCooldownExpired очищает истекшие cooldown'ы
func (r *PostgresRepository) CheckCooldownExpired() error {
	now := time.Now()
	return r.db.Model(&models.WashBox{}).
		Where("cooldown_until IS NOT NULL AND cooldown_until <= ?", now).
		Updates(map[string]interface{}{
			"last_completed_session_user_id":     nil,
			"last_completed_session_car_number":   nil,
			"last_completed_at":                  nil,
			"cooldown_until":                    nil,
			"status":                             models.StatusFree,
		}).Error
}
