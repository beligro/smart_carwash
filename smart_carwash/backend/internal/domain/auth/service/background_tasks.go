package service

import (
	"carwash_backend/internal/logger"
	"context"

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

// DeactivateExpiredShifts деактивирует истекшие смены кассиров.
// Одним UPDATE закрывает все смены с is_active=true и expires_at <= now.
// Раньше здесь использовался GetActiveCashierShifts, который сам фильтровал expires_at > now,
// из-за чего истёкшие смены никогда не попадали в выборку и не закрывались.
func (bt *BackgroundTasks) DeactivateExpiredShifts(ctx context.Context) error {
	deactivatedCount, err := bt.repo.DeactivateExpiredCashierShifts(ctx)
	if err != nil {
		logger.Printf("Ошибка деактивации истекших смен: %v", err)
		return err
	}

	if deactivatedCount > 0 {
		logger.Printf("Задача деактивации завершена. Деактивировано смен: %d", deactivatedCount)
	}
	return nil
}
