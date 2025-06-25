package repository

import (
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
)

// Repository интерфейс для работы с авторизацией в базе данных
type Repository interface {
	// Методы для работы с кассирами
	CreateCashier(cashier *models.Cashier) error
	GetCashierByID(id uuid.UUID) (*models.Cashier, error)
	GetCashierByUsername(username string) (*models.Cashier, error)
	UpdateCashier(cashier *models.Cashier) error
	DeleteCashier(id uuid.UUID) error
	ListCashiers() ([]models.Cashier, error)

	// Методы для работы с сессиями кассиров
	CreateCashierSession(session *models.CashierSession) error
	GetActiveCashierSessions() ([]models.CashierSession, error)
	GetCashierSessionByToken(token string) (*models.CashierSession, error)
	DeleteCashierSession(id uuid.UUID) error
	DeleteExpiredCashierSessions() error

	// Методы для работы с двухфакторной аутентификацией
	GetTwoFactorAuthSettings(userID uuid.UUID) (*models.TwoFactorAuthSettings, error)
	SaveTwoFactorAuthSettings(settings *models.TwoFactorAuthSettings) error
}

// PostgresRepository реализация Repository для PostgreSQL
type PostgresRepository struct {
	db *gorm.DB
}

// NewPostgresRepository создает новый экземпляр PostgresRepository
func NewPostgresRepository(db *gorm.DB) *PostgresRepository {
	return &PostgresRepository{db: db}
}

// CreateCashier создает нового кассира
func (r *PostgresRepository) CreateCashier(cashier *models.Cashier) error {
	// Проверяем, существует ли кассир с таким именем
	var count int64
	if err := r.db.Model(&models.Cashier{}).Where("username = ?", cashier.Username).Count(&count).Error; err != nil {
		return err
	}
	if count > 0 {
		return ErrCashierAlreadyExists
	}

	return r.db.Create(cashier).Error
}

// GetCashierByID получает кассира по ID
func (r *PostgresRepository) GetCashierByID(id uuid.UUID) (*models.Cashier, error) {
	var cashier models.Cashier
	err := r.db.Where("id = ?", id).First(&cashier).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrCashierNotFound
		}
		return nil, err
	}
	return &cashier, nil
}

// GetCashierByUsername получает кассира по имени пользователя
func (r *PostgresRepository) GetCashierByUsername(username string) (*models.Cashier, error) {
	var cashier models.Cashier
	err := r.db.Where("username = ?", username).First(&cashier).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrCashierNotFound
		}
		return nil, err
	}
	return &cashier, nil
}

// UpdateCashier обновляет кассира
func (r *PostgresRepository) UpdateCashier(cashier *models.Cashier) error {
	return r.db.Save(cashier).Error
}

// DeleteCashier удаляет кассира
func (r *PostgresRepository) DeleteCashier(id uuid.UUID) error {
	return r.db.Delete(&models.Cashier{}, id).Error
}

// ListCashiers возвращает список всех кассиров
func (r *PostgresRepository) ListCashiers() ([]models.Cashier, error) {
	var cashiers []models.Cashier
	err := r.db.Find(&cashiers).Error
	return cashiers, err
}

// CreateCashierSession создает новую сессию кассира
func (r *PostgresRepository) CreateCashierSession(session *models.CashierSession) error {
	// Проверяем, есть ли уже активные сессии кассиров
	var activeSessions []models.CashierSession
	err := r.db.Where("expires_at > ?", time.Now()).Find(&activeSessions).Error
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
	if err := r.db.Where("cashier_id = ?", session.CashierID).Delete(&models.CashierSession{}).Error; err != nil {
		return err
	}

	// Создаем новую сессию
	return r.db.Create(session).Error
}

// GetActiveCashierSessions возвращает все активные сессии кассиров
func (r *PostgresRepository) GetActiveCashierSessions() ([]models.CashierSession, error) {
	var sessions []models.CashierSession
	err := r.db.Where("expires_at > ?", time.Now()).Find(&sessions).Error
	return sessions, err
}

// GetCashierSessionByToken получает сессию кассира по токену
func (r *PostgresRepository) GetCashierSessionByToken(token string) (*models.CashierSession, error) {
	var session models.CashierSession
	err := r.db.Where("token = ? AND expires_at > ?", token, time.Now()).First(&session).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil // Сессия не найдена или истекла
		}
		return nil, err
	}
	return &session, nil
}

// DeleteCashierSession удаляет сессию кассира
func (r *PostgresRepository) DeleteCashierSession(id uuid.UUID) error {
	return r.db.Delete(&models.CashierSession{}, id).Error
}

// DeleteExpiredCashierSessions удаляет все истекшие сессии кассиров
func (r *PostgresRepository) DeleteExpiredCashierSessions() error {
	return r.db.Where("expires_at <= ?", time.Now()).Delete(&models.CashierSession{}).Error
}

// GetTwoFactorAuthSettings получает настройки двухфакторной аутентификации для пользователя
func (r *PostgresRepository) GetTwoFactorAuthSettings(userID uuid.UUID) (*models.TwoFactorAuthSettings, error) {
	var settings models.TwoFactorAuthSettings
	err := r.db.Where("user_id = ?", userID).First(&settings).Error
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
func (r *PostgresRepository) SaveTwoFactorAuthSettings(settings *models.TwoFactorAuthSettings) error {
	// Проверяем, существуют ли уже настройки для этого пользователя
	var existingSettings models.TwoFactorAuthSettings
	err := r.db.Where("user_id = ?", settings.UserID).First(&existingSettings).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			// Если настройки не найдены, создаем новые
			return r.db.Create(settings).Error
		}
		return err
	}

	// Если настройки найдены, обновляем их
	existingSettings.IsEnabled = settings.IsEnabled
	existingSettings.Secret = settings.Secret
	return r.db.Save(&existingSettings).Error
}
