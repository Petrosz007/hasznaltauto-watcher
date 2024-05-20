package main

import (
	"log"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

type TelegramNotifier struct {
	bot    *tgbotapi.BotAPI
	chatId int64
}

func NewTelegramNotifier(botToken string, chatId int64) (*TelegramNotifier, error) {
	bot, err := tgbotapi.NewBotAPI(botToken)
	if err != nil {
		return nil, err
	}

	bot.Debug = true

	log.Printf("Authorized on account %s", bot.Self.UserName)

	return &TelegramNotifier{
		bot:    bot,
		chatId: chatId,
	}, nil
}

func (t *TelegramNotifier) Notify(urls []string) {
	for _, url := range urls {
		msg := tgbotapi.NewMessage(t.chatId, url)
		t.bot.Send(msg)
	}
}
