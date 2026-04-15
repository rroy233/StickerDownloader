package handler

import (
	"context"
	"fmt"
	tgbotapi "github.com/OvyFlash/telegram-bot-api"
	"github.com/rroy233/StickerDownloader/config"
	"github.com/rroy233/StickerDownloader/db"
	"github.com/rroy233/StickerDownloader/languages"
	"github.com/rroy233/StickerDownloader/statistics"
	"github.com/rroy233/StickerDownloader/utils"
	"gopkg.in/rroy233/logger.v2"
	"time"
)

func AnimationMessage(update tgbotapi.Update) {
	userInfo := utils.GetLogPrefixMessage(&update)

	// 前置缓存检查，跳过排队和下载转码
	cacheItem, err := db.FindStickerCacheItem(update.Message.Animation.FileUniqueID)
	if err == nil && cacheItem.ConvertedFileID != "" {
		statistics.Statistics.Record("CacheHit", 1)

		if err := utils.SendFileByFileID(&update, cacheItem.ConvertedFileID); err != nil {
			logger.Error.Println(userInfo+"failed to send file via FILE_ID:", err)
			utils.SendPlainText(&update, fmt.Sprintf("%s(TelegramAPI:%s)", languages.Get(&update).BotMsg.ErrSendFileFailed, err.Error()))
			return
		}

		if err = db.ConsumeLimit(&update); err != nil {
			logger.Error.Println(userInfo + err.Error())
		}

		msg := tgbotapi.NewMessage(update.Message.Chat.ID, languages.Get(&update).BotMsg.ConvertCompleted)
		msg.ReplyParameters.MessageID = update.Message.MessageID
		_, err = utils.BotSend(msg)
		if err != nil {
			logger.Error.Println(userInfo+"failed to send completed msg:", err)
		}
		return
	}

	statistics.Statistics.Record("CacheMiss", 1)

	oMsg := tgbotapi.NewMessage(update.Message.Chat.ID, languages.Get(&update).BotMsg.Processing)
	oMsg.ReplyParameters.MessageID = update.Message.MessageID
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

	remoteFile, err := utils.BotGetFile(tgbotapi.FileConfig{
		FileID: update.Message.Animation.FileID,
	})
	if err != nil {
		logger.Error.Println(userInfo+"failed to get file:", err)
		utils.EditMsgText(update.Message.Chat.ID, msg.MessageID, languages.Get(&update).BotMsg.ErrFailedToDownload)
		return
	}

	tempFilePath, err := utils.DownloadFile(remoteFile.Link(config.Get().General.BotToken))
	if err != nil {
		logger.Error.Println(userInfo+"failed to download file:", err)
		utils.EditMsgText(update.Message.Chat.ID, msg.MessageID, languages.Get(&update).BotMsg.ErrFailedToDownload)
		return
	}

	logger.Info.Printf("%sGet Animation => %s", userInfo, tempFilePath)

	//delete temp file
	defer utils.RemoveFile(tempFilePath)

	//check file type
	if utils.GetFileExtName(tempFilePath) != "mp4" {
		utils.EditMsgText(update.Message.Chat.ID, msg.MessageID, languages.Get(&update).BotMsg.ErrStickerNotSupport)
		return
	}

	//path to save converted file
	outPath := fmt.Sprintf("./storage/tmp/convert_%d.gif", time.Now().UnixMicro())
	defer func() {
		utils.RemoveFile(outPath)
	}()

	//init convert task
	convertTask := utils.ConvertTask{
		InputFilePath:  tempFilePath,
		InputExtension: "mp4",
		OutputFilePath: outPath,
	}

	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	err = convertTask.Run(ctx)
	cancel()
	if err != nil {
		logger.Error.Println(userInfo+"failed to convert:", err)
		utils.EditMsgText(update.Message.Chat.ID, msg.MessageID, languages.Get(&update).BotMsg.ErrConvertFailed)
		return
	}

	//Dequeue
	dequeue(qItem)
	//Dequeue

	sentMsg, err := utils.SendFileByPath(&update, outPath)
	if err != nil {
		logger.Error.Println(userInfo+"failed to SendFile:", err)
		utils.EditMsgText(
			update.Message.Chat.ID,
			msg.MessageID,
			fmt.Sprintf("%s(TelegramAPI:%s)", languages.Get(&update).BotMsg.ErrSendFileFailed, err.Error()),
		)
		return
	}

	// 缓存Animation
	if config.Get().Cache.Enabled == true {
		fakeSticker := tgbotapi.Sticker{
			FileID:       update.Message.Animation.FileID,
			FileUniqueID: update.Message.Animation.FileUniqueID,
		}
		cacheItem, err = db.CacheSticker(fakeSticker, convertTask.OutputFilePath)
		if err != nil {
			logger.Error.Println(userInfo+"CacheSticker Error ", err)
		} else {
			cacheItem.ConvertedFileID = sentMsg.Document.FileID
			if err := cacheItem.Update(); err != nil {
				logger.Error.Println(userInfo+"failed to update cache:", err)
			}
		}
	}

	//Consume the current user's daily limit
	if err = db.ConsumeLimit(&update); err != nil {
		logger.Error.Println(userInfo + err.Error())
	}

	utils.EditMsgText(update.Message.Chat.ID, msg.MessageID, languages.Get(&update).BotMsg.ConvertCompleted)
	if err != nil {
		logger.Error.Println(userInfo+"failed to delete msg:", err)
	}

	return
}
