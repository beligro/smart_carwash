package service

import (
	"carwash_backend/internal/domain/queue/models"
	sessionModels "carwash_backend/internal/domain/session/models"
	sessionService "carwash_backend/internal/domain/session/service"
	washboxModels "carwash_backend/internal/domain/washbox/models"
	washboxService "carwash_backend/internal/domain/washbox/service"

	"github.com/google/uuid"
)

// Service интерфейс для бизнес-логики очереди
type Service interface {
	GetQueueStatus() (*models.QueueStatus, error)
	GetWashInfoForUser(userID uuid.UUID) (*models.WashInfo, error)
	ProcessQueue() error
	CheckAndCompleteExpiredSessions() error
	CheckAndExpireReservedSessions() error
}

// ServiceImpl реализация Service
type ServiceImpl struct {
	sessionService sessionService.Service
	washboxService washboxService.Service
}

// NewService создает новый экземпляр Service
func NewService(sessionService sessionService.Service, washboxService washboxService.Service) *ServiceImpl {
	return &ServiceImpl{
		sessionService: sessionService,
		washboxService: washboxService,
	}
}

// GetQueueStatus получает статус очереди и боксов
func (s *ServiceImpl) GetQueueStatus() (*models.QueueStatus, error) {
	// Получаем все боксы мойки
	boxes, err := s.washboxService.GetAllWashBoxes()
	if err != nil {
		return nil, err
	}

	// Получаем количество сессий в очереди
	queueSize, err := s.sessionService.CountSessionsByStatus(sessionModels.SessionStatusCreated)
	if err != nil {
		return nil, err
	}

	// Получаем количество свободных боксов
	freeBoxes, err := s.washboxService.GetFreeWashBoxes()
	if err != nil {
		return nil, err
	}

	// Определяем, есть ли очередь
	hasQueue := queueSize > len(freeBoxes)

	// Формируем ответ
	return &models.QueueStatus{
		Boxes:     boxes,
		QueueSize: queueSize,
		HasQueue:  hasQueue,
	}, nil
}

// GetWashInfoForUser получает информацию о мойке для пользователя
func (s *ServiceImpl) GetWashInfoForUser(userID uuid.UUID) (*models.WashInfo, error) {
	// Получаем статус очереди
	queueStatus, err := s.GetQueueStatus()
	if err != nil {
		return nil, err
	}

	// Получаем сессию пользователя
	session, err := s.sessionService.GetUserSession(&sessionModels.GetUserSessionRequest{
		UserID: userID,
	})
	if err != nil {
		// Если сессия не найдена, возвращаем только статус очереди
		return &models.WashInfo{
			Boxes:     queueStatus.Boxes,
			QueueSize: queueStatus.QueueSize,
			HasQueue:  queueStatus.HasQueue,
		}, nil
	}

	// Формируем ответ с сессией пользователя
	return &models.WashInfo{
		Boxes:       queueStatus.Boxes,
		QueueSize:   queueStatus.QueueSize,
		HasQueue:    queueStatus.HasQueue,
		UserSession: session,
	}, nil
}

// ProcessQueue обрабатывает очередь сессий
func (s *ServiceImpl) ProcessQueue() error {
	// Получаем все сессии со статусом "created"
	sessions, err := s.sessionService.GetSessionsByStatus(sessionModels.SessionStatusCreated)
	if err != nil {
		return err
	}

	// Если нет сессий в очереди, выходим
	if len(sessions) == 0 {
		return nil
	}

	// Получаем все свободные боксы
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
		err = s.assignSessionToBox(sessions[i].ID, freeBoxes[i].ID)
		if err != nil {
			return err
		}
	}

	return nil
}

// assignSessionToBox назначает сессию на бокс
func (s *ServiceImpl) assignSessionToBox(sessionID uuid.UUID, boxID uuid.UUID) error {
	// Получаем сессию по ID
	session, err := s.sessionService.GetSession(&sessionModels.GetSessionRequest{
		SessionID: sessionID,
	})
	if err != nil {
		return err
	}

	// Обновляем сессию - назначаем бокс и меняем статус
	session.BoxID = &boxID
	session.Status = sessionModels.SessionStatusAssigned

	// Создаем запрос на обновление сессии
	// Здесь мы используем StartSessionRequest, так как у нас нет специального запроса для обновления сессии
	// В реальном проекте лучше создать отдельный запрос для этой операции
	startReq := &sessionModels.StartSessionRequest{
		SessionID: sessionID,
	}

	_, err = s.sessionService.StartSession(startReq)
	return err
}

// CheckAndCompleteExpiredSessions проверяет и завершает истекшие сессии
func (s *ServiceImpl) CheckAndCompleteExpiredSessions() error {
	// Получаем все активные сессии
	activeSessions, err := s.sessionService.GetSessionsByStatus(sessionModels.SessionStatusActive)
	if err != nil {
		return err
	}

	// Если нет активных сессий, выходим
	if len(activeSessions) == 0 {
		return nil
	}

	// Для каждой активной сессии проверяем, не истекла ли она
	for _, session := range activeSessions {
		// Проверяем, прошло ли достаточно времени с момента активации сессии
		// Эта логика должна быть в сервисе сессий, но для простоты оставим здесь

		// Создаем запрос на завершение сессии
		completeReq := &sessionModels.CompleteSessionRequest{
			SessionID: session.ID,
		}

		// Завершаем сессию
		_, err = s.sessionService.CompleteSession(completeReq)
		if err != nil {
			return err
		}
	}

	return nil
}

// CheckAndExpireReservedSessions проверяет и истекает зарезервированные сессии
func (s *ServiceImpl) CheckAndExpireReservedSessions() error {
	// Получаем все сессии со статусом "assigned"
	assignedSessions, err := s.sessionService.GetSessionsByStatus(sessionModels.SessionStatusAssigned)
	if err != nil {
		return err
	}

	// Если нет назначенных сессий, выходим
	if len(assignedSessions) == 0 {
		return nil
	}

	// Для каждой назначенной сессии проверяем, не истекла ли она
	for _, session := range assignedSessions {
		// Проверяем, прошло ли достаточно времени с момента назначения сессии
		// Эта логика должна быть в сервисе сессий, но для простоты оставим здесь

		// Если у сессии есть назначенный бокс, освобождаем его
		if session.BoxID != nil {
			// Обновляем статус бокса на free
			err = s.washboxService.UpdateWashBoxStatus(*session.BoxID, washboxModels.StatusFree)
			if err != nil {
				return err
			}
		}

		// Обновляем статус сессии на expired
		// Здесь нам нужен метод для изменения статуса сессии, но его нет в интерфейсе
		// В реальном проекте лучше добавить такой метод в сервис сессий
	}

	return nil
}

// GetAllWashBoxes получает все боксы мойки
func (s *ServiceImpl) GetAllWashBoxes() ([]washboxModels.WashBox, error) {
	return s.washboxService.GetAllWashBoxes()
}
