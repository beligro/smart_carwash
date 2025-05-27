package service

import (
	"carwash_backend/internal/domain/session/models"
	"carwash_backend/internal/domain/session/repository"
	"carwash_backend/internal/domain/telegram"
	"carwash_backend/internal/domain/user/service"
	washboxModels "carwash_backend/internal/domain/washbox/models"
	washboxService "carwash_backend/internal/domain/washbox/service"
	"log"
	"time"
)

// Service интерфейс для бизнес-логики сессий
type Service interface {
	CreateSession(req *models.CreateSessionRequest) (*models.Session, error)
	GetUserSession(req *models.GetUserSessionRequest) (*models.Session, error)
	GetSession(req *models.GetSessionRequest) (*models.Session, error)
	StartSession(req *models.StartSessionRequest) (*models.Session, error)
	CompleteSession(req *models.CompleteSessionRequest) (*models.Session, error)
	ProcessQueue() error
	CheckAndCompleteExpiredSessions() error
	CheckAndExpireReservedSessions() error
	CheckAndNotifyExpiringReservedSessions() error
	CheckAndNotifyCompletingSessions() error
	CountSessionsByStatus(status string) (int, error)
	GetSessionsByStatus(status string) ([]models.Session, error)
	GetUserSessionHistory(req *models.GetUserSessionHistoryRequest) ([]models.Session, error)
}

// ServiceImpl реализация Service
type ServiceImpl struct {
	repo           repository.Repository
	washboxService washboxService.Service
	userService    service.Service
	telegramBot    telegram.NotificationService
}

// NewService создает новый экземпляр Service
func NewService(repo repository.Repository, washboxService washboxService.Service, userService service.Service, telegramBot telegram.NotificationService) *ServiceImpl {
	return &ServiceImpl{
		repo:           repo,
		washboxService: washboxService,
		userService:    userService,
		telegramBot:    telegramBot,
	}
}

// CreateSession создает новую сессию мойки
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
	session := &models.Session{
		UserID:         req.UserID,
		Status:         models.SessionStatusCreated,
		IdempotencyKey: req.IdempotencyKey,
	}

	// Сохраняем сессию в базе данных
	err = s.repo.CreateSession(session)
	if err != nil {
		return nil, err
	}

	return session, nil
}

// GetUserSession получает активную сессию пользователя
func (s *ServiceImpl) GetUserSession(req *models.GetUserSessionRequest) (*models.Session, error) {
	return s.repo.GetActiveSessionByUserID(req.UserID)
}

// GetSession получает сессию по ID
func (s *ServiceImpl) GetSession(req *models.GetSessionRequest) (*models.Session, error) {
	return s.repo.GetSessionByID(req.SessionID)
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

	// Обновляем статус сессии на active
	session.Status = models.SessionStatusActive
	err = s.repo.UpdateSession(session)
	if err != nil {
		// Если не удалось обновить сессию, возвращаем статус бокса обратно
		s.washboxService.UpdateWashBoxStatus(*session.BoxID, box.Status)
		return nil, err
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

	// Обновляем статус сессии на complete
	session.Status = models.SessionStatusComplete
	err = s.repo.UpdateSession(session)
	if err != nil {
		return nil, err
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
		startTime := session.UpdatedAt

		// Проверяем, прошло ли 5 минут с момента начала сессии
		if now.Sub(startTime) >= 5*time.Minute {
			// Если прошло 5 минут, завершаем сессию
			if session.BoxID != nil && s.washboxService != nil {
				// Обновляем статус бокса на free
				err = s.washboxService.UpdateWashBoxStatus(*session.BoxID, washboxModels.StatusFree)
				if err != nil {
					return err
				}
			}

			// Обновляем статус сессии на complete
			session.Status = models.SessionStatusComplete
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

	// Получаем все сессии со статусом "created"
	sessions, err := s.repo.GetSessionsByStatus(models.SessionStatusCreated)
	if err != nil {
		return err
	}

	// Если нет сессий в очереди, выходим
	if len(sessions) == 0 {
		return nil
	}

	// Получаем все свободные боксы через сервис washbox
	freeBoxes, err := s.washboxService.GetFreeWashBoxes()
	if err != nil {
		return err
	}

	// Если нет свободных боксов, выходим
	if len(freeBoxes) == 0 {
		return nil
	}

	// Назначаем сессии на свободные боксы
	for i := 0; i < len(sessions) && i < len(freeBoxes); i++ {
		// Обновляем статус бокса на "reserved"
		err = s.washboxService.UpdateWashBoxStatus(freeBoxes[i].ID, washboxModels.StatusReserved)
		if err != nil {
			return err
		}

		// Обновляем сессию - назначаем бокс и меняем статус
		sessions[i].BoxID = &freeBoxes[i].ID
		sessions[i].BoxNumber = &freeBoxes[i].Number
		sessions[i].Status = models.SessionStatusAssigned
		err = s.repo.UpdateSession(&sessions[i])
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
		assignedTime := session.UpdatedAt

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

			// Обновляем статус сессии на expired
			session.Status = models.SessionStatusExpired
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
		assignedTime := session.UpdatedAt

		// Проверяем, прошло ли 2 минуты с момента назначения сессии (за 1 минуту до истечения)
		if now.Sub(assignedTime) >= 2*time.Minute && now.Sub(assignedTime) < 3*time.Minute {
			// Если прошло 2 минуты, отправляем уведомление
			if session.BoxID != nil && session.BoxNumber != nil {
				// Получаем пользователя
				user, err := s.userService.GetUserByID(session.UserID)
				if err != nil {
					log.Printf("Ошибка получения пользователя: %v", err)
					continue
				}

				// Отправляем уведомление
				err = s.telegramBot.SendSessionNotification(user.TelegramID, telegram.NotificationTypeSessionExpiringSoon, *session.BoxNumber)
				if err != nil {
					log.Printf("Ошибка отправки уведомления: %v", err)
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
		startTime := session.UpdatedAt

		// Проверяем, прошло ли 4 минуты с момента начала сессии (за 1 минуту до завершения)
		if now.Sub(startTime) >= 4*time.Minute && now.Sub(startTime) < 5*time.Minute {
			// Если прошло 4 минуты, отправляем уведомление
			if session.BoxID != nil && session.BoxNumber != nil {
				// Получаем пользователя
				user, err := s.userService.GetUserByID(session.UserID)
				if err != nil {
					log.Printf("Ошибка получения пользователя: %v", err)
					continue
				}

				// Отправляем уведомление
				err = s.telegramBot.SendSessionNotification(user.TelegramID, telegram.NotificationTypeSessionCompletingSoon, *session.BoxNumber)
				if err != nil {
					log.Printf("Ошибка отправки уведомления: %v", err)
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

	return s.repo.GetUserSessionHistory(req.UserID, limit, offset)
}
