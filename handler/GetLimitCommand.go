package handler

import (
	"fmt"
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"github.com/rroy233/tg-stickers-dl/db"
	"github.com/rroy233/tg-stickers-dl/utils"
)

func GetLimitCommand(update tgbotapi.Update) {
	utils.SendPlainText(&update, fmt.Sprintf("您当前可用次数为:%d次", db.GetLimit(update.Message.From.ID)))
	return
}
