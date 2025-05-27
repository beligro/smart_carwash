package repository

import (
	"carwash_backend/internal/domain/washbox/models"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// Repository интерфейс для работы с боксами мойки в базе данных
type Repository interface {
	GetAllWashBoxes() ([]models.WashBox, error)
	GetWashBoxByID(id uuid.UUID) (*models.WashBox, error)
	UpdateWashBoxStatus(id uuid.UUID, status string) error
	CreateWashBox(box *models.WashBox) error
	GetFreeWashBoxes() ([]models.WashBox, error)
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
func (r *PostgresRepository) CreateWashBox(box *models.WashBox) error {
	return r.db.Create(box).Error
}

// GetFreeWashBoxes получает все свободные боксы мойки
func (r *PostgresRepository) GetFreeWashBoxes() ([]models.WashBox, error) {
	var boxes []models.WashBox
	err := r.db.Where("status = ?", models.StatusFree).Find(&boxes).Error
	return boxes, err
}
