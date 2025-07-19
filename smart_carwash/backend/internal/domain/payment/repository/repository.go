package repository

import (
	"errors"
	"time"

	"carwash_backend/internal/domain/payment/models"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

var (
	// ErrPaymentNotFound возвращается, когда платеж не найден
	ErrPaymentNotFound = errors.New("платеж не найден")

	// ErrPaymentAlreadyExists возвращается, когда платеж с таким idempotency_key уже существует
	ErrPaymentAlreadyExists = errors.New("платеж с таким ключом идемпотентности уже существует")

	// ErrRefundNotFound возвращается, когда возврат не найден
	ErrRefundNotFound = errors.New("возврат не найден")

	// ErrRefundAlreadyExists возвращается, когда возврат с таким idempotency_key уже существует
	ErrRefundAlreadyExists = errors.New("возврат с таким ключом идемпотентности уже существует")

	// ErrPaymentEventAlreadyExists возвращается, когда событие с таким idempotency_key уже существует
	ErrPaymentEventAlreadyExists = errors.New("событие с таким ключом идемпотентности уже существует")
)

// Repository интерфейс для работы с платежами в базе данных
type Repository interface {
	// Методы для работы с платежами
	CreatePayment(payment *models.Payment) error
	GetPaymentByID(id uuid.UUID) (*models.Payment, error)
	GetPaymentByIdempotencyKey(key string) (*models.Payment, error)
	GetPaymentByTinkoffID(tinkoffID string) (*models.Payment, error)
	GetPaymentBySessionID(sessionID uuid.UUID) (*models.Payment, error)
	UpdatePayment(payment *models.Payment) error
	GetPaymentsByUserID(userID uuid.UUID, limit, offset int) ([]models.Payment, int, error)
	GetPaymentsByStatus(status string) ([]models.Payment, error)
	GetPaymentsByDateRange(startDate, endDate time.Time) ([]models.Payment, error)

	// Методы для работы с возвратами
	CreateRefund(refund *models.PaymentRefund) error
	GetRefundByID(id uuid.UUID) (*models.PaymentRefund, error)
	GetRefundByIdempotencyKey(key string) (*models.PaymentRefund, error)
	UpdateRefund(refund *models.PaymentRefund) error
	GetPendingRefunds() ([]models.PaymentRefund, error)
	GetRefundsByPaymentID(paymentID uuid.UUID) ([]models.PaymentRefund, error)

	// Методы для работы с событиями платежей
	CreatePaymentEvent(event *models.PaymentEvent) error
	GetPaymentEventByIdempotencyKey(key string) (*models.PaymentEvent, error)
	GetPaymentEventsByPaymentID(paymentID uuid.UUID) ([]models.PaymentEvent, error)

	// Административные методы
	AdminListPayments(userID *uuid.UUID, status *string, paymentType *string, dateFrom *time.Time, dateTo *time.Time, limit int, offset int) ([]models.Payment, int, error)
	AdminGetPaymentWithDetails(id uuid.UUID) (*models.Payment, []models.PaymentRefund, []models.PaymentEvent, error)
}

// PostgresRepository реализация Repository для PostgreSQL
type PostgresRepository struct {
	db *gorm.DB
}

// NewPostgresRepository создает новый экземпляр PostgresRepository
func NewPostgresRepository(db *gorm.DB) *PostgresRepository {
	return &PostgresRepository{db: db}
}

// CreatePayment создает новый платеж
func (r *PostgresRepository) CreatePayment(payment *models.Payment) error {
	// Проверяем идемпотентность
	if payment.IdempotencyKey != "" {
		existingPayment, err := r.GetPaymentByIdempotencyKey(payment.IdempotencyKey)
		if err == nil && existingPayment != nil {
			return ErrPaymentAlreadyExists
		}
	}

	return r.db.Create(payment).Error
}

// GetPaymentByID получает платеж по ID
func (r *PostgresRepository) GetPaymentByID(id uuid.UUID) (*models.Payment, error) {
	var payment models.Payment
	err := r.db.Where("id = ?", id).First(&payment).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrPaymentNotFound
		}
		return nil, err
	}
	return &payment, nil
}

// GetPaymentByIdempotencyKey получает платеж по ключу идемпотентности
func (r *PostgresRepository) GetPaymentByIdempotencyKey(key string) (*models.Payment, error) {
	var payment models.Payment
	err := r.db.Where("idempotency_key = ?", key).First(&payment).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil // Не считаем это ошибкой
		}
		return nil, err
	}
	return &payment, nil
}

// GetPaymentByTinkoffID получает платеж по Tinkoff Payment ID
func (r *PostgresRepository) GetPaymentByTinkoffID(tinkoffID string) (*models.Payment, error) {
	var payment models.Payment
	err := r.db.Where("tinkoff_payment_id = ?", tinkoffID).First(&payment).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrPaymentNotFound
		}
		return nil, err
	}
	return &payment, nil
}

// GetPaymentBySessionID получает платеж по ID сессии
func (r *PostgresRepository) GetPaymentBySessionID(sessionID uuid.UUID) (*models.Payment, error) {
	var payment models.Payment
	err := r.db.Where("session_id = ?", sessionID).First(&payment).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrPaymentNotFound
		}
		return nil, err
	}
	return &payment, nil
}

// UpdatePayment обновляет платеж
func (r *PostgresRepository) UpdatePayment(payment *models.Payment) error {
	return r.db.Save(payment).Error
}

// GetPaymentsByUserID получает платежи пользователя с пагинацией
func (r *PostgresRepository) GetPaymentsByUserID(userID uuid.UUID, limit, offset int) ([]models.Payment, int, error) {
	var payments []models.Payment
	var total int64

	// Получаем общее количество
	err := r.db.Model(&models.Payment{}).Where("user_id = ?", userID).Count(&total).Error
	if err != nil {
		return nil, 0, err
	}

	// Получаем платежи с пагинацией
	err = r.db.Where("user_id = ?", userID).
		Order("created_at DESC").
		Limit(limit).
		Offset(offset).
		Find(&payments).Error

	return payments, int(total), err
}

// GetPaymentsByStatus получает платежи по статусу
func (r *PostgresRepository) GetPaymentsByStatus(status string) ([]models.Payment, error) {
	var payments []models.Payment
	err := r.db.Where("status = ?", status).Find(&payments).Error
	return payments, err
}

// GetPaymentsByDateRange получает платежи за период
func (r *PostgresRepository) GetPaymentsByDateRange(startDate, endDate time.Time) ([]models.Payment, error) {
	var payments []models.Payment
	err := r.db.Where("created_at BETWEEN ? AND ?", startDate, endDate).Find(&payments).Error
	return payments, err
}

// CreateRefund создает новый возврат
func (r *PostgresRepository) CreateRefund(refund *models.PaymentRefund) error {
	// Проверяем идемпотентность
	if refund.IdempotencyKey != "" {
		existingRefund, err := r.GetRefundByIdempotencyKey(refund.IdempotencyKey)
		if err == nil && existingRefund != nil {
			return ErrRefundAlreadyExists
		}
	}

	return r.db.Create(refund).Error
}

// GetRefundByID получает возврат по ID
func (r *PostgresRepository) GetRefundByID(id uuid.UUID) (*models.PaymentRefund, error) {
	var refund models.PaymentRefund
	err := r.db.Where("id = ?", id).First(&refund).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrRefundNotFound
		}
		return nil, err
	}
	return &refund, nil
}

// GetRefundByIdempotencyKey получает возврат по ключу идемпотентности
func (r *PostgresRepository) GetRefundByIdempotencyKey(key string) (*models.PaymentRefund, error) {
	var refund models.PaymentRefund
	err := r.db.Where("idempotency_key = ?", key).First(&refund).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil // Не считаем это ошибкой
		}
		return nil, err
	}
	return &refund, nil
}

// UpdateRefund обновляет возврат
func (r *PostgresRepository) UpdateRefund(refund *models.PaymentRefund) error {
	return r.db.Save(refund).Error
}

// GetPendingRefunds получает ожидающие возвраты
func (r *PostgresRepository) GetPendingRefunds() ([]models.PaymentRefund, error) {
	var refunds []models.PaymentRefund
	err := r.db.Where("status = ? AND retry_count < max_retries", "pending").Find(&refunds).Error
	return refunds, err
}

// GetRefundsByPaymentID получает возвраты по ID платежа
func (r *PostgresRepository) GetRefundsByPaymentID(paymentID uuid.UUID) ([]models.PaymentRefund, error) {
	var refunds []models.PaymentRefund
	err := r.db.Where("payment_id = ?", paymentID).Find(&refunds).Error
	return refunds, err
}

// CreatePaymentEvent создает событие платежа
func (r *PostgresRepository) CreatePaymentEvent(event *models.PaymentEvent) error {
	// Проверяем идемпотентность
	if event.IdempotencyKey != "" {
		existingEvent, err := r.GetPaymentEventByIdempotencyKey(event.IdempotencyKey)
		if err == nil && existingEvent != nil {
			return ErrPaymentEventAlreadyExists
		}
	}

	return r.db.Create(event).Error
}

// GetPaymentEventByIdempotencyKey получает событие по ключу идемпотентности
func (r *PostgresRepository) GetPaymentEventByIdempotencyKey(key string) (*models.PaymentEvent, error) {
	var event models.PaymentEvent
	err := r.db.Where("idempotency_key = ?", key).First(&event).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil // Не считаем это ошибкой
		}
		return nil, err
	}
	return &event, nil
}

// GetPaymentEventsByPaymentID получает события по ID платежа
func (r *PostgresRepository) GetPaymentEventsByPaymentID(paymentID uuid.UUID) ([]models.PaymentEvent, error) {
	var events []models.PaymentEvent
	err := r.db.Where("payment_id = ?", paymentID).Order("created_at ASC").Find(&events).Error
	return events, err
}

// AdminListPayments получает список платежей для администратора
func (r *PostgresRepository) AdminListPayments(userID *uuid.UUID, status *string, paymentType *string, dateFrom *time.Time, dateTo *time.Time, limit int, offset int) ([]models.Payment, int, error) {
	var payments []models.Payment
	var total int64

	query := r.db.Model(&models.Payment{})

	// Применяем фильтры
	if userID != nil {
		query = query.Where("user_id = ?", *userID)
	}
	if status != nil {
		query = query.Where("status = ?", *status)
	}
	if paymentType != nil {
		query = query.Where("type = ?", *paymentType)
	}
	if dateFrom != nil {
		query = query.Where("created_at >= ?", *dateFrom)
	}
	if dateTo != nil {
		query = query.Where("created_at <= ?", *dateTo)
	}

	// Получаем общее количество
	err := query.Count(&total).Error
	if err != nil {
		return nil, 0, err
	}

	// Получаем платежи с пагинацией
	err = query.Order("created_at DESC").
		Limit(limit).
		Offset(offset).
		Find(&payments).Error

	return payments, int(total), err
}

// AdminGetPaymentWithDetails получает платеж с деталями для администратора
func (r *PostgresRepository) AdminGetPaymentWithDetails(id uuid.UUID) (*models.Payment, []models.PaymentRefund, []models.PaymentEvent, error) {
	// Получаем платеж
	payment, err := r.GetPaymentByID(id)
	if err != nil {
		return nil, nil, nil, err
	}

	// Получаем возвраты
	refunds, err := r.GetRefundsByPaymentID(id)
	if err != nil {
		return nil, nil, nil, err
	}

	// Получаем события
	events, err := r.GetPaymentEventsByPaymentID(id)
	if err != nil {
		return nil, nil, nil, err
	}

	return payment, refunds, events, nil
} 