package service

import (
	"carwash_backend/internal/domain/session/models"
	"carwash_backend/internal/domain/session/repository"
	"carwash_backend/internal/domain/telegram"
	"carwash_backend/internal/domain/user/service"
	washboxModels "carwash_backend/internal/domain/washbox/models"
	washboxService "carwash_backend/internal/domain/washbox/service"
	paymentService "carwash_backend/internal/domain/payment/service"
	paymentModels "carwash_backend/internal/domain/payment/models"
	"fmt"
	"time"

	"github.com/google/uuid"
)

// Service интерфейс для бизнес-логики сессий
type Service interface {
	CreateSession(req *models.CreateSessionRequest) (*models.Session, error)
	CreateSessionWithPayment(req *models.CreateSessionWithPaymentRequest) (*models.CreateSessionWithPaymentResponse, error)
	GetUserSession(req *models.GetUserSessionRequest) (*models.GetUserSessionResponse, error)
	GetSession(req *models.GetSessionRequest) (*models.GetSessionResponse, error)
	StartSession(req *models.StartSessionRequest) (*models.Session, error)
	CompleteSession(req *models.CompleteSessionRequest) (*models.Session, error)
	ExtendSession(req *models.ExtendSessionRequest) (*models.Session, error)
	UpdateSessionStatus(sessionID uuid.UUID, status string) error
	ProcessQueue() error
	CheckAndCompleteExpiredSessions() error
	CheckAndExpireReservedSessions() error
	CheckAndNotifyExpiringReservedSessions() error
	CheckAndNotifyCompletingSessions() error
	CountSessionsByStatus(status string) (int, error)
	GetSessionsByStatus(status string) ([]models.Session, error)
	GetUserSessionHistory(req *models.GetUserSessionHistoryRequest) ([]models.Session, error)

	// Административные методы
	AdminListSessions(req *models.AdminListSessionsRequest) (*models.AdminListSessionsResponse, error)
	AdminGetSession(req *models.AdminGetSessionRequest) (*models.AdminGetSessionResponse, error)
}

// ServiceImpl реализация Service
type ServiceImpl struct {
	repo           repository.Repository
	washboxService washboxService.Service
	userService    service.Service
	telegramBot    telegram.NotificationService
	paymentService paymentService.Service
}

// NewService создает новый экземпляр Service
func NewService(repo repository.Repository, washboxService washboxService.Service, userService service.Service, telegramBot telegram.NotificationService, paymentService paymentService.Service) *ServiceImpl {
	return &ServiceImpl{
		repo:           repo,
		washboxService: washboxService,
		userService:    userService,
		telegramBot:    telegramBot,
		paymentService: paymentService,
	}
}

// CreateSession создает новую сессию
func (s *ServiceImpl) CreateSession(req *models.CreateSessionRequest) (*models.Session, error) {
	// Проверяем идемпотентность запроса
	existingSessionByKey, err := s.repo.GetSessionByIdempotencyKey(req.IdempotencyKey)
	if err == nil && existingSessionByKey != nil {
		// Сессия с таким ключом идемпотентности уже существует
		return existingSessionByKey, nil
	}

	// Проверяем, есть ли у пользователя активная сессия
	existingSession, err := s.repo.GetActiveSessionByUserID(req.UserID)
	if err == nil && existingSession != nil {
		// У пользователя уже есть активная сессия
		return existingSession, nil
	}

	// Создаем новую сессию
	now := time.Now()
	session := &models.Session{
		UserID:            req.UserID,
		Status:            models.SessionStatusCreated,
		ServiceType:       req.ServiceType,
		WithChemistry:     req.WithChemistry,
		CarNumber:         req.CarNumber,
		RentalTimeMinutes: req.RentalTimeMinutes,
		IdempotencyKey:    req.IdempotencyKey,
		StatusUpdatedAt:   now, // Инициализируем время изменения статуса
	}

	// Сохраняем сессию в базе данных
	err = s.repo.CreateSession(session)
	if err != nil {
		return nil, err
	}

	return session, nil
}

// CreateSessionWithPayment создает сессию с платежом
func (s *ServiceImpl) CreateSessionWithPayment(req *models.CreateSessionWithPaymentRequest) (*models.CreateSessionWithPaymentResponse, error) {
	// 1. Создаем сессию
	session, err := s.CreateSession(&models.CreateSessionRequest{
		UserID:            req.UserID,
		ServiceType:       req.ServiceType,
		WithChemistry:     req.WithChemistry,
		CarNumber:         req.CarNumber,
		RentalTimeMinutes: req.RentalTimeMinutes,
		IdempotencyKey:    req.IdempotencyKey,
	})
	if err != nil {
		return nil, fmt.Errorf("ошибка создания сессии: %w", err)
	}

	// 2. Рассчитываем цену через Payment Service
	priceResp, err := s.paymentService.CalculatePrice(&paymentModels.CalculatePriceRequest{
		ServiceType:       req.ServiceType,
		WithChemistry:     req.WithChemistry,
		RentalTimeMinutes: req.RentalTimeMinutes,
	})
	if err != nil {
		return nil, fmt.Errorf("ошибка расчета цены: %w", err)
	}

	// 3. Создаем платеж через Payment Service
	paymentResp, err := s.paymentService.CreatePayment(&paymentModels.CreatePaymentRequest{
		SessionID: session.ID,
		Amount:    priceResp.Price,
		Currency:  priceResp.Currency,
	})
	if err != nil {
		return nil, fmt.Errorf("ошибка создания платежа: %w", err)
	}

	// 4. Обновляем сессию с payment_id
	session.PaymentID = &paymentResp.Payment.ID
	err = s.repo.UpdateSession(session)
	if err != nil {
		return nil, fmt.Errorf("ошибка обновления сессии: %w", err)
	}

	// 5. Формируем ответ с информацией о платеже
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

	// Получаем информацию о платеже, если она существует
	var payment *models.Payment
	if session != nil && session.PaymentID != nil {
		paymentResp, err := s.paymentService.GetPaymentByID(*session.PaymentID)
		if err == nil && paymentResp != nil {
			payment = &models.Payment{
				ID:         paymentResp.ID,
				SessionID:  paymentResp.SessionID,
				Amount:     paymentResp.Amount,
				Currency:   paymentResp.Currency,
				Status:     paymentResp.Status,
				PaymentURL: paymentResp.PaymentURL,
				TinkoffID:  paymentResp.TinkoffID,
				ExpiresAt:  paymentResp.ExpiresAt,
				CreatedAt:  paymentResp.CreatedAt,
				UpdatedAt:  paymentResp.UpdatedAt,
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

	// Получаем информацию о платеже, если она существует
	var payment *models.Payment
	if session != nil && session.PaymentID != nil {
		paymentResp, err := s.paymentService.GetPaymentByID(*session.PaymentID)
		if err == nil && paymentResp != nil {
			payment = &models.Payment{
				ID:         paymentResp.ID,
				SessionID:  paymentResp.SessionID,
				Amount:     paymentResp.Amount,
				Currency:   paymentResp.Currency,
				Status:     paymentResp.Status,
				PaymentURL: paymentResp.PaymentURL,
				TinkoffID:  paymentResp.TinkoffID,
				ExpiresAt:  paymentResp.ExpiresAt,
				CreatedAt:  paymentResp.CreatedAt,
				UpdatedAt:  paymentResp.UpdatedAt,
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

	// Получаем информацию о платеже, если она существует
	if session.PaymentID != nil {
		paymentResp, err := s.paymentService.GetPaymentByID(*session.PaymentID)
		if err == nil && paymentResp != nil {
			session.Payment = &models.Payment{
				ID:         paymentResp.ID,
				SessionID:  paymentResp.SessionID,
				Amount:     paymentResp.Amount,
				Currency:   paymentResp.Currency,
				Status:     paymentResp.Status,
				PaymentURL: paymentResp.PaymentURL,
				TinkoffID:  paymentResp.TinkoffID,
				ExpiresAt:  paymentResp.ExpiresAt,
				CreatedAt:  paymentResp.CreatedAt,
				UpdatedAt:  paymentResp.UpdatedAt,
			}
		}
	}

	return session, nil
}

// CompleteSession завершает сессию (переводит в статус complete)
func (s *ServiceImpl) CompleteSession(req *models.CompleteSessionRequest) (*models.Session, error) {
	// Получаем сессию по ID
	session, err := s.repo.GetSessionByID(req.SessionID)
	if err != nil {
		return nil, err
	}

	// Проверяем, что сессия в статусе active
	if session.Status != models.SessionStatusActive {
		return session, nil // Возвращаем сессию без изменений
	}

	// Проверяем, что у сессии есть назначенный бокс
	if session.BoxID == nil {
		return session, nil // Возвращаем сессию без изменений
	}

	// Если сервис боксов не инициализирован, просто обновляем статус сессии
	if s.washboxService == nil {
		// Обновляем статус сессии на complete
		session.Status = models.SessionStatusComplete
		err = s.repo.UpdateSession(session)
		if err != nil {
			return nil, err
		}
		return session, nil
	}

	// Обновляем статус бокса на free
	err = s.washboxService.UpdateWashBoxStatus(*session.BoxID, washboxModels.StatusFree)
	if err != nil {
		return nil, err
	}

	// Обновляем статус сессии на complete, время обновления статуса и сбрасываем флаг уведомления
	session.Status = models.SessionStatusComplete
	session.StatusUpdatedAt = time.Now()         // Обновляем время изменения статуса
	session.IsCompletingNotificationSent = false // Сбрасываем флаг, чтобы уведомление могло быть отправлено снова
	err = s.repo.UpdateSession(session)
	if err != nil {
		return nil, err
	}

	// Получаем информацию о платеже, если она существует
	if session.PaymentID != nil {
		paymentResp, err := s.paymentService.GetPaymentByID(*session.PaymentID)
		if err == nil && paymentResp != nil {
			session.Payment = &models.Payment{
				ID:         paymentResp.ID,
				SessionID:  paymentResp.SessionID,
				Amount:     paymentResp.Amount,
				Currency:   paymentResp.Currency,
				Status:     paymentResp.Status,
				PaymentURL: paymentResp.PaymentURL,
				TinkoffID:  paymentResp.TinkoffID,
				ExpiresAt:  paymentResp.ExpiresAt,
				CreatedAt:  paymentResp.CreatedAt,
				UpdatedAt:  paymentResp.UpdatedAt,
			}
		}
	}

	return session, nil
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

	// Обновляем сессию в базе данных
	err = s.repo.UpdateSession(session)
	if err != nil {
		return nil, err
	}

	// Получаем информацию о платеже, если она существует
	if session.PaymentID != nil {
		paymentResp, err := s.paymentService.GetPaymentByID(*session.PaymentID)
		if err == nil && paymentResp != nil {
			session.Payment = &models.Payment{
				ID:         paymentResp.ID,
				SessionID:  paymentResp.SessionID,
				Amount:     paymentResp.Amount,
				Currency:   paymentResp.Currency,
				Status:     paymentResp.Status,
				PaymentURL: paymentResp.PaymentURL,
				TinkoffID:  paymentResp.TinkoffID,
				ExpiresAt:  paymentResp.ExpiresAt,
				CreatedAt:  paymentResp.CreatedAt,
				UpdatedAt:  paymentResp.UpdatedAt,
			}
		}
	}

	return session, nil
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

		// Получаем свободные боксы для данного типа услуги
		freeBoxes, err := s.washboxService.GetFreeWashBoxesByServiceType(session.ServiceType)
		if err != nil {
			return err
		}

		// Если нет свободных боксов для данного типа услуги, пропускаем сессию
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
				session.IsExpiringNotificationSent = true
				err = s.repo.UpdateSession(&session)
				if err != nil {
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
				session.IsCompletingNotificationSent = true
				err = s.repo.UpdateSession(&session)
				if err != nil {
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
		if sessions[i].PaymentID != nil {
			paymentResp, err := s.paymentService.GetPaymentByID(*sessions[i].PaymentID)
			if err == nil && paymentResp != nil {
				sessions[i].Payment = &models.Payment{
					ID:         paymentResp.ID,
					SessionID:  paymentResp.SessionID,
					Amount:     paymentResp.Amount,
					Currency:   paymentResp.Currency,
					Status:     paymentResp.Status,
					PaymentURL: paymentResp.PaymentURL,
					TinkoffID:  paymentResp.TinkoffID,
					ExpiresAt:  paymentResp.ExpiresAt,
					CreatedAt:  paymentResp.CreatedAt,
					UpdatedAt:  paymentResp.UpdatedAt,
				}
			}
		}
	}

	return sessions, nil
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
