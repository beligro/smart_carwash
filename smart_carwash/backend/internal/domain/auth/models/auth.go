package models

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// Cashier представляет модель кассира
type Cashier struct {
	ID           uuid.UUID      `json:"id" gorm:"primaryKey;type:uuid;default:gen_random_uuid()"`
	Username     string         `json:"username" gorm:"uniqueIndex"`
	PasswordHash string         `json:"-" gorm:"column:password_hash"`
	IsActive     bool           `json:"is_active" gorm:"default:true"`
	LastLogin    *time.Time     `json:"last_login"`
	CreatedAt    time.Time      `json:"created_at"`
	UpdatedAt    time.Time      `json:"updated_at"`
	DeletedAt    gorm.DeletedAt `json:"-" gorm:"index"`
}

// CashierSession представляет активную сессию кассира
type CashierSession struct {
	ID        uuid.UUID      `json:"id" gorm:"primaryKey;type:uuid;default:gen_random_uuid()"`
	CashierID uuid.UUID      `json:"cashier_id" gorm:"type:uuid;index"`
	Token     string         `json:"-"`
	ExpiresAt time.Time      `json:"expires_at"`
	CreatedAt time.Time      `json:"created_at"`
	UpdatedAt time.Time      `json:"updated_at"`
	DeletedAt gorm.DeletedAt `json:"-" gorm:"index"`
}

// Cleaner представляет модель уборщика
type Cleaner struct {
	ID           uuid.UUID      `json:"id" gorm:"primaryKey;type:uuid;default:gen_random_uuid()"`
	Username     string         `json:"username" gorm:"uniqueIndex"`
	PasswordHash string         `json:"-" gorm:"column:password_hash"`
	IsActive     bool           `json:"is_active" gorm:"default:true"`
	LastLogin    *time.Time     `json:"last_login"`
	CreatedAt    time.Time      `json:"created_at"`
	UpdatedAt    time.Time      `json:"updated_at"`
	DeletedAt    gorm.DeletedAt `json:"-" gorm:"index"`
}

// CleanerSession представляет активную сессию уборщика
type CleanerSession struct {
	ID        uuid.UUID      `json:"id" gorm:"primaryKey;type:uuid;default:gen_random_uuid()"`
	CleanerID uuid.UUID      `json:"cleaner_id" gorm:"type:uuid;index"`
	Token     string         `json:"-"`
	ExpiresAt time.Time      `json:"expires_at"`
	CreatedAt time.Time      `json:"created_at"`
	UpdatedAt time.Time      `json:"updated_at"`
	DeletedAt gorm.DeletedAt `json:"-" gorm:"index"`
}

// LoginRequest представляет запрос на авторизацию
type LoginRequest struct {
	Username string `json:"username" binding:"required"`
	Password string `json:"password" binding:"required"`
}

// LoginResponse представляет ответ на успешную авторизацию
type LoginResponse struct {
	Token     string    `json:"token"`
	ExpiresAt time.Time `json:"expires_at"`
	IsAdmin   bool      `json:"is_admin"`
	Role      string    `json:"role,omitempty"`
}

// CreateCashierRequest представляет запрос на создание кассира
type CreateCashierRequest struct {
	Username string `json:"username" binding:"required"`
	Password string `json:"password" binding:"required"`
}

// CreateCashierResponse представляет ответ на создание кассира
type CreateCashierResponse struct {
	ID        uuid.UUID `json:"id"`
	Username  string    `json:"username"`
	CreatedAt time.Time `json:"created_at"`
}

// CreateCleanerRequest представляет запрос на создание уборщика
type CreateCleanerRequest struct {
	Username string `json:"username" binding:"required"`
	Password string `json:"password" binding:"required"`
}

// CreateCleanerResponse представляет ответ на создание уборщика
type CreateCleanerResponse struct {
	ID        uuid.UUID `json:"id"`
	Username  string    `json:"username"`
	CreatedAt time.Time `json:"created_at"`
}

// UpdateCashierRequest представляет запрос на обновление кассира
type UpdateCashierRequest struct {
	ID       uuid.UUID `json:"id" binding:"required"`
	Username string    `json:"username"`
	Password string    `json:"password"`
	IsActive bool      `json:"is_active"`
}

// UpdateCashierResponse представляет ответ на обновление кассира
type UpdateCashierResponse struct {
	ID        uuid.UUID `json:"id"`
	Username  string    `json:"username"`
	IsActive  bool      `json:"is_active"`
	UpdatedAt time.Time `json:"updated_at"`
}

// GetCashiersResponse представляет ответ на получение списка кассиров
type GetCashiersResponse struct {
	Cashiers []Cashier `json:"cashiers"`
}

// UpdateCleanerRequest представляет запрос на обновление уборщика
type UpdateCleanerRequest struct {
	ID       uuid.UUID `json:"id" binding:"required"`
	Username string    `json:"username"`
	Password string    `json:"password"`
	IsActive bool      `json:"is_active"`
}

// UpdateCleanerResponse представляет ответ на обновление уборщика
type UpdateCleanerResponse struct {
	ID        uuid.UUID `json:"id"`
	Username  string    `json:"username"`
	IsActive  bool      `json:"is_active"`
	UpdatedAt time.Time `json:"updated_at"`
}

// GetCleanersResponse представляет ответ на получение списка уборщиков
type GetCleanersResponse struct {
	Cleaners []Cleaner `json:"cleaners"`
}

// DeleteCashierRequest представляет запрос на удаление кассира
type DeleteCashierRequest struct {
	ID uuid.UUID `json:"id" binding:"required"`
}

// DeleteCleanerRequest представляет запрос на удаление уборщика
type DeleteCleanerRequest struct {
	ID uuid.UUID `json:"id" binding:"required"`
}

// TokenClaims представляет данные, хранящиеся в JWT токене
type TokenClaims struct {
	ID       uuid.UUID `json:"id"`
	Username string    `json:"username"`
	IsAdmin  bool      `json:"is_admin"`
	Role     string    `json:"role,omitempty"`
}

// TwoFactorAuthSettings представляет настройки двухфакторной аутентификации
// (подготовка для будущей реализации)
type TwoFactorAuthSettings struct {
	ID        uuid.UUID      `json:"id" gorm:"primaryKey;type:uuid;default:gen_random_uuid()"`
	UserID    uuid.UUID      `json:"user_id" gorm:"type:uuid;uniqueIndex"`
	IsEnabled bool           `json:"is_enabled" gorm:"default:false"`
	Secret    string         `json:"-"`
	CreatedAt time.Time      `json:"created_at"`
	UpdatedAt time.Time      `json:"updated_at"`
	DeletedAt gorm.DeletedAt `json:"-" gorm:"index"`
}

// CashierShift представляет активную смену кассира
type CashierShift struct {
	ID        uuid.UUID      `json:"id" gorm:"primaryKey;type:uuid;default:gen_random_uuid()"`
	CashierID uuid.UUID      `json:"cashier_id" gorm:"type:uuid;index"`
	StartedAt time.Time      `json:"started_at"`
	EndedAt   *time.Time     `json:"ended_at"`
	ExpiresAt time.Time      `json:"expires_at"`
	IsActive  bool           `json:"is_active" gorm:"default:true"`
	CreatedAt time.Time      `json:"created_at"`
	UpdatedAt time.Time      `json:"updated_at"`
	DeletedAt gorm.DeletedAt `json:"-" gorm:"index"`
}

// StartShiftRequest представляет запрос на начало смены
type StartShiftRequest struct {
	CashierID uuid.UUID `json:"cashier_id" binding:"required"`
}

// StartShiftResponse представляет ответ на начало смены
type StartShiftResponse struct {
	ID        uuid.UUID `json:"id"`
	StartedAt time.Time `json:"started_at"`
	ExpiresAt time.Time `json:"expires_at"`
	IsActive  bool      `json:"is_active"`
}

// EndShiftRequest представляет запрос на завершение смены
type EndShiftRequest struct {
	CashierID uuid.UUID `json:"cashier_id" binding:"required"`
}

// EndShiftResponse представляет ответ на завершение смены
type EndShiftResponse struct {
	ID        uuid.UUID `json:"id"`
	StartedAt time.Time `json:"started_at"`
	EndedAt   time.Time `json:"ended_at"`
	IsActive  bool      `json:"is_active"`
}

// ShiftStatusResponse представляет ответ на запрос статуса смены
type ShiftStatusResponse struct {
	HasActiveShift bool       `json:"has_active_shift"`
	Shift          *CashierShift `json:"shift,omitempty"`
}

// CashierSessionsRequest представляет запрос на получение сессий кассира
type CashierSessionsRequest struct {
	CashierID uuid.UUID `json:"cashier_id" binding:"required"`
	Limit     int       `json:"limit"`
	Offset    int       `json:"offset"`
}

// CashierPaymentsRequest представляет запрос на получение платежей кассира
type CashierPaymentsRequest struct {
	CashierID uuid.UUID `json:"cashier_id" binding:"required"`
	Limit     int       `json:"limit"`
	Offset    int       `json:"offset"`
}
