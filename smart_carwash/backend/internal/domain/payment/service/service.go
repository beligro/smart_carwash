package service

import (
	"carwash_backend/internal/domain/payment/models"
	"carwash_backend/internal/domain/payment/repository"
	settingsRepo "carwash_backend/internal/domain/settings/repository"
	"encoding/json"
	"fmt"
	"log"
	"time"

	"github.com/google/uuid"
)

// SessionStatusUpdater интерфейс для обновления статуса сессии
type SessionStatusUpdater interface {
	UpdateSessionStatus(sessionID uuid.UUID, status string) error
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
	CreatePayment(req *models.CreatePaymentRequest) (*models.CreatePaymentResponse, error)
	GetPaymentByID(paymentID uuid.UUID) (*models.Payment, error)
	GetPaymentStatus(req *models.GetPaymentStatusRequest) (*models.GetPaymentStatusResponse, error)
	RetryPayment(req *models.RetryPaymentRequest) (*models.RetryPaymentResponse, error)
	HandleWebhook(req *models.WebhookRequest) error
	ListPayments(req *models.AdminListPaymentsRequest) (*models.AdminListPaymentsResponse, error)
	RefundPayment(req *models.RefundPaymentRequest) (*models.RefundPaymentResponse, error)
	CalculatePartialRefund(req *models.CalculatePartialRefundRequest) (*models.CalculatePartialRefundResponse, error)
}

// service реализация Service
type service struct {
	repository      repository.Repository
	settingsRepo    settingsRepo.Repository
	sessionUpdater  SessionStatusUpdater
	tinkoffClient   TinkoffClient
	terminalKey     string
	secretKey       string
}

// NewService создает новый экземпляр Service
func NewService(repository repository.Repository, settingsRepo settingsRepo.Repository, sessionUpdater SessionStatusUpdater, tinkoffClient TinkoffClient, terminalKey, secretKey string) Service {
	return &service{
		repository:    repository,
		settingsRepo:  settingsRepo,
		sessionUpdater: sessionUpdater,
		tinkoffClient: tinkoffClient,
		terminalKey:   terminalKey,
		secretKey:     secretKey,
	}
}

// CalculatePrice рассчитывает цену для услуги
func (s *service) CalculatePrice(req *models.CalculatePriceRequest) (*models.CalculatePriceResponse, error) {
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

	// Рассчитываем базовую цену
	basePrice := basePricePerMinute * req.RentalTimeMinutes
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

		chemistryPrice := chemistryPricePerMinute * req.RentalTimeMinutes
		breakdown.ChemistryPrice = chemistryPrice
		totalPrice += chemistryPrice
	}

	return &models.CalculatePriceResponse{
		Price:     totalPrice,
		Currency:  "RUB",
		Breakdown: breakdown,
	}, nil
}

// CreatePayment создает платеж в Tinkoff и сохраняет в БД
func (s *service) CreatePayment(req *models.CreatePaymentRequest) (*models.CreatePaymentResponse, error) {
	// Создаем платеж в Tinkoff
	orderID := req.SessionID.String()
	description := fmt.Sprintf("Оплата услуги автомойки (сессия: %s)", orderID)

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
		SessionID:  req.SessionID,
		Amount:     req.Amount,
		Currency:   req.Currency,
		Status:     models.PaymentStatusPending,
		PaymentURL: tinkoffResp.PaymentURL,
		TinkoffID:  tinkoffResp.PaymentId,
		ExpiresAt:  &expiresAt,
	}

	if err := s.repository.CreatePayment(payment); err != nil {
		return nil, fmt.Errorf("ошибка сохранения платежа: %w", err)
	}

	log.Printf("Создан платеж: ID=%s, SessionID=%s, Amount=%d, TinkoffID=%s", 
		payment.ID, payment.SessionID, payment.Amount, payment.TinkoffID)

	return &models.CreatePaymentResponse{
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

// RetryPayment создает новый платеж для существующей сессии
func (s *service) RetryPayment(req *models.RetryPaymentRequest) (*models.RetryPaymentResponse, error) {
	// Получаем существующий платеж
	existingPayment, err := s.repository.GetPaymentBySessionID(req.SessionID)
	if err != nil {
		return nil, fmt.Errorf("платеж для сессии не найден: %w", err)
	}

	// Создаем новый платеж с той же суммой
	retryReq := &models.CreatePaymentRequest{
		SessionID: req.SessionID,
		Amount:    existingPayment.Amount,
		Currency:  existingPayment.Currency,
	}

	createResp, err := s.CreatePayment(retryReq)
	if err != nil {
		return nil, err
	}

	return &models.RetryPaymentResponse{
		Payment: createResp.Payment,
	}, nil
}

// HandleWebhook обрабатывает webhook от Tinkoff
func (s *service) HandleWebhook(req *models.WebhookRequest) error {
	log.Printf("Получен webhook от Tinkoff: PaymentId=%d, Status=%s, Success=%v", 
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
		log.Printf("Платеж подтвержден: ID=%s, Status=%s", payment.ID, payment.Status)
		
	// Неудачные платежи
	case "CANCELED", "REJECTED", "AUTH_FAIL", "DEADLINE_EXPIRED", "ATTEMPTS_EXPIRED":
		payment.Status = models.PaymentStatusFailed
		log.Printf("Платеж неудачен (%s): ID=%s, Status=%s", req.Status, payment.ID, payment.Status)
		
	// Возвраты
	case "REFUNDING", "ASYNC_REFUNDING":
		// Возврат в процессе - не меняем статус платежа, но логируем
		log.Printf("Возврат в процессе (%s): ID=%s, PaymentID=%d", req.Status, payment.ID, req.PaymentId)
		
	case "PARTIAL_REFUNDED":
		// Частичный возврат завершен
		payment.RefundedAmount = req.Amount // сумма возврата в копейках
		now := time.Now()
		payment.RefundedAt = &now
		// Не меняем статус платежа на refunded, так как это частичный возврат
		log.Printf("Частичный возврат завершен: ID=%s, RefundedAmount=%d", 
			payment.ID, payment.RefundedAmount)
		
	case "REFUNDED":
		// Полный возврат завершен
		payment.Status = models.PaymentStatusRefunded
		payment.RefundedAmount = req.Amount // сумма возврата в копейках
		now := time.Now()
		payment.RefundedAt = &now
		log.Printf("Полный возврат завершен: ID=%s, Status=%s, RefundedAmount=%d", 
			payment.ID, payment.Status, payment.RefundedAmount)
		
	// Промежуточные статусы - не меняем статус платежа
	case "NEW", "FORM_SHOWED", "AUTHORIZING", "CONFIRMING":
		log.Printf("Промежуточный статус (%s): ID=%s, PaymentID=%d", req.Status, payment.ID, req.PaymentId)
		return nil // Не обновляем платеж для промежуточных статусов
		
	default:
		// Для неизвестных статусов обновляем на основе Success
		if req.Success {
			payment.Status = models.PaymentStatusSucceeded
		} else {
			payment.Status = models.PaymentStatusFailed
		}
		log.Printf("Неизвестный статус (%s): ID=%s, Status=%s", req.Status, payment.ID, payment.Status)
	}

	// Обновляем платеж только если статус изменился
	if err := s.repository.UpdatePayment(payment); err != nil {
		return fmt.Errorf("ошибка обновления статуса платежа: %w", err)
	}

	log.Printf("Обновлен статус платежа: ID=%s, Status=%s", payment.ID, payment.Status)

	// Обновляем статус связанной сессии (только для успешных/неудачных платежей, не для возвратов)
	if payment.Status != models.PaymentStatusRefunded {
		if err := s.updateSessionStatus(payment); err != nil {
			log.Printf("Ошибка обновления статуса сессии: %v", err)
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
		log.Printf("Платеж успешен, обновляем сессию %s в статус 'in_queue'", payment.SessionID)
	} else if payment.Status == models.PaymentStatusFailed {
		newSessionStatus = "payment_failed"
		log.Printf("Платеж неудачен, обновляем сессию %s в статус 'payment_failed'", payment.SessionID)
	} else {
		// Для других статусов платежа не меняем статус сессии
		log.Printf("Статус платежа %s, статус сессии не изменяется", payment.Status)
		return nil
	}
	
	// Обновляем статус сессии через Session Status Updater
	err := s.sessionUpdater.UpdateSessionStatus(payment.SessionID, newSessionStatus)
	if err != nil {
		return fmt.Errorf("ошибка обновления статуса сессии: %w", err)
	}
	
	log.Printf("Статус сессии %s успешно обновлен на '%s'", payment.SessionID, newSessionStatus)
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

	log.Printf("Успешно выполнен возврат: PaymentID=%s, Amount=%d, TotalRefunded=%d", 
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

	log.Printf("Рассчитан частичный возврат: PaymentID=%s, UsedTime=%ds, UnusedTime=%ds, RefundAmount=%d", 
		payment.ID, usedTimeSeconds, unusedTimeSeconds, refundAmount)

	return &models.CalculatePartialRefundResponse{
		RefundAmount: refundAmount,
		UsedTimeSeconds: usedTimeSeconds,
		UnusedTimeSeconds: unusedTimeSeconds,
		PricePerSecond: basePricePerSecond,
		TotalSessionSeconds: totalSessionSeconds,
	}, nil
} 