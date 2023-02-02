package handler

import (
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"github.com/rroy233/tg-stickers-dl/config"
	"github.com/rroy233/tg-stickers-dl/db"
	"github.com/rroy233/tg-stickers-dl/languages"
	"github.com/rroy233/tg-stickers-dl/utils"
	"time"
)

func ClearCacheCommand(update tgbotapi.Update) {
	if update.Message.Chat.ID != config.Get().General.AdminUID {
		utils.SendPlainText(&update, languages.Get(&update).BotMsg.ErrNoPermission)
		return
	}

	out, err := db.ClearCache()
	if err != nil {
		utils.SendPlainText(&update, languages.Get(&update).BotMsg.ErrSysFailureOccurred)
		time.Sleep(100 * time.Millisecond)
		utils.SendPlainText(&update, err.Error())
		return
	}
	utils.SendPlainText(&update, out)
	return
}
