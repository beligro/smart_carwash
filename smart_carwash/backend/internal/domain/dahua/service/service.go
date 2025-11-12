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
	washboxService WashboxService
}

// SessionService интерфейс для работы с сессиями
type SessionService interface {
	GetActiveSessionByUserID(ctx context.Context, userID uuid.UUID) (*sessionModels.Session, error)
	GetActiveSessionByCarNumber(ctx context.Context, carNumber string) (*sessionModels.Session, error)
	GetLastSessionByCarNumber(ctx context.Context, carNumber string) (*sessionModels.Session, error)
	CompleteSessionWithoutRefund(ctx context.Context, sessionID uuid.UUID) error
}

// WashboxService интерфейс для работы с боксами
type WashboxService interface {
	ClearCooldown(ctx context.Context, boxID uuid.UUID) error
}

// NewService создает новый экземпляр сервиса
func NewService(sessionService SessionService, washboxService WashboxService) Service {
	return &ServiceImpl{
		sessionService: sessionService,
		washboxService: washboxService,
	}
}

// ProcessANPREvent обрабатывает ANPR событие от камеры Dahua
func (s *ServiceImpl) ProcessANPREvent(ctx context.Context, req *models.ProcessANPREventRequest) (*models.ProcessANPREventResponse, error) {
	logger.WithFields(logrus.Fields{
		"service":       "dahua",
		"method":        "ProcessANPREvent",
		"license_plate": req.LicensePlate,
		"direction":     req.Direction,
	}).Info("Обработка события ANPR")

	// Проверяем, что это событие выезда
	if req.Direction != "out" {
		logger.WithFields(logrus.Fields{
			"service":       "dahua",
			"method":        "ProcessANPREvent",
			"license_plate": req.LicensePlate,
			"direction":     req.Direction,
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
		"service":                  "dahua",
		"method":                   "ProcessANPREvent",
		"license_plate":            req.LicensePlate,
		"normalized_license_plate": normalizedLicensePlate,
	}).Info("Номер нормализован")

	// Ищем последнюю сессию по номеру автомобиля (любой статус)
	lastSession, err := s.sessionService.GetLastSessionByCarNumber(ctx, normalizedLicensePlate)
	if err != nil {
		logger.WithFields(logrus.Fields{
			"service":                  "dahua",
			"method":                   "ProcessANPREvent",
			"license_plate":            req.LicensePlate,
			"normalized_license_plate": normalizedLicensePlate,
			"error":                    err,
		}).Info("Сессия с номером не найдена")
		return &models.ProcessANPREventResponse{
			Success:      true,
			Message:      fmt.Sprintf("Сессия с номером %s не найдена", req.LicensePlate),
			UserFound:    false,
			SessionFound: false,
		}, nil
	}

	logger.WithFields(logrus.Fields{
		"service":        "dahua",
		"method":         "ProcessANPREvent",
		"session_id":     lastSession.ID,
		"license_plate":  req.LicensePlate,
		"session_status": lastSession.Status,
		"car_number":     lastSession.CarNumber,
	}).Info("Сессия найдена")

	// Если сессия активна - завершаем её
	if lastSession.Status == sessionModels.SessionStatusActive {
		logger.WithFields(logrus.Fields{
			"service":       "dahua",
			"method":        "ProcessANPREvent",
			"session_id":    lastSession.ID,
			"license_plate": req.LicensePlate,
		}).Info("Сессия активна, завершаем")

		// Завершаем сессию БЕЗ частичного возврата
		err = s.sessionService.CompleteSessionWithoutRefund(ctx, lastSession.ID)
		if err != nil {
			logger.WithFields(logrus.Fields{
				"service":       "dahua",
				"method":        "ProcessANPREvent",
				"session_id":    lastSession.ID,
				"license_plate": req.LicensePlate,
				"error":         err,
			}).Error("Ошибка завершения сессии")
			return &models.ProcessANPREventResponse{
				Success:       false,
				Message:       fmt.Sprintf("Ошибка завершения сессии: %v", err),
				UserFound:     false,
				SessionFound:  true,
				SessionID:     lastSession.ID.String(),
				SessionStatus: lastSession.Status,
			}, err
		}

		logger.WithFields(logrus.Fields{
			"service":       "dahua",
			"method":        "ProcessANPREvent",
			"session_id":    lastSession.ID,
			"license_plate": req.LicensePlate,
		}).Info("Сессия успешно завершена БЕЗ возврата")

		return &models.ProcessANPREventResponse{
			Success:       true,
			Message:       fmt.Sprintf("Сессия успешно завершена для автомобиля %s", req.LicensePlate),
			UserFound:     false,
			SessionFound:  true,
			SessionID:     lastSession.ID.String(),
			SessionStatus: "completed",
		}, nil
	}

	// Если сессия уже завершена - сбрасываем кулдаун и ставим бокс в статус свободен
	if lastSession.Status == sessionModels.SessionStatusComplete && lastSession.BoxID != nil {
		logger.WithFields(logrus.Fields{
			"service":       "dahua",
			"method":        "ProcessANPREvent",
			"session_id":    lastSession.ID,
			"license_plate": req.LicensePlate,
			"box_id":        *lastSession.BoxID,
		}).Info("Сессия уже завершена, сбрасываем кулдаун и ставим бокс в статус свободен")

		// Сбрасываем кулдаун
		if s.washboxService != nil {
			err = s.washboxService.ClearCooldown(ctx, *lastSession.BoxID)
			if err != nil {
				logger.WithFields(logrus.Fields{
					"service":       "dahua",
					"method":        "ProcessANPREvent",
					"session_id":    lastSession.ID,
					"license_plate": req.LicensePlate,
					"box_id":        *lastSession.BoxID,
					"error":         err,
				}).Error("Ошибка сброса кулдауна")
				// Продолжаем выполнение, даже если не удалось сбросить кулдаун
			} else {
				logger.WithFields(logrus.Fields{
					"service":       "dahua",
					"method":        "ProcessANPREvent",
					"session_id":    lastSession.ID,
					"license_plate": req.LicensePlate,
					"box_id":        *lastSession.BoxID,
				}).Info("Кулдаун успешно сброшен")
			}
		}

		return &models.ProcessANPREventResponse{
			Success:       true,
			Message:       fmt.Sprintf("Кулдаун сброшен и бокс переведен в статус свободен для автомобиля %s", req.LicensePlate),
			UserFound:     false,
			SessionFound:  true,
			SessionID:     lastSession.ID.String(),
			SessionStatus: lastSession.Status,
		}, nil
	}

	// Если сессия в другом статусе - просто возвращаем информацию
	logger.WithFields(logrus.Fields{
		"service":        "dahua",
		"method":         "ProcessANPREvent",
		"session_id":     lastSession.ID,
		"license_plate":  req.LicensePlate,
		"session_status": lastSession.Status,
	}).Info("Сессия найдена, но в статусе, не требующем обработки")

	return &models.ProcessANPREventResponse{
		Success:       true,
		Message:       fmt.Sprintf("Сессия с номером %s найдена, но в статусе %s", req.LicensePlate, lastSession.Status),
		UserFound:     false,
		SessionFound:  true,
		SessionID:     lastSession.ID.String(),
		SessionStatus: lastSession.Status,
	}, nil
}
