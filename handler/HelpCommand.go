package handler

import (
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"github.com/rroy233/tg-stickers-dl/utils"
)

func HelpCommand(update tgbotapi.Update) {
	utils.SendPlainText(&update, "help")
	return
}
