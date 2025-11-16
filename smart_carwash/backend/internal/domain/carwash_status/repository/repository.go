package repository

import (
	"carwash_backend/internal/domain/carwash_status/models"
	"context"
	"errors"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// Repository интерфейс для работы со статусом мойки в базе данных
type Repository interface {
	GetCurrentStatus(ctx context.Context) (*models.CarwashStatus, error)
	UpdateStatus(ctx context.Context, isClosed bool, reason *string, updatedBy *uuid.UUID) error
	CreateHistoryRecord(ctx context.Context, isClosed bool, reason *string, createdBy *uuid.UUID) error
	GetHistory(ctx context.Context, limit, offset int) ([]models.CarwashStatusHistory, int, error)
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

// GetCurrentStatus получает текущий статус мойки
func (r *PostgresRepository) GetCurrentStatus(ctx context.Context) (*models.CarwashStatus, error) {
	var status models.CarwashStatus

	// Получаем первую (и единственную) запись
	err := r.db.WithContext(ctx).First(&status).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			// Если записи нет, создаем начальную запись со статусом "открыта"
			status = models.CarwashStatus{
				IsClosed:  false,
				UpdatedAt: r.db.NowFunc(),
				CreatedAt: r.db.NowFunc(),
			}
			if err := r.db.WithContext(ctx).Create(&status).Error; err != nil {
				return nil, err
			}
			return &status, nil
		}
		return nil, err
	}

	return &status, nil
}

// UpdateStatus обновляет текущий статус мойки
func (r *PostgresRepository) UpdateStatus(ctx context.Context, isClosed bool, reason *string, updatedBy *uuid.UUID) error {
	now := r.db.NowFunc()

	// Обновляем статус (так как запись всегда одна, используем FirstOrCreate)
	var status models.CarwashStatus
	err := r.db.WithContext(ctx).First(&status).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			// Создаем новую запись
			status = models.CarwashStatus{
				IsClosed:     isClosed,
				ClosedReason: reason,
				UpdatedAt:    now,
				UpdatedBy:    updatedBy,
				CreatedAt:    now,
			}
			return r.db.WithContext(ctx).Create(&status).Error
		}
		return err
	}

	// Обновляем существующую запись
	updates := map[string]interface{}{
		"is_closed":     isClosed,
		"closed_reason": reason,
		"updated_at":    now,
		"updated_by":    updatedBy,
	}

	return r.db.WithContext(ctx).Model(&status).Updates(updates).Error
}

// CreateHistoryRecord создает запись в истории изменений статуса
func (r *PostgresRepository) CreateHistoryRecord(ctx context.Context, isClosed bool, reason *string, createdBy *uuid.UUID) error {
	history := models.CarwashStatusHistory{
		IsClosed:     isClosed,
		ClosedReason: reason,
		CreatedAt:    r.db.NowFunc(),
		CreatedBy:    createdBy,
	}

	return r.db.WithContext(ctx).Create(&history).Error
}

// GetHistory получает историю изменений статуса мойки
func (r *PostgresRepository) GetHistory(ctx context.Context, limit, offset int) ([]models.CarwashStatusHistory, int, error) {
	var history []models.CarwashStatusHistory
	var total int64

	// Подсчитываем общее количество записей
	if err := r.db.WithContext(ctx).Model(&models.CarwashStatusHistory{}).Count(&total).Error; err != nil {
		return nil, 0, err
	}

	// Применяем лимиты по умолчанию
	if limit <= 0 {
		limit = 50
	}
	if offset < 0 {
		offset = 0
	}

	// Получаем записи с пагинацией
	query := r.db.WithContext(ctx).
		Order("created_at DESC").
		Limit(limit).
		Offset(offset)

	if err := query.Find(&history).Error; err != nil {
		return nil, 0, err
	}

	return history, int(total), nil
}
