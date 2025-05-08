package handler

import (
	"fmt"
	tgbotapi "github.com/OvyFlash/telegram-bot-api"
	"github.com/rroy233/StickerDownloader/config"
	"github.com/rroy233/StickerDownloader/languages"
	"github.com/rroy233/StickerDownloader/utils"
)

func AdminCommand(update tgbotapi.Update) {
	if update.Message.Chat.ID != config.Get().General.AdminUID {
		utils.SendPlainText(&update, languages.Get(&update).BotMsg.ErrNoPermission)
		return
	}

	utils.SendPlainText(&update, fmt.Sprintf("Admin Command\n\nReload Config /reload\nClear Cache /clearcache\nWeek Statistics /statistics"))
	return
}
