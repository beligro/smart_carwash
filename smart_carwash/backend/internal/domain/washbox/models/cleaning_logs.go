package models

import (
	"time"

	"github.com/google/uuid"
)

// Статусы логов уборки
const (
	CleaningLogStatusInProgress = "in_progress" // В процессе
	CleaningLogStatusCompleted  = "completed"   // Завершена
	CleaningLogStatusCancelled  = "cancelled"   // Отменена
)

// CleaningLog представляет лог уборки бокса
type CleaningLog struct {
	ID              uuid.UUID  `json:"id" gorm:"primaryKey;type:uuid;default:gen_random_uuid()"`
	CleanerID       uuid.UUID  `json:"cleaner_id" gorm:"type:uuid;not null;index"`
	WashBoxID       uuid.UUID  `json:"wash_box_id" gorm:"type:uuid;not null;index"`
	StartedAt       time.Time  `json:"started_at" gorm:"not null;index"`
	CompletedAt     *time.Time `json:"completed_at"`
	DurationMinutes *int       `json:"duration_minutes"`
	Status          string     `json:"status" gorm:"not null;default:in_progress;index"`
	CreatedAt       time.Time  `json:"created_at"`
	UpdatedAt       time.Time  `json:"updated_at"`
}

// AdminListCleaningLogsRequest запрос на получение списка логов уборки (админка)
type AdminListCleaningLogsRequest struct {
	Status   *string `json:"status" binding:"omitempty,oneof=in_progress completed cancelled"`
	DateFrom *string `json:"date_from" binding:"omitempty"`
	DateTo   *string `json:"date_to" binding:"omitempty"`
	Limit    *int    `json:"limit"`
	Offset   *int    `json:"offset"`
}

// AdminListCleaningLogsInternalRequest внутренний запрос для работы с репозиторием
type AdminListCleaningLogsInternalRequest struct {
	Status   *string    `json:"status"`
	DateFrom *time.Time `json:"date_from"`
	DateTo   *time.Time `json:"date_to"`
	Limit    *int       `json:"limit"`
	Offset   *int       `json:"offset"`
}

// AdminListCleaningLogsResponse ответ на получение списка логов уборки (админка)
type AdminListCleaningLogsResponse struct {
	CleaningLogs []CleaningLogWithDetails `json:"cleaning_logs"`
	Total        int                      `json:"total"`
	Limit        int                      `json:"limit"`
	Offset       int                      `json:"offset"`
}

// CleaningLogWithDetails представляет лог уборки с дополнительными деталями
type CleaningLogWithDetails struct {
	CleaningLog
	CleanerUsername string `json:"cleaner_username"`
	WashBoxNumber   int    `json:"wash_box_number"`
	WashBoxType     string `json:"wash_box_type"`
}
