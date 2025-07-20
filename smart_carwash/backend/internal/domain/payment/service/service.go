package service

import (
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"strconv"
	"fmt"
	"log"
	"math"
	"strings"
	"time"

	"carwash_backend/internal/config"
	"carwash_backend/internal/domain/payment/models"
	"carwash_backend/internal/domain/payment/repository"
	"carwash_backend/internal/domain/payment/tinkoff"
	settingsService "carwash_backend/internal/domain/settings/service"
	sessionModels "carwash_backend/internal/domain/session/models"
	sessionService "carwash_backend/internal/domain/session/service"
	userService "carwash_backend/internal/domain/user/service"

	"github.com/google/uuid"
)

// Service интерфейс для бизнес-логики платежей
type Service interface {
	// Основные методы платежей
	CreatePayment(req *models.CreatePaymentRequest) (*models.CreatePaymentResponse, error)
	GetPayment(req *models.GetPaymentRequest) (*models.GetPaymentResponse, error)
	GetPaymentStatus(req *models.GetPaymentStatusRequest) (*models.GetPaymentStatusResponse, error)
	HandleTinkoffWebhook(webhook *models.TinkoffWebhook) error

	// Методы возвратов
	CreateRefund(req *models.CreateRefundRequest) (*models.CreateRefundResponse, error)
	ProcessAutomaticRefund(sessionID uuid.UUID) error
	ProcessFullRefund(paymentID uuid.UUID) error
	ProcessPendingRefunds() error

	// Методы для интеграции с другими доменами
	CreateQueuePayment(userID uuid.UUID, serviceType string, rentalTimeMinutes int, withChemistry bool, carNumber string) (*models.Payment, error)
	CreateSessionExtensionPayment(sessionID uuid.UUID, extensionMinutes int) (*models.Payment, error)
	CheckPaymentForSession(sessionID uuid.UUID) (*models.Payment, error)

	// Методы расчёта стоимости
	CalculatePrice(req *models.CalculatePriceRequest) (*models.CalculatePriceResponse, error)

	// Административные методы
	AdminListPayments(req *models.AdminListPaymentsRequest) (*models.AdminListPaymentsResponse, error)
	AdminGetPayment(req *models.AdminGetPaymentRequest) (*models.AdminGetPaymentResponse, error)
	ReconcilePayments(startDate, endDate time.Time) (*ReconciliationReport, error)
}

// ServiceImpl реализация Service
type ServiceImpl struct {
	repo           repository.Repository
	tinkoffClient  *tinkoff.Client
	config         *config.Config
	sessionService sessionService.Service
	userService    userService.Service
	settingsService settingsService.Service
}

// NewService создает новый экземпляр Service
func NewService(repo repository.Repository, tinkoffClient *tinkoff.Client, config *config.Config, sessionService sessionService.Service, userService userService.Service, settingsService settingsService.Service) *ServiceImpl {
	return &ServiceImpl{
		repo:           repo,
		tinkoffClient:  tinkoffClient,
		config:         config,
		sessionService: sessionService,
		userService:    userService,
		settingsService: settingsService,
	}
}

// CreatePayment создает новый платеж
func (s *ServiceImpl) CreatePayment(req *models.CreatePaymentRequest) (*models.CreatePaymentResponse, error) {
	// Проверяем идемпотентность
	existingPayment, err := s.repo.GetPaymentByIdempotencyKey(req.IdempotencyKey)
	if err == nil && existingPayment != nil {
		return &models.CreatePaymentResponse{
			Payment:    existingPayment,
			PaymentURL: existingPayment.PaymentURL,
		}, nil
	}

	// Создаем платеж в нашей БД
	payment := &models.Payment{
		UserID:         req.UserID,
		SessionID:      req.SessionID,
		AmountKopecks:  req.AmountKopecks,
		Type:           req.Type,
		Status:         models.PaymentStatusPending,
		Description:    req.Description,
		IdempotencyKey: req.IdempotencyKey,
	}

	log.Println("payment: ", payment)

	if err := s.repo.CreatePayment(payment); err != nil {
		return nil, err
	}

	log.Println("payment after create: ", payment)

	// Создаем платеж в Tinkoff
	tinkoffReq := &models.InitRequest{
		Amount:          req.AmountKopecks,
		OrderId:         payment.ID.String(),
		Description:     req.Description,
		SuccessURL:      fmt.Sprintf("%s?payment_id=%s", s.config.TelegramReturnSuccess, payment.ID.String()), // URL возврата в Telegram при успехе с ID платежа
		FailURL:         fmt.Sprintf("%s?payment_id=%s", s.config.TelegramReturnFail, payment.ID.String()),    // URL возврата в Telegram при ошибке с ID платежа
		NotificationURL: s.config.PaymentWebhookURL,
	}

	tinkoffResp, err := s.tinkoffClient.Init(tinkoffReq)
	if err != nil {
		log.Println("error: ", err)
		// Обновляем статус на "failed"
		payment.Status = models.PaymentStatusFailed
		payment.LastError = err.Error()
		s.repo.UpdatePayment(payment)
		return nil, err
	}

	// Обновляем платеж с данными от Tinkoff
	payment.TinkoffPaymentID = tinkoffResp.PaymentId
	payment.PaymentURL = tinkoffResp.PaymentURL
	payment.Status = models.PaymentStatusPending
	s.repo.UpdatePayment(payment)

	log.Printf("Tinkoff URL получен: %s", tinkoffResp.PaymentURL)

	// Создаем событие платежа
	event := &models.PaymentEvent{
		PaymentID:      payment.ID,
		EventType:      "charge",
		IdempotencyKey: fmt.Sprintf("charge_%s", payment.ID.String()),
		AmountKopecks:  req.AmountKopecks,
		Status:         "completed",
		TinkoffResponse: fmt.Sprintf(`{"PaymentId":"%s","PaymentURL":"%s"}`, tinkoffResp.PaymentId, tinkoffResp.PaymentURL),
	}
	s.repo.CreatePaymentEvent(event)

	return &models.CreatePaymentResponse{
		Payment:    payment,
		PaymentURL: tinkoffResp.PaymentURL,
	}, nil
}

// GetPayment получает платеж
func (s *ServiceImpl) GetPayment(req *models.GetPaymentRequest) (*models.GetPaymentResponse, error) {
	payment, err := s.repo.GetPaymentByID(req.ID)
	if err != nil {
		return nil, err
	}

	return &models.GetPaymentResponse{
		Payment: payment,
	}, nil
}

// GetPaymentStatus получает статус платежа
func (s *ServiceImpl) GetPaymentStatus(req *models.GetPaymentStatusRequest) (*models.GetPaymentStatusResponse, error) {
	payment, err := s.repo.GetPaymentByID(req.ID)
	if err != nil {
		return nil, err
	}

	return &models.GetPaymentStatusResponse{
		Status: payment.Status,
	}, nil
}

// HandleTinkoffWebhook обрабатывает webhook от Tinkoff
func (s *ServiceImpl) HandleTinkoffWebhook(webhook *models.TinkoffWebhook) error {
	// Проверяем, не обрабатывали ли мы уже этот webhook
	eventKey := fmt.Sprintf("webhook_%s_%s", strconv.FormatInt(webhook.PaymentId, 10), webhook.Status)
	
	existingEvent, err := s.repo.GetPaymentEventByIdempotencyKey(eventKey)
	if err == nil && existingEvent != nil {
		return nil // Уже обработали
	}

	// Находим платеж по Tinkoff PaymentId
	payment, err := s.repo.GetPaymentByTinkoffID(strconv.FormatInt(webhook.PaymentId, 10))
	if err != nil {
		return err
	}

	// Создаем событие webhook'а
	webhookResponse, _ := json.Marshal(webhook)
	event := &models.PaymentEvent{
		PaymentID:      payment.ID,
		EventType:      "webhook",
		IdempotencyKey: eventKey,
		AmountKopecks:  webhook.Amount,
		Status:         webhook.Status,
		TinkoffResponse: string(webhookResponse),
	}
	s.repo.CreatePaymentEvent(event)

	// Обновляем статус платежа
	payment.Status = models.MapTinkoffStatusToInternal(webhook.Status)
	payment.UpdatedAt = time.Now()
	
	if err := s.repo.UpdatePayment(payment); err != nil {
		return err
	}

	// Обрабатываем бизнес-логику в зависимости от статуса
	switch webhook.Status {
	case "CONFIRMED":
		return s.handlePaymentConfirmed(payment)
	case "CANCELED":
		return s.handlePaymentCanceled(payment)
	case "REFUNDED":
		return s.handlePaymentRefunded(payment)
	}

	return nil
}

// CreateRefund создает возврат средств
func (s *ServiceImpl) CreateRefund(req *models.CreateRefundRequest) (*models.CreateRefundResponse, error) {
	// Проверяем идемпотентность
	existingRefund, err := s.repo.GetRefundByIdempotencyKey(req.IdempotencyKey)
	if err == nil && existingRefund != nil {
		return &models.CreateRefundResponse{
			Refund: existingRefund,
		}, nil
	}

	// Получаем платеж
	payment, err := s.repo.GetPaymentByID(req.PaymentID)
	if err != nil {
		return nil, err
	}

	// Проверяем, что платеж можно вернуть
	if payment.Status != models.PaymentStatusCompleted {
		return nil, fmt.Errorf("платеж не может быть возвращен, статус: %s", payment.Status)
	}

	// Создаем возврат в нашей БД
	refund := &models.PaymentRefund{
		PaymentID:     req.PaymentID,
		AmountKopecks: req.AmountKopecks,
		Type:          req.Type,
		Status:        "pending",
		IdempotencyKey: req.IdempotencyKey,
		MaxRetries:    5,
	}

	if err := s.repo.CreateRefund(refund); err != nil {
		return nil, err
	}

	// Отправляем запрос в Tinkoff
	tinkoffReq := &models.RefundRequest{
		PaymentId: payment.TinkoffPaymentID,
		Amount:    req.AmountKopecks,
	}

	tinkoffResp, err := s.tinkoffClient.Refund(tinkoffReq)
	if err != nil {
		// Обновляем статус на "failed"
		refund.Status = "failed"
		refund.LastError = err.Error()
		s.repo.UpdateRefund(refund)
		return nil, err
	}

	// Обновляем возврат
	refund.Status = "completed"
	refund.TinkoffRefundID = tinkoffResp.RefundId
	s.repo.UpdateRefund(refund)

	return &models.CreateRefundResponse{
		Refund: refund,
	}, nil
}

// ProcessAutomaticRefund обрабатывает автоматический возврат при досрочном завершении
func (s *ServiceImpl) ProcessAutomaticRefund(sessionID uuid.UUID) error {
	// Получаем сессию
	session, err := s.sessionService.GetSession(&sessionModels.GetSessionRequest{SessionID: sessionID})
	if err != nil {
		return err
	}

	// Получаем платеж за сессию
	payment, err := s.repo.GetPaymentBySessionID(sessionID)
	if err != nil {
		return err // Возможно, платежа нет
	}

	// Вычисляем неиспользованное время
	unusedMinutes := s.calculateUnusedMinutes(session)
	if unusedMinutes <= 0 {
		return nil // Нечего возвращать
	}

	// Вычисляем сумму возврата
	refundAmount := s.calculateRefundAmount(payment.AmountKopecks, unusedMinutes)
	if refundAmount <= 0 {
		return nil // Нечего возвращать
	}

	// Создаем возврат
	refundReq := &models.CreateRefundRequest{
		PaymentID:      payment.ID,
		AmountKopecks:  refundAmount,
		Type:           "automatic_refund",
		IdempotencyKey: fmt.Sprintf("auto_refund_%s_%d", sessionID.String(), time.Now().Unix()),
	}

	_, err = s.CreateRefund(refundReq)
	return err
}

// ProcessFullRefund обрабатывает полный возврат
func (s *ServiceImpl) ProcessFullRefund(paymentID uuid.UUID) error {
	payment, err := s.repo.GetPaymentByID(paymentID)
	if err != nil {
		return err
	}

	// Проверяем, что платеж можно вернуть
	if payment.Status != models.PaymentStatusCompleted {
		return fmt.Errorf("платеж не может быть возвращен, статус: %s", payment.Status)
	}

	// Создаем возврат на полную сумму
	refundReq := &models.CreateRefundRequest{
		PaymentID:      paymentID,
		AmountKopecks:  payment.AmountKopecks,
		Type:           "full_refund",
		IdempotencyKey: fmt.Sprintf("full_refund_%s", paymentID.String()),
	}

	_, err = s.CreateRefund(refundReq)
	return err
}

// ProcessPendingRefunds обрабатывает ожидающие возвраты
func (s *ServiceImpl) ProcessPendingRefunds() error {
	pendingRefunds, err := s.repo.GetPendingRefunds()
	if err != nil {
		return err
	}

	for _, refund := range pendingRefunds {
		if time.Now().After(*refund.NextRetryAt) {
			s.processRefundWithRetry(&refund)
		}
	}

	return nil
}

// CreateQueuePayment создает платеж за очередь
func (s *ServiceImpl) CreateQueuePayment(userID uuid.UUID, serviceType string, rentalTimeMinutes int, withChemistry bool, carNumber string) (*models.Payment, error) {
	log.Printf("Создание платежа за очередь: userID=%s, serviceType=%s, rentalTime=%d, withChemistry=%v, carNumber=%s", 
		userID, serviceType, rentalTimeMinutes, withChemistry, carNumber)

	// Рассчитываем стоимость услуги с учетом всех параметров
	priceReq := &models.CalculatePriceRequest{
		ServiceType:       serviceType,
		RentalTimeMinutes: rentalTimeMinutes,
		WithChemistry:     withChemistry,
	}

	priceResp, err := s.CalculatePrice(priceReq)
	if err != nil {
		log.Printf("Ошибка расчета стоимости: %v", err)
		return nil, fmt.Errorf("ошибка расчета стоимости: %w", err)
	}

	log.Printf("Рассчитанная стоимость: %d копеек", priceResp.TotalPriceKopecks)

	req := &models.CreatePaymentRequest{
		UserID:         userID,
		AmountKopecks:  priceResp.TotalPriceKopecks,
		Type:           models.PaymentTypeQueueBooking,
		Description:    fmt.Sprintf("Предоплата за очередь: %s (%d мин), номер: %s", serviceType, rentalTimeMinutes, carNumber),
		IdempotencyKey: generateIdempotencyKey(),
	}

	resp, err := s.CreatePayment(req)
	if err != nil {
		log.Printf("Ошибка создания платежа: %v", err)
		return nil, err
	}

	log.Printf("Платеж успешно создан: ID=%s, Amount=%d, URL=%s", resp.Payment.ID, resp.Payment.AmountKopecks, resp.Payment.PaymentURL)
	return resp.Payment, nil
}

// CreateSessionExtensionPayment создает платеж за продление сессии
func (s *ServiceImpl) CreateSessionExtensionPayment(sessionID uuid.UUID, extensionMinutes int) (*models.Payment, error) {
	// Вычисляем стоимость продления
	extensionCost := s.calculateExtensionCost(extensionMinutes)

	req := &models.CreatePaymentRequest{
		UserID:         uuid.Nil, // Будет заполнено из сессии
		SessionID:      &sessionID,
		AmountKopecks:  extensionCost,
		Type:           models.PaymentTypeSessionExtension,
		Description:    fmt.Sprintf("Продление сессии на %d минут", extensionMinutes),
		IdempotencyKey: generateIdempotencyKey(),
	}

	resp, err := s.CreatePayment(req)
	if err != nil {
		return nil, err
	}

	return resp.Payment, nil
}

// CheckPaymentForSession проверяет платеж для сессии
func (s *ServiceImpl) CheckPaymentForSession(sessionID uuid.UUID) (*models.Payment, error) {
	return s.repo.GetPaymentBySessionID(sessionID)
}

// CalculatePrice рассчитывает стоимость услуги
func (s *ServiceImpl) CalculatePrice(req *models.CalculatePriceRequest) (*models.CalculatePriceResponse, error) {
	// Получаем базовую стоимость за минуту
	basePricePerMinute, err := s.getSettingValue(req.ServiceType, "price_per_minute")
	if err != nil {
		return nil, fmt.Errorf("не удалось получить базовую стоимость для %s: %w", req.ServiceType, err)
	}

	// Рассчитываем базовую стоимость
	basePriceKopecks := int64(basePricePerMinute) * int64(req.RentalTimeMinutes)

	// Рассчитываем стоимость химии, если требуется
	var chemistryPriceKopecks int64
	if req.WithChemistry {
		chemistryPricePerMinute, err := s.getSettingValue(req.ServiceType, "chemistry_price_per_minute")
		if err != nil {
			return nil, fmt.Errorf("не удалось получить стоимость химии для %s: %w", req.ServiceType, err)
		}
		chemistryPriceKopecks = int64(chemistryPricePerMinute) * int64(req.RentalTimeMinutes)
	}

	// Рассчитываем общую стоимость
	totalPriceKopecks := basePriceKopecks + chemistryPriceKopecks

	return &models.CalculatePriceResponse{
		TotalPriceKopecks:     totalPriceKopecks,
		BasePriceKopecks:      basePriceKopecks,
		ChemistryPriceKopecks: chemistryPriceKopecks,
		RentalTimeMinutes:     req.RentalTimeMinutes,
		ServiceType:           req.ServiceType,
		WithChemistry:         req.WithChemistry,
	}, nil
}

// AdminListPayments получает список платежей для администратора
func (s *ServiceImpl) AdminListPayments(req *models.AdminListPaymentsRequest) (*models.AdminListPaymentsResponse, error) {
	payments, total, err := s.repo.AdminListPayments(req.UserID, req.Status, req.Type, req.DateFrom, req.DateTo, req.Limit, req.Offset)
	if err != nil {
		return nil, err
	}

	return &models.AdminListPaymentsResponse{
		Payments: payments,
		Total:    total,
	}, nil
}

// AdminGetPayment получает платеж для администратора
func (s *ServiceImpl) AdminGetPayment(req *models.AdminGetPaymentRequest) (*models.AdminGetPaymentResponse, error) {
	payment, refunds, events, err := s.repo.AdminGetPaymentWithDetails(req.ID)
	if err != nil {
		return nil, err
	}

	return &models.AdminGetPaymentResponse{
		Payment: payment,
		Refunds: refunds,
		Events:  events,
	}, nil
}

// ReconciliationReport отчет о сверке платежей
type ReconciliationReport struct {
	Date           time.Time `json:"date"`
	Discrepancies  []Discrepancy `json:"discrepancies"`
	TotalPayments  int       `json:"total_payments"`
	TotalAmount    int64     `json:"total_amount"`
}

// Discrepancy расхождение в данных
type Discrepancy struct {
	PaymentID    uuid.UUID `json:"payment_id"`
	LocalStatus  string    `json:"local_status"`
	TinkoffStatus string   `json:"tinkoff_status"`
	Amount       int64     `json:"amount"`
	Description  string    `json:"description"`
}

// ReconcilePayments сверяет платежи с Tinkoff
func (s *ServiceImpl) ReconcilePayments(startDate, endDate time.Time) (*ReconciliationReport, error) {
	// Получаем локальные платежи
	localPayments, err := s.repo.GetPaymentsByDateRange(startDate, endDate)
	if err != nil {
		return nil, err
	}

	// Получаем платежи из Tinkoff
	tinkoffReq := &models.GetOperationsRequest{
		StartDate: startDate.Format("2006-01-02"),
		EndDate:   endDate.Format("2006-01-02"),
	}

	tinkoffResp, err := s.tinkoffClient.GetOperations(tinkoffReq)
	if err != nil {
		return nil, err
	}

	// Сравниваем данные
	discrepancies := s.findDiscrepancies(localPayments, tinkoffResp.Operations)

	return &ReconciliationReport{
		Date:          startDate,
		Discrepancies: discrepancies,
		TotalPayments: len(localPayments),
		TotalAmount:   s.calculateTotalAmount(localPayments),
	}, nil
}

// Вспомогательные методы

func (s *ServiceImpl) handlePaymentConfirmed(payment *models.Payment) error {
	// Если это платеж за очередь, создаем сессию
	if payment.Type == models.PaymentTypeQueueBooking {
		// Получаем данные пользователя для создания сессии
		user, err := s.userService.GetUserByID(payment.UserID)
		if err != nil {
			return fmt.Errorf("не удалось получить пользователя: %w", err)
		}

		// Извлекаем параметры из описания платежа
		// Описание имеет формат: "Предоплата за очередь: {serviceType} ({rentalTime} мин)"
		description := payment.Description
		var serviceType string
		var rentalTimeMinutes int
		var withChemistry bool

		// Парсим описание для извлечения параметров
		// Это упрощенная версия - в реальности лучше хранить параметры в отдельных полях
		if len(description) > 0 {
			// Извлекаем тип услуги из описания
			if strings.Contains(description, "wash") {
				serviceType = "wash"
			} else if strings.Contains(description, "air_dry") {
				serviceType = "air_dry"
			} else if strings.Contains(description, "vacuum") {
				serviceType = "vacuum"
			} else {
				serviceType = "wash" // по умолчанию
			}

			// Извлекаем время аренды
			if strings.Contains(description, "5 мин") {
				rentalTimeMinutes = 5
			} else if strings.Contains(description, "10 мин") {
				rentalTimeMinutes = 10
			} else if strings.Contains(description, "15 мин") {
				rentalTimeMinutes = 15
			} else if strings.Contains(description, "20 мин") {
				rentalTimeMinutes = 20
			} else {
				rentalTimeMinutes = 5 // по умолчанию
			}

			// Определяем наличие химии по сумме платежа
			// Это упрощенная логика - в реальности лучше хранить параметры отдельно
			basePrice := s.getServicePrice(serviceType) * int64(rentalTimeMinutes)
			if payment.AmountKopecks > basePrice {
				withChemistry = true
			}
		}

		// Создаем сессию в статусе in_queue
		sessionReq := &sessionModels.CreateSessionRequest{
			UserID:            payment.UserID,
			ServiceType:       serviceType,
			WithChemistry:     withChemistry,
			CarNumber:         "A000AA", // Временный номер - в реальности нужно получать от пользователя
			RentalTimeMinutes: rentalTimeMinutes,
			IdempotencyKey:    fmt.Sprintf("payment_session_%s", payment.ID.String()),
		}

		session, err := s.sessionService.CreateSession(sessionReq)
		if err != nil {
			return fmt.Errorf("не удалось создать сессию: %w", err)
		}

		// Обновляем статус сессии на in_queue
		session.Status = sessionModels.SessionStatusInQueue
		session.StatusUpdatedAt = time.Now()
		if err := s.sessionService.UpdateSession(session); err != nil {
			return fmt.Errorf("не удалось обновить статус сессии: %w", err)
		}

		// Обновляем платеж - привязываем к сессии
		payment.SessionID = &session.ID
		if err := s.repo.UpdatePayment(payment); err != nil {
			return fmt.Errorf("не удалось обновить платеж: %w", err)
		}

		log.Printf("Создана сессия в очереди: ID=%s, UserID=%s, ServiceType=%s", 
			session.ID, session.UserID, session.ServiceType)
	}

	return nil
}

func (s *ServiceImpl) handlePaymentCanceled(payment *models.Payment) error {
	// Если это платеж за очередь, создаем сессию в статусе payment_failed
	if payment.Type == models.PaymentTypeQueueBooking {
		// Получаем данные пользователя для создания сессии
		user, err := s.userService.GetUserByID(payment.UserID)
		if err != nil {
			return fmt.Errorf("не удалось получить пользователя: %w", err)
		}

		// Извлекаем параметры из описания платежа (аналогично handlePaymentConfirmed)
		description := payment.Description
		var serviceType string
		var rentalTimeMinutes int
		var withChemistry bool

		if len(description) > 0 {
			if strings.Contains(description, "wash") {
				serviceType = "wash"
			} else if strings.Contains(description, "air_dry") {
				serviceType = "air_dry"
			} else if strings.Contains(description, "vacuum") {
				serviceType = "vacuum"
			} else {
				serviceType = "wash"
			}

			if strings.Contains(description, "5 мин") {
				rentalTimeMinutes = 5
			} else if strings.Contains(description, "10 мин") {
				rentalTimeMinutes = 10
			} else if strings.Contains(description, "15 мин") {
				rentalTimeMinutes = 15
			} else if strings.Contains(description, "20 мин") {
				rentalTimeMinutes = 20
			} else {
				rentalTimeMinutes = 5
			}

			basePrice := s.getServicePrice(serviceType) * int64(rentalTimeMinutes)
			if payment.AmountKopecks > basePrice {
				withChemistry = true
			}
		}

		// Создаем сессию в статусе payment_failed
		sessionReq := &sessionModels.CreateSessionRequest{
			UserID:            payment.UserID,
			ServiceType:       serviceType,
			WithChemistry:     withChemistry,
			CarNumber:         "A000AA", // Временный номер
			RentalTimeMinutes: rentalTimeMinutes,
			IdempotencyKey:    fmt.Sprintf("payment_failed_session_%s", payment.ID.String()),
		}

		session, err := s.sessionService.CreateSession(sessionReq)
		if err != nil {
			return fmt.Errorf("не удалось создать сессию: %w", err)
		}

		// Обновляем статус сессии на payment_failed
		session.Status = sessionModels.SessionStatusPaymentFailed
		session.StatusUpdatedAt = time.Now()
		if err := s.sessionService.UpdateSession(session); err != nil {
			return fmt.Errorf("не удалось обновить статус сессии: %w", err)
		}

		// Обновляем платеж - привязываем к сессии
		payment.SessionID = &session.ID
		if err := s.repo.UpdatePayment(payment); err != nil {
			return fmt.Errorf("не удалось обновить платеж: %w", err)
		}

		log.Printf("Создана сессия с ошибкой оплаты: ID=%s, UserID=%s, ServiceType=%s", 
			session.ID, session.UserID, session.ServiceType)
	}

	return nil
}

func (s *ServiceImpl) handlePaymentRefunded(payment *models.Payment) error {
	// Логика после возврата платежа
	// Если есть связанная сессия, можно обновить её статус
	if payment.SessionID != nil {
		session, err := s.sessionService.GetSession(&sessionModels.GetSessionRequest{SessionID: *payment.SessionID})
		if err == nil && session != nil {
			// Обновляем статус сессии на canceled при возврате
			session.Status = sessionModels.SessionStatusCanceled
			session.StatusUpdatedAt = time.Now()
			s.sessionService.UpdateSession(session)
		}
	}
	return nil
}

func (s *ServiceImpl) processRefundWithRetry(refund *models.PaymentRefund) {
	// Увеличиваем счетчик попыток
	refund.RetryCount++
	
	// Вычисляем время следующей попытки
	nextRetry := s.calculateNextRetry(refund.RetryCount)
	refund.NextRetryAt = &nextRetry

	// Пытаемся выполнить возврат
	payment, err := s.repo.GetPaymentByID(refund.PaymentID)
	if err != nil {
		refund.Status = "failed"
		refund.LastError = err.Error()
		s.repo.UpdateRefund(refund)
		return
	}

	tinkoffReq := &models.RefundRequest{
		PaymentId: payment.TinkoffPaymentID,
		Amount:    refund.AmountKopecks,
	}

	tinkoffResp, err := s.tinkoffClient.Refund(tinkoffReq)
	if err != nil {
		refund.LastError = err.Error()
		if refund.RetryCount >= refund.MaxRetries {
			refund.Status = "failed"
		}
		s.repo.UpdateRefund(refund)
		return
	}

	// Успешно
	refund.Status = "completed"
	refund.TinkoffRefundID = tinkoffResp.RefundId
	s.repo.UpdateRefund(refund)
}

func (s *ServiceImpl) calculateNextRetry(retryCount int) time.Time {
	baseDelay := 30 * time.Second
	maxDelay := 24 * time.Hour
	
	delay := time.Duration(float64(baseDelay) * math.Pow(2, float64(retryCount)))
	if delay > maxDelay {
		delay = maxDelay
	}
	
	return time.Now().Add(delay)
}

func (s *ServiceImpl) getServicePrice(serviceType string) int64 {
	// Цены услуг в копейках
	prices := map[string]int64{
		"wash":     5000,  // 50 рублей
		"air_dry":  3000,  // 30 рублей
		"vacuum":   2000,  // 20 рублей
	}
	
	if price, exists := prices[serviceType]; exists {
		return price
	}
	
	return 5000 // Цена по умолчанию
}

func (s *ServiceImpl) calculateExtensionCost(extensionMinutes int) int64 {
	// 10 рублей за минуту
	return int64(extensionMinutes) * 1000
}

func (s *ServiceImpl) calculateUnusedMinutes(session *sessionModels.Session) int {
	// Логика вычисления неиспользованного времени
	// Упрощенная версия
	return 0
}

func (s *ServiceImpl) calculateRefundAmount(totalAmount int64, unusedMinutes int) int64 {
	// Логика вычисления суммы возврата
	// Упрощенная версия
	return 0
}

func (s *ServiceImpl) findDiscrepancies(localPayments []models.Payment, tinkoffOperations []models.Operation) []Discrepancy {
	// Логика поиска расхождений
	return []Discrepancy{}
}

func (s *ServiceImpl) calculateTotalAmount(payments []models.Payment) int64 {
	var total int64
	for _, payment := range payments {
		total += payment.AmountKopecks
	}
	return total
}

func generateIdempotencyKey() string {
	bytes := make([]byte, 16)
	rand.Read(bytes)
	return hex.EncodeToString(bytes)
} 

// getSettingValue получает значение настройки из service_settings
func (s *ServiceImpl) getSettingValue(serviceType, settingKey string) (int, error) {
	return s.settingsService.GetServicePrice(serviceType, settingKey)
} 