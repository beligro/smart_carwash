package repository

import (
	"carwash_backend/internal/domain/payment/models"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// Repository интерфейс для работы с платежами
type Repository interface {
	CreatePayment(payment *models.Payment) error
	GetPaymentByID(id uuid.UUID) (*models.Payment, error)
	GetPaymentBySessionID(sessionID uuid.UUID) (*models.Payment, error)
	GetPaymentByTinkoffID(tinkoffID string) (*models.Payment, error)
	UpdatePayment(payment *models.Payment) error
	ListPayments(req *models.AdminListPaymentsRequest) ([]models.Payment, int, error)
}

// repository реализация Repository
type repository struct {
	db *gorm.DB
}

// NewRepository создает новый экземпляр Repository
func NewRepository(db *gorm.DB) Repository {
	return &repository{db: db}
}

// CreatePayment создает новый платеж
func (r *repository) CreatePayment(payment *models.Payment) error {
	return r.db.Create(payment).Error
}

// GetPaymentByID получает платеж по ID
func (r *repository) GetPaymentByID(id uuid.UUID) (*models.Payment, error) {
	var payment models.Payment
	err := r.db.Where("id = ?", id).First(&payment).Error
	if err != nil {
		return nil, err
	}
	return &payment, nil
}

// GetPaymentBySessionID получает платеж по ID сессии
func (r *repository) GetPaymentBySessionID(sessionID uuid.UUID) (*models.Payment, error) {
	var payment models.Payment
	err := r.db.Where("session_id = ?", sessionID).First(&payment).Error
	if err != nil {
		return nil, err
	}
	return &payment, nil
}

// GetPaymentByTinkoffID получает платеж по ID в Tinkoff
func (r *repository) GetPaymentByTinkoffID(tinkoffID string) (*models.Payment, error) {
	var payment models.Payment
	err := r.db.Where("tinkoff_id = ?", tinkoffID).First(&payment).Error
	if err != nil {
		return nil, err
	}
	return &payment, nil
}

// UpdatePayment обновляет платеж
func (r *repository) UpdatePayment(payment *models.Payment) error {
	return r.db.Save(payment).Error
}

// ListPayments получает список платежей с фильтрацией
func (r *repository) ListPayments(req *models.AdminListPaymentsRequest) ([]models.Payment, int, error) {
	var payments []models.Payment
	var total int64

	query := r.db.Model(&models.Payment{})

	// Применяем фильтры
	if req.SessionID != nil {
		query = query.Where("session_id = ?", *req.SessionID)
	}
	if req.Status != nil {
		query = query.Where("status = ?", *req.Status)
	}
	if req.DateFrom != nil {
		query = query.Where("created_at >= ?", *req.DateFrom)
	}
	if req.DateTo != nil {
		query = query.Where("created_at <= ?", *req.DateTo)
	}

	// Подсчитываем общее количество
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	// Применяем пагинацию
	limit := 50
	if req.Limit != nil {
		limit = *req.Limit
	}
	offset := 0
	if req.Offset != nil {
		offset = *req.Offset
	}

	query = query.Offset(offset).Limit(limit).Order("created_at DESC")

	if err := query.Find(&payments).Error; err != nil {
		return nil, 0, err
	}

	return payments, int(total), nil
} 