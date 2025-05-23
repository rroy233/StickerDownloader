package handler

import (
	tgbotapi "github.com/OvyFlash/telegram-bot-api"
	"github.com/rroy233/StickerDownloader/config"
)

var bot *tgbotapi.BotAPI

func Init(b *tgbotapi.BotAPI) {

	bot = b

	if config.Get().General.ProcessTimeout == 0 {
		ProcessTimeout = 60
	} else {
		ProcessTimeout = config.Get().General.ProcessTimeout
	}
}
