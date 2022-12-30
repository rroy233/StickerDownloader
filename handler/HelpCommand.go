package handler

import (
	"fmt"
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"github.com/rroy233/tg-stickers-dl/config"
	"github.com/rroy233/tg-stickers-dl/languages"
	"github.com/rroy233/tg-stickers-dl/utils"
)

func HelpCommand(update tgbotapi.Update) {
	utils.SendPlainText(&update, fmt.Sprintf(languages.Get().BotMsg.HelpCommand, config.Get().General.UserDailyLimit))
	return
}
