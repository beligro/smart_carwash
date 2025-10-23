package repository

import (
	"carwash_backend/internal/domain/user/models"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// Repository интерфейс для работы с пользователями в базе данных
type Repository interface {
	CreateUser(user *models.User) error
	GetUserByTelegramID(telegramID int64) (*models.User, error)
	GetUserByID(id uuid.UUID) (*models.User, error)
	GetUserByCarNumber(carNumber string) (*models.User, error)
	UpdateUser(user *models.User) error

	// Административные методы
	GetUsersWithPagination(limit int, offset int) ([]models.User, int, error)
}

// PostgresRepository реализация Repository для PostgreSQL
type PostgresRepository struct {
	db *gorm.DB
}

// NewPostgresRepository создает новый экземпляр PostgresRepository
func NewPostgresRepository(db *gorm.DB) *PostgresRepository {
	return &PostgresRepository{db: db}
}

// CreateUser создает нового пользователя
func (r *PostgresRepository) CreateUser(user *models.User) error {
	return r.db.Create(user).Error
}

// GetUserByTelegramID получает пользователя по Telegram ID
func (r *PostgresRepository) GetUserByTelegramID(telegramID int64) (*models.User, error) {
	var user models.User
	err := r.db.Where("telegram_id = ?", telegramID).First(&user).Error
	if err != nil {
		return nil, err
	}
	return &user, nil
}

// GetUserByID получает пользователя по ID
func (r *PostgresRepository) GetUserByID(id uuid.UUID) (*models.User, error) {
	var user models.User
	err := r.db.Where("id = ?", id).First(&user).Error
	if err != nil {
		return nil, err
	}
	return &user, nil
}

// GetUserByCarNumber получает пользователя по номеру автомобиля
func (r *PostgresRepository) GetUserByCarNumber(carNumber string) (*models.User, error) {
	var user models.User
	err := r.db.Where("car_number = ?", carNumber).First(&user).Error
	if err != nil {
		return nil, err
	}
	return &user, nil
}

// UpdateUser обновляет пользователя
func (r *PostgresRepository) UpdateUser(user *models.User) error {
	return r.db.Save(user).Error
}

// GetUsersWithPagination получает пользователей с пагинацией
func (r *PostgresRepository) GetUsersWithPagination(limit int, offset int) ([]models.User, int, error) {
	var users []models.User
	var total int64

	// Получаем общее количество пользователей
	err := r.db.Model(&models.User{}).Count(&total).Error
	if err != nil {
		return nil, 0, err
	}

	// Получаем пользователей с пагинацией и сортировкой
	err = r.db.Order("created_at DESC").Limit(limit).Offset(offset).Find(&users).Error
	if err != nil {
		return nil, 0, err
	}

	return users, int(total), nil
}
