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
	GetPaymentsBySessionID(sessionID uuid.UUID) ([]models.Payment, error)
	GetPaymentByTinkoffID(tinkoffID string) (*models.Payment, error)
	UpdatePayment(payment *models.Payment) error
	ListPayments(req *models.AdminListPaymentsRequest) ([]models.Payment, int, error)
	GetPaymentStatistics(req *models.PaymentStatisticsRequest) (*models.PaymentStatisticsResponse, error)
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

// GetPaymentBySessionID получает основной платеж по ID сессии
func (r *repository) GetPaymentBySessionID(sessionID uuid.UUID) (*models.Payment, error) {
	var payment models.Payment
	err := r.db.Where("session_id = ? AND payment_type = ?", sessionID, models.PaymentTypeMain).First(&payment).Error
	if err != nil {
		return nil, err
	}
	return &payment, nil
}

// GetPaymentsBySessionID получает все платежи по ID сессии
func (r *repository) GetPaymentsBySessionID(sessionID uuid.UUID) ([]models.Payment, error) {
	var payments []models.Payment
	err := r.db.Where("session_id = ?", sessionID).Order("created_at ASC").Find(&payments).Error
	if err != nil {
		return nil, err
	}
	return payments, nil
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
	if req.PaymentID != nil {
		query = query.Where("id = ?", *req.PaymentID)
	}
	if req.SessionID != nil {
		query = query.Where("session_id = ?", *req.SessionID)
	}
	if req.UserID != nil {
		// Для фильтрации по пользователю нужно присоединить таблицу sessions
		query = query.Joins("JOIN sessions ON payments.session_id = sessions.id").
			Where("sessions.user_id = ?", *req.UserID)
	}
	if req.Status != nil {
		query = query.Where("status = ?", *req.Status)
	}
	if req.PaymentType != nil {
		query = query.Where("payment_type = ?", *req.PaymentType)
	}
	if req.PaymentMethod != nil {
		query = query.Where("payment_method = ?", *req.PaymentMethod)
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

// GetPaymentStatistics получает статистику платежей
func (r *repository) GetPaymentStatistics(req *models.PaymentStatisticsRequest) (*models.PaymentStatisticsResponse, error) {
	// Структура для агрегации данных
	type StatResult struct {
		ServiceType    string `json:"service_type"`
		WithChemistry  bool   `json:"with_chemistry"`
		SessionCount   int64  `json:"session_count"`
		TotalAmount    int64  `json:"total_amount"` // сумма с учетом возвратов (amount - refunded_amount)
	}

	var results []StatResult

	// Базовый запрос с JOIN платежей и сессий
	// Учитываем возвраты: вычитаем refunded_amount из amount
	query := r.db.Table("payments").
		Select(`
			sessions.service_type,
			sessions.with_chemistry,
			COUNT(DISTINCT sessions.id) as session_count,
			COALESCE(SUM(payments.amount - payments.refunded_amount), 0) as total_amount
		`).
		Joins("JOIN sessions ON payments.session_id = sessions.id").
		Group("sessions.service_type, sessions.with_chemistry").
		Order("sessions.service_type, sessions.with_chemistry")

	// Применяем фильтры
	if req.UserID != nil {
		query = query.Where("sessions.user_id = ?", *req.UserID)
	}
	if req.DateFrom != nil {
		query = query.Where("payments.created_at >= ?", *req.DateFrom)
	}
	if req.DateTo != nil {
		query = query.Where("payments.created_at <= ?", *req.DateTo)
	}

	// Выполняем запрос
	if err := query.Find(&results).Error; err != nil {
		return nil, err
	}

	// Преобразуем результаты в нужный формат
	statistics := make([]models.ServiceTypeStatistics, 0, len(results))
	var totalSessions int64
	var totalAmount int64

	for _, result := range results {
		statistics = append(statistics, models.ServiceTypeStatistics{
			ServiceType:   result.ServiceType,
			WithChemistry: result.WithChemistry,
			SessionCount:  int(result.SessionCount),
			TotalAmount:   int(result.TotalAmount),
		})
		totalSessions += result.SessionCount
		totalAmount += result.TotalAmount
	}

	// Формируем итоговую статистику
	total := models.ServiceTypeStatistics{
		ServiceType:   "total",
		WithChemistry: false,
		SessionCount:  int(totalSessions),
		TotalAmount:   int(totalAmount),
	}

	// Формируем описание периода
	period := "все время"
	if req.DateFrom != nil && req.DateTo != nil {
		period = "с " + req.DateFrom.Format("02.01.2006") + " по " + req.DateTo.Format("02.01.2006")
	} else if req.DateFrom != nil {
		period = "с " + req.DateFrom.Format("02.01.2006")
	} else if req.DateTo != nil {
		period = "по " + req.DateTo.Format("02.01.2006")
	}

	return &models.PaymentStatisticsResponse{
		Statistics: statistics,
		Total:      total,
		Period:     period,
	}, nil
} 