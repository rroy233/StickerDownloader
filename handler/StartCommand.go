package handler

import (
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"github.com/rroy233/StickerDownloader/languages"
	"github.com/rroy233/StickerDownloader/utils"
)

func StartCommand(update tgbotapi.Update) {
	utils.SendPlainText(&update, languages.Get(&update).BotMsg.StartCommand)
	return
}
