package handler

import (
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"github.com/rroy233/tg-stickers-dl/languages"
	"github.com/rroy233/tg-stickers-dl/utils"
)

func StartCommand(update tgbotapi.Update) {
	utils.SendPlainText(&update, languages.Get().BotMsg.StartCommand)
	return
}
