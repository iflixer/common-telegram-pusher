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

func UpdateUsersPushID(dbService *Service, userIds []int64, pushID int) (err error) {
	if len(userIds) == 0 {
		return nil
	}
	return dbService.DB.Model(&TelegramUser{}).Where("tg_id IN (?)", userIds).Update("push_id", pushID).Error
}
