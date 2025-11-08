package repository

import (
	"carwash_backend/internal/domain/user/models"
	"carwash_backend/internal/utils"
	"context"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// Repository интерфейс для работы с пользователями в базе данных
type Repository interface {
	CreateUser(ctx context.Context, user *models.User) error
	GetUserByTelegramID(ctx context.Context, telegramID int64) (*models.User, error)
	GetUserByID(ctx context.Context, id uuid.UUID) (*models.User, error)
	GetUsersByIDs(ctx context.Context, ids []uuid.UUID) (map[uuid.UUID]*models.User, error)
	GetUserByCarNumber(ctx context.Context, carNumber string) (*models.User, error)
	UpdateUser(ctx context.Context, user *models.User) error

	// Административные методы
	GetUsersWithPagination(ctx context.Context, limit int, offset int) ([]models.User, int, error)
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

// CreateUser создает нового пользователя
func (r *PostgresRepository) CreateUser(ctx context.Context, user *models.User) error {
	return r.db.WithContext(ctx).Create(user).Error
}

// GetUserByTelegramID получает пользователя по Telegram ID
func (r *PostgresRepository) GetUserByTelegramID(ctx context.Context, telegramID int64) (*models.User, error) {
	var user models.User
	err := r.db.WithContext(ctx).Where("telegram_id = ?", telegramID).First(&user).Error
	if err != nil {
		return nil, err
	}
	return &user, nil
}

// GetUserByID получает пользователя по ID
func (r *PostgresRepository) GetUserByID(ctx context.Context, id uuid.UUID) (*models.User, error) {
	var user models.User
	err := r.db.WithContext(ctx).Where("id = ?", id).First(&user).Error
	if err != nil {
		return nil, err
	}
	return &user, nil
}

// GetUsersByIDs получает пользователей по списку ID
func (r *PostgresRepository) GetUsersByIDs(ctx context.Context, ids []uuid.UUID) (map[uuid.UUID]*models.User, error) {
	if len(ids) == 0 {
		return make(map[uuid.UUID]*models.User), nil
	}

	var users []models.User
	err := r.db.WithContext(ctx).Where("id IN ?", ids).Find(&users).Error
	if err != nil {
		return nil, err
	}

	result := make(map[uuid.UUID]*models.User, len(users))
	for i := range users {
		result[users[i].ID] = &users[i]
	}

	return result, nil
}

// GetUserByCarNumber получает пользователя по номеру автомобиля
func (r *PostgresRepository) GetUserByCarNumber(ctx context.Context, carNumber string) (*models.User, error) {
	// Нормализуем номер для поиска
	normalizedCarNumber := utils.NormalizeLicensePlateForSearch(carNumber)

	var user models.User
	err := r.db.WithContext(ctx).Where("car_number = ?", normalizedCarNumber).First(&user).Error
	if err != nil {
		return nil, err
	}
	return &user, nil
}

// UpdateUser обновляет пользователя
func (r *PostgresRepository) UpdateUser(ctx context.Context, user *models.User) error {
	return r.db.WithContext(ctx).Save(user).Error
}

// GetUsersWithPagination получает пользователей с пагинацией
func (r *PostgresRepository) GetUsersWithPagination(ctx context.Context, limit int, offset int) ([]models.User, int, error) {
	var users []models.User
	var total int64

	// Получаем общее количество пользователей
	err := r.db.WithContext(ctx).Model(&models.User{}).Count(&total).Error
	if err != nil {
		return nil, 0, err
	}

	// Получаем пользователей с пагинацией и сортировкой
	err = r.db.WithContext(ctx).Order("created_at DESC").Limit(limit).Offset(offset).Find(&users).Error
	if err != nil {
		return nil, 0, err
	}

	return users, int(total), nil
}
