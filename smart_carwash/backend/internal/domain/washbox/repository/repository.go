package repository

import (
	"carwash_backend/internal/domain/washbox/models"
	"context"
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// Repository интерфейс для работы с боксами мойки в базе данных
type Repository interface {
	GetAllWashBoxes(ctx context.Context) ([]models.WashBox, error)
	GetWashBoxByID(ctx context.Context, id uuid.UUID) (*models.WashBox, error)
	UpdateWashBoxStatus(ctx context.Context, id uuid.UUID, status string) error
	CreateWashBox(ctx context.Context, box *models.WashBox) (*models.WashBox, error)
	GetFreeWashBoxes(ctx context.Context) ([]models.WashBox, error)
	GetFreeWashBoxesByServiceType(ctx context.Context, serviceType string) ([]models.WashBox, error)
	GetFreeWashBoxesWithChemistry(ctx context.Context, serviceType string) ([]models.WashBox, error)
	GetWashBoxesByServiceType(ctx context.Context, serviceType string) ([]models.WashBox, error)

	// Административные методы
	GetWashBoxByNumber(ctx context.Context, number int) (*models.WashBox, error)
	GetWashBoxByNumberIncludingDeleted(ctx context.Context, number int) (*models.WashBox, error)
	UpdateWashBox(ctx context.Context, box *models.WashBox) (*models.WashBox, error)
	DeleteWashBox(ctx context.Context, id uuid.UUID) error
	GetWashBoxesWithFilters(ctx context.Context, status *string, serviceType *string, limit int, offset int) ([]models.WashBox, int, error)
	RestoreWashBox(ctx context.Context, id uuid.UUID, status string, serviceType string) (*models.WashBox, error)

	// Методы для уборщиков
	GetWashBoxesForCleaner(ctx context.Context, limit int, offset int) ([]models.WashBox, int, error)
	ReserveCleaning(ctx context.Context, washBoxID uuid.UUID, cleanerID uuid.UUID) error
	StartCleaning(ctx context.Context, washBoxID uuid.UUID) error
	CancelCleaning(ctx context.Context, washBoxID uuid.UUID) error
	CompleteCleaning(ctx context.Context, washBoxID uuid.UUID) error
	GetCleaningBoxes(ctx context.Context) ([]models.WashBox, error)
	UpdateCleaningStartedAt(ctx context.Context, washBoxID uuid.UUID, startedAt time.Time) error

	// Методы для логов уборки
	CreateCleaningLog(ctx context.Context, log *models.CleaningLog) error
	UpdateCleaningLog(ctx context.Context, log *models.CleaningLog) error
	GetCleaningLogs(ctx context.Context, req *models.AdminListCleaningLogsInternalRequest) ([]models.CleaningLogWithDetails, error)
	GetCleaningLogsCount(ctx context.Context, req *models.AdminListCleaningLogsInternalRequest) (int64, error)
	GetActiveCleaningLogByCleaner(ctx context.Context, cleanerID uuid.UUID) (*models.CleaningLog, error)
	GetLastCleaningLogByBox(ctx context.Context, washBoxID uuid.UUID) (*models.CleaningLog, error)
	GetExpiredCleaningLogs(ctx context.Context, timeoutMinutes int) ([]models.CleaningLog, error)

	// Методы для работы с cooldown
	SetCooldown(ctx context.Context, boxID uuid.UUID, userID uuid.UUID, cooldownUntil time.Time) error
	SetCooldownByCarNumber(ctx context.Context, boxID uuid.UUID, carNumber string, cooldownUntil time.Time) error
	GetCooldownBoxesForUser(ctx context.Context, userID uuid.UUID) ([]models.WashBox, error)
	GetCooldownBoxesByCarNumber(ctx context.Context, carNumber string) ([]models.WashBox, error)
	ClearCooldown(ctx context.Context, boxID uuid.UUID) error
	CheckCooldownExpired(ctx context.Context) error
}

// PostgresRepository реализация Repository для PostgreSQL
type PostgresRepository struct {
	db *gorm.DB
}

// NewPostgresRepository создает новый экземпляр PostgresRepository
func NewPostgresRepository(db *gorm.DB) *PostgresRepository {
	return &PostgresRepository{
		db: db,
	}
}

// GetAllWashBoxes получает все боксы мойки
func (r *PostgresRepository) GetAllWashBoxes(ctx context.Context) ([]models.WashBox, error) {
	var boxes []models.WashBox
	err := r.db.WithContext(ctx).Find(&boxes).Error
	return boxes, err
}

// GetWashBoxByID получает бокс мойки по ID
func (r *PostgresRepository) GetWashBoxByID(ctx context.Context, id uuid.UUID) (*models.WashBox, error) {
	var box models.WashBox
	err := r.db.WithContext(ctx).First(&box, id).Error
	if err != nil {
		return nil, err
	}
	return &box, nil
}

// UpdateWashBoxStatus обновляет статус бокса мойки
func (r *PostgresRepository) UpdateWashBoxStatus(ctx context.Context, id uuid.UUID, status string) error {
	return r.db.WithContext(ctx).Model(&models.WashBox{}).Where("id = ?", id).Update("status", status).Error
}

// CreateWashBox создает новый бокс мойки
func (r *PostgresRepository) CreateWashBox(ctx context.Context, box *models.WashBox) (*models.WashBox, error) {
	err := r.db.WithContext(ctx).Create(box).Error
	if err != nil {
		return nil, err
	}
	return box, nil
}

// GetFreeWashBoxes получает все свободные боксы мойки, отсортированные по приоритету
func (r *PostgresRepository) GetFreeWashBoxes(ctx context.Context) ([]models.WashBox, error) {
	var boxes []models.WashBox
	err := r.db.WithContext(ctx).Where("status = ?", models.StatusFree).Order("priority ASC, number ASC").Find(&boxes).Error
	return boxes, err
}

// GetFreeWashBoxesByServiceType получает все свободные боксы мойки определенного типа, отсортированные по приоритету
func (r *PostgresRepository) GetFreeWashBoxesByServiceType(ctx context.Context, serviceType string) ([]models.WashBox, error) {
	var boxes []models.WashBox
	err := r.db.WithContext(ctx).Where("status = ? AND service_type = ?", models.StatusFree, serviceType).Order("priority ASC, number ASC").Find(&boxes).Error
	return boxes, err
}

// GetFreeWashBoxesWithChemistry получает все свободные боксы мойки с химией определенного типа, отсортированные по приоритету
func (r *PostgresRepository) GetFreeWashBoxesWithChemistry(ctx context.Context, serviceType string) ([]models.WashBox, error) {
	var boxes []models.WashBox
	err := r.db.WithContext(ctx).Where("status = ? AND service_type = ? AND chemistry_enabled = ?", models.StatusFree, serviceType, true).Order("priority ASC, number ASC").Find(&boxes).Error
	return boxes, err
}

// GetWashBoxesByServiceType получает все боксы мойки определенного типа
func (r *PostgresRepository) GetWashBoxesByServiceType(ctx context.Context, serviceType string) ([]models.WashBox, error) {
	var boxes []models.WashBox
	err := r.db.WithContext(ctx).Where("service_type = ?", serviceType).Find(&boxes).Error
	return boxes, err
}

// GetWashBoxByNumber получает бокс мойки по номеру
func (r *PostgresRepository) GetWashBoxByNumber(ctx context.Context, number int) (*models.WashBox, error) {
	var box models.WashBox
	err := r.db.WithContext(ctx).Where("number = ?", number).First(&box).Error
	if err != nil {
		return nil, err
	}
	return &box, nil
}

// GetWashBoxByNumberIncludingDeleted получает бокс мойки по номеру, включая удаленные
func (r *PostgresRepository) GetWashBoxByNumberIncludingDeleted(ctx context.Context, number int) (*models.WashBox, error) {
	var box models.WashBox
	err := r.db.WithContext(ctx).Unscoped().Where("number = ?", number).First(&box).Error
	if err != nil {
		return nil, err
	}
	return &box, nil
}

// UpdateWashBox обновляет бокс мойки
func (r *PostgresRepository) UpdateWashBox(ctx context.Context, box *models.WashBox) (*models.WashBox, error) {
	err := r.db.WithContext(ctx).Save(box).Error
	if err != nil {
		return nil, err
	}
	return box, nil
}

// DeleteWashBox удаляет бокс мойки
func (r *PostgresRepository) DeleteWashBox(ctx context.Context, id uuid.UUID) error {
	return r.db.WithContext(ctx).Delete(&models.WashBox{}, id).Error
}

// GetWashBoxesWithFilters получает боксы мойки с фильтрацией
func (r *PostgresRepository) GetWashBoxesWithFilters(ctx context.Context, status *string, serviceType *string, limit int, offset int) ([]models.WashBox, int, error) {
	var boxes []models.WashBox
	var total int64

	// Получаем общее количество
	err := func() error {
		query := r.db.WithContext(ctx).Model(&models.WashBox{})

		// Применяем фильтры
		if status != nil {
			query = query.Where("status = ?", *status)
		}
		if serviceType != nil {
			query = query.Where("service_type = ?", *serviceType)
		}

		return query.Count(&total).Error
	}()
	if err != nil {
		return nil, 0, err
	}

	// Получаем данные с пагинацией и сортировкой по номеру
	err = func() error {
		query := r.db.WithContext(ctx).Model(&models.WashBox{})

		// Применяем фильтры
		if status != nil {
			query = query.Where("status = ?", *status)
		}
		if serviceType != nil {
			query = query.Where("service_type = ?", *serviceType)
		}

		return query.Order("number ASC").Limit(limit).Offset(offset).Find(&boxes).Error
	}()
	if err != nil {
		return nil, 0, err
	}

	return boxes, int(total), nil
}

// RestoreWashBox восстанавливает удаленный бокс мойки
func (r *PostgresRepository) RestoreWashBox(ctx context.Context, id uuid.UUID, status string, serviceType string) (*models.WashBox, error) {
	// Получаем удаленный бокс
	var box models.WashBox
	err := r.db.WithContext(ctx).Unscoped().First(&box, id).Error
	if err != nil {
		return nil, err
	}

	// Обновляем бокс: убираем deleted_at и обновляем статус и тип услуги
	box.Status = status
	box.ServiceType = serviceType
	box.DeletedAt = gorm.DeletedAt{}

	// Сохраняем обновленный бокс
	updatedBox, err := r.UpdateWashBox(ctx, &box)
	if err != nil {
		return nil, err
	}

	return updatedBox, nil
}

// GetWashBoxesForCleaner получает список боксов для уборщика
func (r *PostgresRepository) GetWashBoxesForCleaner(ctx context.Context, limit int, offset int) ([]models.WashBox, int, error) {
	var boxes []models.WashBox
	var total int64

	// Подсчитываем общее количество
	err := r.db.WithContext(ctx).Model(&models.WashBox{}).Count(&total).Error
	if err != nil {
		return nil, 0, err
	}

	// Получаем боксы с пагинацией
	query := r.db.WithContext(ctx).Order("number ASC")
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
func (r *PostgresRepository) ReserveCleaning(ctx context.Context, washBoxID uuid.UUID, cleanerID uuid.UUID) error {
	return r.db.WithContext(ctx).Model(&models.WashBox{}).
		Where("id = ?", washBoxID).
		Update("cleaning_reserved_by", cleanerID).Error
}

// StartCleaning начинает уборку бокса
func (r *PostgresRepository) StartCleaning(ctx context.Context, washBoxID uuid.UUID) error {
	now := time.Now()
	return r.db.WithContext(ctx).Model(&models.WashBox{}).
		Where("id = ?", washBoxID).
		Updates(map[string]interface{}{
			"status":               models.StatusCleaning,
			"cleaning_started_at":  now,
			"cleaning_reserved_by": nil,
		}).Error
}

// CancelCleaning отменяет резервирование уборки
func (r *PostgresRepository) CancelCleaning(ctx context.Context, washBoxID uuid.UUID) error {
	return r.db.WithContext(ctx).Model(&models.WashBox{}).
		Where("id = ?", washBoxID).
		Update("cleaning_reserved_by", nil).Error
}

// CompleteCleaning завершает уборку бокса
func (r *PostgresRepository) CompleteCleaning(ctx context.Context, washBoxID uuid.UUID) error {
	return r.db.WithContext(ctx).Model(&models.WashBox{}).
		Where("id = ?", washBoxID).
		Updates(map[string]interface{}{
			"status":               models.StatusFree,
			"cleaning_started_at":  nil,
			"cleaning_reserved_by": nil,
		}).Error
}

// GetCleaningBoxes получает все боксы в статусе уборки
func (r *PostgresRepository) GetCleaningBoxes(ctx context.Context) ([]models.WashBox, error) {
	var boxes []models.WashBox
	err := r.db.WithContext(ctx).Where("status = ?", models.StatusCleaning).Find(&boxes).Error
	return boxes, err
}

// UpdateCleaningStartedAt обновляет время начала уборки
func (r *PostgresRepository) UpdateCleaningStartedAt(ctx context.Context, washBoxID uuid.UUID, startedAt time.Time) error {
	return r.db.WithContext(ctx).Model(&models.WashBox{}).
		Where("id = ?", washBoxID).
		Update("cleaning_started_at", startedAt).Error
}

// CreateCleaningLog создает новый лог уборки
func (r *PostgresRepository) CreateCleaningLog(ctx context.Context, log *models.CleaningLog) error {
	return r.db.WithContext(ctx).Create(log).Error
}

// UpdateCleaningLog обновляет лог уборки
func (r *PostgresRepository) UpdateCleaningLog(ctx context.Context, log *models.CleaningLog) error {
	return r.db.WithContext(ctx).Save(log).Error
}

// GetCleaningLogs получает логи уборки с фильтрами
func (r *PostgresRepository) GetCleaningLogs(ctx context.Context, req *models.AdminListCleaningLogsInternalRequest) ([]models.CleaningLogWithDetails, error) {
	var logs []models.CleaningLogWithDetails

	query := r.db.WithContext(ctx).Table("cleaning_logs cl").
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
func (r *PostgresRepository) GetCleaningLogsCount(ctx context.Context, req *models.AdminListCleaningLogsInternalRequest) (int64, error) {
	var count int64

	query := r.db.WithContext(ctx).Model(&models.CleaningLog{})

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
func (r *PostgresRepository) GetActiveCleaningLogByCleaner(ctx context.Context, cleanerID uuid.UUID) (*models.CleaningLog, error) {
	var log models.CleaningLog
	err := r.db.WithContext(ctx).Where("cleaner_id = ? AND status = ?", cleanerID, models.CleaningLogStatusInProgress).
		First(&log).Error
	if err != nil {
		return nil, err
	}
	return &log, nil
}

// GetLastCleaningLogByBox получает последний лог уборки для бокса
func (r *PostgresRepository) GetLastCleaningLogByBox(ctx context.Context, washBoxID uuid.UUID) (*models.CleaningLog, error) {
	var log models.CleaningLog
	err := r.db.WithContext(ctx).Where("wash_box_id = ?", washBoxID).
		Order("started_at DESC").
		First(&log).Error
	if err != nil {
		return nil, err
	}
	return &log, nil
}

// GetExpiredCleaningLogs получает логи уборки, которые нужно автоматически завершить
func (r *PostgresRepository) GetExpiredCleaningLogs(ctx context.Context, timeoutMinutes int) ([]models.CleaningLog, error) {
	var logs []models.CleaningLog
	timeout := time.Now().Add(-time.Duration(timeoutMinutes) * time.Minute)
	err := r.db.WithContext(ctx).Where("status = ? AND started_at <= ?",
		models.CleaningLogStatusInProgress, timeout).
		Find(&logs).Error
	return logs, err
}

// SetCooldown устанавливает cooldown для бокса после завершения сессии
func (r *PostgresRepository) SetCooldown(ctx context.Context, boxID uuid.UUID, userID uuid.UUID, cooldownUntil time.Time) error {
	return r.db.WithContext(ctx).Model(&models.WashBox{}).
		Where("id = ?", boxID).
		Updates(map[string]interface{}{
			"last_completed_session_user_id": userID,
			"last_completed_at":              time.Now(),
			"cooldown_until":                 cooldownUntil,
		}).Error
}

// GetCooldownBoxesForUser получает боксы в cooldown для конкретного пользователя
func (r *PostgresRepository) GetCooldownBoxesForUser(ctx context.Context, userID uuid.UUID) ([]models.WashBox, error) {
	var boxes []models.WashBox
	now := time.Now()

	err := r.db.WithContext(ctx).Where("last_completed_session_user_id = ? AND cooldown_until > ?",
		userID, now).
		Find(&boxes).Error
	return boxes, err
}

// SetCooldownByCarNumber устанавливает cooldown для бокса по госномеру после завершения сессии
func (r *PostgresRepository) SetCooldownByCarNumber(ctx context.Context, boxID uuid.UUID, carNumber string, cooldownUntil time.Time) error {
	return r.db.WithContext(ctx).Model(&models.WashBox{}).
		Where("id = ?", boxID).
		Updates(map[string]interface{}{
			"last_completed_session_car_number": carNumber,
			"last_completed_at":                 time.Now(),
			"cooldown_until":                    cooldownUntil,
		}).Error
}

// GetCooldownBoxesByCarNumber получает боксы в cooldown для конкретного госномера
func (r *PostgresRepository) GetCooldownBoxesByCarNumber(ctx context.Context, carNumber string) ([]models.WashBox, error) {
	var boxes []models.WashBox
	now := time.Now()

	err := r.db.WithContext(ctx).Where("last_completed_session_car_number = ? AND cooldown_until > ?",
		carNumber, now).
		Find(&boxes).Error
	return boxes, err
}

// ClearCooldown очищает поля кулдауна у бокса
func (r *PostgresRepository) ClearCooldown(ctx context.Context, boxID uuid.UUID) error {
	return r.db.WithContext(ctx).Model(&models.WashBox{}).
		Where("id = ?", boxID).
		Updates(map[string]interface{}{
			"last_completed_session_user_id":    nil,
			"last_completed_session_car_number": nil,
			"last_completed_at":                 nil,
			"cooldown_until":                    nil,
		}).Error
}

// CheckCooldownExpired очищает истекшие cooldown'ы
func (r *PostgresRepository) CheckCooldownExpired(ctx context.Context) error {
	now := time.Now()
	return r.db.WithContext(ctx).Model(&models.WashBox{}).
		Where("cooldown_until IS NOT NULL AND cooldown_until <= ?", now).
		Updates(map[string]interface{}{
			"last_completed_session_user_id":    nil,
			"last_completed_session_car_number": nil,
			"last_completed_at":                 nil,
			"cooldown_until":                    nil,
			"status":                            models.StatusFree,
		}).Error
}
