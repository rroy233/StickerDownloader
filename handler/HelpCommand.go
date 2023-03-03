package handler

import (
	"fmt"
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"github.com/rroy233/StickerDownloader/config"
	"github.com/rroy233/StickerDownloader/languages"
	"github.com/rroy233/StickerDownloader/utils"
)

func HelpCommand(update tgbotapi.Update) {
	utils.SendPlainText(&update, fmt.Sprintf(languages.Get(&update).BotMsg.HelpCommand, config.Get().General.UserDailyLimit))
	return
}
