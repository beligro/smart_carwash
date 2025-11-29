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
	CarNumber            string    `json:"car_number"`         // Опциональный номер машины
	CarNumberCountry     string    `json:"car_number_country"` // Опциональная страна гос номера
}

// CashierPaymentResponse представляет ответ на запрос от 1C
type CashierPaymentResponse struct {
	Success   bool   `json:"success"`
	SessionID string `json:"session_id,omitempty"`
	PaymentID string `json:"payment_id,omitempty"`
	Message   string `json:"message,omitempty"`
}

// ExtendSession1CRequest представляет запрос от 1C для продления сессии без платежа
type ExtendSession1CRequest struct {
	CarNumber                     string `json:"car_number" binding:"required"`              // Номер машины (обязательный)
	ExtensionTimeMinutes          *int   `json:"extension_time_minutes,omitempty"`           // Время продления мойки в минутах (опционально)
	ExtensionChemistryTimeMinutes *int   `json:"extension_chemistry_time_minutes,omitempty"` // Время продления химии в минутах (опционально)
	Amount                        int    `json:"amount"`                                     // Стоимость продления (для логирования, без фактического платежа)
}

// ExtendSession1CResponse представляет ответ на запрос продления сессии от 1C
type ExtendSession1CResponse struct {
	Success   bool   `json:"success"`
	SessionID string `json:"session_id,omitempty"`
	Message   string `json:"message,omitempty"`
}

// GetSessionType1CRequest представляет запрос от 1C для получения типа активной сессии
type GetSessionType1CRequest struct {
	CarNumber string `json:"car_number" binding:"required"` // Номер машины (обязательный)
}

// GetSessionType1CResponse представляет ответ на запрос получения типа сессии от 1C
type GetSessionType1CResponse struct {
	Success     bool   `json:"success"`
	ServiceType string `json:"service_type,omitempty"` // Тип услуги: wash, air_dry, vacuum
	Message     string `json:"message,omitempty"`
}
