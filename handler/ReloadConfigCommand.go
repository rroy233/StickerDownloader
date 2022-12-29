package handler

import (
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"github.com/rroy233/tg-stickers-dl/config"
	"github.com/rroy233/tg-stickers-dl/languages"
	"github.com/rroy233/tg-stickers-dl/utils"
)

func ReloadConfigCommand(update tgbotapi.Update) {
	if update.Message.Chat.ID != config.Get().General.AdminUID {
		utils.SendPlainText(&update, languages.Get().BotMsg.ErrNoPermission)
		return
	}

	config.Init()
	languages.Init(config.Get().General.Language)
	utils.SendPlainText(&update, languages.Get().BotMsg.ReloadConfigSuccess)
	return
}
