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

	cacheItem, err := db.FindStickerCacheItem(update.Message.Sticker.FileUniqueID)
	if err == nil && cacheItem.ConvertedFileID != "" {
		//缓存存在
		statistics.Statistics.Record("CacheHit", 1)

		//通过file_id直接发送文件
		if err := utils.SendFileByFileID(&update, cacheItem.ConvertedFileID); err != nil {
			logger.Error.Println(userInfo+"failed to send file via FILE_ID:", err)
			utils.EditMsgText(update.Message.Chat.ID,
				msg.MessageID,
				fmt.Sprintf("%s(TelegramAPI:%s)", languages.Get(&update).BotMsg.ErrSendFileFailed, err.Error()),
			)
			return
		}
	} else {
		//缓存不存在
		statistics.Statistics.Record("CacheMiss", 1)
		tempFilePath, err := utils.DownloadFile(remoteFile.Link(config.Get().General.BotToken))
		if err != nil {
			logger.Error.Println(userInfo+"failed to download file:", err)
		}

		logger.Info.Printf("%sGet sticker %s.%s", userInfo, update.Message.Sticker.SetName, update.Message.Sticker.Emoji)

		//delete temp file
		defer utils.RemoveFile(tempFilePath)

		//init convert task
		convertTask := utils.ConvertTask{
			InputFilePath:  tempFilePath,
			InputExtension: utils.GetFileExtName(tempFilePath),
		}

		//check file type
		if config.Get().General.SupportTGSFile == false && convertTask.InputExtension == "tgs" {
			utils.EditMsgText(update.Message.Chat.ID, msg.MessageID, languages.Get(&update).BotMsg.ErrStickerNotSupport)
			return
		}

		//generate output file path
		fileExt := "gif"
		if convertTask.InputExtension == "webp" {
			fileExt = "png"
		}
		outPath := fmt.Sprintf("./storage/tmp/convert_%s.%s", utils.RandString(), fileExt)
		convertTask.OutputFilePath = outPath

		//start to convert
		ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
		err = convertTask.Run(ctx)
		cancel()
		if err != nil {
			logger.Error.Println(userInfo+"failed to convert:", err, convertTask.OutputFilePath)
			utils.EditMsgText(update.Message.Chat.ID, msg.MessageID, languages.Get(&update).BotMsg.ErrConvertFailed)
			return
		}

		//upload file
		utils.SendAction(update.Message.Chat.ID, utils.ChatActionSendDocument)
		sentMsg, err := utils.SendFileByPath(&update, outPath)
		if err != nil {
			logger.Error.Println(userInfo+"failed to SendFile:", err)
			utils.EditMsgText(update.Message.Chat.ID,
				msg.MessageID,
				fmt.Sprintf("%s(TelegramAPI:%s)", languages.Get(&update).BotMsg.ErrSendFileFailed, err.Error()),
			)
			return
		}

		//CacheSticker
		if config.Get().Cache.Enabled == true {
			cacheItem, err = db.CacheSticker(*update.Message.Sticker, convertTask.OutputFilePath)
			if err != nil {
				logger.Error.Println(userInfo+"CacheSticker Error ", err)
			} else {
				cacheItem.ConvertedFileID = sentMsg.Document.FileID
				if err := cacheItem.Update(); err != nil {
					logger.Error.Println(userInfo+"failed to update cache:", err)
				}
			}
		}
		utils.RemoveFile(outPath)
	}

	//Consume the current user's daily limit
	if err = db.ConsumeLimit(&update); err != nil {
		logger.Error.Println(userInfo + err.Error())
	}

	err = utils.BotRequest(tgbotapi.NewEditMessageTextAndMarkup(update.Message.Chat.ID, msg.MessageID, languages.Get(&update).BotMsg.ConvertCompleted, tgbotapi.NewInlineKeyboardMarkup(
		tgbotapi.NewInlineKeyboardRow(tgbotapi.NewInlineKeyboardButtonData(languages.Get(&update).BotMsg.DownloadStickerSet, DownloadStickerSetCallbackQuery)),
	)))
	if err != nil {
		logger.Error.Println(userInfo+"failed to delete msg:", err)
	}

	return
}
