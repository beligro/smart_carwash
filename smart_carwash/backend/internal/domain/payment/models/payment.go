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

// Типы платежей
const (
	PaymentTypeMain      = "main"      // Основной платеж за сессию
	PaymentTypeExtension = "extension" // Платеж за продление
)

// Payment представляет платеж
type Payment struct {
	ID             uuid.UUID      `json:"id" gorm:"primaryKey;type:uuid;default:gen_random_uuid()"`
	SessionID      uuid.UUID      `json:"session_id" gorm:"index;type:uuid;not null"`
	Amount         int            `json:"amount" gorm:"not null"` // сумма в копейках
	RefundedAmount int            `json:"refunded_amount" gorm:"default:0"` // сумма возврата в копейках
	Currency       string         `json:"currency" gorm:"default:RUB"`
	Status         string         `json:"status" gorm:"default:pending;index"`
	PaymentType    string         `json:"payment_type" gorm:"default:main;index"` // тип платежа: main или extension
	PaymentMethod  string         `json:"payment_method" gorm:"default:tinkoff"` // метод платежа: tinkoff, cashier
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
	ServiceType          string `json:"service_type" binding:"required"`
	WithChemistry        bool   `json:"with_chemistry"`
	ChemistryTimeMinutes int    `json:"chemistry_time_minutes"` // Выбранное время химии в минутах
	RentalTimeMinutes    int    `json:"rental_time_minutes" binding:"required"`
}

// CalculatePriceResponse представляет ответ на расчет цены
type CalculatePriceResponse struct {
	Price     int           `json:"price"` // в копейках
	Currency  string        `json:"currency"`
	Breakdown PriceBreakdown `json:"breakdown"`
}

// CalculateExtensionPriceRequest представляет запрос на расчет цены продления
type CalculateExtensionPriceRequest struct {
	ServiceType           string `json:"service_type" binding:"required"`
	ExtensionTimeMinutes  int    `json:"extension_time_minutes" binding:"required"`
	WithChemistry         bool   `json:"with_chemistry"`
	ExtensionChemistryTimeMinutes int `json:"extension_chemistry_time_minutes"` // Время химии при продлении (опционально)
}

// CalculateExtensionPriceResponse представляет ответ на расчет цены продления
type CalculateExtensionPriceResponse struct {
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
	Email     string    `json:"email"` // Email для чека
}

// CreatePaymentResponse представляет ответ на создание платежа
type CreatePaymentResponse struct {
	Payment Payment `json:"payment"`
}

// CreateExtensionPaymentRequest представляет запрос на создание платежа продления
type CreateExtensionPaymentRequest struct {
	SessionID uuid.UUID `json:"session_id" binding:"required"`
	Amount    int       `json:"amount" binding:"required"`
	Currency  string    `json:"currency" binding:"required"`
	Email     string    `json:"email"` // Email для чека
}

// CreateExtensionPaymentResponse представляет ответ на создание платежа продления
type CreateExtensionPaymentResponse struct {
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
	PaymentID     *uuid.UUID `json:"payment_id"`
	SessionID     *uuid.UUID `json:"session_id"`
	UserID        *uuid.UUID `json:"user_id"`
	Status        *string    `json:"status" binding:"omitempty,oneof=pending succeeded failed refunded"`
	PaymentType   *string    `json:"payment_type" binding:"omitempty,oneof=main extension"`
	PaymentMethod *string    `json:"payment_method" binding:"omitempty,oneof=tinkoff cashier"`
	DateFrom      *time.Time `json:"date_from"`
	DateTo        *time.Time `json:"date_to"`
	Limit         *int       `json:"limit"`
	Offset        *int       `json:"offset"`
}

// AdminListPaymentsResponse ответ на получение списка платежей
type AdminListPaymentsResponse struct {
	Payments []Payment `json:"payments"`
	Total    int       `json:"total"`
	Limit    int       `json:"limit"`
	Offset   int       `json:"offset"`
}

// AdminRefundPaymentRequest представляет запрос на возврат платежа (админка)
type AdminRefundPaymentRequest struct {
	PaymentID uuid.UUID `json:"payment_id" binding:"required"`
	Amount    int       `json:"amount" binding:"required"` // сумма возврата в копейках
}

// AdminRefundPaymentResponse представляет ответ на возврат платежа (админка)
type AdminRefundPaymentResponse struct {
	Payment Payment `json:"payment"`
	Refund  Refund  `json:"refund"`
	Success bool     `json:"success"`
	Message string   `json:"message"`
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

// CalculatePartialRefundRequest представляет запрос на расчет частичного возврата
type CalculatePartialRefundRequest struct {
	PaymentID         uuid.UUID `json:"payment_id" binding:"required"`
	ServiceType       string    `json:"service_type" binding:"required"`
	RentalTimeMinutes int       `json:"rental_time_minutes" binding:"required"`
	ExtensionTimeMinutes int    `json:"extension_time_minutes"`
	UsedTimeSeconds   int       `json:"used_time_seconds" binding:"required"` // использованное время в секундах
}

// CalculatePartialRefundResponse представляет ответ на расчет частичного возврата
type CalculatePartialRefundResponse struct {
	RefundAmount         int `json:"refund_amount"` // сумма возврата в копейках
	UsedTimeSeconds      int `json:"used_time_seconds"` // использованное время в секундах
	UnusedTimeSeconds    int `json:"unused_time_seconds"` // неиспользованное время в секундах
	PricePerSecond       int `json:"price_per_second"` // цена за секунду в копейках
	TotalSessionSeconds   int `json:"total_session_seconds"` // общее время сессии в секундах
}

// GetPaymentsBySessionRequest представляет запрос на получение платежей сессии
type GetPaymentsBySessionRequest struct {
	SessionID uuid.UUID `json:"session_id" binding:"required"`
}

// GetPaymentsBySessionResponse представляет ответ на получение платежей сессии
type GetPaymentsBySessionResponse struct {
	MainPayment       *Payment   `json:"main_payment,omitempty"`
	ExtensionPayments []Payment  `json:"extension_payments,omitempty"`
}

// PaymentStatisticsRequest представляет запрос на получение статистики платежей
type PaymentStatisticsRequest struct {
	UserID      *uuid.UUID `json:"user_id"`
	DateFrom    *time.Time `json:"date_from"`
	DateTo      *time.Time `json:"date_to"`
	ServiceType *string    `json:"service_type" binding:"omitempty,oneof=wash air_dry vacuum"`
}

// ServiceTypeStatistics представляет статистику по типу услуги
type ServiceTypeStatistics struct {
	ServiceType string `json:"service_type"` // wash, air_dry, vacuum
	WithChemistry bool `json:"with_chemistry"` // true для "мойка+химия", false для "мойка"
	SessionCount int  `json:"session_count"`   // количество сессий
	TotalAmount  int  `json:"total_amount"`    // общая сумма в копейках (с учетом возвратов)
}

// PaymentStatisticsResponse представляет ответ со статистикой платежей
type PaymentStatisticsResponse struct {
	Statistics []ServiceTypeStatistics `json:"statistics"` // статистика по типам услуг (с учетом возвратов)
	Total      ServiceTypeStatistics   `json:"total"`      // итоговая статистика (с учетом возвратов)
	Period     string                 `json:"period"`     // описание периода
}

// CashierPaymentsRequest представляет запрос на получение платежей кассира
type CashierPaymentsRequest struct {
	ShiftStartedAt time.Time `json:"shift_started_at" binding:"required"`
	Limit          *int      `json:"limit"`
	Offset         *int      `json:"offset"`
}

// CashierLastShiftStatisticsRequest представляет запрос на получение статистики последней смены кассира
type CashierLastShiftStatisticsRequest struct {
	CashierID uuid.UUID `json:"cashier_id" binding:"required"`
}

// CashierShiftStatistics представляет статистику смены кассира
type CashierShiftStatistics struct {
	ShiftStartedAt time.Time `json:"shift_started_at"`
	ShiftEndedAt   time.Time `json:"shift_ended_at"`
	CashierSessions []ServiceTypeStatistics `json:"cashier_sessions"` // сессии через кассира
	MiniAppSessions []ServiceTypeStatistics `json:"mini_app_sessions"` // сессии из mini app
	TotalSessions   []ServiceTypeStatistics `json:"total_sessions"`   // общий итог
}

// CashierLastShiftStatisticsResponse представляет ответ со статистикой последней смены кассира
type CashierLastShiftStatisticsResponse struct {
	Statistics *CashierShiftStatistics `json:"statistics,omitempty"`
	Message    string                 `json:"message"`
	HasShift   bool                   `json:"has_shift"`
} 