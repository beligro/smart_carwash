package service

import (
	"context"
	"math"
	"time"

	"carwash_backend/internal/domain/queue/models"
	sessionModels "carwash_backend/internal/domain/session/models"
	sessionService "carwash_backend/internal/domain/session/service"
	userModels "carwash_backend/internal/domain/user/models"
	userService "carwash_backend/internal/domain/user/service"
	washboxModels "carwash_backend/internal/domain/washbox/models"
	washboxService "carwash_backend/internal/domain/washbox/service"
	"carwash_backend/internal/metrics"
)

// Service интерфейс для бизнес-логики очереди
type Service interface {
	GetQueueStatus(ctx context.Context) (*models.QueueStatus, error)

	// Административные методы
	AdminGetQueueStatus(ctx context.Context, req *models.AdminQueueStatusRequest) (*models.AdminQueueStatusResponse, error)
}

// ServiceImpl реализация Service
type ServiceImpl struct {
	sessionService sessionService.Service
	washboxService washboxService.Service
	userService    userService.Service
	metrics        *metrics.Metrics
}

// NewService создает новый экземпляр Service
func NewService(sessionService sessionService.Service, washboxService washboxService.Service, userService userService.Service, metrics *metrics.Metrics) *ServiceImpl {
	return &ServiceImpl{
		sessionService: sessionService,
		washboxService: washboxService,
		userService:    userService,
		metrics:        metrics,
	}
}

// getServiceQueueInfo получает информацию об очереди для конкретного типа услуги
func (s *ServiceImpl) getServiceQueueInfo(ctx context.Context, serviceType string) (*models.ServiceQueueInfo, error) {
	// Получаем боксы для данного типа услуги
	boxes, err := s.washboxService.GetWashBoxesByServiceType(ctx, serviceType)
	if err != nil {
		return nil, err
	}

	// Получаем сессии со статусом "created"
	createdSessions, err := s.sessionService.GetSessionsByStatus(ctx, sessionModels.SessionStatusInQueue)
	if err != nil {
		return nil, err
	}

	// Подсчитываем количество сессий для данного типа услуги и собираем пользователей в очереди
	queueSize := 0
	var usersInQueue []models.QueueUser

	for _, session := range createdSessions {
		if session.ServiceType == serviceType {
			queueSize++

			// Получаем информацию о пользователе
			user, err := s.userService.GetUserByID(ctx, session.UserID)
			if err != nil {
				// Если не удалось получить пользователя, используем базовую информацию
				user = &userModels.User{
					ID:        session.UserID,
					Username:  "Неизвестный пользователь",
					FirstName: "Имя",
					LastName:  "Фамилия",
				}
			}

			// Добавляем пользователя в очередь
			queueUser := models.QueueUser{
				UserID:       session.UserID.String(),
				Username:     user.Username,
				FirstName:    user.FirstName,
				LastName:     user.LastName,
				ServiceType:  serviceType,
				Position:     queueSize, // Позиция в очереди
				WaitingSince: session.CreatedAt.Format("2006-01-02 15:04:05"),
				CarNumber:    session.CarNumber, // Номер машины из сессии
			}
			usersInQueue = append(usersInQueue, queueUser)
		}
	}

	// Получаем свободные боксы для данного типа услуги
	freeBoxes, err := s.washboxService.GetFreeWashBoxesByServiceType(ctx, serviceType)
	if err != nil {
		return nil, err
	}

	// Определяем, есть ли очередь
	hasQueue := queueSize > len(freeBoxes)

	// Вычисляем время ожидания, если есть очередь
	var waitTimeMinutes *int
	if hasQueue && queueSize > 0 {
		activeSessions, err := s.sessionService.GetSessionsByStatus(ctx, sessionModels.SessionStatusActive)
		if err != nil {
			return nil, err
		}

		// Фильтруем сессии по типу услуги и вычисляем оставшееся время
		minRemainingTime := -1.0
		now := time.Now()

		for _, session := range activeSessions {
			if session.ServiceType == serviceType {
				// Общее время сессии в минутах
				totalTimeMinutes := session.RentalTimeMinutes + session.ExtensionTimeMinutes

				// Время старта активной сессии (когда статус стал "active")
				sessionStartTime := session.StatusUpdatedAt

				// Прошедшее время с момента старта сессии в минутах
				elapsedMinutes := now.Sub(sessionStartTime).Minutes()

				// Оставшееся время в минутах
				remainingMinutes := float64(totalTimeMinutes) - elapsedMinutes

				// Если время еще осталось, учитываем его
				if remainingMinutes > 0 {
					if minRemainingTime == -1 || remainingMinutes < minRemainingTime {
						minRemainingTime = remainingMinutes
					}
				}
			}
		}

		// Если нашли хотя бы одну активную сессию с оставшимся временем, округляем вверх
		if minRemainingTime > 0 {
			roundedTime := int(math.Ceil(minRemainingTime))
			waitTimeMinutes = &roundedTime
		}
	}

	// Формируем ответ
	return &models.ServiceQueueInfo{
		ServiceType:     serviceType,
		Boxes:           boxes,
		QueueSize:       queueSize,
		HasQueue:        hasQueue,
		UsersInQueue:    usersInQueue,
		WaitTimeMinutes: waitTimeMinutes,
	}, nil
}

// GetQueueStatus получает статус очереди и боксов
func (s *ServiceImpl) GetQueueStatus(ctx context.Context) (*models.QueueStatus, error) {
	// Получаем все боксы мойки
	allBoxes, err := s.washboxService.GetAllWashBoxes(ctx)
	if err != nil {
		return nil, err
	}

	// Получаем информацию об очереди для каждого типа услуги
	washQueueInfo, err := s.getServiceQueueInfo(ctx, washboxModels.ServiceTypeWash)
	if err != nil {
		return nil, err
	}

	airDryQueueInfo, err := s.getServiceQueueInfo(ctx, washboxModels.ServiceTypeAirDry)
	if err != nil {
		return nil, err
	}

	vacuumQueueInfo, err := s.getServiceQueueInfo(ctx, washboxModels.ServiceTypeVacuum)
	if err != nil {
		return nil, err
	}

	// Подсчитываем общее количество сессий в очереди
	totalQueueSize := washQueueInfo.QueueSize + airDryQueueInfo.QueueSize + vacuumQueueInfo.QueueSize

	// Определяем, есть ли очередь хотя бы для одного типа услуги
	hasAnyQueue := washQueueInfo.HasQueue || airDryQueueInfo.HasQueue || vacuumQueueInfo.HasQueue

	// Обновляем метрики очереди
	if s.metrics != nil {
		s.metrics.UpdateQueueSize("wash", float64(washQueueInfo.QueueSize))
		s.metrics.UpdateQueueSize("air_dry", float64(airDryQueueInfo.QueueSize))
		s.metrics.UpdateQueueSize("vacuum", float64(vacuumQueueInfo.QueueSize))
	}

	// Формируем ответ
	return &models.QueueStatus{
		AllBoxes:       allBoxes,
		WashQueue:      *washQueueInfo,
		AirDryQueue:    *airDryQueueInfo,
		VacuumQueue:    *vacuumQueueInfo,
		TotalQueueSize: totalQueueSize,
		HasAnyQueue:    hasAnyQueue,
	}, nil
}

// AdminGetQueueStatus получает детальный статус очереди для администратора
func (s *ServiceImpl) AdminGetQueueStatus(ctx context.Context, req *models.AdminQueueStatusRequest) (*models.AdminQueueStatusResponse, error) {
	// Получаем базовый статус очереди
	queueStatus, err := s.GetQueueStatus(ctx)
	if err != nil {
		return nil, err
	}

	response := &models.AdminQueueStatusResponse{
		QueueStatus: *queueStatus,
	}

	// Если запрошены детали, добавляем их
	if req.IncludeDetails {
		details, err := s.getQueueDetails(ctx)
		if err != nil {
			return nil, err
		}
		response.Details = details
	}

	return response, nil
}

// getQueueDetails получает детальную информацию об очереди
func (s *ServiceImpl) getQueueDetails(ctx context.Context) (*models.QueueDetails, error) {
	// Получаем все сессии со статусом "created" (в очереди)
	createdSessions, err := s.sessionService.GetSessionsByStatus(ctx, sessionModels.SessionStatusCreated)
	if err != nil {
		return nil, err
	}

	// Группируем сессии по типу услуги
	sessionsByType := make(map[string][]sessionModels.Session)
	for _, session := range createdSessions {
		sessionsByType[session.ServiceType] = append(sessionsByType[session.ServiceType], session)
	}

	var usersInQueue []models.QueueUser
	var queueOrder []string

	// Обрабатываем каждый тип услуги
	for serviceType, sessions := range sessionsByType {
		for i, session := range sessions {
			// Здесь нужно получить информацию о пользователе
			// Пока используем базовую информацию
			user := models.QueueUser{
				UserID:       session.UserID.String(),
				Username:     "Пользователь", // Нужно получить из user service
				FirstName:    "Имя",          // Нужно получить из user service
				LastName:     "Фамилия",      // Нужно получить из user service
				ServiceType:  serviceType,
				Position:     i + 1,
				WaitingSince: session.CreatedAt.Format("2006-01-02 15:04:05"),
				CarNumber:    session.CarNumber, // Номер машины из сессии
			}
			usersInQueue = append(usersInQueue, user)
			queueOrder = append(queueOrder, session.UserID.String())
		}
	}

	return &models.QueueDetails{
		UsersInQueue: usersInQueue,
		QueueOrder:   queueOrder,
	}, nil
}
