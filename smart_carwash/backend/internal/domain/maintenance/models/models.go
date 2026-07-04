package models

import (
	"time"

	"github.com/google/uuid"
)

// Типы боксов для фильтрации справочников
const (
	BoxTypeWash   = "wash"
	BoxTypeVacuum = "vacuum"
	BoxTypeAir    = "air"
	BoxTypeAny    = "any"
)

// Статусы наряда
const (
	TicketStatusOpen   = "open"
	TicketStatusClosed = "closed"
)

// Действия по детали
const (
	ActionReplace = "replace"
	ActionRepair  = "repair"
)

// SymptomType — справочник симптомов кассира.
type SymptomType struct {
	ID        int    `json:"id" gorm:"primaryKey"`
	GroupName string `json:"group_name" gorm:"column:group_name"`
	Name      string `json:"name"`
	BoxType   string `json:"box_type"`
	SortOrder int    `json:"sort_order"`
	IsActive  bool   `json:"is_active"`
}

func (SymptomType) TableName() string { return "symptom_types" }

// ComponentGroup — группа работ мастера.
type ComponentGroup struct {
	ID        int    `json:"id" gorm:"primaryKey"`
	Name      string `json:"name"`
	Carrier   string `json:"carrier"`
	BoxType   string `json:"box_type"`
	SortOrder int    `json:"sort_order"`
}

func (ComponentGroup) TableName() string { return "component_groups" }

// ComponentType — деталь (нижний уровень справочника работ).
type ComponentType struct {
	ID        int    `json:"id" gorm:"primaryKey"`
	GroupID   int    `json:"group_id"`
	Name      string `json:"name"`
	SortOrder int    `json:"sort_order"`
}

func (ComponentType) TableName() string { return "component_types" }

// ServiceTicket — наряд на сервис (двухэтапный).
type ServiceTicket struct {
	ID                uuid.UUID  `json:"id" gorm:"primaryKey;type:uuid;default:uuid_generate_v4()"`
	BoxID             *uuid.UUID `json:"box_id" gorm:"type:uuid"`
	BoxNumber         int        `json:"box_number"`
	BoxType           string     `json:"box_type"`
	Status            string     `json:"status"`
	OpenedAt          time.Time  `json:"opened_at"`
	OpenedBy          string     `json:"opened_by"`
	OpenedByCashierID *uuid.UUID `json:"opened_by_cashier_id" gorm:"type:uuid"`
	SymptomID         *int       `json:"symptom_id"`
	CashierComment    string     `json:"cashier_comment"`
	ClosedAt          *time.Time `json:"closed_at"`
	ClosedBy          string     `json:"closed_by"`
	MasterComment     string     `json:"master_comment"`
	CreatedAt         time.Time  `json:"created_at"`
	UpdatedAt         time.Time  `json:"updated_at"`
}

func (ServiceTicket) TableName() string { return "service_tickets" }

// TicketWork — выполненная работа в рамках наряда.
type TicketWork struct {
	ID          uuid.UUID `json:"id" gorm:"primaryKey;type:uuid;default:uuid_generate_v4()"`
	TicketID    uuid.UUID `json:"ticket_id" gorm:"type:uuid"`
	ComponentID int       `json:"component_id"`
	Action      string    `json:"action"`
	Carrier     string    `json:"carrier"`
	MotorHours  *int      `json:"motor_hours"`
	CreatedAt   time.Time `json:"created_at"`
}

func (TicketWork) TableName() string { return "ticket_works" }

// NotificationRecipient — получатель push-уведомлений (управляется в админке).
type NotificationRecipient struct {
	ID        uuid.UUID `json:"id" gorm:"primaryKey;type:uuid;default:uuid_generate_v4()"`
	Name      string    `json:"name"`
	ChatID    int64     `json:"chat_id"`
	EventType string    `json:"event_type"`
	IsActive  bool      `json:"is_active"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

func (NotificationRecipient) TableName() string { return "notification_recipients" }

// ==================== DTO ====================

// OpenTicketRequest — постановка бокса в сервис кассиром.
type OpenTicketRequest struct {
	BoxID     uuid.UUID `json:"box_id" binding:"required"`
	SymptomID *int      `json:"symptom_id"`
	Comment   string    `json:"comment"`
}

// WorkInput — одна выполненная работа при закрытии наряда.
type WorkInput struct {
	ComponentID int    `json:"component_id" binding:"required"`
	Action      string `json:"action" binding:"required,oneof=replace repair"`
}

// CloseTicketRequest — закрытие наряда мастером.
type CloseTicketRequest struct {
	Works         []WorkInput `json:"works"`
	MasterComment string      `json:"master_comment"`
}

// CatalogGroup — группа работ с вложенными деталями (для формы закрытия).
type CatalogGroup struct {
	ID       int             `json:"id"`
	Name     string          `json:"name"`
	Carrier  string          `json:"carrier"`
	BoxType  string          `json:"box_type"`
	Children []ComponentType `json:"children"`
}

// TicketWorkView — работа наряда с именем детали (для журнала).
type TicketWorkView struct {
	ComponentID   int    `json:"component_id"`
	ComponentName string `json:"component_name"`
	GroupName     string `json:"group_name"`
	Action        string `json:"action"`
	Carrier       string `json:"carrier"`
	MotorHours    *int   `json:"motor_hours"`
}

// TicketView — наряд для журнала (с расшифровкой симптома и работами).
type TicketView struct {
	ServiceTicket
	SymptomName    string           `json:"symptom_name"`
	SymptomGroup   string           `json:"symptom_group"`
	DowntimeMinutes *int            `json:"downtime_minutes"`
	Works          []TicketWorkView `json:"works"`
}

// CreateRecipientRequest — добавление получателя уведомлений.
type CreateRecipientRequest struct {
	Name   string `json:"name" binding:"required"`
	ChatID int64  `json:"chat_id" binding:"required"`
}

// UpdateRecipientRequest — изменение активности получателя.
type UpdateRecipientRequest struct {
	IsActive *bool `json:"is_active"`
}
