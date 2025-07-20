package models

import (
	"time"
)

// Tinkoff API модели

// InitRequest запрос на инициализацию платежа в Tinkoff
type InitRequest struct {
	TerminalKey     string            `json:"TerminalKey"`
	Amount          int64             `json:"Amount"`          // в копейках
	OrderId         string            `json:"OrderId"`         // наш UUID
	Description     string            `json:"Description"`
	SuccessURL      string            `json:"SuccessURL"`
	FailURL         string            `json:"FailURL"`
	NotificationURL string            `json:"NotificationURL"`
	Token           string            `json:"Token"`
}

// InitResponse ответ на инициализацию платежа
type InitResponse struct {
	Success    bool   `json:"Success"`
	ErrorCode  string `json:"ErrorCode"`
	Message    string `json:"Message"`
	PaymentId  string `json:"PaymentId"`
	PaymentURL string `json:"PaymentUrl"`
}

// GetStateRequest запрос на получение статуса платежа
type GetStateRequest struct {
	TerminalKey string `json:"TerminalKey"`
	PaymentId   string `json:"PaymentId"`
}

// GetStateResponse ответ с состоянием платежа
type GetStateResponse struct {
	Success     bool   `json:"Success"`
	ErrorCode   string `json:"ErrorCode"`
	Message     string `json:"Message"`
	PaymentId   string `json:"PaymentId"`
	OrderId     string `json:"OrderId"`
	Status      string `json:"Status"` // NEW, FORM_SHOWED, DEADLINE_EXPIRED, CANCELED, PREAUTHORIZING, AUTHORIZING, AUTH_FAIL, AUTHORIZED, CONFIRMING, CONFIRMED, REFUNDING, PARTIAL_REFUNDED, REFUNDED, REJECTING, REJECTED, UNKNOWN
	Amount      int64  `json:"Amount"`
	Pan         string `json:"Pan"`         // маскированный номер карты
	ExpDate     string `json:"ExpDate"`     // срок действия карты
	CardType    string `json:"CardType"`    // тип карты
	PaymentURL  string `json:"PaymentUrl"`
}

// CancelRequest запрос на отмену платежа
type CancelRequest struct {
	TerminalKey string `json:"TerminalKey"`
	PaymentId   string `json:"PaymentId"`
	Amount      int64  `json:"Amount,omitempty"` // если не указан, отменяется весь платеж
}

// CancelResponse ответ на отмену платежа
type CancelResponse struct {
	Success    bool   `json:"Success"`
	ErrorCode  string `json:"ErrorCode"`
	Message    string `json:"Message"`
	PaymentId  string `json:"PaymentId"`
	OrderId    string `json:"OrderId"`
	Status     string `json:"Status"`
	Amount     int64  `json:"Amount"`
	RefundId   string `json:"RefundId"`
}

// RefundRequest запрос на возврат средств
type RefundRequest struct {
	TerminalKey string   `json:"TerminalKey"`
	PaymentId   string   `json:"PaymentId"`
	Amount      int64    `json:"Amount"`
	Receipt     *Receipt `json:"Receipt,omitempty"`
}

// RefundResponse ответ на возврат средств
type RefundResponse struct {
	Success    bool   `json:"Success"`
	ErrorCode  string `json:"ErrorCode"`
	Message    string `json:"Message"`
	PaymentId  string `json:"PaymentId"`
	OrderId    string `json:"OrderId"`
	Status     string `json:"Status"`
	Amount     int64  `json:"Amount"`
	RefundId   string `json:"RefundId"`
}

// ConfirmRequest запрос на подтверждение платежа
type ConfirmRequest struct {
	TerminalKey string   `json:"TerminalKey"`
	PaymentId   string   `json:"PaymentId"`
	Amount      int64    `json:"Amount"`
	Receipt     *Receipt `json:"Receipt,omitempty"`
}

// ConfirmResponse ответ на подтверждение платежа
type ConfirmResponse struct {
	Success    bool   `json:"Success"`
	ErrorCode  string `json:"ErrorCode"`
	Message    string `json:"Message"`
	PaymentId  string `json:"PaymentId"`
	OrderId    string `json:"OrderId"`
	Status     string `json:"Status"`
	Amount     int64  `json:"Amount"`
}

// GetOperationsRequest запрос на получение операций
type GetOperationsRequest struct {
	TerminalKey string `json:"TerminalKey"`
	StartDate   string `json:"StartDate"` // YYYY-MM-DD
	EndDate     string `json:"EndDate"`   // YYYY-MM-DD
}

// GetOperationsResponse ответ с операциями
type GetOperationsResponse struct {
	Success     bool        `json:"Success"`
	ErrorCode   string      `json:"ErrorCode"`
	Message     string      `json:"Message"`
	Operations  []Operation `json:"Operations"`
}

// Operation операция в Tinkoff
type Operation struct {
	PaymentId   string    `json:"PaymentId"`
	OrderId     string    `json:"OrderId"`
	Status      string    `json:"Status"`
	Amount      int64     `json:"Amount"`
	Operation   string    `json:"Operation"` // PAY, REFUND, CANCEL
	DateTime    time.Time `json:"DateTime"`
	CardType    string    `json:"CardType"`
	CardNumber  string    `json:"CardNumber"`
}

// Receipt чек для фискализации
type Receipt struct {
	Email    string     `json:"Email,omitempty"`
	Phone    string     `json:"Phone,omitempty"`
	Taxation string     `json:"Taxation"` // usn_income, usn_income_outcome, patent, basic, simplified, simplified_without_vat, unified, vat
	Items    []ReceiptItem `json:"Items"`
}

// ReceiptItem товар в чеке
type ReceiptItem struct {
	Name     string `json:"Name"`
	Price    int64  `json:"Price"`    // в копейках
	Quantity int    `json:"Quantity"`
	Amount   int64  `json:"Amount"`   // в копейках
	Tax      string `json:"Tax"`      // none, vat0, vat10, vat18, vat110, vat118
}

// Webhook модели

// TinkoffWebhook webhook от Tinkoff
type TinkoffWebhook struct {
	TerminalKey string            `json:"TerminalKey"`
	OrderId     string            `json:"OrderId"`
	Success     bool              `json:"Success"`
	Status      string            `json:"Status"`
	PaymentId   int64             `json:"PaymentId"`
	ErrorCode   string            `json:"ErrorCode,omitempty"`
	Message     string            `json:"Message,omitempty"`
	Amount      int64             `json:"Amount"`
	Pan         string            `json:"Pan,omitempty"`
	ExpDate     string            `json:"ExpDate,omitempty"`
	CardType    string            `json:"CardType,omitempty"`
	Data        map[string]string `json:"Data,omitempty"`
}

// WebhookSignature подпись webhook'а
type WebhookSignature struct {
	Signature string `json:"Signature"`
	Token     string `json:"Token"`
}

// Утилиты для работы с Tinkoff API

// MapTinkoffStatusToInternal маппинг статусов Tinkoff в внутренние
func MapTinkoffStatusToInternal(tinkoffStatus string) string {
	switch tinkoffStatus {
	case "CONFIRMED":
		return PaymentStatusCompleted
	case "AUTHORIZED":
		return PaymentStatusProcessing
	case "CANCELED", "REJECTED", "AUTH_FAIL":
		return PaymentStatusFailed
	case "REFUNDED", "PARTIAL_REFUNDED":
		return PaymentStatusRefunded
	case "NEW", "FORM_SHOWED", "PREAUTHORIZING", "AUTHORIZING", "CONFIRMING":
		return PaymentStatusPending
	default:
		return PaymentStatusPending
	}
}

// IsTinkoffStatusFinal проверяет, является ли статус Tinkoff финальным
func IsTinkoffStatusFinal(tinkoffStatus string) bool {
	finalStatuses := []string{"CONFIRMED", "CANCELED", "REJECTED", "AUTH_FAIL", "REFUNDED", "PARTIAL_REFUNDED"}
	for _, status := range finalStatuses {
		if tinkoffStatus == status {
			return true
		}
	}
	return false
}

// TinkoffError представляет ошибку Tinkoff API
type TinkoffError struct {
	ErrorCode string `json:"ErrorCode"`
	Message   string `json:"Message"`
	Details   string `json:"Details,omitempty"`
}

func (e *TinkoffError) Error() string {
	return e.Message
} 

func (r *InitRequest) ToMap() map[string]interface{} {
    params := map[string]interface{}{
        "TerminalKey": r.TerminalKey,
        "Amount":      r.Amount,
        "OrderId":     r.OrderId,
    }
    
    // Добавляем обязательные строковые поля
    if r.Description != "" {
        params["Description"] = r.Description
    }
    
    if r.SuccessURL != "" {
        params["SuccessURL"] = r.SuccessURL
    }
    
    if r.FailURL != "" {
        params["FailURL"] = r.FailURL
    }
    
    if r.NotificationURL != "" {
        params["NotificationURL"] = r.NotificationURL
    }

    return params
}


