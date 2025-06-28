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
	CreateWashBox(box *models.WashBox) (*models.WashBox, error)
	GetFreeWashBoxes() ([]models.WashBox, error)
	GetFreeWashBoxesByServiceType(serviceType string) ([]models.WashBox, error)
	GetWashBoxesByServiceType(serviceType string) ([]models.WashBox, error)

	// Административные методы
	GetWashBoxByNumber(number int) (*models.WashBox, error)
	UpdateWashBox(box *models.WashBox) (*models.WashBox, error)
	DeleteWashBox(id uuid.UUID) error
	GetWashBoxesWithFilters(status *string, serviceType *string, limit int, offset int) ([]models.WashBox, int, error)
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

// GetFreeWashBoxes получает все свободные боксы мойки
func (r *PostgresRepository) GetFreeWashBoxes() ([]models.WashBox, error) {
	var boxes []models.WashBox
	err := r.db.Where("status = ?", models.StatusFree).Find(&boxes).Error
	return boxes, err
}

// GetFreeWashBoxesByServiceType получает все свободные боксы мойки определенного типа
func (r *PostgresRepository) GetFreeWashBoxesByServiceType(serviceType string) ([]models.WashBox, error) {
	var boxes []models.WashBox
	err := r.db.Where("status = ? AND service_type = ?", models.StatusFree, serviceType).Find(&boxes).Error
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

	// Получаем данные с пагинацией
	err = query.Limit(limit).Offset(offset).Find(&boxes).Error
	if err != nil {
		return nil, 0, err
	}

	return boxes, int(total), nil
}
