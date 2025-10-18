package models

import (
	"time"
)

// CashierPaymentRequest представляет запрос от 1C для платежа через кассира
type CashierPaymentRequest struct {
	ServiceType          string    `json:"service_type" binding:"required,oneof=wash air_dry vacuum"`
	WithChemistry        bool      `json:"with_chemistry"`
	ChemistryTimeMinutes int       `json:"chemistry_time_minutes"` // Выбранное время химии в минутах
	PaymentTime          time.Time `json:"payment_time" binding:"required"`
	Amount               int       `json:"amount" binding:"required,min=1"` // сумма в копейках
	RentalTimeMinutes    int       `json:"rental_time_minutes" binding:"required,min=1"`
	CarNumber            string    `json:"car_number"` // Опциональный номер машины
}

// CashierPaymentResponse представляет ответ на запрос от 1C
type CashierPaymentResponse struct {
	Success   bool     `json:"success"`
	SessionID string   `json:"session_id,omitempty"`
	PaymentID string   `json:"payment_id,omitempty"`
	Message   string   `json:"message,omitempty"`
} 