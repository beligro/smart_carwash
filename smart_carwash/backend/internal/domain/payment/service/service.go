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

// Service интерфейс для бизнес-логики платежей
type Service interface {
	CalculatePrice(req *models.CalculatePriceRequest) (*models.CalculatePriceResponse, error)
	CreatePayment(req *models.CreatePaymentRequest) (*models.CreatePaymentResponse, error)
	GetPaymentByID(paymentID uuid.UUID) (*models.Payment, error)
	GetPaymentStatus(req *models.GetPaymentStatusRequest) (*models.GetPaymentStatusResponse, error)
	RetryPayment(req *models.RetryPaymentRequest) (*models.RetryPaymentResponse, error)
	HandleWebhook(req *models.WebhookRequest) error
	ListPayments(req *models.AdminListPaymentsRequest) (*models.AdminListPaymentsResponse, error)
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

	// Обновляем статус платежа
	if req.Success {
		payment.Status = models.PaymentStatusSucceeded
	} else {
		payment.Status = models.PaymentStatusFailed
	}

	if err := s.repository.UpdatePayment(payment); err != nil {
		return fmt.Errorf("ошибка обновления статуса платежа: %w", err)
	}

	log.Printf("Обновлен статус платежа: ID=%s, Status=%s", payment.ID, payment.Status)

	// Обновляем статус связанной сессии
	if err := s.updateSessionStatus(payment); err != nil {
		log.Printf("Ошибка обновления статуса сессии: %v", err)
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