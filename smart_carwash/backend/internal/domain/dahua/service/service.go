package service

import (
	"context"
	"fmt"

	"github.com/google/uuid"
	"github.com/sirupsen/logrus"

	"carwash_backend/internal/domain/dahua/models"
	sessionModels "carwash_backend/internal/domain/session/models"
	"carwash_backend/internal/logger"
	"carwash_backend/internal/utils"
)

// Service представляет сервис для обработки событий от камеры Dahua
type Service interface {
	ProcessANPREvent(ctx context.Context, req *models.ProcessANPREventRequest) (*models.ProcessANPREventResponse, error)
}

// ServiceImpl реализует интерфейс Service
type ServiceImpl struct {
	sessionService SessionService
}

// SessionService интерфейс для работы с сессиями
type SessionService interface {
	GetActiveSessionByUserID(ctx context.Context, userID uuid.UUID) (*sessionModels.Session, error)
	GetActiveSessionByCarNumber(ctx context.Context, carNumber string) (*sessionModels.Session, error)
	CompleteSessionWithoutRefund(ctx context.Context, sessionID uuid.UUID) error
}

// NewService создает новый экземпляр сервиса
func NewService(sessionService SessionService) Service {
	return &ServiceImpl{
		sessionService: sessionService,
	}
}

// ProcessANPREvent обрабатывает ANPR событие от камеры Dahua
func (s *ServiceImpl) ProcessANPREvent(ctx context.Context, req *models.ProcessANPREventRequest) (*models.ProcessANPREventResponse, error) {
	logger.WithFields(logrus.Fields{
		"service":      "dahua",
		"method":       "ProcessANPREvent",
		"license_plate": req.LicensePlate,
		"direction":    req.Direction,
	}).Info("Обработка события ANPR")

	// Проверяем, что это событие выезда
	if req.Direction != "out" {
		logger.WithFields(logrus.Fields{
			"service":      "dahua",
			"method":       "ProcessANPREvent",
			"license_plate": req.LicensePlate,
			"direction":    req.Direction,
		}).Info("Событие не является выездом, пропускаем")
		return &models.ProcessANPREventResponse{
			Success:      true,
			Message:      "Событие не является выездом, обработка не требуется",
			UserFound:    false,
			SessionFound: false,
		}, nil
	}

	// Нормализуем номер от камеры для поиска
	normalizedLicensePlate := utils.NormalizeLicensePlateForSearch(req.LicensePlate)
	logger.WithFields(logrus.Fields{
		"service":              "dahua",
		"method":               "ProcessANPREvent",
		"license_plate":        req.LicensePlate,
		"normalized_license_plate": normalizedLicensePlate,
	}).Info("Номер нормализован")

	// Ищем активную сессию напрямую по номеру автомобиля
	activeSession, err := s.sessionService.GetActiveSessionByCarNumber(ctx, normalizedLicensePlate)
	if err != nil {
		logger.WithFields(logrus.Fields{
			"service":      "dahua",
			"method":       "ProcessANPREvent",
			"license_plate": req.LicensePlate,
			"normalized_license_plate": normalizedLicensePlate,
			"error":        err,
		}).Info("Активная сессия с номером не найдена")
		return &models.ProcessANPREventResponse{
			Success:      true,
			Message:      fmt.Sprintf("Активная сессия с номером %s не найдена", req.LicensePlate),
			UserFound:    false,
			SessionFound: false,
		}, nil
	}

	logger.WithFields(logrus.Fields{
		"service":      "dahua",
		"method":       "ProcessANPREvent",
		"session_id":    activeSession.ID,
		"license_plate": req.LicensePlate,
		"session_status": activeSession.Status,
		"car_number":    activeSession.CarNumber,
	}).Info("Активная сессия найдена")

	// Завершаем сессию БЕЗ частичного возврата
	err = s.sessionService.CompleteSessionWithoutRefund(ctx, activeSession.ID)
	if err != nil {
		logger.WithFields(logrus.Fields{
			"service":      "dahua",
			"method":       "ProcessANPREvent",
			"session_id":   activeSession.ID,
			"license_plate": req.LicensePlate,
			"error":        err,
		}).Error("Ошибка завершения сессии")
		return &models.ProcessANPREventResponse{
			Success:       false,
			Message:       fmt.Sprintf("Ошибка завершения сессии: %v", err),
			UserFound:     false,
			SessionFound:  true,
			SessionID:     activeSession.ID.String(),
			SessionStatus: activeSession.Status,
		}, err
	}

	logger.WithFields(logrus.Fields{
		"service":      "dahua",
		"method":       "ProcessANPREvent",
		"session_id":   activeSession.ID,
		"license_plate": req.LicensePlate,
	}).Info("Сессия успешно завершена БЕЗ возврата")

	return &models.ProcessANPREventResponse{
		Success:       true,
		Message:       fmt.Sprintf("Сессия успешно завершена для автомобиля %s", req.LicensePlate),
		UserFound:     false,
		SessionFound:  true,
		SessionID:     activeSession.ID.String(),
		SessionStatus: "completed",
	}, nil
}
