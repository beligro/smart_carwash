package service

import (
	"carwash_backend/internal/logger"
	"context"
	"time"

	"carwash_backend/internal/domain/auth/repository"
)

// BackgroundTasks структура для фоновых задач
type BackgroundTasks struct {
	repo repository.Repository
}

// NewBackgroundTasks создает новый экземпляр BackgroundTasks
func NewBackgroundTasks(repo repository.Repository) *BackgroundTasks {
	return &BackgroundTasks{
		repo: repo,
	}
}

// DeactivateExpiredShifts деактивирует истекшие смены кассиров
func (bt *BackgroundTasks) DeactivateExpiredShifts(ctx context.Context) error {
	logger.Info("Запуск задачи деактивации истекших смен кассиров")

	// Получаем все активные смены
	activeShifts, err := bt.repo.GetActiveCashierShifts(ctx)
	if err != nil {
		logger.Printf("Ошибка получения активных смен: %v", err)
		return err
	}

	now := time.Now()
	deactivatedCount := 0

	for _, shift := range activeShifts {
		// Проверяем, истекла ли смена
		if now.After(shift.ExpiresAt) {
			// Деактивируем смену
			shift.IsActive = false
			endedAt := now
			shift.EndedAt = &endedAt

			if err := bt.repo.UpdateCashierShift(ctx, &shift); err != nil {
				logger.Printf("Ошибка деактивации смены %s: %v", shift.ID, err)
				continue
			}

			logger.Printf("Смена %s деактивирована (истекла в %s)", shift.ID, shift.ExpiresAt.Format(time.RFC3339))
			deactivatedCount++
		}
	}

	logger.Printf("Задача деактивации завершена. Деактивировано смен: %d", deactivatedCount)
	return nil
}
