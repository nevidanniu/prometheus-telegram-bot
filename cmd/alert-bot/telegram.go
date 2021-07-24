package main

import (
	"fmt"
	log "github.com/sirupsen/logrus"
	tgbotapi "gopkg.in/telegram-bot-api.v4"
)

var bot *tgbotapi.BotAPI

func telegramBot(bot *tgbotapi.BotAPI) {
	u := tgbotapi.NewUpdate(0)
	u.Timeout = 60

	updates, err := bot.GetUpdatesChan(u)
	if err != nil {
		log.Fatal(err)
	}

	introduce := func(update tgbotapi.Update) {
		msg := tgbotapi.NewMessage(update.Message.Chat.ID, fmt.Sprintf("Chat id is '%d'", update.Message.Chat.ID))
		bot.Send(msg)
	}

	for update := range updates {
		if update.Message == nil {
			log.Debugf("[UNKNOWN_MESSAGE] [%v]", update)
			continue
		}

		if update.Message.NewChatMembers != nil && len(*update.Message.NewChatMembers) > 0 {
			for _, member := range *update.Message.NewChatMembers {
				if member.UserName == bot.Self.UserName && update.Message.Chat.Type == "group" {
					introduce(update)
				}
			}
		} else if update.Message != nil && update.Message.Text != "" {
			introduce(update)
		}
	}
}
