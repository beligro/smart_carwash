package repository

import (
	"context"
	"errors"
	"time"

	"carwash_backend/internal/domain/auth/models"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

var (
	// ErrCashierNotFound возвращается, когда кассир не найден
	ErrCashierNotFound = errors.New("кассир не найден")

	// ErrCashierAlreadyExists возвращается, когда кассир с таким именем уже существует
	ErrCashierAlreadyExists = errors.New("кассир с таким именем уже существует")

	// ErrActiveCashierSessionExists возвращается, когда уже есть активная сессия кассира
	ErrActiveCashierSessionExists = errors.New("уже есть активная сессия кассира")

	// ErrActiveShiftExists возвращается, когда уже есть активная смена
	ErrActiveShiftExists = errors.New("уже есть активная смена")

	// ErrNoActiveShift возвращается, когда нет активной смены
	ErrNoActiveShift = errors.New("нет активной смены")

	// ErrCleanerNotFound возвращается, когда уборщик не найден
	ErrCleanerNotFound = errors.New("уборщик не найден")

	// ErrCleanerAlreadyExists возвращается, когда уборщик с таким именем уже существует
	ErrCleanerAlreadyExists = errors.New("уборщик с таким именем уже существует")

	// ErrActiveCleanerSessionExists возвращается, когда уже есть активная сессия уборщика
	ErrActiveCleanerSessionExists = errors.New("уже есть активная сессия уборщика")
)

// Repository интерфейс для работы с авторизацией в базе данных
type Repository interface {
	// Методы для работы с кассирами
	CreateCashier(ctx context.Context, cashier *models.Cashier) error
	GetCashierByID(ctx context.Context, id uuid.UUID) (*models.Cashier, error)
	GetCashierByUsername(ctx context.Context, username string) (*models.Cashier, error)
	UpdateCashier(ctx context.Context, cashier *models.Cashier) error
	DeleteCashier(ctx context.Context, id uuid.UUID) error
	ListCashiers(ctx context.Context) ([]models.Cashier, error)

	// Методы для работы с сессиями кассиров
	CreateCashierSession(ctx context.Context, session *models.CashierSession) error
	GetActiveCashierSessions(ctx context.Context) ([]models.CashierSession, error)
	GetCashierSessionByToken(ctx context.Context, token string) (*models.CashierSession, error)
	DeleteCashierSession(ctx context.Context, id uuid.UUID) error
	DeleteExpiredCashierSessions(ctx context.Context) error

	// Методы для работы с двухфакторной аутентификацией
	GetTwoFactorAuthSettings(ctx context.Context, userID uuid.UUID) (*models.TwoFactorAuthSettings, error)
	SaveTwoFactorAuthSettings(ctx context.Context, settings *models.TwoFactorAuthSettings) error

	// Методы для работы со сменами кассиров
	CreateCashierShift(ctx context.Context, shift *models.CashierShift) error
	GetActiveCashierShift(ctx context.Context) (*models.CashierShift, error)
	GetActiveCashierShifts(ctx context.Context) ([]models.CashierShift, error)
	UpdateCashierShift(ctx context.Context, shift *models.CashierShift) error
	DeleteCashierShift(ctx context.Context, id uuid.UUID) error

	// Методы для работы с уборщиками
	CreateCleaner(ctx context.Context, cleaner *models.Cleaner) error
	GetCleanerByID(ctx context.Context, id uuid.UUID) (*models.Cleaner, error)
	GetCleanerByUsername(ctx context.Context, username string) (*models.Cleaner, error)
	UpdateCleaner(ctx context.Context, cleaner *models.Cleaner) error
	DeleteCleaner(ctx context.Context, id uuid.UUID) error
	ListCleaners(ctx context.Context) ([]models.Cleaner, error)

	// Методы для работы с сессиями уборщиков
	CreateCleanerSession(ctx context.Context, session *models.CleanerSession) error
	GetActiveCleanerSessions(ctx context.Context) ([]models.CleanerSession, error)
	GetCleanerSessionByToken(ctx context.Context, token string) (*models.CleanerSession, error)
	DeleteCleanerSession(ctx context.Context, id uuid.UUID) error
	DeleteExpiredCleanerSessions(ctx context.Context) error
}

// PostgresRepository реализация Repository для PostgreSQL
type PostgresRepository struct {
	db *gorm.DB
}

// NewPostgresRepository создает новый экземпляр PostgresRepository
func NewPostgresRepository(db *gorm.DB) *PostgresRepository {
	return &PostgresRepository{
		db: db,
	}
}

// CreateCashier создает нового кассира
func (r *PostgresRepository) CreateCashier(ctx context.Context, cashier *models.Cashier) error {
	// Проверяем, существует ли кассир с таким именем
	var count int64
	err := r.db.WithContext(ctx).Model(&models.Cashier{}).Where("username = ?", cashier.Username).Count(&count).Error
	if err != nil {
		return err
	}
	if count > 0 {
		return ErrCashierAlreadyExists
	}

	return r.db.WithContext(ctx).Create(cashier).Error
}

// GetCashierByID получает кассира по ID
func (r *PostgresRepository) GetCashierByID(ctx context.Context, id uuid.UUID) (*models.Cashier, error) {
	var cashier models.Cashier
	err := r.db.WithContext(ctx).Where("id = ?", id).First(&cashier).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrCashierNotFound
		}
		return nil, err
	}
	return &cashier, nil
}

// GetCashierByUsername получает кассира по имени пользователя
func (r *PostgresRepository) GetCashierByUsername(ctx context.Context, username string) (*models.Cashier, error) {
	var cashier models.Cashier
	err := r.db.WithContext(ctx).Where("username = ?", username).First(&cashier).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrCashierNotFound
		}
		return nil, err
	}
	return &cashier, nil
}

// UpdateCashier обновляет кассира
func (r *PostgresRepository) UpdateCashier(ctx context.Context, cashier *models.Cashier) error {
	return r.db.WithContext(ctx).Save(cashier).Error
}

// DeleteCashier удаляет кассира
func (r *PostgresRepository) DeleteCashier(ctx context.Context, id uuid.UUID) error {
	return r.db.WithContext(ctx).Delete(&models.Cashier{}, id).Error
}

// ListCashiers возвращает список всех кассиров
func (r *PostgresRepository) ListCashiers(ctx context.Context) ([]models.Cashier, error) {
	var cashiers []models.Cashier
	err := r.db.WithContext(ctx).Find(&cashiers).Error
	return cashiers, err
}

// CreateCashierSession создает новую сессию кассира
func (r *PostgresRepository) CreateCashierSession(ctx context.Context, session *models.CashierSession) error {
	// Проверяем, есть ли уже активные сессии кассиров
	var activeSessions []models.CashierSession
	err := r.db.WithContext(ctx).Where("expires_at > ?", time.Now()).Find(&activeSessions).Error
	if err != nil {
		return err
	}

	// Если есть активные сессии других кассиров, возвращаем ошибку
	for _, s := range activeSessions {
		if s.CashierID != session.CashierID {
			return ErrActiveCashierSessionExists
		}
	}

	// Удаляем все предыдущие сессии этого кассира
	err = r.db.WithContext(ctx).Where("cashier_id = ?", session.CashierID).Delete(&models.CashierSession{}).Error
	if err != nil {
		return err
	}

	// Создаем новую сессию
	return r.db.WithContext(ctx).Create(session).Error
}

// GetActiveCashierSessions возвращает все активные сессии кассиров
func (r *PostgresRepository) GetActiveCashierSessions(ctx context.Context) ([]models.CashierSession, error) {
	var sessions []models.CashierSession
	err := r.db.WithContext(ctx).Where("expires_at > ?", time.Now()).Find(&sessions).Error
	return sessions, err
}

// GetCashierSessionByToken получает сессию кассира по токену
func (r *PostgresRepository) GetCashierSessionByToken(ctx context.Context, token string) (*models.CashierSession, error) {
	var session models.CashierSession
	err := r.db.WithContext(ctx).Where("token = ? AND expires_at > ?", token, time.Now()).First(&session).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil // Сессия не найдена или истекла
		}
		return nil, err
	}
	return &session, nil
}

// DeleteCashierSession удаляет сессию кассира
func (r *PostgresRepository) DeleteCashierSession(ctx context.Context, id uuid.UUID) error {
	return r.db.WithContext(ctx).Delete(&models.CashierSession{}, id).Error
}

// DeleteExpiredCashierSessions удаляет все истекшие сессии кассиров
func (r *PostgresRepository) DeleteExpiredCashierSessions(ctx context.Context) error {
	return r.db.WithContext(ctx).Where("expires_at <= ?", time.Now()).Delete(&models.CashierSession{}).Error
}

// GetTwoFactorAuthSettings получает настройки двухфакторной аутентификации для пользователя
func (r *PostgresRepository) GetTwoFactorAuthSettings(ctx context.Context, userID uuid.UUID) (*models.TwoFactorAuthSettings, error) {
	var settings models.TwoFactorAuthSettings
	err := r.db.WithContext(ctx).Where("user_id = ?", userID).First(&settings).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			// Если настройки не найдены, возвращаем настройки по умолчанию
			return &models.TwoFactorAuthSettings{
				UserID:    userID,
				IsEnabled: false,
			}, nil
		}
		return nil, err
	}
	return &settings, nil
}

// SaveTwoFactorAuthSettings сохраняет настройки двухфакторной аутентификации
func (r *PostgresRepository) SaveTwoFactorAuthSettings(ctx context.Context, settings *models.TwoFactorAuthSettings) error {
	// Проверяем, существуют ли уже настройки для этого пользователя
	var existingSettings models.TwoFactorAuthSettings
	err := r.db.WithContext(ctx).Where("user_id = ?", settings.UserID).First(&existingSettings).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			// Если настройки не найдены, создаем новые
			return r.db.WithContext(ctx).Create(settings).Error
		}
		return err
	}

	// Если настройки найдены, обновляем их
	existingSettings.IsEnabled = settings.IsEnabled
	existingSettings.Secret = settings.Secret
	return r.db.WithContext(ctx).Save(&existingSettings).Error
}

// CreateCashierShift создает новую смену для кассира
func (r *PostgresRepository) CreateCashierShift(ctx context.Context, shift *models.CashierShift) error {
	// Проверяем, есть ли уже активная смена
	var activeShift models.CashierShift
	err := r.db.WithContext(ctx).Where("is_active = ? AND expires_at > ?", true, time.Now()).First(&activeShift).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			// Если активной смены нет, создаем новую
			return r.db.WithContext(ctx).Create(shift).Error
		}
		return err
	}
	return ErrActiveShiftExists
}

// GetActiveCashierShift получает активную смену
func (r *PostgresRepository) GetActiveCashierShift(ctx context.Context) (*models.CashierShift, error) {
	var shift models.CashierShift
	err := r.db.WithContext(ctx).Where("is_active = ? AND expires_at > ?", true, time.Now()).First(&shift).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrNoActiveShift
		}
		return nil, err
	}
	return &shift, nil
}

// GetActiveCashierShifts получает все активные смены
func (r *PostgresRepository) GetActiveCashierShifts(ctx context.Context) ([]models.CashierShift, error) {
	var shifts []models.CashierShift
	err := r.db.WithContext(ctx).Where("is_active = ? AND expires_at > ?", true, time.Now()).Find(&shifts).Error
	return shifts, err
}

// UpdateCashierShift обновляет смену
func (r *PostgresRepository) UpdateCashierShift(ctx context.Context, shift *models.CashierShift) error {
	return r.db.WithContext(ctx).Save(shift).Error
}

// DeleteCashierShift удаляет смену
func (r *PostgresRepository) DeleteCashierShift(ctx context.Context, id uuid.UUID) error {
	return r.db.WithContext(ctx).Delete(&models.CashierShift{}, id).Error
}

// CreateCleaner создает нового уборщика
func (r *PostgresRepository) CreateCleaner(ctx context.Context, cleaner *models.Cleaner) error {
	// Проверяем, существует ли уже уборщик с таким именем
	existing, err := r.GetCleanerByUsername(ctx, cleaner.Username)
	if err == nil && existing != nil {
		return ErrCleanerAlreadyExists
	}

	// Создаем уборщика
	return r.db.WithContext(ctx).Create(cleaner).Error
}

// GetCleanerByID получает уборщика по ID
func (r *PostgresRepository) GetCleanerByID(ctx context.Context, id uuid.UUID) (*models.Cleaner, error) {
	var cleaner models.Cleaner
	err := r.db.WithContext(ctx).First(&cleaner, id).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrCleanerNotFound
		}
		return nil, err
	}
	return &cleaner, nil
}

// GetCleanerByUsername получает уборщика по имени пользователя
func (r *PostgresRepository) GetCleanerByUsername(ctx context.Context, username string) (*models.Cleaner, error) {
	var cleaner models.Cleaner
	err := r.db.WithContext(ctx).Where("username = ?", username).First(&cleaner).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrCleanerNotFound
		}
		return nil, err
	}
	return &cleaner, nil
}

// UpdateCleaner обновляет уборщика
func (r *PostgresRepository) UpdateCleaner(ctx context.Context, cleaner *models.Cleaner) error {
	return r.db.WithContext(ctx).Save(cleaner).Error
}

// DeleteCleaner удаляет уборщика
func (r *PostgresRepository) DeleteCleaner(ctx context.Context, id uuid.UUID) error {
	return r.db.WithContext(ctx).Delete(&models.Cleaner{}, id).Error
}

// ListCleaners получает список всех уборщиков
func (r *PostgresRepository) ListCleaners(ctx context.Context) ([]models.Cleaner, error) {
	var cleaners []models.Cleaner
	err := r.db.WithContext(ctx).Find(&cleaners).Error
	return cleaners, err
}

// CreateCleanerSession создает новую сессию уборщика
func (r *PostgresRepository) CreateCleanerSession(ctx context.Context, session *models.CleanerSession) error {
	// Уборщики могут иметь множественные сессии, поэтому не проверяем существующие
	// Просто создаем новую сессию
	return r.db.WithContext(ctx).Create(session).Error
}

// GetActiveCleanerSessions получает все активные сессии уборщиков
func (r *PostgresRepository) GetActiveCleanerSessions(ctx context.Context) ([]models.CleanerSession, error) {
	var sessions []models.CleanerSession
	err := r.db.WithContext(ctx).Where("expires_at > ?", time.Now()).Find(&sessions).Error
	return sessions, err
}

// GetCleanerSessionByToken получает сессию уборщика по токену
func (r *PostgresRepository) GetCleanerSessionByToken(ctx context.Context, token string) (*models.CleanerSession, error) {
	var session models.CleanerSession
	err := r.db.WithContext(ctx).Where("token = ? AND expires_at > ?", token, time.Now()).First(&session).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrCleanerNotFound
		}
		return nil, err
	}
	return &session, nil
}

// DeleteCleanerSession удаляет сессию уборщика
func (r *PostgresRepository) DeleteCleanerSession(ctx context.Context, id uuid.UUID) error {
	return r.db.WithContext(ctx).Delete(&models.CleanerSession{}, id).Error
}

// DeleteExpiredCleanerSessions удаляет истекшие сессии уборщиков
func (r *PostgresRepository) DeleteExpiredCleanerSessions(ctx context.Context) error {
	return r.db.WithContext(ctx).Where("expires_at < ?", time.Now()).Delete(&models.CleanerSession{}).Error
}
