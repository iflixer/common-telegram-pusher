package pusher

import (
	"errors"
	"fmt"
	"log"
	"telegram-pusher/database"
	"telegram-pusher/helper"
	"time"

	"gopkg.in/telebot.v4"
)

type Service struct {
	// mu        sync.RWMutex
	dbService *database.Service
}

func NewService(dbService *database.Service) (s *Service, err error) {
	s = &Service{
		dbService: dbService,
	}
	go s.workerSender()
	return
}

func (s *Service) workerSender() {
	for {
		//log.Println("push worker start")
		affected := 0
		var err error
		if affected, err = s.tryPush(); err != nil {
			log.Println(err)
		}
		if affected > 0 {
			time.Sleep(time.Second)
		} else {
			time.Sleep(time.Second * 10)
		}
	}
}

func (s *Service) tryPush() (affected int, err error) {
	push := &database.TelegramPush{}
	if err = push.SearchTasks(s.dbService); err != nil {
		return
	}

	if push.ID == 0 {
		return 0, errors.New("no ready push found")
	}

	bot := &database.TelegramBot{}
	if err = bot.Load(s.dbService, push.BotID); err != nil {
		log.Printf("bot id %d not found for push id %d, err: %s", push.BotID, push.ID, err)
		return
	}

	audience := &database.TelegramAudience{}
	if err = audience.Load(s.dbService, push.AudienceID); err != nil {
		log.Printf("audience id %d not found for push id %d, err: %s", push.AudienceID, push.ID, err)
		return
	}

	var targetTelegramIDs []int64

	switch audience.Type {
	case database.AudienceTypeUsers:
		var users []*database.TelegramUser
		if audience.Query != "" {
			if err = s.dbService.DB.Where("bot_id=? AND push_id != ? AND disabled=0", push.BotID, push.ID).Where(audience.Query).Limit(20).Find(&users).Error; err != nil {
				log.Println(err)
				return
			}
			if len(users) == 0 {
				log.Println("No users found for query:", audience.Query)
				push.SetStatusDone(s.dbService)
				return
			}
			affected = len(users)
			log.Printf("Found %d users for query: %s", len(users), audience.Query)
			for _, user := range users {
				targetTelegramIDs = append(targetTelegramIDs, user.TgID)
			}
			database.UpdateUsersPushID(s.dbService, targetTelegramIDs, push.ID)
		}
	case database.AudienceTypeChannel:
		if audience.ChannelID != 0 {
			channel := &database.TelegramChannel{}
			channel.Load(s.dbService, audience.ChannelID)
			targetTelegramIDs = append(targetTelegramIDs, channel.TgID)
			// set done status immediately if channel is used
			if err = push.SetStatusDone(s.dbService); err != nil {
				log.Printf("Error setting status to done for push id %d: %s", push.ID, err)
				return 0, err
			}
		}
	default:
		return 0, fmt.Errorf("audienceType not defined or not supported: %d", audience.Type)
	}

	pref := telebot.Settings{
		Token: bot.Token,
	}

	tbot, err := telebot.NewBot(pref)
	if err != nil {
		return
	}

	for _, tgID := range targetTelegramIDs {
		log.Println("Push to target:", tgID)

		switch push.Type {
		case "text":
			if push.Text == "" {
				log.Printf("Push text is empty for target %d, skipping", tgID)
				continue
			}
			if err = s.sendText(tbot, tgID, push.Text, push.InlineButtons); err != nil {
				log.Printf("Error sending text to %d: %s", tgID, err)
				//push.SetStatusError(s.dbService)
				continue
			}
		case "photo":
			if push.ImageURL == "" {
				log.Printf("Push image URL is empty for target %d, skipping", tgID)
				continue
			}
			if push.Text == "" {
				log.Printf("Push text is empty for target %d, using default caption", tgID)
				push.Text = "Image from push notification"
			}
			if err = s.sendPhoto(tbot, push.InlineButtons, tgID, push.Text, push.ImageURL); err != nil {
				log.Printf("Error sending photo to %d: %s", tgID, err)
				//push.SetStatusError(s.dbService)
				continue
			}
		case "video":
			if push.ImageURL == "" {
				log.Printf("Push image URL is empty for target %d, skipping", tgID)
				continue
			}
			if push.Text == "" {
				log.Printf("Push text is empty for target %d, using default caption", tgID)
				push.Text = "Image from push notification"
			}
			if err = s.sendVideo(tbot, push.InlineButtons, tgID, push.Text, push.ImageURL); err != nil {
				log.Printf("Error sending photo to %d: %s", tgID, err)
				//push.SetStatusError(s.dbService)
				continue
			}
		default:
			return 0, errors.New("push type not defined or not supported: " + push.Type)
		}
	}

	return

}

func (s *Service) sendText(tbot *telebot.Bot, chatID int64, msg string, inlineMenuJson string) (err error) {
	recipient := &telebot.User{ID: chatID}
	msg = helper.SanitizeTelegramHTML(msg)
	if inlineMenuJson != "" {
		inlineMenu, err := s.createInlineMenu(inlineMenuJson)
		if err == nil {
			_, err = tbot.Send(recipient, msg, inlineMenu, telebot.ModeHTML)
			return err
		}
	}

	_, err = tbot.Send(recipient, msg, telebot.ModeHTML)
	return
}

func (s *Service) sendPhoto(tbot *telebot.Bot, inlineMenuJson string, chatID int64, msg string, url string) (err error) {
	recipient := &telebot.User{ID: chatID}
	photo := &telebot.Photo{
		File:    telebot.FromURL(url),
		Caption: helper.SanitizeTelegramHTML(msg),
	}

	if inlineMenuJson != "" {
		inlineMenu, err := s.createInlineMenu(inlineMenuJson)
		if err == nil {
			_, err = tbot.Send(recipient, photo, inlineMenu, telebot.ModeHTML)
			return err
		}
	}
	_, err = tbot.Send(recipient, photo, telebot.ModeHTML)
	return
}

func (s *Service) sendVideo(tbot *telebot.Bot, inlineMenuJson string, chatID int64, msg string, url string) (err error) {
	recipient := &telebot.User{ID: chatID}
	video := &telebot.Video{
		File:    telebot.FromURL(url),
		Caption: helper.SanitizeTelegramHTML(msg),
	}

	if inlineMenuJson != "" {
		inlineMenu, err := s.createInlineMenu(inlineMenuJson)
		if err == nil {
			_, err = tbot.Send(recipient, video, inlineMenu, telebot.ModeHTML)
			return err
		}
	}
	_, err = tbot.Send(recipient, video, telebot.ModeHTML)
	return
}
