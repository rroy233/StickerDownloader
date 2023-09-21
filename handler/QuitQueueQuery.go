package handler

import (
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"github.com/rroy233/StickerDownloader/db"
	"github.com/rroy233/StickerDownloader/languages"
	"github.com/rroy233/StickerDownloader/utils"
	"gopkg.in/rroy233/logger.v2"
)

func QuitQueueQuery(update tgbotapi.Update) {
	if len(update.CallbackQuery.Data) < len(QuitQueueCallbackQueryPrefix) {
		logger.FATAL.Println("[QuitQueueQuery]  len(update.CallbackQuery.Data) < len(QuitQueueCallbackQueryPrefix)")
		utils.CallBackWithAlert(update.CallbackQuery.ID, languages.Get(&update).BotMsg.ErrSysFailureOccurred)
		return
	}
	utils.CallBack(update.CallbackQuery.ID, "ok")
	UUID := update.CallbackQuery.Data[len(QuitQueueCallbackQueryPrefix):]
	item, err := db.FindQueueItemByUUID(UUID)
	if err != nil {
		utils.EditMsgText(update.CallbackQuery.Message.Chat.ID, update.CallbackQuery.Message.MessageID, languages.Get(&update).BotMsg.QueueAborted)
	} else {
		dequeue(item)
	}
	return
}
