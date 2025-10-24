package models

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// Статусы сессии
const (
	SessionStatusCreated       = "created"        // Создана
	SessionStatusInQueue       = "in_queue"       // Оплачено, в очереди
	SessionStatusPaymentFailed = "payment_failed" // Ошибка оплаты
	SessionStatusAssigned      = "assigned"       // Назначена на бокс
	SessionStatusActive        = "active"         // Активна (клиент приступил к мойке)
	SessionStatusComplete      = "complete"       // Завершена
	SessionStatusCanceled      = "canceled"       // Отменена
)

// Session представляет сессию мойки
type Session struct {
	ID                            uuid.UUID      `json:"id" gorm:"primaryKey;type:uuid;default:gen_random_uuid()"`
	UserID                        uuid.UUID      `json:"user_id" gorm:"index;type:uuid"`
	BoxID                         *uuid.UUID     `json:"box_id,omitempty" gorm:"index;type:uuid"`
	BoxNumber                     *int           `json:"box_number,omitempty" gorm:"-"` // Виртуальное поле, не хранится в БД
	Status                        string         `json:"status" gorm:"default:created;index"`
	ServiceType                   string         `json:"service_type,omitempty" gorm:"default:null"`
	WithChemistry                 bool           `json:"with_chemistry" gorm:"default:false"`
	WasChemistryOn                bool           `json:"was_chemistry_on" gorm:"default:false"`             // Была ли фактически включена химия
	ChemistryTimeMinutes          int            `json:"chemistry_time_minutes" gorm:"default:0"`           // Выбранное время химии в минутах
	ChemistryStartedAt            *time.Time     `json:"chemistry_started_at,omitempty"`                    // Когда была включена химия
	ChemistryEndedAt              *time.Time     `json:"chemistry_ended_at,omitempty"`                      // Когда была выключена химия
	CarNumber                     string         `json:"car_number"`                                        // Номер машины в сессии
	RentalTimeMinutes             int            `json:"rental_time_minutes" gorm:"default:5"`              // Время мойки в минутах
	ExtensionTimeMinutes          int            `json:"extension_time_minutes" gorm:"default:0"`           // Время продления в минутах
	RequestedExtensionTimeMinutes int            `json:"requested_extension_time_minutes" gorm:"default:0"` // Запрошенное время продления в минутах
	RequestedExtensionChemistryTimeMinutes int `json:"requested_extension_chemistry_time_minutes" gorm:"default:0"` // Запрошенное время химии при продлении в минутах
	ExtensionChemistryTimeMinutes int            `json:"extension_chemistry_time_minutes" gorm:"default:0"` // Время химии при продлении в минутах
	Payment                       *Payment       `json:"payment,omitempty" gorm:"-"`                        // Информация о платеже (не хранится в БД)
	MainPayment                   *Payment       `json:"main_payment,omitempty" gorm:"-"`                   // Основной платеж (не хранится в БД)
	ExtensionPayments             []Payment      `json:"extension_payments,omitempty" gorm:"-"`             // Платежи продления (не хранится в БД)
	IdempotencyKey                string         `json:"idempotency_key,omitempty" gorm:"index"`
	IsExpiringNotificationSent    bool           `json:"is_expiring_notification_sent" gorm:"default:false"`
	IsCompletingNotificationSent  bool           `json:"is_completing_notification_sent" gorm:"default:false"`
	CreatedAt                     time.Time      `json:"created_at"`
	UpdatedAt                     time.Time      `json:"updated_at"`
	StatusUpdatedAt               time.Time      `json:"status_updated_at"` // Время последнего обновления статуса
	SessionTimeoutMinutes         int            `json:"session_timeout_minutes" gorm:"-"` // Время ожидания старта мойки в минутах (виртуальное поле)
	DeletedAt                     gorm.DeletedAt `json:"-" gorm:"index"`
}

// ReassignSessionRequest запрос на переназначение сессии на другой бокс
type ReassignSessionRequest struct {
	SessionID uuid.UUID `json:"session_id" binding:"required"`
}

// ReassignSessionResponse ответ на переназначение сессии
type ReassignSessionResponse struct {
	Success bool    `json:"success"`
	Message string  `json:"message"`
	Session Session `json:"session"`
}

// CreateSessionRequest представляет запрос на создание сессии
type CreateSessionRequest struct {
	UserID               uuid.UUID `json:"user_id" binding:"required"`
	ServiceType          string    `json:"service_type" binding:"required"`
	WithChemistry        bool      `json:"with_chemistry"`
	ChemistryTimeMinutes int       `json:"chemistry_time_minutes"` // Выбранное время химии в минутах
	CarNumber            string    `json:"car_number" binding:"required"`
	RentalTimeMinutes    int       `json:"rental_time_minutes" binding:"required"`
	IdempotencyKey       string    `json:"idempotency_key" binding:"required"`
}

// CreateSessionResponse представляет ответ на создание сессии
type CreateSessionResponse struct {
	Session Session `json:"session"`
}

// CreateSessionWithPaymentRequest представляет запрос на создание сессии с платежом
type CreateSessionWithPaymentRequest struct {
	UserID               uuid.UUID `json:"user_id" binding:"required"`
	ServiceType          string    `json:"service_type" binding:"required"`
	WithChemistry        bool      `json:"with_chemistry"`
	ChemistryTimeMinutes int       `json:"chemistry_time_minutes"` // Выбранное время химии в минутах
	CarNumber            string    `json:"car_number" binding:"required"`
	RentalTimeMinutes    int       `json:"rental_time_minutes" binding:"required"`
	IdempotencyKey       string    `json:"idempotency_key" binding:"required"`
}

// CreateSessionWithPaymentResponse представляет ответ на создание сессии с платежом
type CreateSessionWithPaymentResponse struct {
	Session Session  `json:"session"`
	Payment *Payment `json:"payment,omitempty"`
}

// Payment представляет информацию о платеже (для интеграции с payment доменом)
type Payment struct {
	ID             uuid.UUID  `json:"id"`
	SessionID      uuid.UUID  `json:"session_id"`
	Amount         int        `json:"amount"`          // сумма в копейках
	RefundedAmount int        `json:"refunded_amount"` // сумма возврата в копейках
	Currency       string     `json:"currency"`
	Status         string     `json:"status"`
	PaymentType    string     `json:"payment_type"` // тип платежа: main или extension
	PaymentURL     string     `json:"payment_url"`
	TinkoffID      string     `json:"tinkoff_id" gorm:"index"`
	ExpiresAt      *time.Time `json:"expires_at"`
	RefundedAt     *time.Time `json:"refunded_at,omitempty"` // время возврата
	CreatedAt      time.Time  `json:"created_at"`
	UpdatedAt      time.Time  `json:"updated_at"`
}

// GetUserSessionRequest представляет запрос на получение сессии пользователя
type GetUserSessionRequest struct {
	UserID uuid.UUID `json:"user_id" binding:"required"`
}

// GetUserSessionResponse представляет ответ на получение сессии пользователя
type GetUserSessionResponse struct {
	Session *Session `json:"session"`
	Payment *Payment `json:"payment,omitempty"`
}

// GetSessionRequest представляет запрос на получение сессии по ID
type GetSessionRequest struct {
	SessionID uuid.UUID `json:"session_id" binding:"required"`
}

// GetSessionResponse представляет ответ на получение сессии
type GetSessionResponse struct {
	Session *Session `json:"session"`
	Payment *Payment `json:"payment,omitempty"`
}

// StartSessionRequest представляет запрос на запуск сессии
type StartSessionRequest struct {
	SessionID uuid.UUID `json:"session_id" binding:"required"`
}

// StartSessionResponse представляет ответ на запуск сессии
type StartSessionResponse struct {
	Session *Session `json:"session"`
}

// CompleteSessionRequest представляет запрос на завершение сессии
type CompleteSessionRequest struct {
	SessionID uuid.UUID `json:"session_id" binding:"required"`
}

// CompleteSessionResponse представляет ответ на завершение сессии
type CompleteSessionResponse struct {
	Session *Session `json:"session"`
	Payment *Payment `json:"payment,omitempty"`
}

// GetUserSessionHistoryRequest представляет запрос на получение истории сессий пользователя
type GetUserSessionHistoryRequest struct {
	UserID uuid.UUID `json:"user_id" binding:"required"`
	Limit  int       `json:"limit"`
	Offset int       `json:"offset"`
}

// GetUserSessionHistoryResponse представляет ответ на получение истории сессий пользователя
type GetUserSessionHistoryResponse struct {
	Sessions []Session `json:"sessions"`
}

// ExtendSessionRequest представляет запрос на продление сессии
type ExtendSessionRequest struct {
	SessionID            uuid.UUID `json:"session_id" binding:"required"`
	ExtensionTimeMinutes int       `json:"extension_time_minutes" binding:"required"`
}

// ExtendSessionResponse представляет ответ на продление сессии
type ExtendSessionResponse struct {
	Session *Session `json:"session"`
}

// ExtendSessionWithPaymentRequest представляет запрос на продление сессии с оплатой
type ExtendSessionWithPaymentRequest struct {
	SessionID                     uuid.UUID `json:"session_id" binding:"required"`
	ExtensionTimeMinutes          int       `json:"extension_time_minutes"`
	ExtensionChemistryTimeMinutes int       `json:"extension_chemistry_time_minutes"` // Время химии при продлении (опционально)
}

// ExtendSessionWithPaymentResponse представляет ответ на продление сессии с оплатой
type ExtendSessionWithPaymentResponse struct {
	Session *Session `json:"session"`
	Payment *Payment `json:"payment,omitempty"`
}

// GetSessionPaymentsRequest представляет запрос на получение платежей сессии
type GetSessionPaymentsRequest struct {
	SessionID uuid.UUID `json:"session_id" binding:"required"`
}

// GetSessionPaymentsResponse представляет ответ на получение платежей сессии
type GetSessionPaymentsResponse struct {
	MainPayment       *Payment  `json:"main_payment,omitempty"`
	ExtensionPayments []Payment `json:"extension_payments,omitempty"`
}

// CancelSessionRequest представляет запрос на отмену сессии
type CancelSessionRequest struct {
	SessionID  uuid.UUID `json:"session_id" binding:"required"`
	UserID     uuid.UUID `json:"user_id" binding:"required"`
	SkipRefund bool      `json:"skip_refund"` // Пропустить возврат средств
}

// CancelSessionResponse представляет ответ на отмену сессии
type CancelSessionResponse struct {
	Session Session  `json:"session"`
	Payment *Payment `json:"payment,omitempty"`
	Refund  *Refund  `json:"refund,omitempty"`
}

// Refund представляет информацию о возврате (импорт из payment domain)
type Refund struct {
	ID        uuid.UUID `json:"id"`
	PaymentID uuid.UUID `json:"payment_id"`
	Amount    int       `json:"amount"` // сумма возврата в копейках
	Status    string    `json:"status"`
	CreatedAt time.Time `json:"created_at"`
}

// Административные модели

// AdminListSessionsRequest запрос на получение списка сессий с фильтрацией
type AdminListSessionsRequest struct {
	UserID      *uuid.UUID `json:"user_id"`
	BoxID       *uuid.UUID `json:"box_id"`
	BoxNumber   *int       `json:"box_number"`
	Status      *string    `json:"status" binding:"omitempty,oneof=created in_queue payment_failed assigned active complete canceled"`
	ServiceType *string    `json:"service_type" binding:"omitempty,oneof=wash air_dry vacuum"`
	DateFrom    *time.Time `json:"date_from"`
	DateTo      *time.Time `json:"date_to"`
	Limit       *int       `json:"limit"`
	Offset      *int       `json:"offset"`
}

// AdminGetSessionRequest запрос на получение сессии по ID
type AdminGetSessionRequest struct {
	ID uuid.UUID `json:"id" binding:"required"`
}

// AdminListSessionsResponse ответ на получение списка сессий
type AdminListSessionsResponse struct {
	Sessions []Session `json:"sessions"`
	Total    int       `json:"total"`
	Limit    int       `json:"limit"`
	Offset   int       `json:"offset"`
}

// AdminGetSessionResponse ответ на получение сессии
type AdminGetSessionResponse struct {
	Session Session `json:"session"`
}

// CashierSessionsRequest представляет запрос на получение сессий кассира
type CashierSessionsRequest struct {
	ShiftStartedAt time.Time `json:"shift_started_at" binding:"required"`
	Limit          int       `json:"limit"`
	Offset         int       `json:"offset"`
}

// CashierActiveSessionsRequest представляет запрос на получение активных сессий кассира
type CashierActiveSessionsRequest struct {
	Limit  int `json:"limit"`
	Offset int `json:"offset"`
}

// CashierActiveSessionsResponse представляет ответ на получение активных сессий кассира
type CashierActiveSessionsResponse struct {
	Sessions []Session `json:"sessions"`
	Total    int       `json:"total"`
	Limit    int       `json:"limit"`
	Offset   int       `json:"offset"`
}

// CashierStartSessionRequest представляет запрос на запуск сессии кассиром
type CashierStartSessionRequest struct {
	SessionID uuid.UUID `json:"session_id" binding:"required"`
}

// CashierCompleteSessionRequest представляет запрос на завершение сессии кассиром
type CashierCompleteSessionRequest struct {
	SessionID uuid.UUID `json:"session_id" binding:"required"`
}

// CashierCancelSessionRequest представляет запрос на отмену сессии кассиром
type CashierCancelSessionRequest struct {
	SessionID  uuid.UUID `json:"session_id" binding:"required"`
	SkipRefund bool      `json:"skip_refund"` // Пропустить возврат средств
}

// CashierStartSessionResponse представляет ответ на запуск сессии кассиром
type CashierStartSessionResponse struct {
	Session Session `json:"session"`
}

// CashierCompleteSessionResponse представляет ответ на завершение сессии кассиром
type CashierCompleteSessionResponse struct {
	Session Session `json:"session"`
}

// CashierCancelSessionResponse представляет ответ на отмену сессии кассиром
type CashierCancelSessionResponse struct {
	Session Session `json:"session"`
}

// EnableChemistryRequest представляет запрос на включение химии
type EnableChemistryRequest struct {
	SessionID uuid.UUID `json:"session_id" binding:"required"`
}

// EnableChemistryResponse представляет ответ на включение химии
type EnableChemistryResponse struct {
	Session Session `json:"session"`
}

// ChemistryStats представляет статистику использования химии
type ChemistryStats struct {
	TotalSessionsWithChemistry int     `json:"total_sessions_with_chemistry"` // Общее количество сессий с химией
	TotalChemistryEnabled      int     `json:"total_chemistry_enabled"`       // Количество сессий где химия была включена
	UsagePercentage            float64 `json:"usage_percentage"`              // Процент использования химии
	Period                     string  `json:"period"`                        // Период статистики
}

// GetChemistryStatsRequest представляет запрос на получение статистики химии
type GetChemistryStatsRequest struct {
	DateFrom *time.Time `json:"date_from"`
	DateTo   *time.Time `json:"date_to"`
}

// GetChemistryStatsResponse представляет ответ на получение статистики химии
type GetChemistryStatsResponse struct {
	Stats ChemistryStats `json:"stats"`
}
