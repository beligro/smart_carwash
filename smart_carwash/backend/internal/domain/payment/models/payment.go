package models

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// Статусы платежа
const (
	PaymentStatusPending   = "pending"   // Ожидает оплаты
	PaymentStatusSucceeded = "succeeded" // Оплачен
	PaymentStatusFailed    = "failed"    // Ошибка оплаты
	PaymentStatusRefunded  = "refunded"  // Возвращен
)

// Payment представляет платеж
type Payment struct {
	ID             uuid.UUID      `json:"id" gorm:"primaryKey;type:uuid;default:gen_random_uuid()"`
	SessionID      uuid.UUID      `json:"session_id" gorm:"index;type:uuid;not null"`
	Amount         int            `json:"amount" gorm:"not null"` // сумма в копейках
	RefundedAmount int            `json:"refunded_amount" gorm:"default:0"` // сумма возврата в копейках
	Currency       string         `json:"currency" gorm:"default:RUB"`
	Status         string         `json:"status" gorm:"default:pending;index"`
	PaymentURL     string         `json:"payment_url"`
	TinkoffID      string         `json:"tinkoff_id" gorm:"index"`
	ExpiresAt      *time.Time     `json:"expires_at"`
	RefundedAt     *time.Time     `json:"refunded_at,omitempty"` // время возврата
	CreatedAt      time.Time      `json:"created_at"`
	UpdatedAt      time.Time      `json:"updated_at"`
	DeletedAt      gorm.DeletedAt `json:"-" gorm:"index"`
}

// CalculatePriceRequest представляет запрос на расчет цены
type CalculatePriceRequest struct {
	ServiceType       string `json:"service_type" binding:"required"`
	WithChemistry     bool   `json:"with_chemistry"`
	RentalTimeMinutes int    `json:"rental_time_minutes" binding:"required"`
}

// CalculatePriceResponse представляет ответ на расчет цены
type CalculatePriceResponse struct {
	Price     int           `json:"price"` // в копейках
	Currency  string        `json:"currency"`
	Breakdown PriceBreakdown `json:"breakdown"`
}

// PriceBreakdown представляет разбивку цены
type PriceBreakdown struct {
	BasePrice     int `json:"base_price"`     // базовая цена
	ChemistryPrice int `json:"chemistry_price"` // цена за химию
}

// CreatePaymentRequest представляет запрос на создание платежа
type CreatePaymentRequest struct {
	SessionID uuid.UUID `json:"session_id" binding:"required"`
	Amount    int       `json:"amount" binding:"required"`
	Currency  string    `json:"currency" binding:"required"`
}

// CreatePaymentResponse представляет ответ на создание платежа
type CreatePaymentResponse struct {
	Payment Payment `json:"payment"`
}

// GetPaymentStatusRequest представляет запрос на получение статуса платежа
type GetPaymentStatusRequest struct {
	PaymentID uuid.UUID `json:"payment_id" binding:"required"`
}

// GetPaymentStatusResponse представляет ответ на получение статуса платежа
type GetPaymentStatusResponse struct {
	Payment Payment `json:"payment"`
}

// RetryPaymentRequest представляет запрос на повторную попытку оплаты
type RetryPaymentRequest struct {
	SessionID uuid.UUID `json:"session_id" binding:"required"`
}

// RetryPaymentResponse представляет ответ на повторную попытку оплаты
type RetryPaymentResponse struct {
	Payment Payment `json:"payment"`
}

// WebhookRequest представляет запрос webhook от Tinkoff
type WebhookRequest struct {
	TerminalKey string `json:"TerminalKey"`
	OrderId     string `json:"OrderId"`
	Success     bool   `json:"Success"`
	Status      string `json:"Status"`
	PaymentId   int64  `json:"PaymentId"`
	ErrorCode   string `json:"ErrorCode"`
	Amount      int    `json:"Amount"`
	Signature   string `json:"Signature"`
}

// AdminListPaymentsRequest запрос на получение списка платежей с фильтрацией
type AdminListPaymentsRequest struct {
	SessionID *uuid.UUID `json:"session_id"`
	Status    *string    `json:"status" binding:"omitempty,oneof=pending succeeded failed refunded"`
	DateFrom  *time.Time `json:"date_from"`
	DateTo    *time.Time `json:"date_to"`
	Limit     *int       `json:"limit"`
	Offset    *int       `json:"offset"`
}

// AdminListPaymentsResponse ответ на получение списка платежей
type AdminListPaymentsResponse struct {
	Payments []Payment `json:"payments"`
	Total    int       `json:"total"`
	Limit    int       `json:"limit"`
	Offset   int       `json:"offset"`
}

// RefundPaymentRequest представляет запрос на возврат платежа
type RefundPaymentRequest struct {
	PaymentID uuid.UUID `json:"payment_id" binding:"required"`
	Amount    int       `json:"amount" binding:"required"` // сумма возврата в копейках
}

// RefundPaymentResponse представляет ответ на возврат платежа
type RefundPaymentResponse struct {
	Payment Payment `json:"payment"`
	Refund  Refund  `json:"refund"`
}

// Refund представляет информацию о возврате
type Refund struct {
	ID        uuid.UUID  `json:"id"`
	PaymentID uuid.UUID  `json:"payment_id"`
	Amount    int        `json:"amount"` // сумма возврата в копейках
	Status    string     `json:"status"`
	CreatedAt time.Time  `json:"created_at"`
} 