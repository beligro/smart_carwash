package models

import (
	"time"

	"github.com/google/uuid"
)

// LinkToken запись одноразового токена привязки Telegram к веб-пользователю
type LinkToken struct {
	Token     string     `json:"token" gorm:"primaryKey"`
	UserID    uuid.UUID  `json:"user_id" gorm:"type:uuid;not null;index"`
	ExpiresAt time.Time  `json:"expires_at" gorm:"not null"`
	UsedAt    *time.Time `json:"used_at"`
	CreatedAt time.Time  `json:"created_at"`
}

// CreateLinkTokenRequest запрос на создание токена привязки
type CreateLinkTokenRequest struct {
	UserID uuid.UUID `json:"user_id" binding:"required"`
}

// CreateLinkTokenResponse ответ с ссылкой для привязки
type CreateLinkTokenResponse struct {
	Link string `json:"link"` // https://t.me/BotName?start=link_XXXX
}
