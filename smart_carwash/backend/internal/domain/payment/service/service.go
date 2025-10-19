package service

import (
	"carwash_backend/internal/domain/payment/models"
	"carwash_backend/internal/domain/payment/repository"
	settingsRepo "carwash_backend/internal/domain/settings/repository"
	"carwash_backend/internal/logger"
	"carwash_backend/internal/metrics"
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"sort"
	"time"

	"github.com/google/uuid"
	"github.com/sirupsen/logrus"
)

// SessionStatusUpdater интерфейс для обновления статуса сессии
type SessionStatusUpdater interface {
	UpdateSessionStatus(sessionID uuid.UUID, status string) error
}

// SessionExtensionUpdater интерфейс для обновления времени продления сессии
type SessionExtensionUpdater interface {
	UpdateSessionExtension(sessionID uuid.UUID, extensionTimeMinutes int) error
}

// TinkoffClient интерфейс для работы с Tinkoff API
type TinkoffClient interface {
	CreatePayment(orderID string, amount int, description string) (*TinkoffPaymentResponse, error)
	GetPaymentStatus(paymentID string) (*TinkoffPaymentStatusResponse, error)
	RefundPayment(paymentID string, amount int) (*TinkoffRefundResponse, error)
	VerifyWebhookSignature(data []byte, signature string) bool
}

// TinkoffPaymentResponse ответ от Tinkoff API при создании платежа
type TinkoffPaymentResponse struct {
	Success     bool   `json:"Success"`
	ErrorCode   string `json:"ErrorCode"`
	TerminalKey string `json:"TerminalKey"`
	OrderId     string `json:"OrderId"`
	PaymentId   string `json:"PaymentId"`
	Amount      int    `json:"Amount"`
	PaymentURL  string `json:"PaymentURL"`
}

// TinkoffPaymentStatusResponse ответ от Tinkoff API при проверке статуса
type TinkoffPaymentStatusResponse struct {
	Success   bool   `json:"Success"`
	ErrorCode string `json:"ErrorCode"`
	Status    string `json:"Status"`
	PaymentId string `json:"PaymentId"`
	Amount    int    `json:"Amount"`
}

// TinkoffRefundResponse ответ от Tinkoff API при возврате платежа
type TinkoffRefundResponse struct {
	Success   bool   `json:"Success"`
	ErrorCode string `json:"ErrorCode"`
	Status    string `json:"Status"`
	PaymentId string `json:"PaymentId"`
	Amount    int    `json:"Amount"`
}

// Service интерфейс для бизнес-логики платежей
type Service interface {
	CalculatePrice(req *models.CalculatePriceRequest) (*models.CalculatePriceResponse, error)
	CalculateExtensionPrice(req *models.CalculateExtensionPriceRequest) (*models.CalculateExtensionPriceResponse, error)
	CreatePayment(req *models.CreatePaymentRequest) (*models.CreatePaymentResponse, error)
	CreateExtensionPayment(req *models.CreateExtensionPaymentRequest) (*models.CreateExtensionPaymentResponse, error)
	GetPaymentByID(paymentID uuid.UUID) (*models.Payment, error)
	GetMainPaymentBySessionID(sessionID uuid.UUID) (*models.Payment, error)
	GetPaymentsBySessionID(sessionID uuid.UUID) (*models.GetPaymentsBySessionResponse, error)
	GetPaymentStatus(req *models.GetPaymentStatusRequest) (*models.GetPaymentStatusResponse, error)
	HandleWebhook(req *models.WebhookRequest) error
	ListPayments(req *models.AdminListPaymentsRequest) (*models.AdminListPaymentsResponse, error)
	RefundPayment(req *models.RefundPaymentRequest) (*models.RefundPaymentResponse, error)
	CalculatePartialRefund(req *models.CalculatePartialRefundRequest) (*models.CalculatePartialRefundResponse, error)
	CalculateSessionRefund(req *models.CalculateSessionRefundRequest) (*models.CalculateSessionRefundResponse, error)
	GetPaymentStatistics(req *models.PaymentStatisticsRequest) (*models.PaymentStatisticsResponse, error)
	CreateForCashier(sessionID uuid.UUID, amount int) (*models.Payment, error)
	CashierListPayments(req *models.CashierPaymentsRequest) (*models.AdminListPaymentsResponse, error)
	GetCashierLastShiftStatistics(req *models.CashierLastShiftStatisticsRequest) (*models.CashierLastShiftStatisticsResponse, error)
}

// service реализация Service
type service struct {
	repository           repository.Repository
	settingsRepo         settingsRepo.Repository
	sessionUpdater       SessionStatusUpdater
	sessionExtensionUpdater SessionExtensionUpdater
	tinkoffClient        TinkoffClient
	terminalKey          string
	secretKey            string
	metrics              *metrics.Metrics
}

// generateRandomString генерирует короткую случайную строку
func generateRandomString(length int) string {
	bytes := make([]byte, length)
	rand.Read(bytes)
	return hex.EncodeToString(bytes)[:length]
}

// NewService создает новый экземпляр Service
func NewService(repository repository.Repository, settingsRepo settingsRepo.Repository, sessionUpdater SessionStatusUpdater, sessionExtensionUpdater SessionExtensionUpdater, tinkoffClient TinkoffClient, terminalKey, secretKey string, metrics *metrics.Metrics) Service {
	return &service{
		repository:              repository,
		settingsRepo:            settingsRepo,
		sessionUpdater:          sessionUpdater,
		sessionExtensionUpdater: sessionExtensionUpdater,
		tinkoffClient:           tinkoffClient,
		terminalKey:             terminalKey,
		secretKey:               secretKey,
		metrics:                 metrics,
	}
}

// CalculatePrice рассчитывает цену для услуги
func (s *service) CalculatePrice(req *models.CalculatePriceRequest) (*models.CalculatePriceResponse, error) {
	logger.WithFields(logrus.Fields{
		"service_type":         req.ServiceType,
		"rental_time_minutes":  req.RentalTimeMinutes,
		"with_chemistry":       req.WithChemistry,
	}).Info("Payment Service - CalculatePrice: начало расчета цены")
	
	// Получаем базовую цену за минуту
	basePriceSetting, err := s.settingsRepo.GetServiceSetting(req.ServiceType, "price_per_minute")
	if err != nil {
		logger.WithFields(logrus.Fields{
			"service_type": req.ServiceType,
			"error":        err,
		}).Error("Payment Service - CalculatePrice: ошибка получения базовой цены")
		return nil, fmt.Errorf("не удалось получить базовую цену: %w", err)
	}

	// Проверяем, что настройка найдена
	if basePriceSetting == nil {
		return nil, fmt.Errorf("настройка базовой цены для услуги '%s' не найдена", req.ServiceType)
	}

	var basePricePerMinute int
	if err := json.Unmarshal(basePriceSetting.SettingValue, &basePricePerMinute); err != nil {
		return nil, fmt.Errorf("неверный формат базовой цены в настройках: %w", err)
	}

	// Рассчитываем базовую цену
	basePrice := basePricePerMinute * req.RentalTimeMinutes
	totalPrice := basePrice

	// Разбивка цены
	breakdown := models.PriceBreakdown{
		BasePrice:     basePrice,
		ChemistryPrice: 0,
	}

	// Если используется химия, добавляем стоимость химии
	if req.WithChemistry && req.ChemistryTimeMinutes > 0 {
		chemistryPriceSetting, err := s.settingsRepo.GetServiceSetting(req.ServiceType, "chemistry_price_per_minute")
		if err != nil {
			return nil, fmt.Errorf("не удалось получить цену химии: %w", err)
		}

		// Проверяем, что настройка найдена
		if chemistryPriceSetting == nil {
			return nil, fmt.Errorf("настройка цены химии для услуги '%s' не найдена", req.ServiceType)
		}

		var chemistryPricePerMinute int
		if err := json.Unmarshal(chemistryPriceSetting.SettingValue, &chemistryPricePerMinute); err != nil {
			return nil, fmt.Errorf("неверный формат цены химии в настройках: %w", err)
		}

		// НОВАЯ ФОРМУЛА: цена химии = chemistry_price_per_minute * chemistry_time_minutes
		chemistryPrice := chemistryPricePerMinute * req.ChemistryTimeMinutes
		breakdown.ChemistryPrice = chemistryPrice
		totalPrice += chemistryPrice
	}

	return &models.CalculatePriceResponse{
		Price:     totalPrice,
		Currency:  "RUB",
		Breakdown: breakdown,
	}, nil
}

// CalculateExtensionPrice рассчитывает цену для продления сессии
func (s *service) CalculateExtensionPrice(req *models.CalculateExtensionPriceRequest) (*models.CalculateExtensionPriceResponse, error) {
	// Получаем базовую цену за минуту
	basePriceSetting, err := s.settingsRepo.GetServiceSetting(req.ServiceType, "price_per_minute")
	if err != nil {
		return nil, fmt.Errorf("не удалось получить базовую цену: %w", err)
	}

	// Проверяем, что настройка найдена
	if basePriceSetting == nil {
		return nil, fmt.Errorf("настройка базовой цены для услуги '%s' не найдена", req.ServiceType)
	}

	var basePricePerMinute int
	if err := json.Unmarshal(basePriceSetting.SettingValue, &basePricePerMinute); err != nil {
		return nil, fmt.Errorf("неверный формат базовой цены в настройках: %w", err)
	}

	// Рассчитываем базовую цену продления
	basePrice := basePricePerMinute * req.ExtensionTimeMinutes
	totalPrice := basePrice

	// Разбивка цены
	breakdown := models.PriceBreakdown{
		BasePrice:     basePrice,
		ChemistryPrice: 0,
	}

	// Если используется химия, добавляем стоимость химии
	if req.WithChemistry {
		chemistryPriceSetting, err := s.settingsRepo.GetServiceSetting(req.ServiceType, "chemistry_price_per_minute")
		if err != nil {
			return nil, fmt.Errorf("не удалось получить цену химии: %w", err)
		}

		// Проверяем, что настройка найдена
		if chemistryPriceSetting == nil {
			return nil, fmt.Errorf("настройка цены химии для услуги '%s' не найдена", req.ServiceType)
		}

		var chemistryPricePerMinute int
		if err := json.Unmarshal(chemistryPriceSetting.SettingValue, &chemistryPricePerMinute); err != nil {
			return nil, fmt.Errorf("неверный формат цены химии в настройках: %w", err)
		}

		chemistryPrice := chemistryPricePerMinute * req.ExtensionTimeMinutes
		breakdown.ChemistryPrice = chemistryPrice
		totalPrice += chemistryPrice
	}

	return &models.CalculateExtensionPriceResponse{
		Price:     totalPrice,
		Currency:  "RUB",
		Breakdown: breakdown,
	}, nil
}

// CreatePayment создает платеж в Tinkoff и сохраняет в БД
func (s *service) CreatePayment(req *models.CreatePaymentRequest) (*models.CreatePaymentResponse, error) {
	// Создаем уникальный orderID для основного платежа
	orderID := fmt.Sprintf("%s_main", req.SessionID.String())
	description := fmt.Sprintf("Оплата услуги автомойки (сессия: %s)", req.SessionID.String())

	tinkoffResp, err := s.tinkoffClient.CreatePayment(orderID, req.Amount, description)
	if err != nil {
		return nil, fmt.Errorf("ошибка создания платежа в Tinkoff: %w", err)
	}

	if !tinkoffResp.Success {
		return nil, fmt.Errorf("ошибка Tinkoff: %s", tinkoffResp.ErrorCode)
	}

	// Создаем платеж в БД
	expiresAt := time.Now().Add(15 * time.Minute) // Платеж действителен 15 минут

	payment := &models.Payment{
		SessionID:   req.SessionID,
		Amount:      req.Amount,
		Currency:    req.Currency,
		Status:      models.PaymentStatusPending,
		PaymentType: models.PaymentTypeMain,
		PaymentURL:  tinkoffResp.PaymentURL,
		TinkoffID:  tinkoffResp.PaymentId,
		ExpiresAt:   &expiresAt,
	}

	if err := s.repository.CreatePayment(payment); err != nil {
		return nil, fmt.Errorf("ошибка сохранения платежа: %w", err)
	}

	logger.WithFields(logrus.Fields{
		"payment_id":  payment.ID,
		"session_id":  payment.SessionID,
		"amount":      payment.Amount,
		"tinkoff_id":  payment.TinkoffID,
	}).Info("Создан платеж")

	return &models.CreatePaymentResponse{
		Payment: *payment,
	}, nil
}

// CreateExtensionPayment создает платеж продления в Tinkoff и сохраняет в БД
func (s *service) CreateExtensionPayment(req *models.CreateExtensionPaymentRequest) (*models.CreateExtensionPaymentResponse, error) {
	// Создаем уникальный orderID для платежа продления
	orderID := fmt.Sprintf("ext_%s", generateRandomString(12))
	description := fmt.Sprintf("Продление сессии автомойки (сессия: %s)", req.SessionID.String())

	tinkoffResp, err := s.tinkoffClient.CreatePayment(orderID, req.Amount, description)
	if err != nil {
		return nil, fmt.Errorf("ошибка создания платежа продления в Tinkoff: %w", err)
	}

	if !tinkoffResp.Success {
		return nil, fmt.Errorf("ошибка Tinkoff: %s", tinkoffResp.ErrorCode)
	}

	// Создаем платеж продления в БД
	expiresAt := time.Now().Add(15 * time.Minute) // Платеж действителен 15 минут

	payment := &models.Payment{
		SessionID:   req.SessionID,
		Amount:      req.Amount,
		Currency:    req.Currency,
		Status:      models.PaymentStatusPending,
		PaymentType: models.PaymentTypeExtension,
		PaymentURL:  tinkoffResp.PaymentURL,
		TinkoffID:   tinkoffResp.PaymentId,
		ExpiresAt:   &expiresAt,
	}

	if err := s.repository.CreatePayment(payment); err != nil {
		return nil, fmt.Errorf("ошибка сохранения платежа продления: %w", err)
	}

	logger.Printf("Создан платеж продления: ID=%s, SessionID=%s, Amount=%d, TinkoffID=%s", 
		payment.ID, payment.SessionID, payment.Amount, payment.TinkoffID)

	return &models.CreateExtensionPaymentResponse{
		Payment: *payment,
	}, nil
}

// GetPaymentStatus получает статус платежа
func (s *service) GetPaymentStatus(req *models.GetPaymentStatusRequest) (*models.GetPaymentStatusResponse, error) {
	payment, err := s.repository.GetPaymentByID(req.PaymentID)
	if err != nil {
		return nil, fmt.Errorf("платеж не найден: %w", err)
	}

	return &models.GetPaymentStatusResponse{
		Payment: *payment,
	}, nil
}

// GetPaymentByID получает платеж по ID
func (s *service) GetPaymentByID(paymentID uuid.UUID) (*models.Payment, error) {
	payment, err := s.repository.GetPaymentByID(paymentID)
	if err != nil {
		return nil, fmt.Errorf("платеж не найден: %w", err)
	}

	return payment, nil
}

// GetMainPaymentBySessionID получает основной платеж сессии
func (s *service) GetMainPaymentBySessionID(sessionID uuid.UUID) (*models.Payment, error) {
	payment, err := s.repository.GetPaymentBySessionID(sessionID)
	if err != nil {
		return nil, fmt.Errorf("ошибка получения основного платежа сессии: %w", err)
	}
	return payment, nil
}

// GetPaymentsBySessionID получает все платежи сессии
func (s *service) GetPaymentsBySessionID(sessionID uuid.UUID) (*models.GetPaymentsBySessionResponse, error) {
	payments, err := s.repository.GetPaymentsBySessionID(sessionID)
	if err != nil {
		return nil, fmt.Errorf("ошибка получения платежей сессии: %w", err)
	}

	var mainPayment *models.Payment
	var extensionPayments []models.Payment

	for _, payment := range payments {
		if payment.PaymentType == models.PaymentTypeMain {
			mainPayment = &payment
		} else if payment.PaymentType == models.PaymentTypeExtension {
			extensionPayments = append(extensionPayments, payment)
		}
	}

	return &models.GetPaymentsBySessionResponse{
		MainPayment:       mainPayment,
		ExtensionPayments: extensionPayments,
	}, nil
}


// HandleWebhook обрабатывает webhook от Tinkoff
func (s *service) HandleWebhook(req *models.WebhookRequest) error {
	logger.Printf("Получен webhook от Tinkoff: PaymentId=%d, Status=%s, Success=%v", 
		req.PaymentId, req.Status, req.Success)

	// Получаем платеж по Tinkoff ID
	payment, err := s.repository.GetPaymentByTinkoffID(fmt.Sprintf("%d", req.PaymentId))
	if err != nil {
		return fmt.Errorf("платеж не найден: %w", err)
	}

	// Обрабатываем разные статусы
	switch req.Status {
	// Успешные платежи
	case "CONFIRMED", "AUTHORIZED":
		payment.Status = models.PaymentStatusSucceeded
		logger.Printf("Платеж подтвержден: ID=%s, Status=%s", payment.ID, payment.Status)
		
		// Записываем метрику успешного платежа
		if s.metrics != nil {
			s.metrics.RecordPayment("succeeded", "tinkoff", "unknown") // service_type будет получен из сессии
		}
		
	// Неудачные платежи
	case "CANCELED", "REJECTED", "AUTH_FAIL", "DEADLINE_EXPIRED", "ATTEMPTS_EXPIRED":
		payment.Status = models.PaymentStatusFailed
		logger.Printf("Платеж неудачен (%s): ID=%s, Status=%s", req.Status, payment.ID, payment.Status)
		
		// Записываем метрику неудачного платежа
		if s.metrics != nil {
			s.metrics.RecordPayment("failed", "tinkoff", "unknown") // service_type будет получен из сессии
		}
		
	// Возвраты
	case "REFUNDING", "ASYNC_REFUNDING":
		// Возврат в процессе - не меняем статус платежа, но логируем
		logger.Printf("Возврат в процессе (%s): ID=%s, PaymentID=%d", req.Status, payment.ID, req.PaymentId)
		
	case "PARTIAL_REFUNDED":
		// Частичный возврат завершен
		payment.RefundedAmount = req.Amount // сумма возврата в копейках
		now := time.Now()
		payment.RefundedAt = &now
		// Не меняем статус платежа на refunded, так как это частичный возврат
		logger.Printf("Частичный возврат завершен: ID=%s, RefundedAmount=%d", 
			payment.ID, payment.RefundedAmount)
		
	case "REFUNDED":
		// Полный возврат завершен
		payment.Status = models.PaymentStatusRefunded
		payment.RefundedAmount = req.Amount // сумма возврата в копейках
		now := time.Now()
		payment.RefundedAt = &now
		logger.Printf("Полный возврат завершен: ID=%s, Status=%s, RefundedAmount=%d", 
			payment.ID, payment.Status, payment.RefundedAmount)
		
	// Промежуточные статусы - не меняем статус платежа
	case "NEW", "FORM_SHOWED", "AUTHORIZING", "CONFIRMING":
		logger.Printf("Промежуточный статус (%s): ID=%s, PaymentID=%d", req.Status, payment.ID, req.PaymentId)
		return nil // Не обновляем платеж для промежуточных статусов
		
	default:
		// Для неизвестных статусов обновляем на основе Success
		if req.Success {
			payment.Status = models.PaymentStatusSucceeded
		} else {
			payment.Status = models.PaymentStatusFailed
		}
		logger.Printf("Неизвестный статус (%s): ID=%s, Status=%s", req.Status, payment.ID, payment.Status)
	}

	// Обновляем платеж только если статус изменился
	if err := s.repository.UpdatePayment(payment); err != nil {
		return fmt.Errorf("ошибка обновления статуса платежа: %w", err)
	}

	logger.Printf("Обновлен статус платежа: ID=%s, Status=%s", payment.ID, payment.Status)

	// Обновляем статус связанной сессии (только для успешных/неудачных платежей, не для возвратов)
	if payment.Status != models.PaymentStatusRefunded {
		if err := s.updateSessionStatus(payment); err != nil {
			logger.Printf("Ошибка обновления статуса сессии: %v", err)
		}
		
		// Если платеж успешен и это платеж продления, обновляем время продления сессии
		if payment.Status == models.PaymentStatusSucceeded && payment.PaymentType == models.PaymentTypeExtension {
			if err := s.updateSessionExtension(payment); err != nil {
				logger.Printf("Ошибка обновления времени продления сессии: %v", err)
			}
		}
	}

	return nil
}

// updateSessionStatus обновляет статус сессии в зависимости от статуса платежа
func (s *service) updateSessionStatus(payment *models.Payment) error {
	// Определяем новый статус сессии в зависимости от статуса платежа
	var newSessionStatus string
	
	if payment.Status == models.PaymentStatusSucceeded {
		newSessionStatus = "in_queue"
		logger.Printf("Платеж успешен, обновляем сессию %s в статус 'in_queue'", payment.SessionID)
	} else if payment.Status == models.PaymentStatusFailed {
		newSessionStatus = "payment_failed"
		logger.Printf("Платеж неудачен, обновляем сессию %s в статус 'payment_failed'", payment.SessionID)
	} else {
		// Для других статусов платежа не меняем статус сессии
		logger.Printf("Статус платежа %s, статус сессии не изменяется", payment.Status)
		return nil
	}
	
	// Обновляем статус сессии через Session Status Updater
	err := s.sessionUpdater.UpdateSessionStatus(payment.SessionID, newSessionStatus)
	if err != nil {
		return fmt.Errorf("ошибка обновления статуса сессии: %w", err)
	}
	
	logger.Printf("Статус сессии %s успешно обновлен на '%s'", payment.SessionID, newSessionStatus)
	return nil
}

// updateSessionExtension обновляет время продления сессии при успешном платеже продления
func (s *service) updateSessionExtension(payment *models.Payment) error {
	// Обновляем время продления сессии через Session Extension Updater
	err := s.sessionExtensionUpdater.UpdateSessionExtension(payment.SessionID, 0) // 0 означает использовать requested_extension_time_minutes
	if err != nil {
		return fmt.Errorf("ошибка обновления времени продления сессии: %w", err)
	}
	
	logger.Printf("Время продления сессии %s успешно обновлено", payment.SessionID)
	return nil
}

// ListPayments получает список платежей для админки
func (s *service) ListPayments(req *models.AdminListPaymentsRequest) (*models.AdminListPaymentsResponse, error) {
	payments, total, err := s.repository.ListPayments(req)
	if err != nil {
		return nil, fmt.Errorf("ошибка получения списка платежей: %w", err)
	}

	limit := 50
	if req.Limit != nil {
		limit = *req.Limit
	}
	offset := 0
	if req.Offset != nil {
		offset = *req.Offset
	}

	return &models.AdminListPaymentsResponse{
		Payments: payments,
		Total:    total,
		Limit:    limit,
		Offset:   offset,
	}, nil
}

// RefundPayment возвращает деньги за платеж
func (s *service) RefundPayment(req *models.RefundPaymentRequest) (*models.RefundPaymentResponse, error) {
	// Получаем платеж по ID
	payment, err := s.repository.GetPaymentByID(req.PaymentID)
	if err != nil {
		return nil, fmt.Errorf("платеж не найден: %w", err)
	}

	// Проверяем, что платеж успешен
	if payment.Status != models.PaymentStatusSucceeded {
		return nil, fmt.Errorf("невозможно вернуть деньги за платеж со статусом '%s'", payment.Status)
	}

	// Проверяем, что сумма возврата не превышает оплаченную сумму
	if req.Amount > payment.Amount {
		return nil, fmt.Errorf("сумма возврата (%d) не может превышать оплаченную сумму (%d)", req.Amount, payment.Amount)
	}

	// Проверяем, что сумма возврата не превышает уже возвращенную сумму
	if req.Amount > (payment.Amount - payment.RefundedAmount) {
		return nil, fmt.Errorf("сумма возврата (%d) не может превышать оставшуюся сумму (%d)", req.Amount, payment.Amount-payment.RefundedAmount)
	}

	// Выполняем возврат через Tinkoff API
	refundResp, err := s.tinkoffClient.RefundPayment(payment.TinkoffID, req.Amount)
	if err != nil {
		return nil, fmt.Errorf("ошибка возврата через Tinkoff API: %w", err)
	}

	// Проверяем успешность возврата
	if !refundResp.Success {
		return nil, fmt.Errorf("ошибка возврата в Tinkoff: %s", refundResp.ErrorCode)
	}

	// Обновляем информацию о платеже
	now := time.Now()
	payment.RefundedAmount += req.Amount
	payment.RefundedAt = &now

	// Если возвращена вся сумма, меняем статус на refunded
	if payment.RefundedAmount >= payment.Amount {
		payment.Status = models.PaymentStatusRefunded
	}

	// Сохраняем обновленный платеж
	if err := s.repository.UpdatePayment(payment); err != nil {
		return nil, fmt.Errorf("ошибка обновления платежа: %w", err)
	}

	// Создаем информацию о возврате
	refund := models.Refund{
		ID:        uuid.New(),
		PaymentID: payment.ID,
		Amount:    req.Amount,
		Status:    "succeeded",
		CreatedAt: now,
	}

	logger.Printf("Успешно выполнен возврат: PaymentID=%s, Amount=%d, TotalRefunded=%d", 
		payment.ID, req.Amount, payment.RefundedAmount)

	return &models.RefundPaymentResponse{
		Payment: *payment,
		Refund:  refund,
	}, nil
}

// CalculatePartialRefund рассчитывает сумму частичного возврата при досрочном завершении сессии
func (s *service) CalculatePartialRefund(req *models.CalculatePartialRefundRequest) (*models.CalculatePartialRefundResponse, error) {
	// Получаем платеж по ID
	payment, err := s.repository.GetPaymentByID(req.PaymentID)
	if err != nil {
		return nil, fmt.Errorf("платеж не найден: %w", err)
	}

	// Проверяем, что платеж успешен
	if payment.Status != models.PaymentStatusSucceeded {
		return nil, fmt.Errorf("невозможно рассчитать возврат для платежа со статусом '%s'", payment.Status)
	}

	// Получаем базовую цену за минуту для типа услуги
	basePriceSetting, err := s.settingsRepo.GetServiceSetting(req.ServiceType, "price_per_minute")
	if err != nil {
		return nil, fmt.Errorf("не удалось получить базовую цену: %w", err)
	}

	// Проверяем, что настройка найдена
	if basePriceSetting == nil {
		return nil, fmt.Errorf("настройка базовой цены для услуги '%s' не найдена", req.ServiceType)
	}

	var basePricePerMinute int
	if err := json.Unmarshal(basePriceSetting.SettingValue, &basePricePerMinute); err != nil {
		return nil, fmt.Errorf("неверный формат базовой цены в настройках: %w", err)
	}

	// Рассчитываем цену за секунду (цена за минуту / 60)
	basePricePerSecond := basePricePerMinute / 60

	// Рассчитываем общую продолжительность сессии в секундах
	totalSessionSeconds := (req.RentalTimeMinutes + req.ExtensionTimeMinutes) * 60

	// Используем переданное использованное время в секундах
	usedTimeSeconds := req.UsedTimeSeconds

	// Рассчитываем неиспользованное время в секундах
	unusedTimeSeconds := totalSessionSeconds - usedTimeSeconds

	// Если неиспользованное время отрицательное или нулевое, возврат не нужен
	if unusedTimeSeconds <= 0 {
		return &models.CalculatePartialRefundResponse{
			RefundAmount: 0,
			UsedTimeSeconds: usedTimeSeconds,
			UnusedTimeSeconds: 0,
			PricePerSecond: basePricePerSecond,
			TotalSessionSeconds: totalSessionSeconds,
		}, nil
	}

	// Рассчитываем сумму возврата в копейках (неиспользованные секунды * цена за секунду)
	refundAmountKopecks := unusedTimeSeconds * basePricePerSecond

	// Округляем в пользу системы (до рублей)
	// 1 рубль = 100 копеек
	refundAmountRubles := refundAmountKopecks / 100

	// Конвертируем обратно в копейки для точности
	refundAmount := refundAmountRubles * 100

	// Проверяем, что сумма возврата не превышает оплаченную сумму
	if refundAmount > payment.Amount {
		refundAmount = payment.Amount
	}

	// Проверяем, что сумма возврата не превышает уже возвращенную сумму
	remainingAmount := payment.Amount - payment.RefundedAmount
	if refundAmount > remainingAmount {
		refundAmount = remainingAmount
	}

	logger.Printf("Рассчитан частичный возврат: PaymentID=%s, UsedTime=%ds, UnusedTime=%ds, RefundAmount=%d", 
		payment.ID, usedTimeSeconds, unusedTimeSeconds, refundAmount)

	return &models.CalculatePartialRefundResponse{
		RefundAmount: refundAmount,
		UsedTimeSeconds: usedTimeSeconds,
		UnusedTimeSeconds: unusedTimeSeconds,
		PricePerSecond: basePricePerSecond,
		TotalSessionSeconds: totalSessionSeconds,
	}, nil
}

// CalculateSessionRefund рассчитывает возврат по всем платежам сессии
func (s *service) CalculateSessionRefund(req *models.CalculateSessionRefundRequest) (*models.CalculateSessionRefundResponse, error) {
	// Получаем все платежи сессии
	paymentsResp, err := s.GetPaymentsBySessionID(req.SessionID)
	if err != nil {
		return nil, fmt.Errorf("ошибка получения платежей сессии: %w", err)
	}

	// Собираем все платежи в один слайс для удобства обработки
	var allPayments []models.Payment
	
	if paymentsResp.MainPayment != nil {
		allPayments = append(allPayments, *paymentsResp.MainPayment)
	}
	
	for _, payment := range paymentsResp.ExtensionPayments {
		allPayments = append(allPayments, payment)
	}

	if len(allPayments) == 0 {
		return &models.CalculateSessionRefundResponse{
			TotalRefundAmount: 0,
			Refunds:           []models.SessionRefundBreakdown{},
		}, nil
	}

	// Сортируем платежи по времени создания (хронологический порядок)
	// Основной платеж всегда первый, затем платежи продления по времени создания
	sort.Slice(allPayments, func(i, j int) bool {
		return allPayments[i].CreatedAt.Before(allPayments[j].CreatedAt)
	})

	// Получаем базовую цену за минуту для типа услуги
	basePriceSetting, err := s.settingsRepo.GetServiceSetting(req.ServiceType, "price_per_minute")
	if err != nil {
		return nil, fmt.Errorf("не удалось получить базовую цену: %w", err)
	}

	if basePriceSetting == nil {
		return nil, fmt.Errorf("настройка базовой цены для услуги '%s' не найдена", req.ServiceType)
	}

	var basePricePerMinute int
	if err := json.Unmarshal(basePriceSetting.SettingValue, &basePricePerMinute); err != nil {
		return nil, fmt.Errorf("неверный формат базовой цены в настройках: %w", err)
	}

	// Рассчитываем цену за секунду
	basePricePerSecond := basePricePerMinute / 60

	// Рассчитываем общее время сессии в секундах для каждого платежа
	var totalSessionSeconds []int
	var totalTimeSeconds int

	// Распределяем время между платежами
	// Основной платеж покрывает RentalTimeMinutes
	// Платежи продления покрывают ExtensionTimeMinutes
	mainPaymentTimeSeconds := req.RentalTimeMinutes * 60

	logger.Printf("Расчет времени платежей: RentalTime=%dmin, ExtensionTime=%dmin, ExtensionPayments=%d", 
		req.RentalTimeMinutes, req.ExtensionTimeMinutes, len(paymentsResp.ExtensionPayments))

	for _, payment := range allPayments {
		var paymentTimeSeconds int
		if payment.PaymentType == models.PaymentTypeMain {
			paymentTimeSeconds = mainPaymentTimeSeconds
		} else {
			// Для платежей продления используем время из ExtensionTimeMinutes
			// Распределяем общее время продления между всеми платежами продления
			if len(paymentsResp.ExtensionPayments) > 0 {
				// Простое распределение: каждому платежу продления одинаковое время
				extensionTimeSeconds := req.ExtensionTimeMinutes * 60
				paymentTimeSeconds = extensionTimeSeconds / len(paymentsResp.ExtensionPayments)
			} else {
				paymentTimeSeconds = 0
			}
		}
		
		totalSessionSeconds = append(totalSessionSeconds, paymentTimeSeconds)
		totalTimeSeconds += paymentTimeSeconds
		
		logger.Printf("  Платеж %s: Type=%s, Amount=%d, Time=%ds", 
			payment.ID, payment.PaymentType, payment.Amount, paymentTimeSeconds)
	}

	// Рассчитываем возврат последовательно по времени
	var refunds []models.SessionRefundBreakdown
	totalRefundAmount := 0
	remainingUsedTime := req.UsedTimeSeconds
	
	logger.Printf("Расчет возврата: UsedTime=%ds, RemainingUsedTime=%ds", req.UsedTimeSeconds, remainingUsedTime)

	for i, payment := range allPayments {
		paymentTimeSeconds := totalSessionSeconds[i]
		
		// Определяем, сколько времени использовано из этого платежа
		var usedTimeForPayment int
		if remainingUsedTime >= paymentTimeSeconds {
			// Использовано все время этого платежа
			usedTimeForPayment = paymentTimeSeconds
			remainingUsedTime -= paymentTimeSeconds
		} else if remainingUsedTime > 0 {
			// Использована часть времени этого платежа
			usedTimeForPayment = remainingUsedTime
			remainingUsedTime = 0
		} else {
			// Время этого платежа не использовано
			usedTimeForPayment = 0
		}
		
		logger.Printf("  Платеж %d: PaymentTime=%ds, Used=%ds, RemainingUsedTime=%ds", 
			i+1, paymentTimeSeconds, usedTimeForPayment, remainingUsedTime)
		
		// Рассчитываем неиспользованное время для этого платежа
		unusedTimeForPayment := paymentTimeSeconds - usedTimeForPayment
		
		// Если неиспользованное время нулевое, возврат не нужен
		if unusedTimeForPayment <= 0 {
			refunds = append(refunds, models.SessionRefundBreakdown{
				PaymentID:         payment.ID,
				PaymentType:       payment.PaymentType,
				OriginalAmount:    payment.Amount,
				RefundAmount:      0,
				UsedTimeSeconds:   usedTimeForPayment,
				UnusedTimeSeconds: 0,
				PricePerSecond:    basePricePerSecond,
			})
			continue
		}

		// Рассчитываем сумму возврата для этого платежа
		refundAmountForPayment := unusedTimeForPayment * basePricePerSecond
		
		// Округляем в пользу системы (до рублей)
		refundAmountRubles := refundAmountForPayment / 100
		refundAmountForPayment = refundAmountRubles * 100

		// Проверяем, что сумма возврата не превышает оплаченную сумму
		if refundAmountForPayment > payment.Amount {
			refundAmountForPayment = payment.Amount
		}

		// Проверяем, что сумма возврата не превышает уже возвращенную сумму
		remainingAmount := payment.Amount - payment.RefundedAmount
		if refundAmountForPayment > remainingAmount {
			refundAmountForPayment = remainingAmount
		}

		refunds = append(refunds, models.SessionRefundBreakdown{
			PaymentID:         payment.ID,
			PaymentType:       payment.PaymentType,
			OriginalAmount:    payment.Amount,
			RefundAmount:      refundAmountForPayment,
			UsedTimeSeconds:   usedTimeForPayment,
			UnusedTimeSeconds: unusedTimeForPayment,
			PricePerSecond:    basePricePerSecond,
		})

		totalRefundAmount += refundAmountForPayment
	}

	logger.Printf("Рассчитан возврат по сессии: SessionID=%s, TotalRefundAmount=%d, UsedTime=%ds, Payments=%d", 
		req.SessionID, totalRefundAmount, req.UsedTimeSeconds, len(allPayments))
	
	// Детальное логирование для отладки
	logger.Printf("Детали расчета:")
	for i, refund := range refunds {
		logger.Printf("  Платеж %d: ID=%s, Type=%s, Original=%d, Refund=%d, Used=%ds, Unused=%ds", 
			i+1, refund.PaymentID, refund.PaymentType, refund.OriginalAmount, refund.RefundAmount, 
			refund.UsedTimeSeconds, refund.UnusedTimeSeconds)
	}
	
	// Логируем общее время сессии
	logger.Printf("Общее время сессии: %d секунд (%d минут)", totalTimeSeconds, totalTimeSeconds/60)

	return &models.CalculateSessionRefundResponse{
		TotalRefundAmount: totalRefundAmount,
		Refunds:           refunds,
	}, nil
} 

// GetPaymentStatistics получает статистику платежей
func (s *service) GetPaymentStatistics(req *models.PaymentStatisticsRequest) (*models.PaymentStatisticsResponse, error) {
	logger.Printf("Getting payment statistics with filters: user_id=%v, date_from=%v, date_to=%v, service_type=%v",
		req.UserID, req.DateFrom, req.DateTo, req.ServiceType)

	statistics, err := s.repository.GetPaymentStatistics(req)
	if err != nil {
		logger.Printf("Error getting payment statistics: %v", err)
		return nil, fmt.Errorf("failed to get payment statistics: %w", err)
	}

	logger.Printf("Successfully retrieved payment statistics (with refunds): %d service types, total sessions: %d, total amount: %d",
		len(statistics.Statistics), statistics.Total.SessionCount, statistics.Total.TotalAmount)

	return statistics, nil
} 

// CreateForCashier создает платеж для кассира
func (s *service) CreateForCashier(sessionID uuid.UUID, amount int) (*models.Payment, error) {
	logger.Printf("Creating payment for cashier: SessionID=%s, Amount=%d", sessionID, amount)

	// Создаем платеж
	payment := &models.Payment{
		SessionID:     sessionID,
		Amount:        amount,
		Currency:      "RUB",
		Status:        models.PaymentStatusSucceeded, // Статус "оплачен" как указано в требованиях
		PaymentType:   models.PaymentTypeMain,        // Тип "основной" как указано в требованиях
		PaymentMethod: "cashier",                     // Метод "кассир" как указано в требованиях
		TinkoffID:     "",                           // Пустой TinkoffID как указано в требованиях
	}

	// Сохраняем платеж в базе данных
	err := s.repository.CreatePayment(payment)
	if err != nil {
		logger.Printf("Error creating payment for cashier: %v", err)
		return nil, fmt.Errorf("failed to create payment for cashier: %w", err)
	}

	logger.Printf("Successfully created payment for cashier: PaymentID=%s, SessionID=%s, Amount=%d", 
		payment.ID, sessionID, amount)

	return payment, nil
} 

// CashierListPayments получает список платежей для кассира
func (s *service) CashierListPayments(req *models.CashierPaymentsRequest) (*models.AdminListPaymentsResponse, error) {
	payments, total, err := s.repository.CashierListPayments(req)
	if err != nil {
		return nil, fmt.Errorf("ошибка получения списка платежей для кассира: %w", err)
	}

	limit := 50
	if req.Limit != nil {
		limit = *req.Limit
	}
	offset := 0
	if req.Offset != nil {
		offset = *req.Offset
	}

	return &models.AdminListPaymentsResponse{
		Payments: payments,
		Total:    total,
		Limit:    limit,
		Offset:   offset,
	}, nil
} 

// GetCashierLastShiftStatistics получает статистику последней смены кассира
func (s *service) GetCashierLastShiftStatistics(req *models.CashierLastShiftStatisticsRequest) (*models.CashierLastShiftStatisticsResponse, error) {
	logger.Printf("Getting cashier last shift statistics: CashierID=%s", req.CashierID)

	statistics, err := s.repository.GetCashierLastShiftStatistics(req)
	if err != nil {
		logger.Printf("Error getting cashier last shift statistics: %v", err)
		return nil, fmt.Errorf("failed to get cashier last shift statistics: %w", err)
	}

	if statistics.HasShift {
		logger.Printf("Successfully retrieved cashier last shift statistics: ShiftStartedAt=%v, ShiftEndedAt=%v, CashierSessions=%d, MiniAppSessions=%d, TotalSessions=%d",
			statistics.Statistics.ShiftStartedAt, statistics.Statistics.ShiftEndedAt,
			len(statistics.Statistics.CashierSessions), len(statistics.Statistics.MiniAppSessions), len(statistics.Statistics.TotalSessions))
	} else {
		logger.Printf("No completed shifts found for cashier: CashierID=%s", req.CashierID)
	}

	return statistics, nil
} 