package repository

import (
	"carwash_backend/internal/domain/washboxlog/models"
	"context"
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type Repository interface {
	Insert(ctx context.Context, log *models.WashBoxChangeLog) error
	List(ctx context.Context, boxNumber *int, actorType *string, action *string, since *time.Time, until *time.Time, limit int, offset int) ([]models.WashBoxChangeLog, int64, error)
}

type PostgresRepository struct {
	db *gorm.DB
}

func NewPostgresRepository(db *gorm.DB) *PostgresRepository {
	return &PostgresRepository{db: db}
}

func (r *PostgresRepository) Insert(ctx context.Context, logItem *models.WashBoxChangeLog) error {
	if logItem.ID == uuid.Nil {
		logItem.ID = uuid.New()
	}
	logItem.CreatedAt = time.Now()
	return r.db.WithContext(ctx).Create(logItem).Error
}

func (r *PostgresRepository) List(ctx context.Context, boxNumber *int, actorType *string, action *string, since *time.Time, until *time.Time, limit int, offset int) ([]models.WashBoxChangeLog, int64, error) {
	var rows []models.WashBoxChangeLog
	q := r.db.WithContext(ctx).Model(&models.WashBoxChangeLog{})

	if boxNumber != nil {
		q = q.Where("box_number = ?", *boxNumber)
	}
	if actorType != nil && *actorType != "" {
		q = q.Where("actor_type = ?", *actorType)
	}
	if action != nil && *action != "" {
		q = q.Where("action = ?", *action)
	}
	if since != nil {
		q = q.Where("created_at >= ?", *since)
	}
	if until != nil {
		q = q.Where("created_at <= ?", *until)
	}

	var total int64
	if err := q.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	if limit <= 0 || limit > 200 {
		limit = 50
	}
	if offset < 0 {
		offset = 0
	}

	if err := q.Order("created_at DESC").Limit(limit).Offset(offset).Find(&rows).Error; err != nil {
		return nil, 0, err
	}
	return rows, total, nil
}



