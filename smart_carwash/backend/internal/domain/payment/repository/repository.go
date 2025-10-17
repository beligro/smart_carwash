package repository

import (
	"carwash_backend/internal/domain/payment/models"
	"fmt"
	"carwash_backend/internal/logger"
	"sort"
	"time"

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
	CashierListPayments(req *models.CashierPaymentsRequest) ([]models.Payment, int, error)
	GetCashierLastShiftStatistics(req *models.CashierLastShiftStatisticsRequest) (*models.CashierLastShiftStatisticsResponse, error)
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

// CashierListPayments получает список платежей для кассира с начала смены
func (r *repository) CashierListPayments(req *models.CashierPaymentsRequest) ([]models.Payment, int, error) {
	var payments []models.Payment
	var total int64

	query := r.db.Model(&models.Payment{})

	// Фильтруем по дате начала смены
	query = query.Where("created_at >= ?", req.ShiftStartedAt)

	// Получаем общее количество
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

	// Получаем платежи
	if err := query.Find(&payments).Error; err != nil {
		return nil, 0, err
	}

	return payments, int(total), nil
}

// GetCashierLastShiftStatistics получает статистику последней смены кассира
func (r *repository) GetCashierLastShiftStatistics(req *models.CashierLastShiftStatisticsRequest) (*models.CashierLastShiftStatisticsResponse, error) {
	// Сначала найдем последнюю смену кассира
	var lastShift struct {
		StartedAt time.Time `json:"started_at"`
		EndedAt   time.Time `json:"ended_at"`
	}

	logger.Printf("Searching for last shift for cashier: %s", req.CashierID)

	err := r.db.Table("cashier_shifts").
		Select("started_at, COALESCE(ended_at, expires_at) as ended_at").
		Where("cashier_id = ? AND (ended_at IS NOT NULL OR expires_at < ?)", req.CashierID, time.Now()).
		Order("started_at DESC").
		Limit(1).
		Scan(&lastShift).Error

	if err != nil {
		if err == gorm.ErrRecordNotFound {
			logger.Printf("No completed shifts found for cashier: %s", req.CashierID)
			return &models.CashierLastShiftStatisticsResponse{
				Statistics: nil,
				Message:    "У кассира нет завершенных смен",
				HasShift:   false,
			}, nil
		}
		logger.Printf("Error finding last shift: %v", err)
		return nil, err
	}

	logger.Printf("Found last shift: StartedAt=%v, EndedAt=%v", lastShift.StartedAt, lastShift.EndedAt)

	// Теперь получим статистику для этой смены
	type StatResult struct {
		ServiceType    string `json:"service_type"`
		WithChemistry  bool   `json:"with_chemistry"`
		PaymentMethod  string `json:"payment_method"`
		SessionCount   int64  `json:"session_count"`
		TotalAmount    int64  `json:"total_amount"`
	}

	var results []StatResult

	// Запрос для получения статистики с группировкой по типу услуги, химии и методу платежа
	logger.Printf("Querying statistics for period: %v to %v", lastShift.StartedAt, lastShift.EndedAt)
	
	query := r.db.Table("payments").
		Select(`
			sessions.service_type,
			sessions.with_chemistry,
			payments.payment_method,
			COUNT(DISTINCT sessions.id) as session_count,
			COALESCE(SUM(payments.amount - payments.refunded_amount), 0) as total_amount
		`).
		Joins("JOIN sessions ON payments.session_id = sessions.id").
		Where("payments.created_at >= ? AND payments.created_at <= ?", lastShift.StartedAt, lastShift.EndedAt).
		Group("sessions.service_type, sessions.with_chemistry, payments.payment_method").
		Order("sessions.service_type, sessions.with_chemistry, payments.payment_method")

	if err := query.Find(&results).Error; err != nil {
		logger.Printf("Error querying statistics: %v", err)
		return nil, err
	}

	logger.Printf("Found %d statistics results", len(results))

	// Группируем результаты по методу платежа
	cashierStats := make(map[string]models.ServiceTypeStatistics)
	miniAppStats := make(map[string]models.ServiceTypeStatistics)
	totalStats := make(map[string]models.ServiceTypeStatistics)

	for _, result := range results {
		key := result.ServiceType + "_" + fmt.Sprintf("%t", result.WithChemistry)
		
		stat := models.ServiceTypeStatistics{
			ServiceType:   result.ServiceType,
			WithChemistry: result.WithChemistry,
			SessionCount:  int(result.SessionCount),
			TotalAmount:   int(result.TotalAmount),
		}

		// Добавляем в соответствующую группу
		if result.PaymentMethod == "cashier" {
			if existing, exists := cashierStats[key]; exists {
				existing.SessionCount += stat.SessionCount
				existing.TotalAmount += stat.TotalAmount
				cashierStats[key] = existing
			} else {
				cashierStats[key] = stat
			}
		} else if result.PaymentMethod == "tinkoff" {
			if existing, exists := miniAppStats[key]; exists {
				existing.SessionCount += stat.SessionCount
				existing.TotalAmount += stat.TotalAmount
				miniAppStats[key] = existing
			} else {
				miniAppStats[key] = stat
			}
		}

		// Добавляем в общую статистику
		if existing, exists := totalStats[key]; exists {
			existing.SessionCount += stat.SessionCount
			existing.TotalAmount += stat.TotalAmount
			totalStats[key] = existing
		} else {
			totalStats[key] = stat
		}
	}

	// Преобразуем map в slice
	cashierSessions := make([]models.ServiceTypeStatistics, 0, len(cashierStats))
	miniAppSessions := make([]models.ServiceTypeStatistics, 0, len(miniAppStats))
	totalSessions := make([]models.ServiceTypeStatistics, 0, len(totalStats))

	for _, stat := range cashierStats {
		cashierSessions = append(cashierSessions, stat)
	}
	for _, stat := range miniAppStats {
		miniAppSessions = append(miniAppSessions, stat)
	}
	for _, stat := range totalStats {
		totalSessions = append(totalSessions, stat)
	}

	// Сортируем результаты
	sort.Slice(cashierSessions, func(i, j int) bool {
		if cashierSessions[i].ServiceType != cashierSessions[j].ServiceType {
			return cashierSessions[i].ServiceType < cashierSessions[j].ServiceType
		}
		return !cashierSessions[i].WithChemistry
	})

	sort.Slice(miniAppSessions, func(i, j int) bool {
		if miniAppSessions[i].ServiceType != miniAppSessions[j].ServiceType {
			return miniAppSessions[i].ServiceType < miniAppSessions[j].ServiceType
		}
		return !miniAppSessions[i].WithChemistry
	})

	sort.Slice(totalSessions, func(i, j int) bool {
		if totalSessions[i].ServiceType != totalSessions[j].ServiceType {
			return totalSessions[i].ServiceType < totalSessions[j].ServiceType
		}
		return !totalSessions[i].WithChemistry
	})

	statistics := &models.CashierShiftStatistics{
		ShiftStartedAt:   lastShift.StartedAt,
		ShiftEndedAt:     lastShift.EndedAt,
		CashierSessions:  cashierSessions,
		MiniAppSessions:  miniAppSessions,
		TotalSessions:    totalSessions,
	}

	return &models.CashierLastShiftStatisticsResponse{
		Statistics: statistics,
		Message:    "Статистика последней смены",
		HasShift:   true,
	}, nil
} 