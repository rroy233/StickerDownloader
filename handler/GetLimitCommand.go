package handler

import (
	"fmt"
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"github.com/rroy233/tg-stickers-dl/db"
	"github.com/rroy233/tg-stickers-dl/languages"
	"github.com/rroy233/tg-stickers-dl/utils"
)

func GetLimitCommand(update tgbotapi.Update) {
	num := db.GetLimit(update.Message.From.ID)
	text := fmt.Sprintf(languages.Get().BotMsg.GetLimitCommand, num)
	utils.SendPlainText(&update,
		text,
		utils.EntityBold(fmt.Sprintf(languages.Get().BotMsg.GetLimitCommand, num), fmt.Sprintf("%d", num)),
	)
	return
}
