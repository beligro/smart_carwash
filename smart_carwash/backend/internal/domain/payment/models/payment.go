package models

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// Статусы платежей
const (
	PaymentStatusPending    = "pending"
	PaymentStatusProcessing = "processing"
	PaymentStatusCompleted  = "completed"
	PaymentStatusFailed     = "failed"
	PaymentStatusRefunded   = "refunded"
	PaymentStatusCancelled  = "cancelled"
)

// Типы платежей
const (
	PaymentTypeQueueBooking    = "queue_booking"
	PaymentTypeSessionExtension = "session_extension"
	PaymentTypeRefund          = "refund"
)

// Payment представляет платеж в системе
type Payment struct {
	ID                uuid.UUID      `json:"id" gorm:"primaryKey;type:uuid;default:gen_random_uuid()"`
	UserID            uuid.UUID      `json:"user_id" gorm:"index;type:uuid"`
	SessionID         *uuid.UUID     `json:"session_id,omitempty" gorm:"index;type:uuid"`
	AmountKopecks     int64          `json:"amount_kopecks"` // Сумма в копейках
	Type              string         `json:"type" gorm:"index"`
	Status            string         `json:"status" gorm:"index"`
	Description       string         `json:"description"`
	IdempotencyKey    string         `json:"idempotency_key,omitempty" gorm:"uniqueIndex"`
	TinkoffPaymentID  string         `json:"tinkoff_payment_id,omitempty"`
	PaymentURL        string         `json:"payment_url,omitempty"`
	LastError         string         `json:"last_error,omitempty"`
	CreatedAt         time.Time      `json:"created_at"`
	UpdatedAt         time.Time      `json:"updated_at"`
	DeletedAt         gorm.DeletedAt `json:"-" gorm:"index"`
}

// PaymentRefund представляет возврат средств
type PaymentRefund struct {
	ID                uuid.UUID      `json:"id" gorm:"primaryKey;type:uuid;default:gen_random_uuid()"`
	PaymentID         uuid.UUID      `json:"payment_id" gorm:"index;type:uuid"`
	AmountKopecks     int64          `json:"amount_kopecks"`
	Type              string         `json:"type"` // automatic_refund, full_refund, partial_refund
	Status            string         `json:"status" gorm:"index"`
	RetryCount        int            `json:"retry_count" gorm:"default:0"`
	MaxRetries        int            `json:"max_retries" gorm:"default:5"`
	NextRetryAt       *time.Time     `json:"next_retry_at,omitempty"`
	LastError         string         `json:"last_error,omitempty"`
	IdempotencyKey    string         `json:"idempotency_key,omitempty" gorm:"uniqueIndex"`
	TinkoffRefundID   string         `json:"tinkoff_refund_id,omitempty"`
	CreatedAt         time.Time      `json:"created_at"`
	UpdatedAt         time.Time      `json:"updated_at"`
	DeletedAt         gorm.DeletedAt `json:"-" gorm:"index"`
}

// PaymentEvent представляет событие платежа для идемпотентности
type PaymentEvent struct {
	ID                uuid.UUID      `json:"id" gorm:"primaryKey;type:uuid;default:gen_random_uuid()"`
	PaymentID         uuid.UUID      `json:"payment_id" gorm:"index;type:uuid"`
	EventType         string         `json:"event_type"` // charge, refund, cancel, webhook
	IdempotencyKey    string         `json:"idempotency_key,omitempty" gorm:"uniqueIndex"`
	AmountKopecks     int64          `json:"amount_kopecks"`
	Status            string         `json:"status"`
	TinkoffResponse   string         `json:"tinkoff_response,omitempty"` // JSON ответ от Tinkoff
	CreatedAt         time.Time      `json:"created_at"`
	UpdatedAt         time.Time      `json:"updated_at"`
	DeletedAt         gorm.DeletedAt `json:"-" gorm:"index"`
}

// Запросы и ответы API

// CreatePaymentRequest запрос на создание платежа
type CreatePaymentRequest struct {
	UserID         uuid.UUID `json:"user_id" binding:"required"`
	SessionID      *uuid.UUID `json:"session_id,omitempty"`
	AmountKopecks  int64     `json:"amount_kopecks" binding:"required"`
	Type           string    `json:"type" binding:"required"`
	Description    string    `json:"description"`
	IdempotencyKey string    `json:"idempotency_key" binding:"required"`
}

// CreatePaymentResponse ответ на создание платежа
type CreatePaymentResponse struct {
	Payment    *Payment `json:"payment"`
	PaymentURL string   `json:"payment_url"`
}

// GetPaymentRequest запрос на получение платежа
type GetPaymentRequest struct {
	ID uuid.UUID `json:"id" binding:"required"`
}

// GetPaymentResponse ответ с данными платежа
type GetPaymentResponse struct {
	Payment *Payment `json:"payment"`
}

// GetPaymentStatusRequest запрос на получение статуса платежа
type GetPaymentStatusRequest struct {
	ID uuid.UUID `json:"id" binding:"required"`
}

// GetPaymentStatusResponse ответ со статусом платежа
type GetPaymentStatusResponse struct {
	Status string `json:"status"`
}

// CreateRefundRequest запрос на создание возврата
type CreateRefundRequest struct {
	PaymentID      uuid.UUID `json:"payment_id" binding:"required"`
	AmountKopecks  int64     `json:"amount_kopecks" binding:"required"`
	Type           string    `json:"type" binding:"required"`
	IdempotencyKey string    `json:"idempotency_key" binding:"required"`
}

// CreateRefundResponse ответ на создание возврата
type CreateRefundResponse struct {
	Refund *PaymentRefund `json:"refund"`
}

// AdminListPaymentsRequest запрос на получение списка платежей (админ)
type AdminListPaymentsRequest struct {
	UserID    *uuid.UUID `json:"user_id,omitempty"`
	Status    *string    `json:"status,omitempty"`
	Type      *string    `json:"type,omitempty"`
	DateFrom  *time.Time `json:"date_from,omitempty"`
	DateTo    *time.Time `json:"date_to,omitempty"`
	Limit     int        `json:"limit" binding:"required"`
	Offset    int        `json:"offset"`
}

// AdminListPaymentsResponse ответ со списком платежей (админ)
type AdminListPaymentsResponse struct {
	Payments []Payment `json:"payments"`
	Total    int       `json:"total"`
}

// AdminGetPaymentRequest запрос на получение платежа (админ)
type AdminGetPaymentRequest struct {
	ID uuid.UUID `json:"id" binding:"required"`
}

// AdminGetPaymentResponse ответ с данными платежа (админ)
type AdminGetPaymentResponse struct {
	Payment *Payment        `json:"payment"`
	Refunds []PaymentRefund `json:"refunds,omitempty"`
	Events  []PaymentEvent  `json:"events,omitempty"`
}

// CalculatePriceRequest запрос на расчёт стоимости
type CalculatePriceRequest struct {
	ServiceType       string `json:"service_type" binding:"required"`
	RentalTimeMinutes int    `json:"rental_time_minutes" binding:"required"`
	WithChemistry     bool   `json:"with_chemistry"`
}

// CalculatePriceResponse ответ с расчётом стоимости
type CalculatePriceResponse struct {
	TotalPriceKopecks     int64  `json:"total_price_kopecks"`
	BasePriceKopecks      int64  `json:"base_price_kopecks"`
	ChemistryPriceKopecks int64  `json:"chemistry_price_kopecks"`
	RentalTimeMinutes     int    `json:"rental_time_minutes"`
	ServiceType           string `json:"service_type"`
	WithChemistry         bool   `json:"with_chemistry"`
} 