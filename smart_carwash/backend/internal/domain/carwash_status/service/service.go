package service

import (
	"carwash_backend/internal/domain/carwash_status/models"
	"carwash_backend/internal/domain/carwash_status/repository"
	sessionModels "carwash_backend/internal/domain/session/models"
	sessionService "carwash_backend/internal/domain/session/service"
	"carwash_backend/internal/logger"
	"context"
	"fmt"

	"github.com/google/uuid"
)

// Service интерфейс для бизнес-логики статуса мойки
type Service interface {
	GetCurrentStatus(ctx context.Context, req *models.GetCurrentStatusRequest) (*models.GetCurrentStatusResponse, error)
	CloseCarwash(ctx context.Context, req *models.CloseCarwashRequest, adminID uuid.UUID) (*models.CloseCarwashResponse, error)
	OpenCarwash(ctx context.Context, req *models.OpenCarwashRequest, adminID uuid.UUID) (*models.OpenCarwashResponse, error)
	GetHistory(ctx context.Context, req *models.GetHistoryRequest) (*models.GetHistoryResponse, error)
}

// ServiceImpl реализация Service
type ServiceImpl struct {
	repo           repository.Repository
	sessionService sessionService.Service
}

// NewService создает новый экземпляр Service
func NewService(repo repository.Repository, sessionService sessionService.Service) *ServiceImpl {
	return &ServiceImpl{
		repo:           repo,
		sessionService: sessionService,
	}
}

// GetCurrentStatus получает текущий статус мойки
func (s *ServiceImpl) GetCurrentStatus(ctx context.Context, req *models.GetCurrentStatusRequest) (*models.GetCurrentStatusResponse, error) {
	status, err := s.repo.GetCurrentStatus(ctx)
	if err != nil {
		return nil, fmt.Errorf("ошибка получения статуса мойки: %w", err)
	}

	return &models.GetCurrentStatusResponse{
		IsClosed:     status.IsClosed,
		ClosedReason: status.ClosedReason,
		UpdatedAt:    status.UpdatedAt,
	}, nil
}

// CloseCarwash закрывает мойку
func (s *ServiceImpl) CloseCarwash(ctx context.Context, req *models.CloseCarwashRequest, adminID uuid.UUID) (*models.CloseCarwashResponse, error) {
	logger.Printf("CloseCarwash: начало закрытия мойки, admin_id: %s, reason: %v", adminID, req.Reason)

	completedSessions := 0
	canceledSessions := 0

	// 1. Завершаем все активные сессии (status = 'active')
	activeSessions, err := s.sessionService.GetSessionsByStatus(ctx, sessionModels.SessionStatusActive)
	if err != nil {
		logger.Printf("CloseCarwash: ошибка получения активных сессий: %v", err)
		// Продолжаем выполнение, даже если не удалось получить сессии
	} else {
		logger.Printf("CloseCarwash: найдено %d активных сессий", len(activeSessions))

		for _, session := range activeSessions {
			// Завершаем сессию без возврата (время уже использовано)
			err := s.sessionService.CompleteSessionWithoutRefund(ctx, session.ID)
			if err != nil {
				logger.Printf("CloseCarwash: ошибка завершения активной сессии %s: %v", session.ID, err)
				// Продолжаем выполнение для других сессий
			} else {
				completedSessions++
				logger.Printf("CloseCarwash: активная сессия %s завершена", session.ID)
			}
		}
	}

	// 2. Отменяем сессии в статусах created, in_queue, assigned с возвратом денег
	statusesToCancel := []string{
		sessionModels.SessionStatusCreated,
		sessionModels.SessionStatusInQueue,
		sessionModels.SessionStatusAssigned,
	}

	for _, status := range statusesToCancel {
		sessions, err := s.sessionService.GetSessionsByStatus(ctx, status)
		if err != nil {
			logger.Printf("CloseCarwash: ошибка получения сессий со статусом %s: %v", status, err)
			continue
		}

		logger.Printf("CloseCarwash: найдено %d сессий со статусом %s", len(sessions), status)

		for _, session := range sessions {
			// Для created сессий возврат не нужен (они не оплачены)
			// Для in_queue и assigned - возвращаем деньги
			skipRefund := (status == sessionModels.SessionStatusCreated)

			cancelReq := &sessionModels.CancelSessionRequest{
				SessionID:  session.ID,
				UserID:     session.UserID, // Используем ID владельца сессии
				SkipRefund: skipRefund,
			}

			_, err := s.sessionService.CancelSession(ctx, cancelReq)
			if err != nil {
				logger.Printf("CloseCarwash: ошибка отмены сессии %s (статус: %s): %v", session.ID, status, err)
				// Продолжаем выполнение для других сессий
			} else {
				canceledSessions++
				logger.Printf("CloseCarwash: сессия %s (статус: %s) отменена", session.ID, status)
			}
		}
	}

	// 3. Обновляем статус мойки в БД
	err = s.repo.UpdateStatus(ctx, true, req.Reason, &adminID)
	if err != nil {
		return nil, fmt.Errorf("ошибка обновления статуса мойки: %w", err)
	}

	// 4. Создаем запись в истории
	err = s.repo.CreateHistoryRecord(ctx, true, req.Reason, &adminID)
	if err != nil {
		logger.Printf("CloseCarwash: ошибка создания записи в истории: %v", err)
		// Не возвращаем ошибку, так как статус уже обновлен
	}

	logger.Printf("CloseCarwash: мойка закрыта, завершено сессий: %d, отменено сессий: %d", completedSessions, canceledSessions)

	return &models.CloseCarwashResponse{
		Success:           true,
		Message:           "Мойка успешно закрыта",
		CompletedSessions: completedSessions,
		CanceledSessions:  canceledSessions,
	}, nil
}

// OpenCarwash открывает мойку
func (s *ServiceImpl) OpenCarwash(ctx context.Context, req *models.OpenCarwashRequest, adminID uuid.UUID) (*models.OpenCarwashResponse, error) {
	logger.Printf("OpenCarwash: начало открытия мойки, admin_id: %s", adminID)

	// 1. Обновляем статус мойки в БД
	err := s.repo.UpdateStatus(ctx, false, nil, &adminID)
	if err != nil {
		return nil, fmt.Errorf("ошибка обновления статуса мойки: %w", err)
	}

	// 2. Создаем запись в истории
	err = s.repo.CreateHistoryRecord(ctx, false, nil, &adminID)
	if err != nil {
		logger.Printf("OpenCarwash: ошибка создания записи в истории: %v", err)
		// Не возвращаем ошибку, так как статус уже обновлен
	}

	logger.Printf("OpenCarwash: мойка открыта")

	return &models.OpenCarwashResponse{
		Success: true,
		Message: "Мойка успешно открыта",
	}, nil
}

// GetHistory получает историю изменений статуса мойки
func (s *ServiceImpl) GetHistory(ctx context.Context, req *models.GetHistoryRequest) (*models.GetHistoryResponse, error) {
	// Применяем лимиты по умолчанию
	limit := req.Limit
	if limit <= 0 {
		limit = 50
	}
	offset := req.Offset
	if offset < 0 {
		offset = 0
	}

	history, total, err := s.repo.GetHistory(ctx, limit, offset)
	if err != nil {
		return nil, fmt.Errorf("ошибка получения истории статуса мойки: %w", err)
	}

	return &models.GetHistoryResponse{
		History: history,
		Total:   total,
	}, nil
}
