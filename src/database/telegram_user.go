package database

import (
	"time"
)

type TelegramUser struct {
	ID             int
	BotID          int
	TgID           int64
	LastCommand    string
	Counter        int
	UpdatedAt      *time.Time
	LastActivityAt *time.Time
	PushID         int
	Disabled       bool
	PushTime       *time.Time
}

func (c *TelegramUser) TableName() string {
	return "telegram_user"
}
