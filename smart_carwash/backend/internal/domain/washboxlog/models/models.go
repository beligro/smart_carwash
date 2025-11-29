package models

import (
	"time"

	"github.com/google/uuid"
)

// ActorType фиксирует источник действия
// Возможные значения: super_admin, limited_admin, cashier, cleaner, user, anpr, system
type ActorType string

const (
	ActorSuperAdmin   ActorType = "super_admin"
	ActorLimitedAdmin ActorType = "limited_admin"
	ActorCashier      ActorType = "cashier"
	ActorCleaner      ActorType = "cleaner"
	ActorUser         ActorType = "user"
	ActorANPR         ActorType = "anpr"
	ActorSystem       ActorType = "system"
)

// ActionType тип изменения
type ActionType string

const (
	ActionStatusChange    ActionType = "status_change"
	ActionLightOn         ActionType = "light_on"
	ActionLightOff        ActionType = "light_off"
	ActionChemistryOn     ActionType = "chemistry_on"
	ActionChemistryOff    ActionType = "chemistry_off"
)

// WashBoxChangeLog запись об изменении состояния бокса
type WashBoxChangeLog struct {
	ID         uuid.UUID `gorm:"type:uuid;primaryKey"`
	BoxID      uuid.UUID `gorm:"type:uuid;index"`
	BoxNumber  int       `gorm:"index"`
	Action     string    `gorm:"type:varchar(64);index"`
	PrevStatus *string   `gorm:"type:varchar(32)"`
	NewStatus  *string   `gorm:"type:varchar(32)"`
	PrevValue  *bool
	NewValue   *bool
	ActorType  string    `gorm:"type:varchar(32);index"`
	Source     *string   `gorm:"type:varchar(64)"` // произвольная пометка источника
	CreatedAt  time.Time `gorm:"index"`
}

// TableName переопределяет имя таблицы для соответствия миграции
func (WashBoxChangeLog) TableName() string {
	return "washbox_change_logs"
}


