package repository

import (
	"context"
	"time"

	"carwash_backend/internal/domain/linktoken/models"

	"gorm.io/gorm"
)

// Repository интерфейс для link_tokens
type Repository interface {
	Create(ctx context.Context, t *models.LinkToken) error
	GetByToken(ctx context.Context, token string) (*models.LinkToken, error)
	MarkUsed(ctx context.Context, token string, usedAt time.Time) error
}

// PostgresRepository реализация для PostgreSQL
type PostgresRepository struct {
	db *gorm.DB
}

// NewPostgresRepository создает репозиторий
func NewPostgresRepository(db *gorm.DB) *PostgresRepository {
	return &PostgresRepository{db: db}
}

// Create создает запись токена
func (r *PostgresRepository) Create(ctx context.Context, t *models.LinkToken) error {
	return r.db.WithContext(ctx).Create(t).Error
}

// GetByToken возвращает токен по строке
func (r *PostgresRepository) GetByToken(ctx context.Context, token string) (*models.LinkToken, error) {
	var t models.LinkToken
	err := r.db.WithContext(ctx).Where("token = ?", token).First(&t).Error
	if err != nil {
		return nil, err
	}
	return &t, nil
}

// MarkUsed помечает токен использованным
func (r *PostgresRepository) MarkUsed(ctx context.Context, token string, usedAt time.Time) error {
	return r.db.WithContext(ctx).Model(&models.LinkToken{}).
		Where("token = ?", token).
		Update("used_at", usedAt).Error
}
