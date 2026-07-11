package models

import (
	"database/sql/driver"
	"strings"
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// Sections — список разрешённых разделов админки. В БД хранится как TEXT
// (ключи через запятую), в JSON — как массив строк. Драйвер-независимо.
type Sections []string

// Value сериализует список в строку через запятую для сохранения в БД.
func (s Sections) Value() (driver.Value, error) {
	return strings.Join([]string(s), ","), nil
}

// Scan разбирает строку из БД обратно в список разделов.
func (s *Sections) Scan(value interface{}) error {
	if value == nil {
		*s = Sections{}
		return nil
	}
	var raw string
	switch v := value.(type) {
	case string:
		raw = v
	case []byte:
		raw = string(v)
	default:
		*s = Sections{}
		return nil
	}
	raw = strings.TrimSpace(raw)
	if raw == "" {
		*s = Sections{}
		return nil
	}
	parts := strings.Split(raw, ",")
	out := make(Sections, 0, len(parts))
	for _, p := range parts {
		if t := strings.TrimSpace(p); t != "" {
			out = append(out, t)
		}
	}
	*s = out
	return nil
}

// Has проверяет наличие раздела в списке.
func (s Sections) Has(section string) bool {
	for _, x := range s {
		if x == section {
			return true
		}
	}
	return false
}

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

// Admin представляет персональную учётку администратора (Макс/Костя/Стас и др.).
// Доступ к разделам админки — через AllowedSections.
type Admin struct {
	ID              uuid.UUID      `json:"id" gorm:"primaryKey;type:uuid;default:gen_random_uuid()"`
	Username        string         `json:"username" gorm:"uniqueIndex"`
	PasswordHash    string         `json:"-" gorm:"column:password_hash"`
	DisplayName     string         `json:"display_name" gorm:"column:display_name"`
	AllowedSections Sections       `json:"allowed_sections" gorm:"column:allowed_sections;type:text"`
	IsActive        bool           `json:"is_active" gorm:"default:true"`
	LastLogin       *time.Time     `json:"last_login"`
	CreatedAt       time.Time      `json:"created_at"`
	UpdatedAt       time.Time      `json:"updated_at"`
	DeletedAt       gorm.DeletedAt `json:"-" gorm:"index"`
}

// CreateAdminRequest — запрос на создание админ-учётки.
type CreateAdminRequest struct {
	Username        string   `json:"username" binding:"required"`
	Password        string   `json:"password" binding:"required"`
	DisplayName     string   `json:"display_name"`
	AllowedSections []string `json:"allowed_sections"`
}

// UpdateAdminRequest — обновление админ-учётки (пароль/разделы/активность/имя).
type UpdateAdminRequest struct {
	ID              uuid.UUID `json:"id" binding:"required"`
	Password        string    `json:"password"`
	DisplayName     string    `json:"display_name"`
	AllowedSections []string  `json:"allowed_sections"`
	IsActive        *bool     `json:"is_active"`
}

// GetAdminsResponse — список админ-учёток.
type GetAdminsResponse struct {
	Admins []Admin `json:"admins"`
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
	Token           string    `json:"token"`
	ExpiresAt       time.Time `json:"expires_at"`
	IsAdmin         bool      `json:"is_admin"`
	Role            string    `json:"role,omitempty"`
	AllowedSections []string  `json:"allowed_sections,omitempty"`
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
	ID              uuid.UUID `json:"id"`
	Username        string    `json:"username"`
	IsAdmin         bool      `json:"is_admin"`
	Role            string    `json:"role,omitempty"` // super_admin, limited_admin, cashier, cleaner, web
	AllowedSections []string  `json:"allowed_sections,omitempty"`
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

// WebRegisterSendCodeRequest запрос на отправку кода при регистрации
type WebRegisterSendCodeRequest struct {
	Email           string `json:"email" binding:"required"`
	Password        string `json:"password" binding:"required"`
	PasswordConfirm string `json:"password_confirm" binding:"required"`
}

// WebRegisterVerifyRequest запрос на верификацию email и создание пользователя
type WebRegisterVerifyRequest struct {
	Email           string `json:"email" binding:"required"`
	Code            string `json:"code" binding:"required"`
	Password        string `json:"password" binding:"required"`
	PasswordConfirm string `json:"password_confirm" binding:"required"`
}

// WebLoginRequest запрос на вход по email и паролю
type WebLoginRequest struct {
	Email    string `json:"email" binding:"required"`
	Password string `json:"password" binding:"required"`
}

// WebAuthResponse ответ с JWT и пользователем (регистрация, вход)
type WebAuthResponse struct {
	Token     string    `json:"token"`
	ExpiresAt time.Time `json:"expires_at"`
	User      WebUser   `json:"user"`
}

// WebUser минимальные данные пользователя для веб-клиента
type WebUser struct {
	ID               uuid.UUID `json:"id"`
	Email            string    `json:"email"`
	TelegramID       *int64    `json:"telegram_id,omitempty"`
	CarNumber        string    `json:"car_number"`
	CarNumberCountry string    `json:"car_number_country"`
}

// WebChangePasswordSendCodeRequest запрос на отправку кода для смены пароля
type WebChangePasswordSendCodeRequest struct {
	Email string `json:"email" binding:"required"`
}

// WebChangePasswordConfirmRequest запрос на смену пароля после ввода кода
type WebChangePasswordConfirmRequest struct {
	Email             string `json:"email" binding:"required"`
	Code              string `json:"code" binding:"required"`
	NewPassword       string `json:"new_password" binding:"required"`
	NewPasswordConfirm string `json:"new_password_confirm" binding:"required"`
}
