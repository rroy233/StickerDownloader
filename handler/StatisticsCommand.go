package handler

import (
	"fmt"
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"github.com/rroy233/StickerDownloader/config"
	"github.com/rroy233/StickerDownloader/languages"
	"github.com/rroy233/StickerDownloader/statistics"
	"github.com/rroy233/StickerDownloader/utils"
	"time"
)

func StatisticsCommand(update tgbotapi.Update) {
	if update.Message.Chat.ID != config.Get().General.AdminUID {
		utils.SendPlainText(&update, languages.Get(&update).BotMsg.ErrNoPermission)
		return
	}

	data := statistics.Statistics.Printf()
	text := fmt.Sprintf("Statistics\nStart: %s\nEnd: %s\n%s",
		statistics.Statistics.StartTime.Format("2006-01-02 15:04:05"),
		time.Now().Format("2006-01-02 15:04:05"),
		data)

	utils.SendPlainText(&update, text, utils.EntityCode(text, data))
	return
}
