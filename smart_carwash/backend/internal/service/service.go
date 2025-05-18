package service

import (
	"carwash_backend/internal/models"
	"carwash_backend/internal/repository"
	"time"

	"github.com/google/uuid"
)

// Service интерфейс для бизнес-логики
type Service interface {
	// Пользователи
	CreateUser(req *models.CreateUserRequest) (*models.User, error)
	GetUserByTelegramID(telegramID int64) (*models.User, error)

	// Информация о мойке
	GetWashInfo() (*models.WashInfo, error)
	GetWashInfoForUser(userID uuid.UUID) (*models.WashInfo, error)
	GetQueueStatus() (*models.GetQueueStatusResponse, error)

	// Сессии мойки
	CreateSession(req *models.CreateSessionRequest) (*models.Session, error)
	GetUserSession(req *models.GetUserSessionRequest) (*models.Session, error)
	GetSession(req *models.GetSessionRequest) (*models.Session, error)
	StartSession(req *models.StartSessionRequest) (*models.Session, error)
	CompleteSession(req *models.CompleteSessionRequest) (*models.Session, error)
	ProcessQueue() error
	CheckAndCompleteExpiredSessions() error
	CheckAndExpireReservedSessions() error
}

// ServiceImpl реализация Service
type ServiceImpl struct {
	repo repository.Repository
}

// NewService создает новый экземпляр Service
func NewService(repo repository.Repository) *ServiceImpl {
	return &ServiceImpl{repo: repo}
}

// CreateUser создает нового пользователя
func (s *ServiceImpl) CreateUser(req *models.CreateUserRequest) (*models.User, error) {
	// Проверяем, существует ли пользователь
	existingUser, err := s.repo.GetUserByTelegramID(req.TelegramID)
	if err == nil {
		// Пользователь уже существует, возвращаем его
		return existingUser, nil
	}

	// Создаем нового пользователя
	user := &models.User{
		TelegramID: req.TelegramID,
		Username:   req.Username,
		FirstName:  req.FirstName,
		LastName:   req.LastName,
		IsAdmin:    false, // По умолчанию пользователь не админ
	}

	// Сохраняем пользователя в базе данных
	err = s.repo.CreateUser(user)
	if err != nil {
		return nil, err
	}

	return user, nil
}

// GetUserByTelegramID получает пользователя по Telegram ID
func (s *ServiceImpl) GetUserByTelegramID(telegramID int64) (*models.User, error) {
	return s.repo.GetUserByTelegramID(telegramID)
}

// GetWashInfo получает информацию о мойке
func (s *ServiceImpl) GetWashInfo() (*models.WashInfo, error) {
	// Получаем все боксы мойки
	boxes, err := s.repo.GetAllWashBoxes()
	if err != nil {
		return nil, err
	}

	// Получаем количество сессий в очереди
	queueSize, err := s.repo.CountSessionsByStatus(models.SessionStatusCreated)
	if err != nil {
		return nil, err
	}

	// Получаем количество свободных боксов
	freeBoxes, err := s.repo.GetFreeWashBoxes()
	if err != nil {
		return nil, err
	}

	// Определяем, есть ли очередь
	hasQueue := queueSize > len(freeBoxes)

	// Формируем ответ
	return &models.WashInfo{
		Boxes:     boxes,
		QueueSize: queueSize,
		HasQueue:  hasQueue,
	}, nil
}

// GetWashInfoForUser получает информацию о мойке для конкретного пользователя
func (s *ServiceImpl) GetWashInfoForUser(userID uuid.UUID) (*models.WashInfo, error) {
	// Получаем общую информацию о мойке
	washInfo, err := s.GetWashInfo()
	if err != nil {
		return nil, err
	}

	// Получаем активную сессию пользователя, если есть
	session, err := s.repo.GetActiveSessionByUserID(userID)
	if err == nil {
		// Если сессия найдена, добавляем ее в ответ
		washInfo.UserSession = session
	}

	return washInfo, nil
}

// GetQueueStatus получает статус очереди и боксов
func (s *ServiceImpl) GetQueueStatus() (*models.GetQueueStatusResponse, error) {
	// Получаем все боксы мойки
	boxes, err := s.repo.GetAllWashBoxes()
	if err != nil {
		return nil, err
	}

	// Получаем количество сессий в очереди
	queueSize, err := s.repo.CountSessionsByStatus(models.SessionStatusCreated)
	if err != nil {
		return nil, err
	}

	// Получаем количество свободных боксов
	freeBoxes, err := s.repo.GetFreeWashBoxes()
	if err != nil {
		return nil, err
	}

	// Определяем, есть ли очередь
	hasQueue := queueSize > len(freeBoxes)

	// Формируем ответ
	return &models.GetQueueStatusResponse{
		Boxes:     boxes,
		QueueSize: queueSize,
		HasQueue:  hasQueue,
	}, nil
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

	// Получаем информацию о боксе
	box, err := s.repo.GetWashBoxByID(*session.BoxID)
	if err != nil {
		return nil, err
	}

	// Обновляем статус бокса на busy
	err = s.repo.UpdateWashBoxStatus(*session.BoxID, models.StatusBusy)
	if err != nil {
		return nil, err
	}

	// Обновляем статус сессии на active
	session.Status = models.SessionStatusActive
	err = s.repo.UpdateSession(session)
	if err != nil {
		// Если не удалось обновить сессию, возвращаем статус бокса обратно
		s.repo.UpdateWashBoxStatus(*session.BoxID, box.Status)
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

	// Обновляем статус бокса на free
	err = s.repo.UpdateWashBoxStatus(*session.BoxID, models.StatusFree)
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
			if session.BoxID != nil {
				// Обновляем статус бокса на free
				err = s.repo.UpdateWashBoxStatus(*session.BoxID, models.StatusFree)
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
	// Получаем все сессии со статусом "created"
	sessions, err := s.repo.GetSessionsByStatus(models.SessionStatusCreated)
	if err != nil {
		return err
	}

	// Если нет сессий в очереди, выходим
	if len(sessions) == 0 {
		return nil
	}

	// Получаем все свободные боксы
	freeBoxes, err := s.repo.GetFreeWashBoxes()
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
		err = s.repo.UpdateWashBoxStatus(freeBoxes[i].ID, models.StatusReserved)
		if err != nil {
			return err
		}

		// Обновляем сессию - назначаем бокс и меняем статус
		sessions[i].BoxID = &freeBoxes[i].ID
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
			if session.BoxID != nil {
				// Обновляем статус бокса на free
				err = s.repo.UpdateWashBoxStatus(*session.BoxID, models.StatusFree)
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
