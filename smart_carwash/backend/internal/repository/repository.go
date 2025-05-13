package repository

import (
	"carwash_backend/internal/models"

	"gorm.io/gorm"
)

// Repository интерфейс для работы с базой данных
type Repository interface {
	// Пользователи
	CreateUser(user *models.User) error
	GetUserByTelegramID(telegramID int64) (*models.User, error)
	UpdateUser(user *models.User) error

	// Боксы мойки
	GetAllWashBoxes() ([]models.WashBox, error)
	GetWashBoxByID(id uint) (*models.WashBox, error)
	UpdateWashBoxStatus(id uint, status string) error
	CreateWashBox(box *models.WashBox) error
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

// UpdateUser обновляет пользователя
func (r *PostgresRepository) UpdateUser(user *models.User) error {
	return r.db.Save(user).Error
}

// GetAllWashBoxes получает все боксы мойки
func (r *PostgresRepository) GetAllWashBoxes() ([]models.WashBox, error) {
	var boxes []models.WashBox
	err := r.db.Find(&boxes).Error
	return boxes, err
}

// GetWashBoxByID получает бокс мойки по ID
func (r *PostgresRepository) GetWashBoxByID(id uint) (*models.WashBox, error) {
	var box models.WashBox
	err := r.db.First(&box, id).Error
	if err != nil {
		return nil, err
	}
	return &box, nil
}

// UpdateWashBoxStatus обновляет статус бокса мойки
func (r *PostgresRepository) UpdateWashBoxStatus(id uint, status string) error {
	return r.db.Model(&models.WashBox{}).Where("id = ?", id).Update("status", status).Error
}

// CreateWashBox создает новый бокс мойки
func (r *PostgresRepository) CreateWashBox(box *models.WashBox) error {
	return r.db.Create(box).Error
}
