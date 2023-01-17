package handler

import (
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"github.com/rroy233/logger"
	"github.com/rroy233/tg-stickers-dl/db"
	"github.com/rroy233/tg-stickers-dl/languages"
	"github.com/rroy233/tg-stickers-dl/utils"
)

func QuitQueueQuery(update tgbotapi.Update) {
	if len(update.CallbackQuery.Data) < len(QuitQueueCallbackQueryPrefix) {
		logger.FATAL.Println("[QuitQueueQuery]  len(update.CallbackQuery.Data) < len(QuitQueueCallbackQueryPrefix)")
		utils.CallBackWithAlert(update.CallbackQuery.ID, languages.Get(&update).BotMsg.ErrSysFailureOccurred)
		return
	}
	UUID := update.CallbackQuery.Data[len(QuitQueueCallbackQueryPrefix):]
	item, err := db.FindQueueItemByUUID(UUID)
	if err != nil {
		utils.CallBack(update.CallbackQuery.ID, err.Error())
		utils.EditMsgText(update.CallbackQuery.Message.Chat.ID, update.CallbackQuery.Message.MessageID, languages.Get(&update).BotMsg.QueueAborted)
	} else {
		utils.CallBack(update.CallbackQuery.ID, "ok")
		dequeue(item)
	}
	return
}
