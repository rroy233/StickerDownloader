package handler

import (
	tgbotapi "github.com/OvyFlash/telegram-bot-api"
	"github.com/rroy233/StickerDownloader/config"
	"github.com/rroy233/StickerDownloader/languages"
	"github.com/rroy233/StickerDownloader/utils"
)

func ReloadConfigCommand(update tgbotapi.Update) {
	if update.Message.Chat.ID != config.Get().General.AdminUID {
		utils.SendPlainText(&update, languages.Get(&update).BotMsg.ErrNoPermission)
		return
	}

	config.Init()
	languages.Init()
	utils.SendPlainText(&update, languages.Get(&update).BotMsg.ReloadConfigSuccess)
	return
}
