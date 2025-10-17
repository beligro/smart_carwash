package service

import (
	"carwash_backend/internal/logger"
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
func (bt *BackgroundTasks) DeactivateExpiredShifts() error {
	logger.Info("Запуск задачи деактивации истекших смен кассиров")

	// Получаем все активные смены
	activeShifts, err := bt.repo.GetActiveCashierShifts()
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

			if err := bt.repo.UpdateCashierShift(&shift); err != nil {
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

// StartPeriodicTasks запускает периодические задачи
func (bt *BackgroundTasks) StartPeriodicTasks() {
	// Запускаем задачу деактивации смен каждые 5 минут
	ticker := time.NewTicker(5 * time.Minute)
	defer ticker.Stop()

	logger.Info("Запуск периодических задач для кассиров")

	for {
		select {
		case <-ticker.C:
			if err := bt.DeactivateExpiredShifts(); err != nil {
				logger.Printf("Ошибка выполнения периодической задачи: %v", err)
			}
		}
	}
} 