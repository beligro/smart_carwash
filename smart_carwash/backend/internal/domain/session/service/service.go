package service

import (
	"carwash_backend/internal/domain/modbus"
	paymentModels "carwash_backend/internal/domain/payment/models"
	paymentService "carwash_backend/internal/domain/payment/service"
	"carwash_backend/internal/domain/session/models"
	"carwash_backend/internal/domain/session/repository"
	settingsModels "carwash_backend/internal/domain/settings/models"
	settingsService "carwash_backend/internal/domain/settings/service"
	"carwash_backend/internal/domain/telegram"
	userService "carwash_backend/internal/domain/user/service"
	washboxModels "carwash_backend/internal/domain/washbox/models"
	washboxService "carwash_backend/internal/domain/washbox/service"
	"carwash_backend/internal/logger"
	"carwash_backend/internal/metrics"
	"carwash_backend/internal/utils"
	"errors"
	"fmt"
	"sort"
	"strconv"
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// Service интерфейс для бизнес-логики сессий
type Service interface {
	CreateSession(req *models.CreateSessionRequest) (*models.Session, error)
	CreateSessionWithPayment(req *models.CreateSessionWithPaymentRequest) (*models.CreateSessionWithPaymentResponse, error)
	GetUserSession(req *models.GetUserSessionRequest) (*models.GetUserSessionResponse, error)
	GetUserSessionForPayment(req *models.GetUserSessionRequest) (*models.GetUserSessionResponse, error)
	CheckActiveSession(req *models.CheckActiveSessionRequest) (*models.CheckActiveSessionResponse, error)
	GetSession(req *models.GetSessionRequest) (*models.GetSessionResponse, error)
	StartSession(req *models.StartSessionRequest) (*models.Session, error)
	CompleteSession(req *models.CompleteSessionRequest) (*models.CompleteSessionResponse, error)
	ExtendSessionWithPayment(req *models.ExtendSessionWithPaymentRequest) (*models.ExtendSessionWithPaymentResponse, error)
	GetSessionPayments(req *models.GetSessionPaymentsRequest) (*models.GetSessionPaymentsResponse, error)
	CancelSession(req *models.CancelSessionRequest) (*models.CancelSessionResponse, error)
	UpdateSessionStatus(sessionID uuid.UUID, status string) error
	ProcessQueue() error
	CheckAndCompleteExpiredSessions() error
	CheckAndExpireReservedSessions() error
	CheckAndNotifyExpiringReservedSessions() error
	CheckAndNotifyCompletingSessions() error
	CountSessionsByStatus(status string) (int, error)
	GetSessionsByStatus(status string) ([]models.Session, error)
	GetUserSessionHistory(req *models.GetUserSessionHistoryRequest) ([]models.Session, error)
	CreateFromCashier(req *models.CashierPaymentRequest) (*models.Session, error)
	GetActiveSessionByCarNumber(carNumber string) (*models.Session, error)

	// Административные методы
	AdminListSessions(req *models.AdminListSessionsRequest) (*models.AdminListSessionsResponse, error)
	AdminGetSession(req *models.AdminGetSessionRequest) (*models.AdminGetSessionResponse, error)

	// Методы для кассира
	CashierListSessions(req *models.CashierSessionsRequest) (*models.AdminListSessionsResponse, error)
	CashierGetActiveSessions(req *models.CashierActiveSessionsRequest) (*models.CashierActiveSessionsResponse, error)
	CashierStartSession(req *models.CashierStartSessionRequest) (*models.Session, error)
	CashierCompleteSession(req *models.CashierCompleteSessionRequest) (*models.Session, error)
	CashierCancelSession(req *models.CashierCancelSessionRequest) (*models.Session, error)

	// Методы для химии
	EnableChemistry(req *models.EnableChemistryRequest) (*models.EnableChemistryResponse, error)
	GetChemistryStats(req *models.GetChemistryStatsRequest) (*models.GetChemistryStatsResponse, error)

	// Методы для переназначения сессий
	ReassignSession(req *models.ReassignSessionRequest) (*models.ReassignSessionResponse, error)

	// Методы для Dahua интеграции
	GetActiveSessionByUserID(userID uuid.UUID) (*models.Session, error)
	CompleteSessionWithoutRefund(sessionID uuid.UUID) error
}

// ServiceImpl реализация Service
type ServiceImpl struct {
	repo            repository.Repository
	washboxService  washboxService.Service
	userService     userService.Service
	telegramBot     telegram.NotificationService
	paymentService  paymentService.Service
	modbusService   modbus.ModbusServiceInterface
	settingsService settingsService.Service
	cashierUserID   string
	metrics         *metrics.Metrics
}

// NewService создает новый экземпляр Service
func NewService(repo repository.Repository, washboxService washboxService.Service, userService userService.Service, telegramBot telegram.NotificationService, paymentService paymentService.Service, modbusService modbus.ModbusServiceInterface, settingsService settingsService.Service, cashierUserID string, metrics *metrics.Metrics) *ServiceImpl {
	return &ServiceImpl{
		repo:            repo,
		washboxService:  washboxService,
		userService:     userService,
		telegramBot:     telegramBot,
		paymentService:  paymentService,
		modbusService:   modbusService,
		settingsService: settingsService,
		cashierUserID:   cashierUserID,
		metrics:         metrics,
	}
}

// CreateSession создает новую сессию
func (s *ServiceImpl) CreateSession(req *models.CreateSessionRequest) (*models.Session, error) {
	logger.Printf("Service - CreateSession: начало создания сессии, user_id: %s, service_type: %s, with_chemistry: %t", req.UserID.String(), req.ServiceType, req.WithChemistry)

	// Нормализация госномера (без валидации)
	normalizedCarNumber := utils.NormalizeLicensePlate(req.CarNumber)
	logger.Printf("Service - CreateSession: госномер нормализован '%s' -> '%s', user_id: %s", req.CarNumber, normalizedCarNumber, req.UserID.String())

	// Валидация химии
	if req.WithChemistry {
		switch req.ServiceType {
		case "wash":
			// Химия разрешена для мойки
			logger.Printf("Service - CreateSession: химия разрешена для мойки, user_id: %s", req.UserID.String())

			// Валидация времени химии
			if req.ChemistryTimeMinutes <= 0 {
				logger.Printf("Service - CreateSession: не указано время химии, user_id: %s", req.UserID.String())
				return nil, fmt.Errorf("chemistry_time_minutes is required when with_chemistry is true")
			}

			// Проверяем, что выбранное время химии находится в списке доступных
			availableTimesReq := &settingsModels.GetAvailableChemistryTimesRequest{
				ServiceType: req.ServiceType,
			}
			availableTimesResp, err := s.settingsService.GetAvailableChemistryTimes(availableTimesReq)
			if err != nil {
				logger.Printf("Service - CreateSession: ошибка получения доступного времени химии: %v, user_id: %s", err, req.UserID.String())
				return nil, fmt.Errorf("не удалось получить доступное время химии: %w", err)
			}

			// Проверяем, что выбранное время есть в списке доступных
			isValidTime := false
			for _, availableTime := range availableTimesResp.AvailableChemistryTimes {
				if availableTime == req.ChemistryTimeMinutes {
					isValidTime = true
					break
				}
			}
			if !isValidTime {
				logger.Printf("Service - CreateSession: недопустимое время химии %d минут, user_id: %s", req.ChemistryTimeMinutes, req.UserID.String())
				return nil, fmt.Errorf("недопустимое время химии: %d минут", req.ChemistryTimeMinutes)
			}

		case "air_dry", "vacuum":
			logger.Printf("Service - CreateSession: химия не разрешена для типа услуги '%s', user_id: %s", req.ServiceType, req.UserID.String())
			return nil, fmt.Errorf("chemistry is not available for service type: %s", req.ServiceType)
		default:
			logger.Printf("Service - CreateSession: неверный тип услуги '%s', user_id: %s", req.ServiceType, req.UserID.String())
			return nil, fmt.Errorf("invalid service type: %s", req.ServiceType)
		}
	}

	// Проверяем идемпотентность запроса
	existingSessionByKey, err := s.repo.GetSessionByIdempotencyKey(req.IdempotencyKey)
	if err == nil && existingSessionByKey != nil {
		// Сессия с таким ключом идемпотентности уже существует
		logger.Printf("Service - CreateSession: найдена существующая сессия по ключу идемпотентности, session_id: %s, user_id: %s", existingSessionByKey.ID.String(), req.UserID.String())
		return existingSessionByKey, nil
	}

	// Проверяем, есть ли у пользователя активная сессия
	existingSession, err := s.repo.GetActiveSessionByUserID(req.UserID)
	if err == nil && existingSession != nil {
		// У пользователя уже есть активная сессия
		logger.Printf("Service - CreateSession: у пользователя уже есть активная сессия, session_id: %s, user_id: %s, status: %s", existingSession.ID.String(), req.UserID.String(), existingSession.Status)

		// Дополнительная проверка: если сессия создана недавно (в последние 30 секунд),
		// это может быть попытка создания множественных сессий
		now := time.Now()
		if now.Sub(existingSession.CreatedAt) < 30*time.Second {
			logger.Printf("Service - CreateSession: обнаружена попытка создания множественных сессий, session_id: %s, user_id: %s, created_at: %s", existingSession.ID.String(), req.UserID.String(), existingSession.CreatedAt.Format(time.RFC3339))

			// Записываем метрику попытки создания множественных сессий
			if s.metrics != nil {
				timeDiff := strconv.FormatFloat(now.Sub(existingSession.CreatedAt).Seconds(), 'f', 0, 64)
				s.metrics.RecordMultipleSession("attempt", req.UserID.String(), timeDiff)
			}
		}

		// Возвращаем ошибку вместо существующей сессии
		return nil, fmt.Errorf("у вас уже есть активная сессия")
	}

	// Создаем новую сессию
	now := time.Now()
	session := &models.Session{
		UserID:               req.UserID,
		Status:               models.SessionStatusCreated,
		ServiceType:          req.ServiceType,
		WithChemistry:        req.WithChemistry,
		ChemistryTimeMinutes: req.ChemistryTimeMinutes, // Сохраняем выбранное время химии
		CarNumber:            normalizedCarNumber,      // Используем нормализованный госномер
		CarNumberCountry:     req.CarNumberCountry,     // Сохраняем страну гос номера
		Email:                req.Email,                // Сохраняем email для чека
		RentalTimeMinutes:    req.RentalTimeMinutes,
		IdempotencyKey:       req.IdempotencyKey,
		StatusUpdatedAt:      now, // Инициализируем время изменения статуса
	}

	// Сохраняем сессию в базе данных
	err = s.repo.CreateSession(session)
	if err != nil {
		logger.Printf("Service - CreateSession: ошибка сохранения сессии в БД, user_id: %s, error: %v", req.UserID.String(), err)
		return nil, err
	}

	logger.Printf("Service - CreateSession: сессия успешно создана, session_id: %s, user_id: %s, service_type: %s", session.ID.String(), req.UserID.String(), req.ServiceType)

	// Записываем метрику создания сессии
	if s.metrics != nil {
		s.metrics.RecordSession("created", req.ServiceType, strconv.FormatBool(req.WithChemistry))
	}

	return session, nil
}

// CreateSessionWithPayment создает сессию с платежом
func (s *ServiceImpl) CreateSessionWithPayment(req *models.CreateSessionWithPaymentRequest) (*models.CreateSessionWithPaymentResponse, error) {
	logger.Printf("Service - CreateSessionWithPayment: начало создания сессии с платежом, user_id: %s, service_type: %s", req.UserID.String(), req.ServiceType)

	// 1. Создаем сессию
	session, err := s.CreateSession(&models.CreateSessionRequest{
		UserID:               req.UserID,
		ServiceType:          req.ServiceType,
		WithChemistry:        req.WithChemistry,
		ChemistryTimeMinutes: req.ChemistryTimeMinutes,
		CarNumber:            req.CarNumber,
		RentalTimeMinutes:    req.RentalTimeMinutes,
		IdempotencyKey:       req.IdempotencyKey,
	})
	if err != nil {
		logger.Printf("Service - CreateSessionWithPayment: ошибка создания сессии, user_id: %s, error: %v", req.UserID.String(), err)
		return nil, fmt.Errorf("ошибка создания сессии: %w", err)
	}

	// 2. Рассчитываем цену через Payment Service
	priceResp, err := s.paymentService.CalculatePrice(&paymentModels.CalculatePriceRequest{
		ServiceType:          req.ServiceType,
		WithChemistry:        req.WithChemistry,
		ChemistryTimeMinutes: req.ChemistryTimeMinutes,
		RentalTimeMinutes:    req.RentalTimeMinutes,
	})
	if err != nil {
		logger.Printf("Service - CreateSessionWithPayment: ошибка расчета цены, session_id: %s, error: %v", session.ID.String(), err)
		return nil, fmt.Errorf("ошибка расчета цены: %w", err)
	}

	logger.Printf("Service - CreateSessionWithPayment: цена рассчитана, session_id: %s, price: %d %s", session.ID.String(), priceResp.Price, priceResp.Currency)

	// 3. Создаем платеж через Payment Service
	paymentResp, err := s.paymentService.CreatePayment(&paymentModels.CreatePaymentRequest{
		SessionID: session.ID,
		Amount:    priceResp.Price,
		Currency:  priceResp.Currency,
		Email:     session.Email, // Передаем email из сессии
	})
	if err != nil {
		return nil, fmt.Errorf("ошибка создания платежа: %w", err)
	}

	// 4. Формируем ответ с информацией о платеже
	payment := &models.Payment{
		ID:         paymentResp.Payment.ID,
		SessionID:  paymentResp.Payment.SessionID,
		Amount:     paymentResp.Payment.Amount,
		Currency:   paymentResp.Payment.Currency,
		Status:     paymentResp.Payment.Status,
		PaymentURL: paymentResp.Payment.PaymentURL,
		TinkoffID:  paymentResp.Payment.TinkoffID,
		ExpiresAt:  paymentResp.Payment.ExpiresAt,
		CreatedAt:  paymentResp.Payment.CreatedAt,
		UpdatedAt:  paymentResp.Payment.UpdatedAt,
	}

	return &models.CreateSessionWithPaymentResponse{
		Session: *session,
		Payment: payment,
	}, nil
}

// CheckActiveSession проверяет активную сессию пользователя с учетом временной блокировки
func (s *ServiceImpl) CheckActiveSession(req *models.CheckActiveSessionRequest) (*models.CheckActiveSessionResponse, error) {
	logger.Printf("Service - CheckActiveSession: проверка активной сессии, user_id: %s", req.UserID.String())

	// Проверяем активную сессию
	session, err := s.repo.GetActiveSessionByUserID(req.UserID)
	if err != nil {
		// Если сессии нет, это нормально
		if errors.Is(err, gorm.ErrRecordNotFound) {
			logger.Printf("Service - CheckActiveSession: активных сессий не найдено, user_id: %s", req.UserID.String())
			return &models.CheckActiveSessionResponse{
				HasActiveSession: false,
				Session:          nil,
			}, nil
		}
		return nil, err
	}

	// Если сессия найдена, получаем информацию о платеже
	var payment *models.Payment
	if session != nil {
		paymentResp, err := s.paymentService.GetMainPaymentBySessionID(session.ID)
		if err == nil && paymentResp != nil {
			payment = &models.Payment{
				ID:             paymentResp.ID,
				SessionID:      paymentResp.SessionID,
				Amount:         paymentResp.Amount,
				RefundedAmount: paymentResp.RefundedAmount,
				Currency:       paymentResp.Currency,
				Status:         paymentResp.Status,
				PaymentType:    paymentResp.PaymentType,
				PaymentURL:     paymentResp.PaymentURL,
				TinkoffID:      paymentResp.TinkoffID,
				ExpiresAt:      paymentResp.ExpiresAt,
				CreatedAt:      paymentResp.CreatedAt,
				UpdatedAt:      paymentResp.UpdatedAt,
			}
		}
	}

	logger.Printf("Service - CheckActiveSession: найдена активная сессия, session_id: %s, user_id: %s, status: %s", session.ID.String(), req.UserID.String(), session.Status)

	return &models.CheckActiveSessionResponse{
		HasActiveSession: true,
		Session:          session,
		Payment:          payment,
	}, nil
}

// GetUserSession получает активную сессию пользователя
func (s *ServiceImpl) GetUserSession(req *models.GetUserSessionRequest) (*models.GetUserSessionResponse, error) {
	session, err := s.repo.GetActiveSessionByUserID(req.UserID)
	if err != nil {
		return nil, err
	}

	// Получаем основной платеж сессии, если сессия существует
	var payment *models.Payment
	if session != nil {
		paymentResp, err := s.paymentService.GetMainPaymentBySessionID(session.ID)
		if err == nil && paymentResp != nil {
			payment = &models.Payment{
				ID:             paymentResp.ID,
				SessionID:      paymentResp.SessionID,
				Amount:         paymentResp.Amount,
				RefundedAmount: paymentResp.RefundedAmount,
				Currency:       paymentResp.Currency,
				Status:         paymentResp.Status,
				PaymentType:    paymentResp.PaymentType,
				PaymentURL:     paymentResp.PaymentURL,
				TinkoffID:      paymentResp.TinkoffID,
				ExpiresAt:      paymentResp.ExpiresAt,
				RefundedAt:     paymentResp.RefundedAt,
				CreatedAt:      paymentResp.CreatedAt,
				UpdatedAt:      paymentResp.UpdatedAt,
			}
		}
	}

	// Заполняем session_timeout_minutes из настроек
	if session != nil {
		sessionTimeout, err := s.settingsService.GetSessionTimeout()
		if err != nil {
			// Если не удалось получить настройку, используем значение по умолчанию
			sessionTimeout = 3
		}
		session.SessionTimeoutMinutes = sessionTimeout
	}

	return &models.GetUserSessionResponse{
		Session: session,
		Payment: payment,
	}, nil
}

// GetUserSessionForPayment получает сессию пользователя для PaymentPage (включая payment_failed)
func (s *ServiceImpl) GetUserSessionForPayment(req *models.GetUserSessionRequest) (*models.GetUserSessionResponse, error) {
	session, err := s.repo.GetUserSessionForPayment(req.UserID)
	if err != nil {
		return nil, err
	}

	// Получаем последний платеж сессии, если сессия существует
	var payment *models.Payment
	if session != nil {
		paymentResp, err := s.paymentService.GetLastPaymentBySessionID(session.ID)
		if err == nil && paymentResp != nil {
			payment = &models.Payment{
				ID:             paymentResp.ID,
				SessionID:      paymentResp.SessionID,
				Amount:         paymentResp.Amount,
				RefundedAmount: paymentResp.RefundedAmount,
				Currency:       paymentResp.Currency,
				Status:         paymentResp.Status,
				PaymentType:    paymentResp.PaymentType,
				PaymentURL:     paymentResp.PaymentURL,
				TinkoffID:      paymentResp.TinkoffID,
				ExpiresAt:      paymentResp.ExpiresAt,
				RefundedAt:     paymentResp.RefundedAt,
				CreatedAt:      paymentResp.CreatedAt,
				UpdatedAt:      paymentResp.UpdatedAt,
			}
		}
	}

	// Заполняем session_timeout_minutes из настроек
	if session != nil {
		sessionTimeout, err := s.settingsService.GetSessionTimeout()
		if err != nil {
			// Если не удалось получить настройку, используем значение по умолчанию
			sessionTimeout = 3
		}
		session.SessionTimeoutMinutes = sessionTimeout
	}

	return &models.GetUserSessionResponse{
		Session: session,
		Payment: payment,
	}, nil
}

// GetSession получает сессию по ID
func (s *ServiceImpl) GetSession(req *models.GetSessionRequest) (*models.GetSessionResponse, error) {
	session, err := s.repo.GetSessionByID(req.SessionID)
	if err != nil {
		return nil, err
	}

	// Получаем основной платеж сессии, если сессия существует
	var payment *models.Payment
	if session != nil {
		paymentResp, err := s.paymentService.GetMainPaymentBySessionID(session.ID)
		if err == nil && paymentResp != nil {
			payment = &models.Payment{
				ID:             paymentResp.ID,
				SessionID:      paymentResp.SessionID,
				Amount:         paymentResp.Amount,
				RefundedAmount: paymentResp.RefundedAmount,
				Currency:       paymentResp.Currency,
				Status:         paymentResp.Status,
				PaymentType:    paymentResp.PaymentType,
				PaymentURL:     paymentResp.PaymentURL,
				TinkoffID:      paymentResp.TinkoffID,
				ExpiresAt:      paymentResp.ExpiresAt,
				RefundedAt:     paymentResp.RefundedAt,
				CreatedAt:      paymentResp.CreatedAt,
				UpdatedAt:      paymentResp.UpdatedAt,
			}
		}
	}

	// Заполняем session_timeout_minutes из настроек
	if session != nil {
		sessionTimeout, err := s.settingsService.GetSessionTimeout()
		if err != nil {
			// Если не удалось получить настройку, используем значение по умолчанию
			sessionTimeout = 3
		}
		session.SessionTimeoutMinutes = sessionTimeout
	}

	return &models.GetSessionResponse{
		Session: session,
		Payment: payment,
	}, nil
}

// StartSession запускает сессию (переводит в статус active)
func (s *ServiceImpl) StartSession(req *models.StartSessionRequest) (*models.Session, error) {
	// Получаем сессию по ID
	session, err := s.repo.GetSessionByID(req.SessionID)
	if err != nil {
		return nil, err
	}

	// Проверяем, что сессия в статусе assigned
	if session.Status != models.SessionStatusAssigned {
		logger.Printf("StartSessionError: сессия не в статусе assigned %s", session.ID)
		return session, nil // Возвращаем сессию без изменений
	}

	// Проверяем, что у сессии есть назначенный бокс
	if session.BoxID == nil {
		logger.Printf("StartSessionError: сессия не имеет назначенного бокса %s", session.ID)
		return session, nil // Возвращаем сессию без изменений
	}

	// Если сервис боксов не инициализирован, просто обновляем статус сессии
	if s.washboxService == nil {
		// Обновляем статус сессии на active
		session.Status = models.SessionStatusActive
		err = s.repo.UpdateSession(session)
		if err != nil {
			logger.Printf("StartSessionError: ошибка обновления статуса сессии %s", session.ID)
			return nil, err
		}
		return session, nil
	}

	// Получаем информацию о боксе
	box, err := s.washboxService.GetWashBoxByID(*session.BoxID)
	if err != nil {
		logger.Printf("StartSessionError: ошибка получения информации о боксе %s", *session.BoxID)
		return nil, err
	}

	// Обновляем статус бокса на busy
	err = s.washboxService.UpdateWashBoxStatus(*session.BoxID, washboxModels.StatusBusy)
	if err != nil {
		logger.Printf("StartSessionError: ошибка обновления статуса бокса %s", *session.BoxID)
		return nil, err
	}

	// Обновляем статус сессии на active, время обновления статуса и сбрасываем флаг уведомления
	session.Status = models.SessionStatusActive
	session.StatusUpdatedAt = time.Now()       // Обновляем время изменения статуса
	session.IsExpiringNotificationSent = false // Сбрасываем флаг, чтобы уведомление могло быть отправлено снова
	err = s.repo.UpdateSession(session)
	if err != nil {
		// Если не удалось обновить сессию, возвращаем статус бокса обратно
		s.washboxService.UpdateWashBoxStatus(*session.BoxID, box.Status)
		logger.Printf("StartSessionError: ошибка обновления статуса сессии %s", session.ID)
		return nil, err
	}

	// Включаем свет в боксе через Modbus
	if s.modbusService != nil {
		// Используем уже полученную информацию о боксе
		if box.LightCoilRegister != nil {
			err = s.modbusService.WriteLightCoil(*session.BoxID, *box.LightCoilRegister, true)
			if err != nil {
				// Логируем ошибку (HandleModbusError теперь только логирует, не продлевает время)
				s.modbusService.HandleModbusError(*session.BoxID, "light_on", session.ID, err)
				logger.Printf("StartSessionError: ошибка включения света в боксе %s: %v", *session.BoxID, err)
			}
		} else {
			logger.Printf("StartSessionError: не найден регистр света для бокса %s", *session.BoxID)
		}
	} else {
		logger.Printf("StartSessionError: ModbusService не инициализирован, не включаем свет в боксе %s", *session.BoxID)
	}

	// Получаем основной платеж сессии
	paymentResp, err := s.paymentService.GetMainPaymentBySessionID(session.ID)
	if err == nil && paymentResp != nil {
		session.Payment = &models.Payment{
			ID:             paymentResp.ID,
			SessionID:      paymentResp.SessionID,
			Amount:         paymentResp.Amount,
			RefundedAmount: paymentResp.RefundedAmount,
			Currency:       paymentResp.Currency,
			Status:         paymentResp.Status,
			PaymentType:    paymentResp.PaymentType,
			PaymentURL:     paymentResp.PaymentURL,
			TinkoffID:      paymentResp.TinkoffID,
			ExpiresAt:      paymentResp.ExpiresAt,
			RefundedAt:     paymentResp.RefundedAt,
			CreatedAt:      paymentResp.CreatedAt,
			UpdatedAt:      paymentResp.UpdatedAt,
		}
	}

	return session, nil
}

// CompleteSession завершает сессию (переводит в статус complete) с возможным частичным возвратом
func (s *ServiceImpl) CompleteSession(req *models.CompleteSessionRequest) (*models.CompleteSessionResponse, error) {
	// Получаем сессию по ID
	session, err := s.repo.GetSessionByID(req.SessionID)
	if err != nil {
		return nil, err
	}

	// Проверяем, что сессия в статусе active
	if session.Status != models.SessionStatusActive {
		return &models.CompleteSessionResponse{Session: session}, nil // Возвращаем сессию без изменений
	}

	// Проверяем, что у сессии есть назначенный бокс
	if session.BoxID == nil {
		return &models.CompleteSessionResponse{Session: session}, nil // Возвращаем сессию без изменений
	}

	// Если сервис боксов не инициализирован, просто обновляем статус сессии
	if s.washboxService == nil {
		// Обновляем статус сессии на complete
		session.Status = models.SessionStatusComplete
		err = s.repo.UpdateSession(session)
		if err != nil {
			return nil, err
		}
		return &models.CompleteSessionResponse{Session: session}, nil
	}

	// Проверяем, есть ли резерв уборки для этого бокса
	box, err := s.washboxService.GetWashBoxByID(*session.BoxID)
	if err != nil {
		return nil, err
	}

	// Если есть резерв уборки, переводим бокс в статус cleaning
	if box.CleaningReservedBy != nil {
		err = s.washboxService.UpdateWashBoxStatus(*session.BoxID, washboxModels.StatusCleaning)
		if err != nil {
			return nil, err
		}
		// Устанавливаем время начала уборки
		now := time.Now()
		err = s.washboxService.UpdateCleaningStartedAt(*session.BoxID, now)
		if err != nil {
			return nil, err
		}
		// Очищаем резерв
		err = s.washboxService.ClearCleaningReservation(*session.BoxID)
		if err != nil {
			return nil, err
		}
	} else {
		// Если нет резерва уборки, переводим в статус free
		err = s.washboxService.UpdateWashBoxStatus(*session.BoxID, washboxModels.StatusFree)
		if err != nil {
			return nil, err
		}
	}

	// Рассчитываем использованное время сессии в секундах
	startTime := session.StatusUpdatedAt
	now := time.Now()
	usedTimeSeconds := int(now.Sub(startTime).Seconds())

	// Обновляем статус сессии на complete, время обновления статуса и сбрасываем флаг уведомления
	session.Status = models.SessionStatusComplete
	session.StatusUpdatedAt = time.Now()         // Обновляем время изменения статуса
	session.IsCompletingNotificationSent = false // Сбрасываем флаг, чтобы уведомление могло быть отправлено снова
	err = s.repo.UpdateSession(session)
	if err != nil {
		return nil, err
	}

	// Если химия была включена, но еще не выключена - выключаем ее
	if session.ChemistryStartedAt != nil && session.ChemistryEndedAt == nil {
		now := time.Now()

		// Выключаем химию через Modbus
		if s.modbusService != nil {
			// Используем уже полученную информацию о боксе
			if box.ChemistryCoilRegister != nil {
				err = s.modbusService.WriteChemistryCoil(*session.BoxID, *box.ChemistryCoilRegister, false)
				if err != nil {
					logger.Printf("CompleteSession: ошибка выключения химии в боксе %s: %v", *session.BoxID, err)
				} else {
					logger.Printf("CompleteSession: химия выключена в боксе %s, SessionID=%s", *session.BoxID, session.ID)
				}
			} else {
				logger.Printf("CompleteSession: не найден регистр химии для бокса %s", *session.BoxID)
			}
		}

		// Обновляем только поле chemistry_ended_at
		err = s.repo.UpdateSessionFields(session.ID, map[string]interface{}{
			"chemistry_ended_at": now,
			"updated_at":         now,
		})
		if err != nil {
			logger.Printf("CompleteSession: ошибка обновления времени выключения химии: %v", err)
		}
	}

	// Выключаем свет в боксе через Modbus
	if s.modbusService != nil {
		// Используем уже полученную информацию о боксе
		if box.LightCoilRegister != nil {
			err = s.modbusService.WriteLightCoil(*session.BoxID, *box.LightCoilRegister, false)
			if err != nil {
				logger.Printf("Ошибка выключения света в боксе %s: %v", *session.BoxID, err)
			}
		} else {
			logger.Printf("CompleteSession: не найден регистр света для бокса %s", *session.BoxID)
		}
	}

	logger.Printf("Завершение сессии: SessionID=%s, RentalTime=%dmin, ExtensionTime=%dmin, UsedTime=%ds",
		session.ID, session.RentalTimeMinutes, session.ExtensionTimeMinutes, usedTimeSeconds)

	// Получаем обновленную информацию о платежах для отображения
	paymentsResp, err := s.paymentService.GetPaymentsBySessionID(session.ID)
	if err == nil && paymentsResp != nil {
		// Используем основной платеж для отображения в сессии
		if paymentsResp.MainPayment != nil {
			session.Payment = &models.Payment{
				ID:             paymentsResp.MainPayment.ID,
				SessionID:      paymentsResp.MainPayment.SessionID,
				Amount:         paymentsResp.MainPayment.Amount,
				RefundedAmount: paymentsResp.MainPayment.RefundedAmount,
				Currency:       paymentsResp.MainPayment.Currency,
				Status:         paymentsResp.MainPayment.Status,
				PaymentType:    paymentsResp.MainPayment.PaymentType,
				PaymentURL:     paymentsResp.MainPayment.PaymentURL,
				TinkoffID:      paymentsResp.MainPayment.TinkoffID,
				ExpiresAt:      paymentsResp.MainPayment.ExpiresAt,
				RefundedAt:     paymentsResp.MainPayment.RefundedAt,
				CreatedAt:      paymentsResp.MainPayment.CreatedAt,
				UpdatedAt:      paymentsResp.MainPayment.UpdatedAt,
			}
		}
	}

	// Записываем метрику завершения сессии
	if s.metrics != nil {
		s.metrics.RecordSession("completed", session.ServiceType, strconv.FormatBool(session.WithChemistry))
	}

	// Отправляем уведомление о завершении сессии
	if s.telegramBot != nil {
		user, err := s.userService.GetUserByID(session.UserID)
		if err == nil && user != nil {
			err = s.telegramBot.SendSessionNotification(user.TelegramID, telegram.NotificationTypeSessionCompleted, nil)
			if err != nil {
				logger.Printf("CompleteSession: ошибка отправки уведомления о завершении сессии: %v", err)
			} else {
				logger.Printf("CompleteSession: уведомление о завершении сессии отправлено пользователю %d, SessionID=%s", user.TelegramID, session.ID)
			}
		} else {
			logger.Printf("CompleteSession: не удалось получить данные пользователя для отправки уведомления: %v", err)
		}
	}

	return &models.CompleteSessionResponse{
		Session: session,
		Payment: session.Payment,
	}, nil
}

// CompleteSessionWithoutRefund завершает сессию БЕЗ частичного возврата (для Dahua webhook)
func (s *ServiceImpl) CompleteSessionWithoutRefund(sessionID uuid.UUID) error {
	// Получаем сессию по ID
	session, err := s.repo.GetSessionByID(sessionID)
	if err != nil {
		return err
	}

	// Проверяем, что сессия в статусе active
	if session.Status != models.SessionStatusActive {
		return nil // Возвращаем nil, если сессия уже не активна
	}

	// Проверяем, что у сессии есть назначенный бокс
	if session.BoxID == nil {
		return nil // Возвращаем nil, если бокс не назначен
	}

	// Если сервис боксов не инициализирован, просто обновляем статус сессии
	if s.washboxService == nil {
		// Обновляем статус сессии на complete
		session.Status = models.SessionStatusComplete
		err = s.repo.UpdateSession(session)
		if err != nil {
			return err
		}
		return nil
	}

	// Проверяем, есть ли резерв уборки для этого бокса
	box, err := s.washboxService.GetWashBoxByID(*session.BoxID)
	if err != nil {
		return err
	}

	// Если есть резерв уборки, переводим бокс в статус cleaning
	if box.CleaningReservedBy != nil {
		err = s.washboxService.UpdateWashBoxStatus(*session.BoxID, washboxModels.StatusCleaning)
		if err != nil {
			return err
		}
		// Устанавливаем время начала уборки
		now := time.Now()
		err = s.washboxService.UpdateCleaningStartedAt(*session.BoxID, now)
		if err != nil {
			return err
		}
		// Очищаем резерв
		err = s.washboxService.ClearCleaningReservation(*session.BoxID)
		if err != nil {
			return err
		}
	} else {
		// Если нет резерва уборки, переводим в статус free
		err = s.washboxService.UpdateWashBoxStatus(*session.BoxID, washboxModels.StatusFree)
		if err != nil {
			return err
		}
	}

	// Обновляем статус сессии на complete, время обновления статуса и сбрасываем флаг уведомления
	session.Status = models.SessionStatusComplete
	session.StatusUpdatedAt = time.Now()         // Обновляем время изменения статуса
	session.IsCompletingNotificationSent = false // Сбрасываем флаг, чтобы уведомление могло быть отправлено снова
	err = s.repo.UpdateSession(session)
	if err != nil {
		return err
	}

	// Если химия была включена, но еще не выключена - выключаем ее
	if session.ChemistryStartedAt != nil && session.ChemistryEndedAt == nil {
		now := time.Now()

		// Выключаем химию через Modbus
		if s.modbusService != nil {
			// Используем уже полученную информацию о боксе
			if box.ChemistryCoilRegister != nil {
				err = s.modbusService.WriteChemistryCoil(*session.BoxID, *box.ChemistryCoilRegister, false)
				if err != nil {
					logger.Printf("CompleteSessionWithoutRefund: ошибка выключения химии в боксе %s: %v", *session.BoxID, err)
				} else {
					logger.Printf("CompleteSessionWithoutRefund: химия выключена в боксе %s, SessionID=%s", *session.BoxID, session.ID)
				}
			} else {
				logger.Printf("CompleteSessionWithoutRefund: не найден регистр химии для бокса %s", *session.BoxID)
			}
		}

		// Обновляем только поле chemistry_ended_at
		err = s.repo.UpdateSessionFields(session.ID, map[string]interface{}{
			"chemistry_ended_at": now,
			"updated_at":         now,
		})
		if err != nil {
			logger.Printf("CompleteSessionWithoutRefund: ошибка обновления времени выключения химии: %v", err)
		}
	}

	// Выключаем свет в боксе через Modbus
	if s.modbusService != nil {
		// Используем уже полученную информацию о боксе
		if box.LightCoilRegister != nil {
			err = s.modbusService.WriteLightCoil(*session.BoxID, *box.LightCoilRegister, false)
			if err != nil {
				logger.Printf("CompleteSessionWithoutRefund: ошибка выключения света в боксе %s: %v", *session.BoxID, err)
			} else {
				logger.Printf("CompleteSessionWithoutRefund: свет выключен в боксе %s, SessionID=%s", *session.BoxID, session.ID)
			}
		} else {
			logger.Printf("CompleteSessionWithoutRefund: не найден регистр света для бокса %s", *session.BoxID)
		}
	}

	// Записываем метрику завершения сессии
	if s.metrics != nil {
		s.metrics.RecordSession("completed", session.ServiceType, strconv.FormatBool(session.WithChemistry))
	}

	// Отправляем уведомление о завершении сессии
	if s.telegramBot != nil {
		user, err := s.userService.GetUserByID(session.UserID)
		if err == nil && user != nil {
			err = s.telegramBot.SendSessionNotification(user.TelegramID, telegram.NotificationTypeSessionCompleted, nil)
			if err != nil {
				logger.Printf("CompleteSessionWithoutRefund: ошибка отправки уведомления о завершении сессии: %v", err)
			} else {
				logger.Printf("CompleteSessionWithoutRefund: уведомление о завершении сессии отправлено пользователю %d, SessionID=%s", user.TelegramID, session.ID)
			}
		} else {
			logger.Printf("CompleteSessionWithoutRefund: не удалось получить данные пользователя для отправки уведомления: %v", err)
		}
	}

	logger.Printf("CompleteSessionWithoutRefund: сессия завершена БЕЗ возврата - SessionID=%s, BoxID=%s", session.ID, *session.BoxID)
	return nil
}

// GetActiveSessionByUserID получает активную сессию пользователя
func (s *ServiceImpl) GetActiveSessionByUserID(userID uuid.UUID) (*models.Session, error) {
	return s.repo.GetActiveSessionByUserID(userID)
}

// ExtendSessionWithPayment создает платеж для продления сессии
func (s *ServiceImpl) ExtendSessionWithPayment(req *models.ExtendSessionWithPaymentRequest) (*models.ExtendSessionWithPaymentResponse, error) {
	// Получаем сессию по ID
	session, err := s.repo.GetSessionByID(req.SessionID)
	if err != nil {
		return nil, err
	}

	// Проверяем, что сессия в статусе active
	if session.Status != models.SessionStatusActive {
		return nil, fmt.Errorf("сессия должна быть в статусе active для продления")
	}

	// Проверяем, что время продления не отрицательное (разрешаем 0 для докупки только химии)
	if req.ExtensionTimeMinutes < 0 {
		return nil, fmt.Errorf("время продления не может быть отрицательным")
	}

	// Сохраняем запрошенное время продления и время химии в сессии
	updateFields := map[string]interface{}{
		"requested_extension_time_minutes": req.ExtensionTimeMinutes,
		"updated_at":                       time.Now(),
	}

	// Если передано время химии при продлении, сохраняем его в запрошенное поле
	if req.ExtensionChemistryTimeMinutes > 0 {
		updateFields["requested_extension_chemistry_time_minutes"] = req.ExtensionChemistryTimeMinutes
	}

	if err := s.repo.UpdateSessionFields(session.ID, updateFields); err != nil {
		return nil, fmt.Errorf("ошибка обновления сессии: %w", err)
	}

	// Обновляем локальную копию
	session.RequestedExtensionTimeMinutes = req.ExtensionTimeMinutes
	session.ExtensionChemistryTimeMinutes = req.ExtensionChemistryTimeMinutes

	// Рассчитываем цену продления через Payment Service
	priceResp, err := s.paymentService.CalculateExtensionPrice(&paymentModels.CalculateExtensionPriceRequest{
		ServiceType:                   session.ServiceType,
		ExtensionTimeMinutes:          req.ExtensionTimeMinutes,
		WithChemistry:                 session.WithChemistry || req.ExtensionChemistryTimeMinutes > 0,
		ExtensionChemistryTimeMinutes: req.ExtensionChemistryTimeMinutes,
	})
	if err != nil {
		return nil, fmt.Errorf("ошибка расчета цены продления: %w", err)
	}

	// Создаем платеж продления через Payment Service
	paymentResp, err := s.paymentService.CreateExtensionPayment(&paymentModels.CreateExtensionPaymentRequest{
		SessionID: session.ID,
		Amount:    priceResp.Price,
		Currency:  priceResp.Currency,
		Email:     session.Email, // Передаем email из сессии
	})
	if err != nil {
		return nil, fmt.Errorf("ошибка создания платежа продления: %w", err)
	}

	// Подготавливаем ответ с информацией о платеже
	payment := &models.Payment{
		ID:             paymentResp.Payment.ID,
		SessionID:      paymentResp.Payment.SessionID,
		Amount:         paymentResp.Payment.Amount,
		RefundedAmount: paymentResp.Payment.RefundedAmount,
		Currency:       paymentResp.Payment.Currency,
		Status:         paymentResp.Payment.Status,
		PaymentType:    paymentResp.Payment.PaymentType,
		PaymentURL:     paymentResp.Payment.PaymentURL,
		TinkoffID:      paymentResp.Payment.TinkoffID,
		ExpiresAt:      paymentResp.Payment.ExpiresAt,
		RefundedAt:     paymentResp.Payment.RefundedAt,
		CreatedAt:      paymentResp.Payment.CreatedAt,
		UpdatedAt:      paymentResp.Payment.UpdatedAt,
	}

	return &models.ExtendSessionWithPaymentResponse{
		Session: session,
		Payment: payment,
	}, nil
}

// GetSessionPayments получает все платежи сессии
func (s *ServiceImpl) GetSessionPayments(req *models.GetSessionPaymentsRequest) (*models.GetSessionPaymentsResponse, error) {
	// Получаем сессию по ID
	_, err := s.repo.GetSessionByID(req.SessionID)
	if err != nil {
		return nil, err
	}

	// Получаем все платежи сессии через Payment Service
	paymentsResp, err := s.paymentService.GetPaymentsBySessionID(req.SessionID)
	if err != nil {
		return nil, fmt.Errorf("ошибка получения платежей сессии: %w", err)
	}

	// Подготавливаем ответ
	response := &models.GetSessionPaymentsResponse{
		MainPayment:       nil,
		ExtensionPayments: []models.Payment{},
	}

	// Маппим основной платеж
	if paymentsResp.MainPayment != nil {
		response.MainPayment = &models.Payment{
			ID:             paymentsResp.MainPayment.ID,
			SessionID:      paymentsResp.MainPayment.SessionID,
			Amount:         paymentsResp.MainPayment.Amount,
			RefundedAmount: paymentsResp.MainPayment.RefundedAmount,
			Currency:       paymentsResp.MainPayment.Currency,
			Status:         paymentsResp.MainPayment.Status,
			PaymentType:    paymentsResp.MainPayment.PaymentType,
			PaymentURL:     paymentsResp.MainPayment.PaymentURL,
			TinkoffID:      paymentsResp.MainPayment.TinkoffID,
			ExpiresAt:      paymentsResp.MainPayment.ExpiresAt,
			RefundedAt:     paymentsResp.MainPayment.RefundedAt,
			CreatedAt:      paymentsResp.MainPayment.CreatedAt,
			UpdatedAt:      paymentsResp.MainPayment.UpdatedAt,
		}
	}

	// Маппим платежи продления
	for _, extPayment := range paymentsResp.ExtensionPayments {
		response.ExtensionPayments = append(response.ExtensionPayments, models.Payment{
			ID:             extPayment.ID,
			SessionID:      extPayment.SessionID,
			Amount:         extPayment.Amount,
			RefundedAmount: extPayment.RefundedAmount,
			Currency:       extPayment.Currency,
			Status:         extPayment.Status,
			PaymentType:    extPayment.PaymentType,
			PaymentURL:     extPayment.PaymentURL,
			TinkoffID:      extPayment.TinkoffID,
			ExpiresAt:      extPayment.ExpiresAt,
			RefundedAt:     extPayment.RefundedAt,
			CreatedAt:      extPayment.CreatedAt,
			UpdatedAt:      extPayment.UpdatedAt,
		})
	}

	return response, nil
}

// CancelSession отменяет сессию с возможным возвратом денег
func (s *ServiceImpl) CancelSession(req *models.CancelSessionRequest) (*models.CancelSessionResponse, error) {
	// Получаем сессию по ID
	session, err := s.repo.GetSessionByID(req.SessionID)
	if err != nil {
		return nil, fmt.Errorf("сессия не найдена: %w", err)
	}

	// Проверяем, что пользователь является владельцем сессии
	if session.UserID != req.UserID {
		return nil, fmt.Errorf("недостаточно прав для отмены сессии")
	}

	// Проверяем, что сессия может быть отменена
	allowedStatuses := []string{
		models.SessionStatusCreated,
		models.SessionStatusInQueue,
		models.SessionStatusAssigned,
	}

	isAllowedStatus := false
	for _, status := range allowedStatuses {
		if session.Status == status {
			isAllowedStatus = true
			break
		}
	}

	if !isAllowedStatus {
		return nil, fmt.Errorf("сессия в статусе '%s' не может быть отменена", session.Status)
	}

	// Подготавливаем ответ
	response := &models.CancelSessionResponse{
		Session: *session,
	}

	// Если сессия оплачена (in_queue или assigned), возвращаем деньги
	if (session.Status == models.SessionStatusInQueue || session.Status == models.SessionStatusAssigned) && !req.SkipRefund {
		// Получаем основной платеж сессии
		paymentResp, err := s.paymentService.GetMainPaymentBySessionID(session.ID)
		if err == nil && paymentResp != nil {
			// Выполняем возврат денег через payment service
			refundReq := &paymentModels.RefundPaymentRequest{
				PaymentID: paymentResp.ID,
				Amount:    paymentResp.Amount, // Возвращаем полную сумму
			}

			refundResp, err := s.paymentService.RefundPayment(refundReq)
			if err != nil {
				return nil, fmt.Errorf("ошибка возврата денег: %w", err)
			}

			// Добавляем информацию о платеже и возврате в ответ
			response.Payment = &models.Payment{
				ID:             refundResp.Payment.ID,
				SessionID:      refundResp.Payment.SessionID,
				Amount:         refundResp.Payment.Amount,
				RefundedAmount: refundResp.Payment.RefundedAmount,
				Currency:       refundResp.Payment.Currency,
				Status:         refundResp.Payment.Status,
				PaymentType:    refundResp.Payment.PaymentType,
				PaymentURL:     refundResp.Payment.PaymentURL,
				TinkoffID:      refundResp.Payment.TinkoffID,
				ExpiresAt:      refundResp.Payment.ExpiresAt,
				RefundedAt:     refundResp.Payment.RefundedAt,
				CreatedAt:      refundResp.Payment.CreatedAt,
				UpdatedAt:      refundResp.Payment.UpdatedAt,
			}

			response.Refund = &models.Refund{
				ID:        refundResp.Refund.ID,
				PaymentID: refundResp.Refund.PaymentID,
				Amount:    refundResp.Refund.Amount,
				Status:    refundResp.Refund.Status,
				CreatedAt: refundResp.Refund.CreatedAt,
			}
		}
	}

	// Обновляем статус сессии на canceled
	session.Status = models.SessionStatusCanceled
	session.StatusUpdatedAt = time.Now()
	err = s.repo.UpdateSession(session)
	if err != nil {
		return nil, fmt.Errorf("ошибка обновления статуса сессии: %w", err)
	}

	if session.BoxID == nil {
		// Обновляем сессию в ответе
		response.Session = *session

		return response, nil
	}

	// Обновляем статус бокса на free
	s.washboxService.UpdateWashBoxStatus(*session.BoxID, washboxModels.StatusFree)

	// Выключаем все койлы через Modbus при отмене сессии
	if s.modbusService != nil {
		logger.Printf("Выключение всех койлов для отмененной сессии - session_id: %s, box_id: %s", session.ID, *session.BoxID)

		// Выключаем свет
		lightRegister := s.getLightRegisterForBox(*session.BoxID)
		if lightRegister != "" {
			if err := s.modbusService.WriteLightCoil(*session.BoxID, lightRegister, false); err != nil {
				logger.Printf("Ошибка выключения света при отмене сессии - session_id: %s, box_id: %s, error: %v",
					session.ID, *session.BoxID, err)
			}
		} else {
			logger.Printf("CancelSession: не найден регистр света для бокса %s", *session.BoxID)
		}

		// Выключаем химию (если была включена)
		if session.WasChemistryOn {
			chemistryRegister := s.getChemistryRegisterForBox(*session.BoxID)
			if chemistryRegister != "" {
				if err := s.modbusService.WriteChemistryCoil(*session.BoxID, chemistryRegister, false); err != nil {
					logger.Printf("Ошибка выключения химии при отмене сессии - session_id: %s, box_id: %s, error: %v",
						session.ID, *session.BoxID, err)
				}
			} else {
				logger.Printf("CancelSession: не найден регистр химии для бокса %s", *session.BoxID)
			}
		}
	}

	// Отправляем уведомление о возврате денег при отмене сессии
	if s.telegramBot != nil {
		user, err := s.userService.GetUserByID(session.UserID)
		if err == nil && user != nil {
			err = s.telegramBot.SendSessionNotification(user.TelegramID, telegram.NotificationTypeSessionExpiredOrCanceled, nil)
			if err != nil {
				logger.Printf("CancelSession: ошибка отправки уведомления о возврате денег: %v", err)
			} else {
				logger.Printf("CancelSession: уведомление о возврате денег отправлено пользователю %d, SessionID=%s", user.TelegramID, session.ID)
			}
		} else {
			logger.Printf("CancelSession: не удалось получить данные пользователя для отправки уведомления: %v", err)
		}
	}

	// Обновляем сессию в ответе
	response.Session = *session

	return response, nil
}

// CheckAndCompleteExpiredSessions проверяет и завершает истекшие сессии
func (s *ServiceImpl) CheckAndCompleteExpiredSessions() error {
	// Получаем все активные сессии
	activeSessions, err := s.repo.GetSessionsByStatus(models.SessionStatusActive)
	if err != nil {
		return err
	}

	// Если нет активных сессий, выходим
	if len(activeSessions) == 0 {
		return nil
	}

	// Текущее время
	now := time.Now()

	// Проверяем каждую активную сессию
	for _, session := range activeSessions {
		// Время начала сессии - это время последнего обновления статуса на active
		startTime := session.StatusUpdatedAt

		// Получаем время мойки в минутах (по умолчанию 5 минут)
		rentalTime := session.RentalTimeMinutes
		if rentalTime <= 0 {
			rentalTime = 5
		}

		// Учитываем время продления, если оно есть
		totalTime := rentalTime + session.ExtensionTimeMinutes

		cooldownTimeout, err := s.settingsService.GetCooldownTimeout()
		if err != nil {
			// Если не удалось получить настройку, используем значение по умолчанию
			cooldownTimeout = 5
		}

		// Проверяем, прошло ли выбранное время с момента начала сессии
		if now.Sub(startTime) >= time.Duration(totalTime)*time.Minute {
			// Если прошло время, завершаем сессию
			if session.BoxID != nil && s.washboxService != nil {
				// Исключаем сессии кассира из кулдауна
				if s.cashierUserID != "" {
					cashierUserID, err := uuid.Parse(s.cashierUserID)
					if err == nil && session.UserID != cashierUserID {
						// Устанавливаем cooldown для бокса
						cooldownUntil := now.Add(time.Duration(cooldownTimeout) * time.Minute)
						err = s.washboxService.SetCooldown(*session.BoxID, session.UserID, cooldownUntil)
						if err != nil {
							return err
						}
					} else if err == nil && session.UserID == cashierUserID {
						// Для сессий кассира устанавливаем cooldown по госномеру
						if session.CarNumber != "" {
							// Устанавливаем cooldown для бокса по госномеру
							cooldownUntil := now.Add(time.Duration(cooldownTimeout) * time.Minute)
							err = s.washboxService.SetCooldownByCarNumber(*session.BoxID, session.CarNumber, cooldownUntil)
							if err != nil {
								return err
							}
						} else {
							// Если госномер не указан, переводим бокс в статус "свободен"
							err = s.washboxService.UpdateWashBoxStatus(*session.BoxID, washboxModels.StatusFree)
							if err != nil {
								return err
							}
						}
					}
				} else {
					// Если CASHIER_USER_ID не настроен, устанавливаем cooldown для всех
					cooldownTimeout, err := s.settingsService.GetCooldownTimeout()
					if err != nil {
						// Если не удалось получить настройку, используем значение по умолчанию
						cooldownTimeout = 5
					}

					// Устанавливаем cooldown для бокса
					cooldownUntil := now.Add(time.Duration(cooldownTimeout) * time.Minute)
					err = s.washboxService.SetCooldown(*session.BoxID, session.UserID, cooldownUntil)
					if err != nil {
						return err
					}
				}
			}

			// Выключаем все койлы через Modbus при истечении сессии
			if session.BoxID != nil && s.modbusService != nil {
				logger.Printf("Выключение всех койлов для истекшей сессии - session_id: %s, box_id: %s", session.ID, *session.BoxID)

				// Выключаем свет
				lightRegister := s.getLightRegisterForBox(*session.BoxID)
				if lightRegister != "" {
					if err := s.modbusService.WriteLightCoil(*session.BoxID, lightRegister, false); err != nil {
						logger.Printf("Ошибка выключения света при истечении сессии - session_id: %s, box_id: %s, error: %v",
							session.ID, *session.BoxID, err)
					}
				} else {
					logger.Printf("ExpireSession: не найден регистр света для бокса %s", *session.BoxID)
				}

				// Выключаем химию (если была включена)
				if session.WasChemistryOn {
					chemistryRegister := s.getChemistryRegisterForBox(*session.BoxID)
					if chemistryRegister != "" {
						if err := s.modbusService.WriteChemistryCoil(*session.BoxID, chemistryRegister, false); err != nil {
							logger.Printf("Ошибка выключения химии при истечении сессии - session_id: %s, box_id: %s, error: %v",
								session.ID, *session.BoxID, err)
						}
					} else {
						logger.Printf("ExpireSession: не найден регистр химии для бокса %s", *session.BoxID)
					}
				}
			}

			// Обновляем статус сессии на complete, время обновления статуса и сбрасываем флаг уведомления
			session.Status = models.SessionStatusComplete
			session.StatusUpdatedAt = time.Now()         // Обновляем время изменения статуса
			session.IsCompletingNotificationSent = false // Сбрасываем флаг, чтобы уведомление могло быть отправлено снова
			err = s.repo.UpdateSession(&session)
			if err != nil {
				return err
			}

			// Отправляем уведомление о завершении сессии с информацией о кулдауне
			if s.telegramBot != nil {
				user, err := s.userService.GetUserByID(session.UserID)
				if err == nil && user != nil {
					// Определяем, нужно ли показывать информацию о кулдауне в уведомлении
					var cooldownMinutes *int
					if s.cashierUserID != "" {
						cashierUserID, err := uuid.Parse(s.cashierUserID)
						if err == nil && session.UserID != cashierUserID {
							// Это обычный пользователь, получаем время кулдауна для уведомления
							cooldownTimeout, err := s.settingsService.GetCooldownTimeout()
							if err == nil {
								cooldownMinutes = &cooldownTimeout
							}
						}
					} else {
						// Если CASHIER_USER_ID не настроен, считаем что все сессии имеют кулдаун
						cooldownTimeout, err := s.settingsService.GetCooldownTimeout()
						if err == nil {
							cooldownMinutes = &cooldownTimeout
						}
					}

					err = s.telegramBot.SendSessionNotification(user.TelegramID, telegram.NotificationTypeSessionCompleted, cooldownMinutes)
					if err != nil {
						logger.Printf("CheckAndCompleteExpiredSessions: ошибка отправки уведомления о завершении сессии: %v", err)
					} else {
						logger.Printf("CheckAndCompleteExpiredSessions: уведомление о завершении сессии отправлено пользователю %d, SessionID=%s", user.TelegramID, session.ID)
					}
				} else {
					logger.Printf("CheckAndCompleteExpiredSessions: не удалось получить данные пользователя для отправки уведомления: %v", err)
				}
			}
		}
	}

	return nil
}

// ProcessQueue обрабатывает очередь сессий
func (s *ServiceImpl) ProcessQueue() error {
	// Если сервис боксов не инициализирован, выходим
	if s.washboxService == nil {
		return nil
	}

	// Получаем все сессии со статусом "in_queue" (оплаченные сессии)
	sessions, err := s.repo.GetSessionsByStatus(models.SessionStatusInQueue)
	if err != nil {
		return err
	}

	// Если нет сессий в очереди, выходим
	if len(sessions) == 0 {
		return nil
	}

	// Обрабатываем каждую сессию
	for _, session := range sessions {
		// Если у сессии не указан тип услуги, пропускаем её
		if session.ServiceType == "" {
			continue
		}

		// Получаем доступные боксы с учетом cooldown и приоритетов
		var availableBoxes []washboxModels.WashBox
		var err error

		// Проверяем, является ли это кассирской сессией
		isCashierSession := false
		if s.cashierUserID != "" {
			cashierUserID, err := uuid.Parse(s.cashierUserID)
			if err == nil && session.UserID == cashierUserID {
				isCashierSession = true
			}
		}

		if isCashierSession && session.CarNumber != "" {
			// Для кассирских сессий используем логику по госномеру
			switch session.ServiceType {
			case "wash":
				if session.WithChemistry {
					// Ищем боксы с химией для мойки с учетом cooldown по госномеру
					availableBoxes, err = s.washboxService.GetAvailableBoxesByServiceTypeForCashier("wash", session.CarNumber)
					// Фильтруем только боксы с химией
					var filteredBoxes []washboxModels.WashBox
					for _, box := range availableBoxes {
						if box.ChemistryEnabled {
							filteredBoxes = append(filteredBoxes, box)
						}
					}
					availableBoxes = filteredBoxes
				} else {
					// Ищем любые боксы для мойки с учетом cooldown по госномеру
					availableBoxes, err = s.washboxService.GetAvailableBoxesByServiceTypeForCashier("wash", session.CarNumber)
				}
			case "air_dry":
				// Химия недоступна для air_dry
				availableBoxes, err = s.washboxService.GetAvailableBoxesByServiceTypeForCashier("air_dry", session.CarNumber)
			case "vacuum":
				// Химия недоступна для vacuum
				availableBoxes, err = s.washboxService.GetAvailableBoxesByServiceTypeForCashier("vacuum", session.CarNumber)
			default:
				// Для неизвестных типов услуг
				availableBoxes, err = s.washboxService.GetAvailableBoxesByServiceTypeForCashier(session.ServiceType, session.CarNumber)
			}
		} else {
			// Для обычных пользователей используем существующую логику
			switch session.ServiceType {
			case "wash":
				if session.WithChemistry {
					// Ищем боксы с химией для мойки с учетом cooldown
					availableBoxes, err = s.washboxService.GetAvailableBoxesByServiceType("wash", session.UserID)
					// Фильтруем только боксы с химией
					var filteredBoxes []washboxModels.WashBox
					for _, box := range availableBoxes {
						if box.ChemistryEnabled {
							filteredBoxes = append(filteredBoxes, box)
						}
					}
					availableBoxes = filteredBoxes
				} else {
					// Ищем любые боксы для мойки с учетом cooldown
					availableBoxes, err = s.washboxService.GetAvailableBoxesByServiceType("wash", session.UserID)
				}
			case "air_dry":
				// Химия недоступна для air_dry
				availableBoxes, err = s.washboxService.GetAvailableBoxesByServiceType("air_dry", session.UserID)
			case "vacuum":
				// Химия недоступна для vacuum
				availableBoxes, err = s.washboxService.GetAvailableBoxesByServiceType("vacuum", session.UserID)
			default:
				// Для неизвестных типов услуг
				availableBoxes, err = s.washboxService.GetAvailableBoxesByServiceType(session.ServiceType, session.UserID)
			}
		}

		if err != nil {
			return err
		}

		// Если нет подходящих боксов, пропускаем сессию
		if len(availableBoxes) == 0 {
			continue
		}

		// Берем первый доступный бокс
		box := availableBoxes[0]

		// Обновляем статус бокса на "reserved"
		err = s.washboxService.UpdateWashBoxStatus(box.ID, washboxModels.StatusReserved)
		if err != nil {
			return err
		}

		// Обновляем сессию - назначаем бокс, меняем статус и обновляем время изменения статуса
		session.BoxID = &box.ID
		session.BoxNumber = &box.Number
		session.Status = models.SessionStatusAssigned
		session.StatusUpdatedAt = time.Now() // Обновляем время изменения статуса
		err = s.repo.UpdateSession(&session)
		if err != nil {
			return err
		}

		// Проверяем, был ли бокс в кулдауне для этого пользователя или госномера
		// Если да, то сразу запускаем сессию, пропуская статус assigned
		if isCashierSession && session.CarNumber != "" {
			// Для кассирских сессий проверяем кулдаун по госномеру
			cooldownBoxes, err := s.washboxService.GetCooldownBoxesByCarNumber(session.CarNumber)
			if err == nil {
				// Проверяем, есть ли среди боксов в кулдауне тот, который мы только что назначили
				for _, cooldownBox := range cooldownBoxes {
					if cooldownBox.ID == box.ID && cooldownBox.ServiceType == session.ServiceType {
						// Бокс был в кулдауне для этого госномера - сразу запускаем сессию
						logger.Printf("ProcessQueue: бокс %s был в кулдауне для госномера %s, запускаем сессию %s автоматически", box.ID, session.CarNumber, session.ID)

						// Запускаем сессию
						_, err = s.StartSession(&models.StartSessionRequest{
							SessionID: session.ID,
						})
						if err != nil {
							logger.Printf("ProcessQueue: ошибка автоматического запуска сессии %s: %v", session.ID, err)
						} else {
							logger.Printf("ProcessQueue: сессия %s автоматически запущена для бокса в кулдауне по госномеру", session.ID)
						}
						break
					}
				}
			}
		} else {
			// Для обычных пользователей проверяем кулдаун по user_id
			cooldownBoxes, err := s.washboxService.GetCooldownBoxesForUser(session.UserID)
			if err == nil {
				// Проверяем, есть ли среди боксов в кулдауне тот, который мы только что назначили
				for _, cooldownBox := range cooldownBoxes {
					if cooldownBox.ID == box.ID && cooldownBox.ServiceType == session.ServiceType {
						// Бокс был в кулдауне для этого пользователя - сразу запускаем сессию
						logger.Printf("ProcessQueue: бокс %s был в кулдауне для пользователя %s, запускаем сессию %s автоматически", box.ID, session.UserID, session.ID)

						// Запускаем сессию
						_, err = s.StartSession(&models.StartSessionRequest{
							SessionID: session.ID,
						})
						if err != nil {
							logger.Printf("ProcessQueue: ошибка автоматического запуска сессии %s: %v", session.ID, err)
						} else {
							logger.Printf("ProcessQueue: сессия %s автоматически запущена для бокса в кулдауне", session.ID)
						}
						break
					}
				}
			}
		}

		// Отправляем уведомление о назначении бокса (только если сессия не была автоматически запущена)
		if session.Status == models.SessionStatusAssigned && s.userService != nil && s.telegramBot != nil {
			// Получаем пользователя
			user, err := s.userService.GetUserByID(session.UserID)
			if err == nil {
				// Отправляем уведомление с номером бокса
				err = s.telegramBot.SendBoxAssignmentNotification(user.TelegramID, box.Number)
				if err != nil {
					logger.Printf("ProcessQueue: ошибка отправки уведомления о назначении бокса для сессии %s: %v", session.ID, err)
				}
			} else {
				logger.Printf("ProcessQueue: ошибка получения пользователя для сессии %s: %v", session.ID, err)
			}
		}
	}

	return nil
}

// CheckAndExpireReservedSessions проверяет и автоматически запускает сессии, которые не были стартованы в течение 3 минут
func (s *ServiceImpl) CheckAndExpireReservedSessions() error {
	// Получаем время ожидания старта мойки из настроек
	sessionTimeout, err := s.settingsService.GetSessionTimeout()
	if err != nil {
		// Если не удалось получить настройку, используем значение по умолчанию
		sessionTimeout = 3
	}

	// Получаем все сессии со статусом "assigned"
	assignedSessions, err := s.repo.GetSessionsByStatus(models.SessionStatusAssigned)
	if err != nil {
		return err
	}

	// Если нет назначенных сессий, выходим
	if len(assignedSessions) == 0 {
		return nil
	}

	// Текущее время
	now := time.Now()

	// Проверяем каждую назначенную сессию
	for _, session := range assignedSessions {
		// Время назначения сессии - это время последнего обновления статуса на assigned
		assignedTime := session.StatusUpdatedAt

		// Проверяем, прошло ли время ожидания с момента назначения сессии
		if now.Sub(assignedTime) >= time.Duration(sessionTimeout)*time.Minute {
			// Если прошло время ожидания, автоматически запускаем сессию
			if session.BoxID != nil {
				// Используем существующий метод StartSession для запуска сессии
				_, err = s.StartSession(&models.StartSessionRequest{
					SessionID: session.ID,
				})
				if err != nil {
					logger.Printf("CheckAndExpireReservedSessions: ошибка автоматического запуска сессии %s: %v", session.ID, err)
					continue
				}

				// Отправляем уведомление об автоматическом запуске сессии
				if s.telegramBot != nil {
					user, err := s.userService.GetUserByID(session.UserID)
					if err == nil && user != nil {
						err = s.telegramBot.SendSessionNotification(user.TelegramID, telegram.NotificationTypeSessionAutoStarted, nil)
						if err != nil {
							logger.Printf("CheckAndExpireReservedSessions: ошибка отправки уведомления об автоматическом запуске: %v", err)
						} else {
							logger.Printf("CheckAndExpireReservedSessions: уведомление об автоматическом запуске отправлено пользователю %d, SessionID=%s", user.TelegramID, session.ID)
						}
					} else {
						logger.Printf("CheckAndExpireReservedSessions: не удалось получить данные пользователя для отправки уведомления: %v", err)
					}
				}

				logger.Printf("CheckAndExpireReservedSessions: сессия автоматически запущена - SessionID=%s, BoxID=%s", session.ID, *session.BoxID)
			}
		}
	}

	return nil
}

// CheckAndNotifyExpiringReservedSessions проверяет и отправляет уведомления для сессий, которые скоро истекут
func (s *ServiceImpl) CheckAndNotifyExpiringReservedSessions() error {
	// Если сервис пользователей или телеграм бот не инициализированы, выходим
	if s.userService == nil || s.telegramBot == nil {
		return nil
	}

	// Получаем время ожидания старта мойки из настроек
	sessionTimeout, err := s.settingsService.GetSessionTimeout()
	if err != nil {
		// Если не удалось получить настройку, используем значение по умолчанию
		sessionTimeout = 3
	}

	// Получаем все сессии со статусом "assigned"
	assignedSessions, err := s.repo.GetSessionsByStatus(models.SessionStatusAssigned)
	if err != nil {
		return err
	}

	// Если нет назначенных сессий, выходим
	if len(assignedSessions) == 0 {
		return nil
	}

	// Текущее время
	now := time.Now()

	// Проверяем каждую назначенную сессию
	for _, session := range assignedSessions {
		// Время назначения сессии - это время последнего обновления статуса на assigned
		assignedTime := session.StatusUpdatedAt

		// Проверяем, прошло ли время для отправки уведомления (за 1 минуту до истечения)
		notificationTime := sessionTimeout - 1
		if now.Sub(assignedTime) >= time.Duration(notificationTime)*time.Minute && now.Sub(assignedTime) < time.Duration(sessionTimeout)*time.Minute {
			// Если прошло 2 минуты и уведомление еще не отправлено, отправляем его
			if !session.IsExpiringNotificationSent {
				// Получаем пользователя
				user, err := s.userService.GetUserByID(session.UserID)
				if err != nil {
					continue
				}

				// Отправляем уведомление
				err = s.telegramBot.SendSessionNotification(user.TelegramID, telegram.NotificationTypeSessionExpiringSoon, nil)
				if err != nil {
					continue
				}

				// Помечаем, что уведомление отправлено
				err = s.repo.UpdateSessionFields(session.ID, map[string]interface{}{
					"is_expiring_notification_sent": true,
					"updated_at":                    time.Now(),
				})
				if err != nil {
					logger.Printf("CheckAndNotifyExpiringSessions: ошибка обновления флага уведомления для сессии %s: %v", session.ID, err)
					continue
				}
			}
		}
	}

	return nil
}

// CheckAndNotifyCompletingSessions проверяет и отправляет уведомления для сессий, которые скоро завершатся
func (s *ServiceImpl) CheckAndNotifyCompletingSessions() error {
	// Если сервис пользователей или телеграм бот не инициализированы, выходим
	if s.userService == nil || s.telegramBot == nil {
		return nil
	}

	// Получаем все сессии со статусом "active"
	activeSessions, err := s.repo.GetSessionsByStatus(models.SessionStatusActive)
	if err != nil {
		return err
	}

	// Если нет активных сессий, выходим
	if len(activeSessions) == 0 {
		return nil
	}

	// Текущее время
	now := time.Now()

	// Проверяем каждую активную сессию
	for _, session := range activeSessions {
		// Время начала сессии - это время последнего обновления статуса на active
		startTime := session.StatusUpdatedAt

		// Получаем время мойки в минутах (по умолчанию 5 минут)
		rentalTime := session.RentalTimeMinutes
		if rentalTime <= 0 {
			rentalTime = 5
		}

		// Учитываем время продления, если оно есть
		totalTime := rentalTime + session.ExtensionTimeMinutes

		// Проверяем, прошло ли время с момента начала сессии (за 5 минут до завершения)
		if now.Sub(startTime) >= time.Duration(totalTime-5)*time.Minute && now.Sub(startTime) < time.Duration(totalTime)*time.Minute {
			if !session.IsCompletingNotificationSent {
				// Получаем пользователя
				user, err := s.userService.GetUserByID(session.UserID)
				if err != nil {
					continue
				}

				// Отправляем уведомление
				err = s.telegramBot.SendSessionNotification(user.TelegramID, telegram.NotificationTypeSessionCompletingSoon, nil)
				if err != nil {
					continue
				}

				// Помечаем, что уведомление отправлено
				err = s.repo.UpdateSessionFields(session.ID, map[string]interface{}{
					"is_completing_notification_sent": true,
					"updated_at":                      time.Now(),
				})
				if err != nil {
					logger.Printf("CheckAndNotifyCompletingSessions: ошибка обновления флага уведомления для сессии %s: %v", session.ID, err)
					continue
				}
			}
		}
	}

	return nil
}

// CountSessionsByStatus подсчитывает количество сессий с определенным статусом
func (s *ServiceImpl) CountSessionsByStatus(status string) (int, error) {
	return s.repo.CountSessionsByStatus(status)
}

// GetSessionsByStatus получает сессии по статусу
func (s *ServiceImpl) GetSessionsByStatus(status string) ([]models.Session, error) {
	return s.repo.GetSessionsByStatus(status)
}

// GetUserSessionHistory получает историю сессий пользователя
func (s *ServiceImpl) GetUserSessionHistory(req *models.GetUserSessionHistoryRequest) ([]models.Session, error) {
	// Устанавливаем значения по умолчанию, если не указаны
	limit := req.Limit
	if limit <= 0 {
		limit = 10 // По умолчанию возвращаем 10 сессий
	}

	offset := req.Offset
	if offset < 0 {
		offset = 0
	}

	sessions, err := s.repo.GetUserSessionHistory(req.UserID, limit, offset)
	if err != nil {
		return nil, err
	}

	// Заполняем информацию о платежах для каждой сессии
	for i := range sessions {
		// Получаем полную информацию о платежах сессии
		paymentsResp, err := s.paymentService.GetPaymentsBySessionID(sessions[i].ID)
		if err == nil && paymentsResp != nil {
			// Основной платеж
			if paymentsResp.MainPayment != nil {
				sessions[i].MainPayment = &models.Payment{
					ID:             paymentsResp.MainPayment.ID,
					SessionID:      paymentsResp.MainPayment.SessionID,
					Amount:         paymentsResp.MainPayment.Amount,
					RefundedAmount: paymentsResp.MainPayment.RefundedAmount,
					Currency:       paymentsResp.MainPayment.Currency,
					Status:         paymentsResp.MainPayment.Status,
					PaymentType:    paymentsResp.MainPayment.PaymentType,
					PaymentURL:     paymentsResp.MainPayment.PaymentURL,
					TinkoffID:      paymentsResp.MainPayment.TinkoffID,
					ExpiresAt:      paymentsResp.MainPayment.ExpiresAt,
					RefundedAt:     paymentsResp.MainPayment.RefundedAt,
					CreatedAt:      paymentsResp.MainPayment.CreatedAt,
					UpdatedAt:      paymentsResp.MainPayment.UpdatedAt,
				}
				// Для обратной совместимости оставляем Payment
				sessions[i].Payment = sessions[i].MainPayment
			}

			// Платежи продления
			if len(paymentsResp.ExtensionPayments) > 0 {
				sessions[i].ExtensionPayments = make([]models.Payment, len(paymentsResp.ExtensionPayments))
				for j, extPayment := range paymentsResp.ExtensionPayments {
					sessions[i].ExtensionPayments[j] = models.Payment{
						ID:             extPayment.ID,
						SessionID:      extPayment.SessionID,
						Amount:         extPayment.Amount,
						RefundedAmount: extPayment.RefundedAmount,
						Currency:       extPayment.Currency,
						Status:         extPayment.Status,
						PaymentType:    extPayment.PaymentType,
						PaymentURL:     extPayment.PaymentURL,
						TinkoffID:      extPayment.TinkoffID,
						ExpiresAt:      extPayment.ExpiresAt,
						RefundedAt:     extPayment.RefundedAt,
						CreatedAt:      extPayment.CreatedAt,
						UpdatedAt:      extPayment.UpdatedAt,
					}
				}
			}
		}

		// Заполняем session_timeout_minutes из настроек
		sessionTimeout, err := s.settingsService.GetSessionTimeout()
		if err != nil {
			// Если не удалось получить настройку, используем значение по умолчанию
			sessionTimeout = 3
		}
		sessions[i].SessionTimeoutMinutes = sessionTimeout

		// Заполняем cooldown_minutes только для завершенных сессий с реальным активным кулдауном
		if sessions[i].Status == models.SessionStatusComplete && sessions[i].BoxID != nil {
			// Проверяем, что это не сессия кассира
			if s.cashierUserID != "" {
				cashierUserID, err := uuid.Parse(s.cashierUserID)
				if err == nil && sessions[i].UserID != cashierUserID {
					// Это обычный пользователь, проверяем реальный активный кулдаун
					if s.washboxService != nil {
						box, err := s.washboxService.GetWashBoxByID(*sessions[i].BoxID)
						if err == nil && box != nil {
							// Проверяем, что последний пользователь - это пользователь сессии и кулдаун еще активен
							if box.LastCompletedSessionUserID != nil &&
								*box.LastCompletedSessionUserID == sessions[i].UserID &&
								box.CooldownUntil != nil &&
								box.CooldownUntil.After(time.Now()) {
								// Получаем время кулдауна из настроек
								cooldownTimeout, err := s.settingsService.GetCooldownTimeout()
								if err == nil {
									sessions[i].CooldownMinutes = &cooldownTimeout
								}
							}
						}
					}
				}
			} else {
				// Если CASHIER_USER_ID не настроен, проверяем реальный активный кулдаун для всех
				if s.washboxService != nil {
					box, err := s.washboxService.GetWashBoxByID(*sessions[i].BoxID)
					if err == nil && box != nil {
						// Проверяем, что последний пользователь - это пользователь сессии и кулдаун еще активен
						if box.LastCompletedSessionUserID != nil &&
							*box.LastCompletedSessionUserID == sessions[i].UserID &&
							box.CooldownUntil != nil &&
							box.CooldownUntil.After(time.Now()) {
							// Получаем время кулдауна из настроек
							cooldownTimeout, err := s.settingsService.GetCooldownTimeout()
							if err == nil {
								sessions[i].CooldownMinutes = &cooldownTimeout
							}
						}
					}
				}
			}
		}
	}

	return sessions, nil
}

// CreateFromCashier создает сессию из запроса кассира
func (s *ServiceImpl) CreateFromCashier(req *models.CashierPaymentRequest) (*models.Session, error) {
	// Проверяем, что ID кассира настроен
	if s.cashierUserID == "" {
		return nil, fmt.Errorf("CASHIER_USER_ID не настроен")
	}

	cashierUserID, err := uuid.Parse(s.cashierUserID)
	if err != nil {
		return nil, fmt.Errorf("неверный формат CASHIER_USER_ID: %v", err)
	}

	// Нормализация госномера (если указан)
	var normalizedCarNumber string
	if req.CarNumber != "" {
		normalizedCarNumber = utils.NormalizeLicensePlate(req.CarNumber)
		logger.Printf("Service - CreateFromCashier: госномер нормализован '%s' -> '%s'", req.CarNumber, normalizedCarNumber)
	}

	// Валидация химии
	if req.WithChemistry {
		switch req.ServiceType {
		case "wash":
			// Химия разрешена для мойки
		case "air_dry", "vacuum":
			return nil, fmt.Errorf("chemistry is not available for service type: %s", req.ServiceType)
		default:
			return nil, fmt.Errorf("invalid service type: %s", req.ServiceType)
		}
	}

	// Создаем новую сессию
	now := time.Now()
	session := &models.Session{
		UserID:               cashierUserID,
		Status:               models.SessionStatusInQueue, // Статус "в очереди" как указано в требованиях
		ServiceType:          req.ServiceType,
		WithChemistry:        req.WithChemistry,
		ChemistryTimeMinutes: req.ChemistryTimeMinutes, // Сохраняем выбранное время химии
		CarNumber:            normalizedCarNumber,      // Используем нормализованный госномер
		CarNumberCountry:     req.CarNumberCountry,     // Сохраняем страну гос номера
		RentalTimeMinutes:    req.RentalTimeMinutes,
		StatusUpdatedAt:      now,
	}

	// Сохраняем сессию в базе данных
	err = s.repo.CreateSession(session)
	if err != nil {
		return nil, err
	}

	return session, nil
}

// AdminListSessions список сессий для администратора
func (s *ServiceImpl) AdminListSessions(req *models.AdminListSessionsRequest) (*models.AdminListSessionsResponse, error) {
	// Устанавливаем значения по умолчанию
	limit := 50
	offset := 0

	if req.Limit != nil {
		limit = *req.Limit
	}
	if req.Offset != nil {
		offset = *req.Offset
	}

	// Получаем сессии с фильтрацией
	sessions, total, err := s.repo.GetSessionsWithFilters(
		req.UserID,
		req.BoxID,
		req.BoxNumber,
		req.Status,
		req.ServiceType,
		req.DateFrom,
		req.DateTo,
		limit,
		offset,
	)
	if err != nil {
		return nil, err
	}

	// Заполняем session_timeout_minutes из настроек для каждой сессии
	sessionTimeout, err := s.settingsService.GetSessionTimeout()
	if err != nil {
		// Если не удалось получить настройку, используем значение по умолчанию
		sessionTimeout = 3
	}
	for i := range sessions {
		sessions[i].SessionTimeoutMinutes = sessionTimeout
	}

	return &models.AdminListSessionsResponse{
		Sessions: sessions,
		Total:    total,
		Limit:    limit,
		Offset:   offset,
	}, nil
}

// AdminGetSession получение информации о сессии для администратора
func (s *ServiceImpl) AdminGetSession(req *models.AdminGetSessionRequest) (*models.AdminGetSessionResponse, error) {
	session, err := s.repo.GetSessionByID(req.ID)
	if err != nil {
		return nil, err
	}

	// Заполняем session_timeout_minutes из настроек
	if session != nil {
		sessionTimeout, err := s.settingsService.GetSessionTimeout()
		if err != nil {
			// Если не удалось получить настройку, используем значение по умолчанию
			sessionTimeout = 3
		}
		session.SessionTimeoutMinutes = sessionTimeout
	}

	return &models.AdminGetSessionResponse{
		Session: *session,
	}, nil
}

// UpdateSessionStatus обновляет статус сессии
func (s *ServiceImpl) UpdateSessionStatus(sessionID uuid.UUID, status string) error {
	// Получаем сессию по ID
	session, err := s.repo.GetSessionByID(sessionID)
	if err != nil {
		return fmt.Errorf("сессия не найдена: %w", err)
	}

	if status == models.SessionStatusInQueue && session.Status != models.SessionStatusCreated {
		logger.Printf("сессия не может быть переведена в очередь, так как находится не в статусе created")
		return nil
	}

	// Обновляем статус и время обновления
	session.Status = status
	session.StatusUpdatedAt = time.Now()

	// Сохраняем изменения
	err = s.repo.UpdateSession(session)
	if err != nil {
		return fmt.Errorf("ошибка обновления статуса сессии: %w", err)
	}

	return nil
}

// UpdateSessionExtension обновляет время продления сессии
func (s *ServiceImpl) UpdateSessionExtension(sessionID uuid.UUID, extensionTimeMinutes int) error {
	// Получаем сессию по ID
	session, err := s.repo.GetSessionByID(sessionID)
	if err != nil {
		return fmt.Errorf("сессия не найдена: %w", err)
	}

	// Если extensionTimeMinutes = 0, используем запрошенное время продления
	if extensionTimeMinutes == 0 {
		extensionTimeMinutes = session.RequestedExtensionTimeMinutes
	}

	// Обновляем время продления сессии
	session.ExtensionTimeMinutes += extensionTimeMinutes
	session.RequestedExtensionTimeMinutes = 0 // Очищаем запрошенное время

	// НОВАЯ ЛОГИКА: Обрабатываем химию при продлении
	if session.RequestedExtensionChemistryTimeMinutes > 0 {
		// Если это докупка химии (ExtensionTimeMinutes = 0) и изначально химия не была куплена
		if extensionTimeMinutes == 0 && !session.WithChemistry {
			// Докупка химии для мойки без химии - добавляем химию
			session.WithChemistry = true
			session.ChemistryTimeMinutes = session.RequestedExtensionChemistryTimeMinutes
			logger.Printf("UpdateSessionExtension: докупка химии для мойки без химии, время %d минут, SessionID=%s",
				session.RequestedExtensionChemistryTimeMinutes, session.ID)
		} else if session.WithChemistry {
			// Применяем логику химии при продлении для существующей химии
			if session.WasChemistryOn && session.ChemistryEndedAt == nil {
				// Химия активна - продлеваем на докупленное время
				session.ChemistryTimeMinutes += session.RequestedExtensionChemistryTimeMinutes
				logger.Printf("UpdateSessionExtension: химия активна, продлеваем на %d минут, общее время %d минут, SessionID=%s",
					session.RequestedExtensionChemistryTimeMinutes, session.ChemistryTimeMinutes, session.ID)
			} else if session.ChemistryStartedAt == nil {
				// Химия не использована - даем на первоначальное + докупленное время
				totalChemistryTime := session.ChemistryTimeMinutes + session.RequestedExtensionChemistryTimeMinutes
				session.ChemistryTimeMinutes = totalChemistryTime
				logger.Printf("UpdateSessionExtension: химия не использована, общее время %d минут, SessionID=%s",
					totalChemistryTime, session.ID)
			} else {
				// Химия уже использована - даем только на докупленное время и сбрасываем флаги
				session.ChemistryTimeMinutes = session.RequestedExtensionChemistryTimeMinutes
				session.ChemistryStartedAt = nil
				session.ChemistryEndedAt = nil
				session.WasChemistryOn = false
				logger.Printf("UpdateSessionExtension: химия использована, новое время %d минут, сброшены флаги, SessionID=%s",
					session.RequestedExtensionChemistryTimeMinutes, session.ID)
			}
		}

		// Очищаем запрошенное время химии при продлении
		session.RequestedExtensionChemistryTimeMinutes = 0
	}

	// Сохраняем изменения
	err = s.repo.UpdateSession(session)
	if err != nil {
		return fmt.Errorf("ошибка обновления времени продления сессии: %w", err)
	}

	return nil
}

// CashierListSessions возвращает список сессий для кассира с начала смены
func (s *ServiceImpl) CashierListSessions(req *models.CashierSessionsRequest) (*models.AdminListSessionsResponse, error) {
	// Устанавливаем значения по умолчанию для пагинации
	limit := 50
	if req.Limit > 0 {
		limit = req.Limit
	}
	offset := 0
	if req.Offset > 0 {
		offset = req.Offset
	}

	// Получаем сессии с начала смены
	sessions, total, err := s.repo.GetSessionsWithFilters(nil, nil, nil, nil, nil, &req.ShiftStartedAt, nil, limit, offset)
	if err != nil {
		return nil, err
	}

	// Загружаем платежи для каждой сессии
	for i := range sessions {
		payments, err := s.paymentService.GetPaymentsBySessionID(sessions[i].ID)
		if err != nil {
			// Логируем ошибку, но продолжаем работу
			continue
		}

		// Устанавливаем основной платеж
		if payments.MainPayment != nil {
			sessions[i].MainPayment = &models.Payment{
				ID:             payments.MainPayment.ID,
				SessionID:      payments.MainPayment.SessionID,
				Amount:         payments.MainPayment.Amount,
				RefundedAmount: payments.MainPayment.RefundedAmount,
				Currency:       payments.MainPayment.Currency,
				Status:         payments.MainPayment.Status,
				PaymentType:    payments.MainPayment.PaymentType,
				PaymentURL:     payments.MainPayment.PaymentURL,
				TinkoffID:      payments.MainPayment.TinkoffID,
				ExpiresAt:      payments.MainPayment.ExpiresAt,
				RefundedAt:     payments.MainPayment.RefundedAt,
				CreatedAt:      payments.MainPayment.CreatedAt,
				UpdatedAt:      payments.MainPayment.UpdatedAt,
			}
		}

		// Устанавливаем платежи продления
		if len(payments.ExtensionPayments) > 0 {
			sessions[i].ExtensionPayments = make([]models.Payment, len(payments.ExtensionPayments))
			for j, payment := range payments.ExtensionPayments {
				sessions[i].ExtensionPayments[j] = models.Payment{
					ID:             payment.ID,
					SessionID:      payment.SessionID,
					Amount:         payment.Amount,
					RefundedAmount: payment.RefundedAmount,
					Currency:       payment.Currency,
					Status:         payment.Status,
					PaymentType:    payment.PaymentType,
					PaymentURL:     payment.PaymentURL,
					TinkoffID:      payment.TinkoffID,
					ExpiresAt:      payment.ExpiresAt,
					RefundedAt:     payment.RefundedAt,
					CreatedAt:      payment.CreatedAt,
					UpdatedAt:      payment.UpdatedAt,
				}
			}
		}

		// Заполняем session_timeout_minutes из настроек
		sessionTimeout, err := s.settingsService.GetSessionTimeout()
		if err != nil {
			// Если не удалось получить настройку, используем значение по умолчанию
			sessionTimeout = 3
		}
		sessions[i].SessionTimeoutMinutes = sessionTimeout
	}

	return &models.AdminListSessionsResponse{
		Sessions: sessions,
		Total:    total,
		Limit:    limit,
		Offset:   offset,
	}, nil
}

// CashierGetActiveSessions возвращает активные сессии кассира
func (s *ServiceImpl) CashierGetActiveSessions(req *models.CashierActiveSessionsRequest) (*models.CashierActiveSessionsResponse, error) {
	// Устанавливаем значения по умолчанию для пагинации
	limit := 50
	if req.Limit > 0 {
		limit = req.Limit
	}
	offset := 0
	if req.Offset > 0 {
		offset = req.Offset
	}

	// Получаем ID кассира из конфигурации
	cashierUserID, err := uuid.Parse(s.cashierUserID)
	if err != nil {
		return nil, fmt.Errorf("некорректный ID кассира: %w", err)
	}

	// Получаем активные сессии кассира (не завершенные)
	// Активные статусы: created, in_queue, assigned, active
	// Терминальные статусы: complete, canceled, expired, payment_failed
	sessions, _, err := s.repo.GetSessionsWithFilters(&cashierUserID, nil, nil, nil, nil, nil, nil, limit, offset)
	if err != nil {
		return nil, err
	}

	// Фильтруем только активные сессии
	var activeSessions []models.Session
	for _, session := range sessions {
		if session.Status == "created" || session.Status == "in_queue" ||
			session.Status == "assigned" || session.Status == "active" {
			activeSessions = append(activeSessions, session)
		}
	}

	// Сортируем сессии: сначала in_queue, потом active, потом остальные
	sort.Slice(activeSessions, func(i, j int) bool {
		if activeSessions[i].Status == "in_queue" && activeSessions[j].Status != "in_queue" {
			return true
		}
		if activeSessions[i].Status == "active" && activeSessions[j].Status != "in_queue" && activeSessions[j].Status != "active" {
			return true
		}
		return false
	})

	// Загружаем платежи для каждой сессии
	for i := range activeSessions {
		payments, err := s.paymentService.GetPaymentsBySessionID(activeSessions[i].ID)
		if err != nil {
			// Логируем ошибку, но продолжаем работу
			continue
		}

		// Устанавливаем основной платеж
		if payments.MainPayment != nil {
			activeSessions[i].MainPayment = &models.Payment{
				ID:             payments.MainPayment.ID,
				SessionID:      payments.MainPayment.SessionID,
				Amount:         payments.MainPayment.Amount,
				RefundedAmount: payments.MainPayment.RefundedAmount,
				Currency:       payments.MainPayment.Currency,
				Status:         payments.MainPayment.Status,
				PaymentType:    payments.MainPayment.PaymentType,
				PaymentURL:     payments.MainPayment.PaymentURL,
				TinkoffID:      payments.MainPayment.TinkoffID,
				ExpiresAt:      payments.MainPayment.ExpiresAt,
				RefundedAt:     payments.MainPayment.RefundedAt,
				CreatedAt:      payments.MainPayment.CreatedAt,
				UpdatedAt:      payments.MainPayment.UpdatedAt,
			}
		}

		// Устанавливаем платежи продления
		if len(payments.ExtensionPayments) > 0 {
			activeSessions[i].ExtensionPayments = make([]models.Payment, len(payments.ExtensionPayments))
			for j, payment := range payments.ExtensionPayments {
				activeSessions[i].ExtensionPayments[j] = models.Payment{
					ID:             payment.ID,
					SessionID:      payment.SessionID,
					Amount:         payment.Amount,
					RefundedAmount: payment.RefundedAmount,
					Currency:       payment.Currency,
					Status:         payment.Status,
					PaymentType:    payment.PaymentType,
					PaymentURL:     payment.PaymentURL,
					TinkoffID:      payment.TinkoffID,
					ExpiresAt:      payment.ExpiresAt,
					RefundedAt:     payment.RefundedAt,
					CreatedAt:      payment.CreatedAt,
					UpdatedAt:      payment.UpdatedAt,
				}
			}
		}

		// Заполняем session_timeout_minutes из настроек
		sessionTimeout, err := s.settingsService.GetSessionTimeout()
		if err != nil {
			// Если не удалось получить настройку, используем значение по умолчанию
			sessionTimeout = 3
		}
		activeSessions[i].SessionTimeoutMinutes = sessionTimeout
	}

	return &models.CashierActiveSessionsResponse{
		Sessions: activeSessions,
		Total:    len(activeSessions),
		Limit:    limit,
		Offset:   offset,
	}, nil
}

// CashierStartSession запускает сессию кассиром
func (s *ServiceImpl) CashierStartSession(req *models.CashierStartSessionRequest) (*models.Session, error) {
	// Получаем сессию
	session, err := s.repo.GetSessionByID(req.SessionID)
	if err != nil {
		return nil, fmt.Errorf("сессия не найдена: %w", err)
	}

	// Проверяем, что это сессия кассира
	cashierUserID, err := uuid.Parse(s.cashierUserID)
	if err != nil {
		return nil, fmt.Errorf("некорректный ID кассира: %w", err)
	}

	if session.UserID != cashierUserID {
		return nil, fmt.Errorf("доступ запрещен: сессия не принадлежит кассиру")
	}

	// Проверяем статус сессии
	if session.Status != "assigned" {
		return nil, fmt.Errorf("нельзя запустить сессию со статусом: %s", session.Status)
	}

	// Запускаем сессию
	return s.StartSession(&models.StartSessionRequest{
		SessionID: req.SessionID,
	})
}

// CashierCompleteSession завершает сессию кассиром
func (s *ServiceImpl) CashierCompleteSession(req *models.CashierCompleteSessionRequest) (*models.Session, error) {
	// Получаем сессию
	session, err := s.repo.GetSessionByID(req.SessionID)
	if err != nil {
		return nil, fmt.Errorf("сессия не найдена: %w", err)
	}

	// Кассир может завершать любые активные сессии

	// Проверяем статус сессии
	if session.Status != "active" {
		return nil, fmt.Errorf("нельзя завершить сессию со статусом: %s", session.Status)
	}

	// Завершаем сессию
	response, err := s.CompleteSession(&models.CompleteSessionRequest{
		SessionID: req.SessionID,
	})
	if err != nil {
		return nil, err
	}

	return response.Session, nil
}

// CashierCancelSession отменяет сессию кассиром
func (s *ServiceImpl) CashierCancelSession(req *models.CashierCancelSessionRequest) (*models.Session, error) {
	// Получаем сессию
	session, err := s.repo.GetSessionByID(req.SessionID)
	if err != nil {
		return nil, fmt.Errorf("сессия не найдена: %w", err)
	}

	// Кассир может отменять любые сессии в статусе in_queue или assigned

	// Проверяем статус сессии
	if session.Status != "in_queue" && session.Status != "assigned" {
		return nil, fmt.Errorf("нельзя отменить сессию со статусом: %s", session.Status)
	}

	// Отменяем сессию
	response, err := s.CancelSession(&models.CancelSessionRequest{
		SessionID:  req.SessionID,
		UserID:     session.UserID, // Используем ID владельца сессии
		SkipRefund: req.SkipRefund, // Передаем признак пропуска возврата
	})
	if err != nil {
		return nil, err
	}

	return &response.Session, nil
}

// EnableChemistry включает химию в сессии
func (s *ServiceImpl) EnableChemistry(req *models.EnableChemistryRequest) (*models.EnableChemistryResponse, error) {
	// Получаем сессию
	session, err := s.repo.GetSessionByID(req.SessionID)
	if err != nil {
		return nil, fmt.Errorf("сессия не найдена: %w", err)
	}

	// Проверяем, что химия была оплачена
	if !session.WithChemistry {
		return nil, fmt.Errorf("химия не была оплачена для этой сессии")
	}

	// Проверяем, что химия еще не была включена
	if session.WasChemistryOn {
		return nil, fmt.Errorf("химия уже была включена для этой сессии")
	}

	// Проверяем статус сессии
	if session.Status != "active" {
		return nil, fmt.Errorf("химию можно включить только в активной сессии")
	}

	// Включаем химию
	now := time.Now()

	// ВАЖНО: Обновляем только нужные поля, не трогаем StatusUpdatedAt, RentalTimeMinutes, ExtensionTimeMinutes!
	logger.Printf("EnableChemistry: обновление химии - SessionID=%s", session.ID)

	// Используем Updates для обновления только нужных полей
	err = s.repo.UpdateSessionFields(session.ID, map[string]interface{}{
		"was_chemistry_on":     true,
		"chemistry_started_at": now,
		"updated_at":           now,
	})
	if err != nil {
		return nil, fmt.Errorf("ошибка при обновлении сессии: %w", err)
	}

	// Перечитываем обновленную сессию
	session, err = s.repo.GetSessionByID(session.ID)
	if err != nil {
		return nil, fmt.Errorf("ошибка получения обновленной сессии: %w", err)
	}

	// Включаем химию в боксе через Modbus
	if s.modbusService != nil && session.BoxID != nil {
		// Получаем регистр химии для бокса
		chemistryRegister := s.getChemistryRegisterForBox(*session.BoxID)
		if chemistryRegister != "" {
			err = s.modbusService.WriteChemistryCoil(*session.BoxID, chemistryRegister, true)
			if err != nil {
				// Логируем ошибку (HandleModbusError теперь только логирует, не продлевает время)
				s.modbusService.HandleModbusError(*session.BoxID, "chemistry_on", session.ID, err)
				logger.Printf("Ошибка включения химии в боксе %s: %v", *session.BoxID, err)
			}
		} else {
			logger.Printf("EnableChemistry: не найден регистр химии для бокса %s", *session.BoxID)
		}
	}

	logger.Printf("Химия включена: SessionID=%s, ChemistryTimeMinutes=%d", session.ID, session.ChemistryTimeMinutes)

	// Запускаем автоматическое выключение химии через указанное время
	if session.ChemistryTimeMinutes > 0 {
		s.AutoDisableChemistry(session.ID, session.ChemistryTimeMinutes)
	}

	return &models.EnableChemistryResponse{
		Session: *session,
	}, nil
}

// AutoDisableChemistry автоматически выключает химию через указанное количество минут
func (s *ServiceImpl) AutoDisableChemistry(sessionID uuid.UUID, chemistryTimeMinutes int) {
	// Запускаем в отдельной горутине
	go func() {
		logger.Printf("AutoDisableChemistry: запланировано автовыключение химии через %d минут, SessionID=%s", chemistryTimeMinutes, sessionID)

		// Ждем указанное время
		time.Sleep(time.Duration(chemistryTimeMinutes) * time.Minute)

		// Получаем сессию
		session, err := s.repo.GetSessionByID(sessionID)
		if err != nil {
			logger.Printf("AutoDisableChemistry: ошибка получения сессии для автовыключения химии: %v, SessionID=%s", err, sessionID)
			return
		}

		// Проверяем что химия все еще активна (была включена, но не выключена)
		if session.ChemistryStartedAt != nil && session.ChemistryEndedAt == nil {
			// Выключаем химию через Modbus
			if s.modbusService != nil && session.BoxID != nil {
				// Получаем регистр химии для бокса
				chemistryRegister := s.getChemistryRegisterForBox(*session.BoxID)
				if chemistryRegister != "" {
					err := s.modbusService.WriteChemistryCoil(*session.BoxID, chemistryRegister, false)
					if err != nil {
						logger.Printf("AutoDisableChemistry: ошибка автовыключения химии через Modbus: %v, SessionID=%s, BoxID=%s", err, sessionID, *session.BoxID)
					} else {
						logger.Printf("AutoDisableChemistry: химия успешно выключена через Modbus, SessionID=%s, BoxID=%s", sessionID, *session.BoxID)
					}
				} else {
					logger.Printf("AutoDisableChemistry: не найден регистр химии для бокса %s", *session.BoxID)
				}
			}

			// Обновляем время выключения химии
			now := time.Now()

			// ВАЖНО: Обновляем только нужные поля, не трогаем StatusUpdatedAt, RentalTimeMinutes, ExtensionTimeMinutes!
			logger.Printf("AutoDisableChemistry: выключение химии - SessionID=%s", sessionID)

			// Используем Updates для обновления только нужных полей
			err = s.repo.UpdateSessionFields(sessionID, map[string]interface{}{
				"chemistry_ended_at": now,
				"updated_at":         now,
			})
			if err != nil {
				logger.Printf("AutoDisableChemistry: ошибка обновления сессии: %v, SessionID=%s", err, sessionID)
			} else {
				logger.Printf("AutoDisableChemistry: химия автоматически выключена, SessionID=%s", sessionID)
			}
		} else {
			logger.Printf("AutoDisableChemistry: химия уже была выключена, SessionID=%s", sessionID)
		}
	}()
}

// GetChemistryStats получает статистику использования химии
func (s *ServiceImpl) GetChemistryStats(req *models.GetChemistryStatsRequest) (*models.GetChemistryStatsResponse, error) {
	stats, err := s.repo.GetChemistryStats(req.DateFrom, req.DateTo)
	if err != nil {
		return nil, fmt.Errorf("ошибка при получении статистики химии: %w", err)
	}

	return &models.GetChemistryStatsResponse{
		Stats: *stats,
	}, nil
}

// ReassignSession переназначает сессию на другой бокс
func (s *ServiceImpl) ReassignSession(req *models.ReassignSessionRequest) (*models.ReassignSessionResponse, error) {
	logger.Printf("ReassignSession: начало переназначения сессии, SessionID=%s", req.SessionID)

	// Получаем сессию по ID
	session, err := s.repo.GetSessionByID(req.SessionID)
	if err != nil {
		logger.Printf("ReassignSession: ошибка получения сессии, SessionID=%s, error=%v", req.SessionID, err)
		return nil, fmt.Errorf("не удалось найти сессию: %w", err)
	}

	// Проверяем, что сессия в подходящем статусе для переназначения
	if session.Status != models.SessionStatusAssigned && session.Status != models.SessionStatusActive {
		logger.Printf("ReassignSession: сессия не может быть переназначена, SessionID=%s, status=%s", req.SessionID, session.Status)
		return &models.ReassignSessionResponse{
			Success: false,
			Message: "Сессия не может быть переназначена в текущем статусе",
			Session: *session,
		}, nil
	}

	// Если у сессии есть бокс, переводим его в статус maintenance и выключаем оборудование
	if session.BoxID != nil {
		logger.Printf("ReassignSession: перевод бокса в maintenance, SessionID=%s, BoxID=%s", req.SessionID, *session.BoxID)

		// Переводим бокс в статус maintenance
		if s.washboxService != nil {
			err = s.washboxService.UpdateWashBoxStatus(*session.BoxID, washboxModels.StatusMaintenance)
			if err != nil {
				logger.Printf("ReassignSession: ошибка перевода бокса в maintenance, SessionID=%s, BoxID=%s, error=%v", req.SessionID, *session.BoxID, err)
				return nil, fmt.Errorf("не удалось перевести бокс в статус обслуживания: %w", err)
			}
		}

		// Выключаем свет и химию через Modbus
		if s.modbusService != nil {
			logger.Printf("ReassignSession: выключение оборудования, SessionID=%s, BoxID=%s", req.SessionID, *session.BoxID)

			// Выключаем свет
			lightRegister := s.getLightRegisterForBox(*session.BoxID)
			if lightRegister != "" {
				if err := s.modbusService.WriteLightCoil(*session.BoxID, lightRegister, false); err != nil {
					logger.Printf("ReassignSession: ошибка выключения света, SessionID=%s, BoxID=%s, error=%v", req.SessionID, *session.BoxID, err)
					// Не прерываем выполнение, логируем ошибку
				}
			} else {
				logger.Printf("ReassignSession: не найден регистр света для бокса %s", *session.BoxID)
			}

			// Выключаем химию (если была включена)
			if session.WasChemistryOn {
				chemistryRegister := s.getChemistryRegisterForBox(*session.BoxID)
				if chemistryRegister != "" {
					if err := s.modbusService.WriteChemistryCoil(*session.BoxID, chemistryRegister, false); err != nil {
						logger.Printf("ReassignSession: ошибка выключения химии, SessionID=%s, BoxID=%s, error=%v", req.SessionID, *session.BoxID, err)
						// Не прерываем выполнение, логируем ошибку
					}
				} else {
					logger.Printf("ReassignSession: не найден регистр химии для бокса %s", *session.BoxID)
				}
			}
		}

		// Обнуляем связь с боксом
		session.BoxID = nil
		session.BoxNumber = nil
	}

	// Сбрасываем флаги химии - будто химия не была использована вообще
	session.WasChemistryOn = false
	session.ChemistryStartedAt = nil
	session.ChemistryEndedAt = nil
	session.ChemistryTimeMinutes = 0

	// Возвращаем сессию в очередь
	session.Status = models.SessionStatusInQueue
	session.StatusUpdatedAt = time.Now() // Сбрасываем таймер

	// Обновляем сессию в БД
	err = s.repo.UpdateSession(session)
	if err != nil {
		logger.Printf("ReassignSession: ошибка обновления сессии, SessionID=%s, error=%v", req.SessionID, err)
		return nil, fmt.Errorf("не удалось обновить сессию: %w", err)
	}

	logger.Printf("ReassignSession: сессия возвращена в очередь, SessionID=%s", req.SessionID)

	// Запускаем обработку очереди для назначения на новый бокс
	if s.washboxService != nil {
		go func() {
			err := s.ProcessQueue()
			if err != nil {
				logger.Printf("ReassignSession: ошибка обработки очереди после переназначения, SessionID=%s, error=%v", req.SessionID, err)
			}
		}()
	}

	logger.Printf("ReassignSession: успешно переназначена, SessionID=%s", req.SessionID)

	return &models.ReassignSessionResponse{
		Success: true,
		Message: "Сессия успешно переназначена на другой бокс",
		Session: *session,
	}, nil
}

// CheckAndAutoEnableChemistry проверяет и автоматически включает химию для сессий, где время подходит к концу
func (s *ServiceImpl) CheckAndAutoEnableChemistry() error {
	// Если сервис пользователей или телеграм бот не инициализированы, выходим
	if s.userService == nil || s.telegramBot == nil {
		return nil
	}

	// Получаем все сессии со статусом "active"
	activeSessions, err := s.repo.GetSessionsByStatus(models.SessionStatusActive)
	if err != nil {
		return err
	}

	// Если нет активных сессий, выходим
	if len(activeSessions) == 0 {
		return nil
	}

	// Текущее время
	now := time.Now()

	// Проверяем каждую активную сессию
	for _, session := range activeSessions {
		// Проверяем, что это мойка с химией
		if session.ServiceType != "wash" || !session.WithChemistry {
			continue
		}

		// Проверяем, что химия еще не была включена
		if session.WasChemistryOn {
			continue
		}

		// Проверяем, что есть время химии для использования
		if session.ChemistryTimeMinutes <= 0 {
			continue
		}

		// Время начала сессии - это время последнего обновления статуса на active
		startTime := session.StatusUpdatedAt

		// Получаем время мойки в минутах (по умолчанию 5 минут)
		rentalTime := session.RentalTimeMinutes
		if rentalTime <= 0 {
			rentalTime = 5
		}

		// Учитываем время продления, если оно есть
		totalTime := rentalTime + session.ExtensionTimeMinutes

		// Прошедшее время с момента начала сессии
		elapsedTime := now.Sub(startTime)

		// Оставшееся время мойки
		remainingTime := time.Duration(totalTime)*time.Minute - elapsedTime

		// Если оставшегося времени меньше чем время химии, включаем химию автоматически
		if remainingTime < time.Duration(session.ChemistryTimeMinutes)*time.Minute {
			logger.Printf("CheckAndAutoEnableChemistry: автоматическое включение химии - SessionID=%s, оставшееся время=%v, время химии=%d минут",
				session.ID, remainingTime, session.ChemistryTimeMinutes)

			// Включаем химию
			enableReq := &models.EnableChemistryRequest{
				SessionID: session.ID,
			}

			_, err := s.EnableChemistry(enableReq)
			if err != nil {
				logger.Printf("CheckAndAutoEnableChemistry: ошибка автоматического включения химии - SessionID=%s, error=%v", session.ID, err)
				continue
			}

			// Отправляем уведомление пользователю
			user, err := s.userService.GetUserByID(session.UserID)
			if err != nil {
				logger.Printf("CheckAndAutoEnableChemistry: ошибка получения пользователя для уведомления - UserID=%s, error=%v", session.UserID, err)
				continue
			}

			// Отправляем уведомление через Telegram
			err = s.telegramBot.SendSessionNotification(user.TelegramID, telegram.NotificationTypeChemistryAutoEnabled, nil)
			if err != nil {
				logger.Printf("CheckAndAutoEnableChemistry: ошибка отправки уведомления - UserID=%s, TelegramID=%s, error=%v",
					user.ID, user.TelegramID, err)
			} else {
				logger.Printf("CheckAndAutoEnableChemistry: уведомление отправлено - UserID=%s, TelegramID=%s",
					user.ID, user.TelegramID)
			}
		}
	}

	return nil
}

// getLightRegisterForBox получает регистр света для бокса
func (s *ServiceImpl) getLightRegisterForBox(boxID uuid.UUID) string {
	if s.washboxService == nil {
		return ""
	}

	box, err := s.washboxService.GetWashBoxByID(boxID)
	if err != nil {
		logger.Printf("getLightRegisterForBox: ошибка получения бокса %s: %v", boxID, err)
		return ""
	}

	if box.LightCoilRegister != nil {
		return *box.LightCoilRegister
	}

	return ""
}

// getChemistryRegisterForBox получает регистр химии для бокса
func (s *ServiceImpl) getChemistryRegisterForBox(boxID uuid.UUID) string {
	if s.washboxService == nil {
		return ""
	}

	box, err := s.washboxService.GetWashBoxByID(boxID)
	if err != nil {
		logger.Printf("getChemistryRegisterForBox: ошибка получения бокса %s: %v", boxID, err)
		return ""
	}

	if box.ChemistryCoilRegister != nil {
		return *box.ChemistryCoilRegister
	}

	return ""
}

// GetActiveSessionByCarNumber получает активную сессию по номеру автомобиля
func (s *ServiceImpl) GetActiveSessionByCarNumber(carNumber string) (*models.Session, error) {
	logger.Printf("Service - GetActiveSessionByCarNumber: поиск активной сессии по номеру %s", carNumber)

	session, err := s.repo.GetActiveSessionByCarNumber(carNumber)
	if err != nil {
		logger.Printf("Service - GetActiveSessionByCarNumber: активная сессия с номером %s не найдена: %v", carNumber, err)
		return nil, err
	}

	logger.Printf("Service - GetActiveSessionByCarNumber: найдена активная сессия - SessionID=%s, CarNumber=%s, Status=%s",
		session.ID, session.CarNumber, session.Status)

	return session, nil
}
