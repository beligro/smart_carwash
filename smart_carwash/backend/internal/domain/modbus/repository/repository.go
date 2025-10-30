package repository

import (
	"context"
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"

	"carwash_backend/internal/domain/modbus/models"
	washboxModels "carwash_backend/internal/domain/washbox/models"
)

// ModbusRepository предоставляет методы для работы с данными Modbus в БД
type ModbusRepository struct {
	db *gorm.DB
}

// NewModbusRepository создает новый экземпляр ModbusRepository
func NewModbusRepository(db *gorm.DB) *ModbusRepository {
	return &ModbusRepository{
		db: db,
	}
}

// GetWashBoxByID получает бокс по ID
func (r *ModbusRepository) GetWashBoxByID(ctx context.Context, boxID uuid.UUID) (*washboxModels.WashBox, error) {
	var washbox washboxModels.WashBox
	err := r.db.WithContext(ctx).Where("id = ?", boxID).First(&washbox).Error
	if err != nil {
		return nil, err
	}
	return &washbox, nil
}

// SaveModbusOperation сохраняет операцию Modbus в БД
func (r *ModbusRepository) SaveModbusOperation(ctx context.Context, operation *models.ModbusOperation) error {
	return r.db.WithContext(ctx).Create(operation).Error
}

// GetModbusOperations получает историю операций Modbus
func (r *ModbusRepository) GetModbusOperations(ctx context.Context, boxID *uuid.UUID, limit, offset int) ([]models.ModbusOperation, error) {
	var operations []models.ModbusOperation
	query := r.db.WithContext(ctx).Order("created_at DESC")

	if boxID != nil {
		query = query.Where("box_id = ?", *boxID)
	}

	if limit > 0 {
		query = query.Limit(limit)
	}

	if offset > 0 {
		query = query.Offset(offset)
	}

	err := query.Find(&operations).Error
	return operations, err
}

// GetModbusOperationsCount получает общее количество операций Modbus
func (r *ModbusRepository) GetModbusOperationsCount(ctx context.Context, boxID *uuid.UUID) (int64, error) {
	var count int64
	query := r.db.WithContext(ctx).Model(&models.ModbusOperation{})

	if boxID != nil {
		query = query.Where("box_id = ?", *boxID)
	}

	err := query.Count(&count).Error
	return count, err
}

// GetModbusStats получает статистику по операциям Modbus
func (r *ModbusRepository) GetModbusStats(ctx context.Context, boxID *uuid.UUID, since time.Time) (*models.ModbusStats, error) {
	var stats models.ModbusStats

	query := r.db.WithContext(ctx).Model(&models.ModbusOperation{}).Where("created_at >= ?", since)
	if boxID != nil {
		query = query.Where("box_id = ?", *boxID)
	}

	// Общее количество операций
	err := query.Count(&stats.TotalOperations).Error
	if err != nil {
		return nil, err
	}

	// Успешные операции
	err = query.Where("success = ?", true).Count(&stats.SuccessfulOperations).Error
	if err != nil {
		return nil, err
	}

	// Операции с ошибками
	err = query.Where("success = ?", false).Count(&stats.FailedOperations).Error
	if err != nil {
		return nil, err
	}

	// Последняя операция
	var lastOp models.ModbusOperation
	err = query.Order("created_at DESC").First(&lastOp).Error
	if err != nil && err != gorm.ErrRecordNotFound {
		return nil, err
	}
	if err != gorm.ErrRecordNotFound {
		stats.LastOperation = &lastOp
	}

	// Последняя успешная операция
	var lastSuccessOp models.ModbusOperation
	err = query.Where("success = ?", true).Order("created_at DESC").First(&lastSuccessOp).Error
	if err != nil && err != gorm.ErrRecordNotFound {
		return nil, err
	}
	if err != gorm.ErrRecordNotFound {
		stats.LastSuccessfulOperation = &lastSuccessOp
	}

	return &stats, nil
}

// GetModbusErrorsByType получает ошибки по типам
func (r *ModbusRepository) GetModbusErrorsByType(ctx context.Context, boxID *uuid.UUID, since time.Time) (map[string]int64, error) {
	var results []struct {
		Error string
		Count int64
	}

	query := r.db.WithContext(ctx).Model(&models.ModbusOperation{}).
		Select("error, COUNT(*) as count").
		Where("success = ? AND created_at >= ? AND error != ''", false, since).
		Group("error")

	if boxID != nil {
		query = query.Where("box_id = ?", *boxID)
	}

	err := query.Find(&results).Error
	if err != nil {
		return nil, err
	}

	errorStats := make(map[string]int64)
	for _, result := range results {
		errorStats[result.Error] = result.Count
	}

	return errorStats, nil
}

// GetActiveBoxesWithModbusConfig получает боксы с настроенным modbus
func (r *ModbusRepository) GetActiveBoxesWithModbusConfig(ctx context.Context) ([]washboxModels.WashBox, error) {
	var boxes []washboxModels.WashBox
	err := r.db.WithContext(ctx).Where("(light_coil_register IS NOT NULL AND light_coil_register != '') OR (chemistry_coil_register IS NOT NULL AND chemistry_coil_register != '')").Find(&boxes).Error
	return boxes, err
}

// UpdateModbusConnectionStatus обновляет статус подключения для бокса
func (r *ModbusRepository) UpdateModbusConnectionStatus(ctx context.Context, boxID uuid.UUID, connected bool, lastError string) error {
	// Создаем или обновляем запись статуса
	status := models.ModbusConnectionStatus{
		BoxID:     boxID,
		Connected: connected,
		LastError: lastError,
		LastSeen:  time.Now(),
	}

	err := r.db.WithContext(ctx).Model(&models.ModbusConnectionStatus{}).
		Where("box_id = ?", boxID).
		Assign(status).
		FirstOrCreate(&status).Error
	return err
}

// GetModbusConnectionStatus получает статус подключения для бокса
func (r *ModbusRepository) GetModbusConnectionStatus(ctx context.Context, boxID uuid.UUID) (*models.ModbusConnectionStatus, error) {
	var status models.ModbusConnectionStatus
	err := r.db.WithContext(ctx).Where("box_id = ?", boxID).First(&status).Error
	if err != nil {
		return nil, err
	}
	return &status, nil
}

// GetAllModbusConnectionStatuses получает статусы подключения для всех боксов
func (r *ModbusRepository) GetAllModbusConnectionStatuses(ctx context.Context) ([]models.ModbusConnectionStatus, error) {
	var statuses []models.ModbusConnectionStatus
	err := r.db.WithContext(ctx).Find(&statuses).Error
	return statuses, err
}

// UpdateModbusCoilStatus обновляет статус конкретного койла (света или химии) для бокса
func (r *ModbusRepository) UpdateModbusCoilStatus(ctx context.Context, boxID uuid.UUID, coilType string, value bool) error {
	// Получаем или создаем запись статуса
	var status models.ModbusConnectionStatus
	err := r.db.WithContext(ctx).Where("box_id = ?", boxID).First(&status).Error

	// Если запись не найдена, создаем новую
	if err == gorm.ErrRecordNotFound {
		status = models.ModbusConnectionStatus{
			BoxID:    boxID,
			LastSeen: time.Now(),
		}

		// Устанавливаем соответствующий статус
		if coilType == "light" {
			status.LightStatus = &value
		} else if coilType == "chemistry" {
			status.ChemistryStatus = &value
		}

		return r.db.WithContext(ctx).Create(&status).Error
	}

	if err != nil {
		return err
	}

	// Обновляем соответствующий статус
	updates := map[string]interface{}{
		"last_seen": time.Now(),
	}

	if coilType == "light" {
		updates["light_status"] = value
	} else if coilType == "chemistry" {
		updates["chemistry_status"] = value
	}

	err = r.db.WithContext(ctx).Model(&models.ModbusConnectionStatus{}).
		Where("box_id = ?", boxID).
		Updates(updates).Error

	return err
}
