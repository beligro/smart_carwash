package service

import (
	"fmt"
	"log"

	"github.com/google/uuid"

	"carwash_backend/internal/domain/dahua/models"
	sessionModels "carwash_backend/internal/domain/session/models"
	userModels "carwash_backend/internal/domain/user/models"
	"carwash_backend/internal/utils"
)

// Service представляет сервис для обработки событий от камеры Dahua
type Service interface {
	ProcessANPREvent(req *models.ProcessANPREventRequest) (*models.ProcessANPREventResponse, error)
}

// ServiceImpl реализует интерфейс Service
type ServiceImpl struct {
	userService    UserService
	sessionService SessionService
}

// UserService интерфейс для работы с пользователями
type UserService interface {
	GetUserByCarNumber(carNumber string) (*userModels.User, error)
}

// SessionService интерфейс для работы с сессиями
type SessionService interface {
	GetActiveSessionByUserID(userID uuid.UUID) (*sessionModels.Session, error)
	CompleteSessionWithoutRefund(sessionID uuid.UUID) error
}

// NewService создает новый экземпляр сервиса
func NewService(userService UserService, sessionService SessionService) Service {
	return &ServiceImpl{
		userService:    userService,
		sessionService: sessionService,
	}
}

// ProcessANPREvent обрабатывает ANPR событие от камеры Dahua
func (s *ServiceImpl) ProcessANPREvent(req *models.ProcessANPREventRequest) (*models.ProcessANPREventResponse, error) {
	log.Printf("ProcessANPREvent: обработка события ANPR - LicensePlate=%s, Direction=%s", req.LicensePlate, req.Direction)

	// Проверяем, что это событие выезда
	if req.Direction != "out" {
		log.Printf("ProcessANPREvent: событие не является выездом (direction=%s), пропускаем", req.Direction)
		return &models.ProcessANPREventResponse{
			Success:      true,
			Message:      "Событие не является выездом, обработка не требуется",
			UserFound:    false,
			SessionFound: false,
		}, nil
	}

	// Нормализуем номер от камеры для поиска
	normalizedLicensePlate := utils.NormalizeLicensePlateForSearch(req.LicensePlate)
	log.Printf("ProcessANPREvent: номер нормализован '%s' -> '%s'", req.LicensePlate, normalizedLicensePlate)

	// Ищем пользователя по нормализованному номеру автомобиля
	user, err := s.userService.GetUserByCarNumber(normalizedLicensePlate)
	if err != nil {
		log.Printf("ProcessANPREvent: пользователь с номером %s не найден: %v", req.LicensePlate, err)
		return &models.ProcessANPREventResponse{
			Success:      true,
			Message:      fmt.Sprintf("Пользователь с номером %s не найден", req.LicensePlate),
			UserFound:    false,
			SessionFound: false,
		}, nil
	}

	log.Printf("ProcessANPREvent: пользователь найден - UserID=%s, CarNumber=%s", user.ID, user.CarNumber)

	// Ищем активную сессию пользователя
	activeSession, err := s.sessionService.GetActiveSessionByUserID(user.ID)
	if err != nil {
		log.Printf("ProcessANPREvent: активная сессия для пользователя %s не найдена: %v", user.ID, err)
		return &models.ProcessANPREventResponse{
			Success:       true,
			Message:       fmt.Sprintf("Активная сессия для пользователя %s не найдена", user.CarNumber),
			UserFound:     true,
			SessionFound:  false,
			SessionStatus: "not_found",
		}, nil
	}

	log.Printf("ProcessANPREvent: активная сессия найдена - SessionID=%s, Status=%s", activeSession.ID, activeSession.Status)

	// Завершаем сессию БЕЗ частичного возврата
	err = s.sessionService.CompleteSessionWithoutRefund(activeSession.ID)
	if err != nil {
		log.Printf("ProcessANPREvent: ошибка завершения сессии %s: %v", activeSession.ID, err)
		return &models.ProcessANPREventResponse{
			Success:       false,
			Message:       fmt.Sprintf("Ошибка завершения сессии: %v", err),
			UserFound:     true,
			SessionFound:  true,
			SessionID:     activeSession.ID.String(),
			SessionStatus: activeSession.Status,
		}, err
	}

	log.Printf("ProcessANPREvent: сессия успешно завершена БЕЗ возврата - SessionID=%s, UserID=%s, CarNumber=%s", 
		activeSession.ID, user.ID, req.LicensePlate)

	return &models.ProcessANPREventResponse{
		Success:       true,
		Message:       fmt.Sprintf("Сессия успешно завершена для автомобиля %s", req.LicensePlate),
		UserFound:     true,
		SessionFound:  true,
		SessionID:     activeSession.ID.String(),
		SessionStatus: "completed",
	}, nil
}
