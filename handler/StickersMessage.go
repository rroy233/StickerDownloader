package handler

import (
	"context"
	"fmt"
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"github.com/rroy233/StickerDownloader/config"
	"github.com/rroy233/StickerDownloader/db"
	"github.com/rroy233/StickerDownloader/languages"
	"github.com/rroy233/StickerDownloader/statistics"
	"github.com/rroy233/StickerDownloader/utils"
	"gopkg.in/rroy233/logger.v2"
	"time"
)

func StickerMessage(update tgbotapi.Update) {
	userInfo := utils.GetLogPrefixMessage(&update)

	oMsg := tgbotapi.NewMessage(update.Message.Chat.ID, languages.Get(&update).BotMsg.Processing)
	oMsg.ReplyToMessageID = update.Message.MessageID
	msg, err := utils.BotSend(oMsg)
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
	//Dequeue
	defer dequeue(qItem)
	//Dequeue

	remoteFile, err := utils.BotGetFile(tgbotapi.FileConfig{
		FileID: update.Message.Sticker.FileID,
	})
	if err != nil {
		logger.Error.Println(userInfo+"failed to get file:", err)
	}

	cacheFile, err := db.FindStickerCache(update.Message.Sticker.FileUniqueID)
	outPath := ""
	if err == nil {
		//缓存存在
		statistics.Statistics.Record("CacheHit", 1)
		outPath = cacheFile
	} else {
		//缓存不存在
		statistics.Statistics.Record("CacheMiss", 1)
		tempFilePath, err := utils.DownloadFile(remoteFile.Link(config.Get().General.BotToken))
		if err != nil {
			logger.Error.Println(userInfo+"failed to download file:", err)
		}

		logger.Info.Printf("%sGet sticker %s.%s", userInfo, update.Message.Sticker.SetName, update.Message.Sticker.Emoji)

		defer utils.RemoveFile(tempFilePath) //delete temp file

		//check file type
		if utils.GetFileExtName(tempFilePath) != "webp" && utils.GetFileExtName(tempFilePath) != "webm" {
			utils.EditMsgText(update.Message.Chat.ID, msg.MessageID, languages.Get(&update).BotMsg.ErrStickerNotSupport)
			return
		}

		fileExt := "gif"
		if utils.GetFileExtName(tempFilePath) == "webp" {
			fileExt = "png"
		}
		outPath = fmt.Sprintf("./storage/tmp/convert_%s.%s", utils.RandString(), fileExt)
		ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
		err = utils.ConvertToGif(ctx, tempFilePath, outPath)
		cancel()
		if err != nil {
			logger.Error.Println(userInfo+"failed to convert:", err)
			utils.EditMsgText(update.Message.Chat.ID, msg.MessageID, languages.Get(&update).BotMsg.ErrConvertFailed)
			return
		}

		db.CacheSticker(*update.Message.Sticker, outPath)
	}

	utils.SendAction(update.Message.Chat.ID, utils.ChatActionSendDocument)
	err = utils.SendFile(&update, outPath)
	if err != nil {
		logger.Error.Println(userInfo+"failed to SendFile:", err)
		utils.EditMsgText(update.Message.Chat.ID,
			msg.MessageID,
			fmt.Sprintf("%s(TelegramAPI:%s)", languages.Get(&update).BotMsg.ErrSendFileFailed, err.Error()),
		)
		return
	}

	err = utils.BotRequest(tgbotapi.NewEditMessageTextAndMarkup(update.Message.Chat.ID, msg.MessageID, languages.Get(&update).BotMsg.ConvertCompleted, tgbotapi.NewInlineKeyboardMarkup(
		tgbotapi.NewInlineKeyboardRow(tgbotapi.NewInlineKeyboardButtonData(languages.Get(&update).BotMsg.DownloadStickerSet, DownloadStickerSetCallbackQuery)),
	)))
	if err != nil {
		logger.Error.Println(userInfo+"failed to delete msg:", err)
	}

	utils.RemoveFile(outPath)
	return
}
