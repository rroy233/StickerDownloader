package handler

import (
	"fmt"
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"github.com/rroy233/StickerDownloader/db"
	"github.com/rroy233/StickerDownloader/languages"
	"github.com/rroy233/StickerDownloader/utils"
)

func GetLimitCommand(update tgbotapi.Update) {
	num := db.GetLimit(update.Message.From.ID)
	text := fmt.Sprintf(languages.Get(&update).BotMsg.GetLimitCommand, num)
	utils.SendPlainText(&update,
		text,
		utils.EntityBold(fmt.Sprintf(languages.Get(&update).BotMsg.GetLimitCommand, num), fmt.Sprintf("%d", num)),
	)
	return
}
