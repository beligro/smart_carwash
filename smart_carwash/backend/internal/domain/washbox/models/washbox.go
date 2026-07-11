package models

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// Статусы бокса мойки
const (
	StatusFree        = "free"        // Свободен
	StatusReserved    = "reserved"    // Зарезервирован (назначен на пользователя, но сессия не запущена)
	StatusBusy        = "busy"        // Занят (сессия активна)
	StatusMaintenance = "maintenance" // На обслуживании
	StatusCleaning    = "cleaning"    // На уборке
)

// Типы услуг
const (
	ServiceTypeWash   = "wash"    // Мойка
	ServiceTypeAirDry = "air_dry" // Обдув воздухом
	ServiceTypeVacuum = "vacuum"  // Пылесос
)

// WashBox представляет бокс автомойки
type WashBox struct {
	ID                            uuid.UUID      `json:"id" gorm:"primaryKey;type:uuid;default:gen_random_uuid()"`
	Number                        int            `json:"number" gorm:"uniqueIndex"`
	Status                        string         `json:"status" gorm:"default:free"`
	ServiceType                   string         `json:"service_type" gorm:"default:wash"`
	ChemistryEnabled              bool           `json:"chemistry_enabled" gorm:"default:true"`
	Priority                      string         `json:"priority" gorm:"type:varchar(1);default:'A';check:priority ~ '^[A-Z]$'"`
	LightCoilRegister             *string        `json:"light_coil_register"`
	ChemistryCoilRegister         *string        `json:"chemistry_coil_register"`
	Comment                       *string        `json:"comment" gorm:"type:varchar(1000)"`
	CleaningReservedBy            *uuid.UUID     `json:"cleaning_reserved_by" gorm:"type:uuid;index"`
	CleaningStartedAt             *time.Time     `json:"cleaning_started_at"`
	LastCompletedSessionUserID    *uuid.UUID     `json:"last_completed_session_user_id" gorm:"type:uuid;index"`
	LastCompletedSessionCarNumber *string        `json:"last_completed_session_car_number"`
	LastCompletedAt               *time.Time     `json:"last_completed_at"`
	CooldownUntil                 *time.Time     `json:"cooldown_until"`
	ServiceUntil                  *time.Time     `json:"service_until"` // Таймерный сервис (чистка пылесоса): авто-возврат в работу после этого момента
	SecondsLeft                   *int           `json:"seconds_left,omitempty" gorm:"-"`     // Остаток времени активной сессии в секундах (вирт., для табло)
	ReservedSecondsLeft           *int           `json:"reserved_seconds_left,omitempty" gorm:"-"` // Остаток до авто-старта назначенного бокса (вирт., для табло)
	LightStatus                   *bool          `json:"light_status,omitempty" gorm:"-"`     // Статус света (не хранится в БД, заполняется из modbus_connection_statuses)
	ChemistryStatus               *bool          `json:"chemistry_status,omitempty" gorm:"-"` // Статус химии (не хранится в БД, заполняется из modbus_connection_statuses)
	CanBeCleaned                  *bool          `json:"can_be_cleaned,omitempty" gorm:"-"`   // Можно ли убирать бокс (не хранится в БД, вычисляется динамически)
	CreatedAt                     time.Time      `json:"created_at"`
	UpdatedAt                     time.Time      `json:"updated_at"`
	DeletedAt                     gorm.DeletedAt `json:"-" gorm:"index"`
}

// Причины админского включения коилов
const (
	AdminCoilReasonPersonalWash = "personal_wash"
	AdminCoilReasonTest         = "test"
)

// Причины завершения админского включения коилов
const (
	AdminCoilEndedAuto   = "auto"
	AdminCoilEndedManual = "manual"
)

// Тип бокса для лимитов личного включения
const (
	BoxTypeWash   = "wash"
	BoxTypeVacuum = "vacuum"
	BoxTypeAir    = "air"
)

// AdminCoilAction — запись о ручном включении коилов бокса администратором
// (личная мойка/тест). Хранит окно действия и причину завершения.
type AdminCoilAction struct {
	ID            uuid.UUID  `json:"id" gorm:"primaryKey;type:uuid;default:gen_random_uuid()"`
	AdminUsername string     `json:"admin_username" gorm:"type:varchar(255);not null;default:''"`
	BoxID         *uuid.UUID `json:"box_id" gorm:"type:uuid"`
	BoxNumber     int        `json:"box_number" gorm:"not null"`
	BoxType       string     `json:"box_type" gorm:"type:varchar(20);not null"`
	Reason        string     `json:"reason" gorm:"type:varchar(20);not null"`
	Coil          *string    `json:"coil" gorm:"type:varchar(20)"`
	TicketID      *uuid.UUID `json:"ticket_id" gorm:"type:uuid"`
	StartedAt     time.Time  `json:"started_at" gorm:"not null;default:now()"`
	ExpiresAt     *time.Time `json:"expires_at"`
	EndedAt       *time.Time `json:"ended_at"`
	EndedReason   *string    `json:"ended_reason" gorm:"type:varchar(20)"`
	CreatedAt     time.Time  `json:"created_at" gorm:"not null;default:now()"`
}

// TableName задаёт имя таблицы для AdminCoilAction.
func (AdminCoilAction) TableName() string {
	return "admin_coil_actions"
}

// PersonalUseBox — информация о боксе для раздела «Моя мойка».
type PersonalUseBox struct {
	ID          uuid.UUID  `json:"id"`
	Number      int        `json:"number"`
	ServiceType string     `json:"service_type"`
	BoxType     string     `json:"box_type"`
	Status      string     `json:"status"`
	Remaining   *int       `json:"remaining,omitempty"`    // остаток лимита по типу (для свободных)
	ExpiresAt   *time.Time `json:"expires_at,omitempty"`   // если сейчас активно личное включение
	ActionID    *uuid.UUID `json:"action_id,omitempty"`    // id открытого AdminCoilAction
	AdminUsername *string  `json:"admin_username,omitempty"` // кто включил (для активных)
}

// PersonalUseAvailabilityResponse — ответ по доступности личного включения.
type PersonalUseAvailabilityResponse struct {
	Boxes  []PersonalUseBox `json:"boxes"`
	Active []PersonalUseBox `json:"active"`
}

// AdminPersonalUseResponse — ответ на личное включение бокса.
type AdminPersonalUseResponse struct {
	BoxNumber int        `json:"box_number"`
	ExpiresAt *time.Time `json:"expires_at"`
	Message   string     `json:"message"`
}

// AdminPersonalUseRequest — запрос на личное включение/возврат бокса.
type AdminPersonalUseRequest struct {
	BoxID uuid.UUID `json:"box_id" binding:"required"`
}

// AdminTestCoilRequest — запрос на тестовое включение/выключение коила из наряда.
type AdminTestCoilRequest struct {
	BoxID    uuid.UUID  `json:"box_id" binding:"required"`
	Coil     string     `json:"coil" binding:"required"`
	Value    bool       `json:"value"`
	TicketID *uuid.UUID `json:"ticket_id"`
}

// GetQueueStatusResponse представляет ответ на получение статуса очереди и боксов
type GetQueueStatusResponse struct {
	Boxes     []WashBox `json:"boxes"`
	QueueSize int       `json:"queue_size"`
	HasQueue  bool      `json:"has_queue"`
}

// AdminCreateWashBoxRequest запрос на создание бокса мойки
type AdminCreateWashBoxRequest struct {
	Number                int     `json:"number" binding:"required"`
	Status                string  `json:"status" binding:"required,oneof=free reserved busy maintenance cleaning"`
	ServiceType           string  `json:"service_type" binding:"required,oneof=wash air_dry vacuum"`
	ChemistryEnabled      *bool   `json:"chemistry_enabled"`
	Priority              string  `json:"priority" binding:"required,len=1"`
	LightCoilRegister     *string `json:"light_coil_register"`
	ChemistryCoilRegister *string `json:"chemistry_coil_register"`
	Comment               *string `json:"comment" binding:"omitempty,max=1000"`
}

// AdminUpdateWashBoxRequest запрос на обновление бокса мойки
type AdminUpdateWashBoxRequest struct {
	ID                    uuid.UUID `json:"id" binding:"required"`
	Number                *int      `json:"number"`
	Status                *string   `json:"status" binding:"omitempty,oneof=free reserved busy maintenance cleaning"`
	ServiceType           *string   `json:"service_type" binding:"omitempty,oneof=wash air_dry vacuum"`
	ChemistryEnabled      *bool     `json:"chemistry_enabled"`
	Priority              *string   `json:"priority" binding:"omitempty,len=1"`
	LightCoilRegister     *string   `json:"light_coil_register"`
	ChemistryCoilRegister *string   `json:"chemistry_coil_register"`
	Comment               *string   `json:"comment" binding:"omitempty,max=1000"`
}

// AdminDeleteWashBoxRequest запрос на удаление бокса мойки
type AdminDeleteWashBoxRequest struct {
	ID uuid.UUID `json:"id" binding:"required"`
}

// AdminGetWashBoxRequest запрос на получение бокса мойки
type AdminGetWashBoxRequest struct {
	ID uuid.UUID `json:"id" binding:"required"`
}

// AdminListWashBoxesRequest запрос на получение списка боксов мойки
type AdminListWashBoxesRequest struct {
	Status      *string `json:"status" binding:"omitempty,oneof=free reserved busy maintenance cleaning"`
	ServiceType *string `json:"service_type" binding:"omitempty,oneof=wash air_dry vacuum"`
	Limit       *int    `json:"limit"`
	Offset      *int    `json:"offset"`
}

// AdminCreateWashBoxResponse ответ на создание бокса мойки
type AdminCreateWashBoxResponse struct {
	WashBox WashBox `json:"wash_box"`
}

// CleanerListWashBoxesRequest запрос на получение списка боксов для уборщика
type CleanerListWashBoxesRequest struct {
	Limit  *int `json:"limit"`
	Offset *int `json:"offset"`
}

// CleanerListWashBoxesResponse ответ на получение списка боксов для уборщика
type CleanerListWashBoxesResponse struct {
	WashBoxes []WashBox `json:"wash_boxes"`
	Total     int       `json:"total"`
}

// CleanerStartCleaningRequest запрос на начало уборки
type CleanerStartCleaningRequest struct {
	WashBoxID uuid.UUID `json:"wash_box_id" binding:"required"`
}

// CleanerStartCleaningResponse ответ на начало уборки
type CleanerStartCleaningResponse struct {
	Success bool `json:"success"`
}

// CleanerCompleteCleaningRequest запрос на завершение уборки
type CleanerCompleteCleaningRequest struct {
	WashBoxID uuid.UUID `json:"wash_box_id" binding:"required"`
}

// CleanerCompleteCleaningResponse ответ на завершение уборки
type CleanerCompleteCleaningResponse struct {
	Success bool `json:"success"`
}

// CleanerBoxStateResponse ответ с текущим состоянием спецбокса уборщика
type CleanerBoxStateResponse struct {
	WashBox WashBox `json:"wash_box"`
}

// CleanerCleaningHistoryRequest запрос истории уборок уборщика
type CleanerCleaningHistoryRequest struct {
	Limit  *int `json:"limit"`
	Offset *int `json:"offset"`
}

// CleanerCleaningHistoryItem элемент истории уборщика
type CleanerCleaningHistoryItem struct {
	ID          uuid.UUID  `json:"id"`
	WashBoxID   uuid.UUID  `json:"wash_box_id"`
	StartedAt   time.Time  `json:"started_at"`
	CompletedAt *time.Time `json:"completed_at"`
}

// CleanerCleaningHistoryResponse ответ на запрос истории уборщика
type CleanerCleaningHistoryResponse struct {
	Logs   []CleanerCleaningHistoryItem `json:"logs"`
	Total  int                          `json:"total"`
	Limit  int                          `json:"limit"`
	Offset int                          `json:"offset"`
}

// AdminUpdateWashBoxResponse ответ на обновление бокса мойки
type AdminUpdateWashBoxResponse struct {
	WashBox WashBox `json:"wash_box"`
}

// AdminGetWashBoxResponse ответ на получение бокса мойки
type AdminGetWashBoxResponse struct {
	WashBox WashBox `json:"wash_box"`
}

// AdminListWashBoxesResponse ответ на получение списка боксов мойки
type AdminListWashBoxesResponse struct {
	WashBoxes []WashBox `json:"wash_boxes"`
	Total     int       `json:"total"`
	Limit     int       `json:"limit"`
	Offset    int       `json:"offset"`
}

// AdminDeleteWashBoxResponse ответ на удаление бокса мойки
type AdminDeleteWashBoxResponse struct {
	Message string `json:"message"`
}

// CashierListWashBoxesRequest запрос на получение списка боксов мойки для кассира
type CashierListWashBoxesRequest struct {
	Status      *string `json:"status" binding:"omitempty,oneof=free reserved busy maintenance"`
	ServiceType *string `json:"service_type" binding:"omitempty,oneof=wash air_dry vacuum"`
	Limit       *int    `json:"limit"`
	Offset      *int    `json:"offset"`
}

// CashierListWashBoxesResponse ответ на получение списка боксов мойки для кассира
type CashierListWashBoxesResponse struct {
	WashBoxes []WashBox `json:"wash_boxes"`
	Total     int       `json:"total"`
	Limit     int       `json:"limit"`
	Offset    int       `json:"offset"`
}

// CashierSetMaintenanceRequest запрос на перевод бокса в режим обслуживания
type CashierSetMaintenanceRequest struct {
	ID uuid.UUID `json:"id" binding:"required"`
}

// CashierStartTimedServiceRequest запрос на таймерный сервис бокса (чистка пылесоса)
type CashierStartTimedServiceRequest struct {
	ID      uuid.UUID `json:"id" binding:"required"`
	Minutes int       `json:"minutes"` // 0 => значение по умолчанию (12 мин)
}

// CashierStartTimedServiceResponse ответ на таймерный сервис
type CashierStartTimedServiceResponse struct {
	WashBox WashBox `json:"wash_box"`
	Message string  `json:"message"`
}

// CashierSetMaintenanceResponse ответ на перевод бокса в режим обслуживания
type CashierSetMaintenanceResponse struct {
	WashBox WashBox `json:"wash_box"`
	Message string  `json:"message"`
}
