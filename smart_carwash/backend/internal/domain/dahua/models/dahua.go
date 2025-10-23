package models

import (
	"time"
)

// DahuaWebhookRequest представляет входящий webhook от камеры Dahua
type DahuaWebhookRequest struct {
	LicensePlate string `json:"licensePlate" binding:"required"` // Номер автомобиля
	Confidence   int    `json:"confidence"`                     // Уровень уверенности распознавания
	Direction    string `json:"direction" binding:"required"`   // Направление движения (in/out)
	EventType    string `json:"eventType"`                      // Тип события (ANPR)
	CaptureTime  string `json:"captureTime"`                    // Время захвата
	ImagePath    string `json:"imagePath"`                      // Путь к изображению
}

// DahuaWebhookResponse представляет ответ на webhook от Dahua
type DahuaWebhookResponse struct {
	Success bool   `json:"success"`
	Message string `json:"message"`
}

// ProcessANPREventRequest представляет запрос на обработку ANPR события
type ProcessANPREventRequest struct {
	LicensePlate string `json:"license_plate" binding:"required"`
	Direction    string `json:"direction" binding:"required"`
	Confidence   int    `json:"confidence"`
	EventType    string `json:"event_type"`
	CaptureTime  string `json:"capture_time"`
	ImagePath    string `json:"image_path"`
}

// ProcessANPREventResponse представляет ответ на обработку ANPR события
type ProcessANPREventResponse struct {
	Success       bool   `json:"success"`
	Message       string `json:"message"`
	UserFound     bool   `json:"user_found"`
	SessionFound  bool   `json:"session_found"`
	SessionID     string `json:"session_id,omitempty"`
	SessionStatus string `json:"session_status,omitempty"`
}

// DahuaConfig представляет конфигурацию для Dahua интеграции
type DahuaConfig struct {
	WebhookUsername string   `json:"webhook_username"`
	WebhookPassword string   `json:"webhook_password"`
	AllowedIPs      []string `json:"allowed_ips"`
}

// ValidateDirection проверяет, что направление корректное
func (r *DahuaWebhookRequest) ValidateDirection() bool {
	return r.Direction == "in" || r.Direction == "out"
}

// IsExitEvent проверяет, является ли событие выездом
func (r *DahuaWebhookRequest) IsExitEvent() bool {
	return r.Direction == "out"
}

// ParseCaptureTime парсит время захвата из строки
func (r *DahuaWebhookRequest) ParseCaptureTime() (time.Time, error) {
	// Формат времени от Dahua: "2025-10-20T14:05:33"
	return time.Parse("2006-01-02T15:04:05", r.CaptureTime)
}
