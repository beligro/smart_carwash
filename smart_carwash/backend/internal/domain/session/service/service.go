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
	"fmt"
	"sort"
	"strconv"
	"time"

	"github.com/google/uuid"
)

// Service интерфейс для бизнес-логики сессий
type Service interface {
	CreateSession(req *models.CreateSessionRequest) (*models.Session, error)
	CreateSessionWithPayment(req *models.CreateSessionWithPaymentRequest) (*models.CreateSessionWithPaymentResponse, error)
	GetUserSession(req *models.GetUserSessionRequest) (*models.GetUserSessionResponse, error)
	GetUserSessionForPayment(req *models.GetUserSessionRequest) (*models.GetUserSessionResponse, error)
	GetSession(req *models.GetSessionRequest) (*models.GetSessionResponse, error)
	StartSession(req *models.StartSessionRequest) (*models.Session, error)
	CompleteSession(req *models.CompleteSessionRequest) (*models.CompleteSessionResponse, error)
	ExtendSession(req *models.ExtendSessionRequest) (*models.Session, error)
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
		logger.Printf("Service - CreateSession: у пользователя уже есть активная сессия, session_id: %s, user_id: %s", existingSession.ID.String(), req.UserID.String())
		return existingSession, nil
	}

	// Создаем новую сессию
	now := time.Now()
	session := &models.Session{
		UserID:               req.UserID,
		Status:               models.SessionStatusCreated,
		ServiceType:          req.ServiceType,
		WithChemistry:        req.WithChemistry,
		ChemistryTimeMinutes: req.ChemistryTimeMinutes, // Сохраняем выбранное время химии
		CarNumber:            req.CarNumber,
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
		return session, nil // Возвращаем сессию без изменений
	}

	// Проверяем, что у сессии есть назначенный бокс
	if session.BoxID == nil {
		return session, nil // Возвращаем сессию без изменений
	}

	// Если сервис боксов не инициализирован, просто обновляем статус сессии
	if s.washboxService == nil {
		// Обновляем статус сессии на active
		session.Status = models.SessionStatusActive
		err = s.repo.UpdateSession(session)
		if err != nil {
			return nil, err
		}
		return session, nil
	}

	// Получаем информацию о боксе
	box, err := s.washboxService.GetWashBoxByID(*session.BoxID)
	if err != nil {
		return nil, err
	}

	// Обновляем статус бокса на busy
	err = s.washboxService.UpdateWashBoxStatus(*session.BoxID, washboxModels.StatusBusy)
	if err != nil {
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
		return nil, err
	}

	// Включаем свет в боксе через Modbus
	if s.modbusService != nil {
		err = s.modbusService.WriteLightCoil(*session.BoxID, true)
		if err != nil {
			// Логируем ошибку (HandleModbusError теперь только логирует, не продлевает время)
			s.modbusService.HandleModbusError(*session.BoxID, "light_on", session.ID, err)
			logger.Printf("Ошибка включения света в боксе %s: %v", *session.BoxID, err)
		}
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
			err = s.modbusService.WriteChemistryCoil(*session.BoxID, false)
			if err != nil {
				logger.Printf("CompleteSession: ошибка выключения химии в боксе %s: %v", *session.BoxID, err)
			} else {
				logger.Printf("CompleteSession: химия выключена в боксе %s, SessionID=%s", *session.BoxID, session.ID)
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
		err = s.modbusService.WriteLightCoil(*session.BoxID, false)
		if err != nil {
			logger.Printf("Ошибка выключения света в боксе %s: %v", *session.BoxID, err)
		}
	}

	// Рассчитываем возврат по всем платежам сессии
	refundReq := &paymentModels.CalculateSessionRefundRequest{
		SessionID:            session.ID,
		ServiceType:          session.ServiceType,
		RentalTimeMinutes:    session.RentalTimeMinutes,
		ExtensionTimeMinutes: session.ExtensionTimeMinutes,
		UsedTimeSeconds:      usedTimeSeconds,
	}

	logger.Printf("Завершение сессии: SessionID=%s, RentalTime=%dmin, ExtensionTime=%dmin, UsedTime=%ds",
		session.ID, session.RentalTimeMinutes, session.ExtensionTimeMinutes, usedTimeSeconds)

	refundCalcResp, err := s.paymentService.CalculateSessionRefund(refundReq)
	if err != nil {
		logger.Printf("Ошибка расчета возврата по сессии: %v", err)
	} else if refundCalcResp.TotalRefundAmount > 0 {
		// Выполняем возврат по каждому платежу
		for _, refund := range refundCalcResp.Refunds {
			if refund.RefundAmount > 0 {
				refundPaymentReq := &paymentModels.RefundPaymentRequest{
					PaymentID: refund.PaymentID,
					Amount:    refund.RefundAmount,
				}

				_, err := s.paymentService.RefundPayment(refundPaymentReq)
				if err != nil {
					logger.Printf("Ошибка выполнения возврата для платежа %s: %v", refund.PaymentID, err)
				} else {
					logger.Printf("Успешно выполнен возврат для платежа %s: Amount=%d",
						refund.PaymentID, refund.RefundAmount)
				}
			}
		}

		logger.Printf("Успешно выполнен возврат по сессии: SessionID=%s, TotalRefundAmount=%d",
			session.ID, refundCalcResp.TotalRefundAmount)
	}

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

	return &models.CompleteSessionResponse{
		Session: session,
		Payment: session.Payment,
	}, nil
}

// ExtendSession продлевает сессию (добавляет время к активной сессии)
func (s *ServiceImpl) ExtendSession(req *models.ExtendSessionRequest) (*models.Session, error) {
	// Получаем сессию по ID
	session, err := s.repo.GetSessionByID(req.SessionID)
	if err != nil {
		return nil, err
	}

	// Проверяем, что сессия в статусе active
	if session.Status != models.SessionStatusActive {
		return nil, fmt.Errorf("сессия должна быть в статусе active для продления")
	}

	// Проверяем, что время продления положительное
	if req.ExtensionTimeMinutes <= 0 {
		return nil, fmt.Errorf("время продления должно быть положительным числом")
	}

	// Обновляем время продления сессии
	session.ExtensionTimeMinutes += req.ExtensionTimeMinutes
	session.StatusUpdatedAt = time.Now()
	err = s.repo.UpdateSession(session)
	if err != nil {
		return nil, err
	}

	return session, nil
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

	// Проверяем, что время продления положительное
	if req.ExtensionTimeMinutes <= 0 {
		return nil, fmt.Errorf("время продления должно быть положительным числом")
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
		WithChemistry:                 session.WithChemistry,
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
		if err := s.modbusService.WriteLightCoil(*session.BoxID, false); err != nil {
			logger.Printf("Ошибка выключения света при отмене сессии - session_id: %s, box_id: %s, error: %v",
				session.ID, *session.BoxID, err)
		}

		// Выключаем химию (если была включена)
		if session.WasChemistryOn {
			if err := s.modbusService.WriteChemistryCoil(*session.BoxID, false); err != nil {
				logger.Printf("Ошибка выключения химии при отмене сессии - session_id: %s, box_id: %s, error: %v",
					session.ID, *session.BoxID, err)
			}
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

		// Получаем время аренды в минутах (по умолчанию 5 минут)
		rentalTime := session.RentalTimeMinutes
		if rentalTime <= 0 {
			rentalTime = 5
		}

		// Учитываем время продления, если оно есть
		totalTime := rentalTime + session.ExtensionTimeMinutes

		// Проверяем, прошло ли выбранное время с момента начала сессии
		if now.Sub(startTime) >= time.Duration(totalTime)*time.Minute {
			// Если прошло 5 минут, завершаем сессию
			if session.BoxID != nil && s.washboxService != nil {
				// Обновляем статус бокса на free
				err = s.washboxService.UpdateWashBoxStatus(*session.BoxID, washboxModels.StatusFree)
				if err != nil {
					return err
				}
			}

			// Выключаем все койлы через Modbus при истечении сессии
			if session.BoxID != nil && s.modbusService != nil {
				logger.Printf("Выключение всех койлов для истекшей сессии - session_id: %s, box_id: %s", session.ID, *session.BoxID)

				// Выключаем свет
				if err := s.modbusService.WriteLightCoil(*session.BoxID, false); err != nil {
					logger.Printf("Ошибка выключения света при истечении сессии - session_id: %s, box_id: %s, error: %v",
						session.ID, *session.BoxID, err)
				}

				// Выключаем химию (если была включена)
				if session.WasChemistryOn {
					if err := s.modbusService.WriteChemistryCoil(*session.BoxID, false); err != nil {
						logger.Printf("Ошибка выключения химии при истечении сессии - session_id: %s, box_id: %s, error: %v",
							session.ID, *session.BoxID, err)
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

		// Получаем подходящие боксы с учетом химии
		var freeBoxes []washboxModels.WashBox
		var err error

		switch session.ServiceType {
		case "wash":
			if session.WithChemistry {
				// Ищем боксы с химией для мойки
				freeBoxes, err = s.washboxService.GetFreeWashBoxesWithChemistry("wash")
			} else {
				// Ищем любые боксы для мойки
				freeBoxes, err = s.washboxService.GetFreeWashBoxesByServiceType("wash")
			}
		case "air_dry":
			// Химия недоступна для air_dry
			freeBoxes, err = s.washboxService.GetFreeWashBoxesByServiceType("air_dry")
		case "vacuum":
			// Химия недоступна для vacuum
			freeBoxes, err = s.washboxService.GetFreeWashBoxesByServiceType("vacuum")
		default:
			// Для неизвестных типов услуг
			freeBoxes, err = s.washboxService.GetFreeWashBoxesByServiceType(session.ServiceType)
		}

		if err != nil {
			return err
		}

		// Если нет подходящих боксов, пропускаем сессию
		if len(freeBoxes) == 0 {
			continue
		}

		// Берем первый свободный бокс
		box := freeBoxes[0]

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
	}

	return nil
}

// CheckAndExpireReservedSessions проверяет и истекает сессии, которые не были стартованы в течение 3 минут
func (s *ServiceImpl) CheckAndExpireReservedSessions() error {
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

		// Проверяем, прошло ли 3 минуты с момента назначения сессии
		if now.Sub(assignedTime) >= 3*time.Minute {
			// Если прошло 3 минуты, истекаем сессию
			if session.BoxID != nil && s.washboxService != nil {
				// Обновляем статус бокса на free
				err = s.washboxService.UpdateWashBoxStatus(*session.BoxID, washboxModels.StatusFree)
				if err != nil {
					return err
				}
			}

			// Выключаем все койлы через Modbus при истечении резервирования
			if session.BoxID != nil && s.modbusService != nil {
				logger.Printf("Выключение всех койлов для истекшего резервирования - session_id: %s, box_id: %s", session.ID, *session.BoxID)

				// Выключаем свет
				if err := s.modbusService.WriteLightCoil(*session.BoxID, false); err != nil {
					logger.Printf("Ошибка выключения света при истечении резервирования - session_id: %s, box_id: %s, error: %v",
						session.ID, *session.BoxID, err)
				}

				// Выключаем химию (если была включена)
				if session.WasChemistryOn {
					if err := s.modbusService.WriteChemistryCoil(*session.BoxID, false); err != nil {
						logger.Printf("Ошибка выключения химии при истечении резервирования - session_id: %s, box_id: %s, error: %v",
							session.ID, *session.BoxID, err)
					}
				}
			}

			// Обновляем статус сессии на expired, время обновления статуса и сбрасываем флаг уведомления
			session.Status = models.SessionStatusExpired
			session.StatusUpdatedAt = time.Now()       // Обновляем время изменения статуса
			session.IsExpiringNotificationSent = false // Сбрасываем флаг, чтобы уведомление могло быть отправлено снова
			err = s.repo.UpdateSession(&session)
			if err != nil {
				return err
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

		// Проверяем, прошло ли 2 минуты с момента назначения сессии (за 1 минуту до истечения)
		if now.Sub(assignedTime) >= 2*time.Minute && now.Sub(assignedTime) < 3*time.Minute {
			// Если прошло 2 минуты и уведомление еще не отправлено, отправляем его
			if !session.IsExpiringNotificationSent {
				// Получаем пользователя
				user, err := s.userService.GetUserByID(session.UserID)
				if err != nil {
					continue
				}

				// Отправляем уведомление
				err = s.telegramBot.SendSessionNotification(user.TelegramID, telegram.NotificationTypeSessionExpiringSoon)
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

		// Получаем время аренды в минутах (по умолчанию 5 минут)
		rentalTime := session.RentalTimeMinutes
		if rentalTime <= 0 {
			rentalTime = 5
		}

		// Учитываем время продления, если оно есть
		totalTime := rentalTime + session.ExtensionTimeMinutes

		// Проверяем, прошло ли время с момента начала сессии (за 1 минуту до завершения)
		if now.Sub(startTime) >= time.Duration(totalTime-1)*time.Minute && now.Sub(startTime) < time.Duration(totalTime)*time.Minute {
			// Если прошло 4 минуты и уведомление еще не отправлено, отправляем его
			if !session.IsCompletingNotificationSent {
				// Получаем пользователя
				user, err := s.userService.GetUserByID(session.UserID)
				if err != nil {
					continue
				}

				// Отправляем уведомление
				err = s.telegramBot.SendSessionNotification(user.TelegramID, telegram.NotificationTypeSessionCompletingSoon)
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
		CarNumber:            req.CarNumber,            // Используем переданный номер машины или пустую строку
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
	if session.WithChemistry && session.RequestedExtensionChemistryTimeMinutes > 0 {
		// Применяем логику химии при продлении
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
		err = s.modbusService.WriteChemistryCoil(*session.BoxID, true)
		if err != nil {
			// Логируем ошибку (HandleModbusError теперь только логирует, не продлевает время)
			s.modbusService.HandleModbusError(*session.BoxID, "chemistry_on", session.ID, err)
			logger.Printf("Ошибка включения химии в боксе %s: %v", *session.BoxID, err)
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
				err := s.modbusService.WriteChemistryCoil(*session.BoxID, false)
				if err != nil {
					logger.Printf("AutoDisableChemistry: ошибка автовыключения химии через Modbus: %v, SessionID=%s, BoxID=%s", err, sessionID, *session.BoxID)
				} else {
					logger.Printf("AutoDisableChemistry: химия успешно выключена через Modbus, SessionID=%s, BoxID=%s", sessionID, *session.BoxID)
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
