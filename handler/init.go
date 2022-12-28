package handler

import tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"

var bot *tgbotapi.BotAPI

func Init(b *tgbotapi.BotAPI) {
	bot = b
}
