package handler

import (
	"fmt"
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"github.com/rroy233/logger"
	"github.com/rroy233/tg-stickers-dl/config"
	"github.com/rroy233/tg-stickers-dl/languages"
	"github.com/rroy233/tg-stickers-dl/utils"
	"time"
)

func StickerMessage(update tgbotapi.Update) {
	userInfo := utils.GetLogPrefixMessage(&update)

	oMsg := tgbotapi.NewMessage(update.Message.Chat.ID, languages.Get(&update).BotMsg.Processing)
	oMsg.ReplyToMessageID = update.Message.MessageID
	msg, err := bot.Send(oMsg)
	if err != nil {
		logger.Error.Println(userInfo+"failed to send msg:", err)
		return
	}

	//Enqueue
	qItem, quit := enqueue(&update, &msg)
	if quit == true {
		return
	}
	//Enqueue

	remoteFile, err := bot.GetFile(tgbotapi.FileConfig{
		FileID: update.Message.Sticker.FileID,
	})
	if err != nil {
		logger.Error.Println(userInfo+"failed to get file:", err)
	}

	tempFilePath, err := utils.DownloadFile(remoteFile.Link(config.Get().General.BotToken))
	if err != nil {
		logger.Error.Println(userInfo+"failed to download file:", err)
	}

	logger.Info.Printf("%sGet sticker %s.%s", userInfo, update.Message.Sticker.SetName, update.Message.Sticker.Emoji)

	//delete temp file
	defer utils.RemoveFile(tempFilePath)

	//check file type
	if utils.GetFileExtName(tempFilePath) != "webp" && utils.GetFileExtName(tempFilePath) != "webm" {
		utils.EditMsgText(update.Message.Chat.ID, msg.MessageID, languages.Get(&update).BotMsg.ErrStickerNotSupport)
		return
	}

	outPath := fmt.Sprintf("./storage/tmp/convert_%d.gif", time.Now().UnixMicro())
	err = utils.ConvertToGif(tempFilePath, outPath)
	if err != nil {
		logger.Error.Println(userInfo+"failed to convert:", err)
		utils.EditMsgText(update.Message.Chat.ID, msg.MessageID, languages.Get(&update).BotMsg.ErrConvertFailed)
		return
	}

	//Dequeue
	dequeue(qItem)
	//Dequeue

	err = utils.SendFile(&update, outPath)
	if err != nil {
		logger.Error.Println(userInfo+"failed to SendFile:", err)
		utils.EditMsgText(update.Message.Chat.ID, msg.MessageID, languages.Get(&update).BotMsg.ErrSendFileFailed)
		return
	}

	_, err = bot.Request(tgbotapi.NewEditMessageTextAndMarkup(update.Message.Chat.ID, msg.MessageID, languages.Get(&update).BotMsg.ConvertCompleted, tgbotapi.NewInlineKeyboardMarkup(
		tgbotapi.NewInlineKeyboardRow(tgbotapi.NewInlineKeyboardButtonData(languages.Get(&update).BotMsg.DownloadStickerSet, DownloadStickerSetCallbackQuery)),
	)))
	if err != nil {
		logger.Error.Println(userInfo+"failed to delete msg:", err)
	}

	utils.RemoveFile(outPath)
	return
}
