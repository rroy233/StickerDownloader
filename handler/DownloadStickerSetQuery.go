package handler

import (
	"context"
	"fmt"
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"github.com/rroy233/logger"
	"github.com/rroy233/tg-stickers-dl/config"
	"github.com/rroy233/tg-stickers-dl/utils"
	"os"
	"sync/atomic"
	"time"
)

const MB = 1 << 20

type downloadTask struct {
	finished   int32
	failed     int32
	total      int32
	folderName string
}

func DownloadStickerSetQuery(update tgbotapi.Update) {
	stickerSet, err := bot.GetStickerSet(tgbotapi.GetStickerSetConfig{
		Name: update.CallbackQuery.Message.ReplyToMessage.Sticker.SetName,
	})
	if err != nil {
		logger.Error.Println("获取表情包-GetStickerSet失败:", err)
		utils.CallBackWithAlert(update.CallbackQuery.ID, "获取失败")
	}

	utils.CallBack(update.CallbackQuery.ID, "ok")

	oMsg := tgbotapi.NewMessage(update.CallbackQuery.Message.Chat.ID, "正在处理...")
	oMsg.ReplyToMessageID = update.CallbackQuery.Message.MessageID
	msg, err := bot.Send(oMsg)
	if err != nil {
		logger.Error.Println("获取表情包-发送处理中信息失败:", err)
		utils.SendPlainText(&update, "发生错误")
		return
	}

	//创建目录
	folderName := fmt.Sprintf("./storage/tmp/stickers_%s_%d", stickerSet.Name, time.Now().UnixMicro())
	err = os.Mkdir(folderName, 0777)
	if err != nil || utils.IsExist(folderName) == false {
		logger.Error.Println("获取表情包-创建目录失败:", err)
		utils.EditMsgText(update.CallbackQuery.Message.Chat.ID, msg.MessageID, "失败-1001")
		return
	}
	//删除临时目录
	defer func() {
		err = os.RemoveAll(folderName)
		if err != nil {
			logger.Error.Println("获取表情包-删除临时目录失败:", folderName, err)
		}
	}()

	cancelCtx, cancel := context.WithCancel(context.Background())
	queue := make(chan tgbotapi.Sticker, 10)
	task := new(downloadTask)
	for i := 0; i < config.Get().General.DownloadWorkerNum; i++ {
		go downloadWorker(cancelCtx, queue, task)
	}
	task.total = int32(len(stickerSet.Stickers))
	task.folderName = folderName
	timeStart := time.Now()
	go func() {
		for _, sticker := range stickerSet.Stickers {
			queue <- sticker
		}
	}()
	go func() {
		for true {
			select {
			case <-cancelCtx.Done():
				return
			default:
				//实时更新进度
				utils.EditMsgText(update.CallbackQuery.Message.Chat.ID, msg.MessageID, fmt.Sprintf("正在下载[%d/%d]...", task.finished+task.failed, task.total))
				time.Sleep(1 * time.Second)
			}
		}
	}()

	//等待
	success := true
	for {
		if time.Now().Sub(timeStart).Seconds() > 60 { //60秒超时
			success = false
			break
		}
		if task.finished+task.failed == task.total {
			break
		}
		logger.Debug.Println("获取表情包-等待中-已用时", time.Now().Sub(timeStart).Seconds())
		time.Sleep(1 * time.Second)
	}
	cancel()
	if success {
		zipFilePath := fmt.Sprintf("./storage/tmp/%s_%d.zip", stickerSet.Name, time.Now().UnixMicro())
		err = utils.Compress(folderName, zipFilePath)
		if err != nil {
			logger.Error.Println("获取表情包-压缩失败:", err)
			utils.EditMsgText(update.CallbackQuery.Message.Chat.ID, msg.MessageID, "失败-1002")
			return
		}
		utils.EditMsgText(update.CallbackQuery.Message.Chat.ID, msg.MessageID, fmt.Sprintf("任务完成(成功%d/失败%d)，正在上传文件。。。", task.finished, task.failed))

		//删除
		defer utils.RemoveFile(zipFilePath)

		fileStat, err := os.Stat(zipFilePath)
		if err != nil {
			logger.Error.Println("获取表情包-读取压缩包信息失败:", err)
			utils.EditMsgText(update.CallbackQuery.Message.Chat.ID, msg.MessageID, "失败-1003")
			return
		}
		if fileStat.Size() > 50*MB {
			logger.Info.Println("获取表情包-正在上传文件(第三方文件托管平台)")
			uploadTask := utils.NewUploadFile(zipFilePath)
			err = uploadTask.Upload2FileHost()
			if err != nil {
				logger.Error.Println("获取表情包-上传失败:", err)
				utils.EditMsgText(update.CallbackQuery.Message.Chat.ID, msg.MessageID, "上传失败！")
				return
			}
			utils.EditMsgText(update.CallbackQuery.Message.Chat.ID, msg.MessageID, fmt.Sprintf("上传成功！！\n表情包名:%s\n文件大小:%s\n下载地址:%s\n", stickerSet.Name, uploadTask.InfoRes.Data.File.Metadata.Size.Readable, uploadTask.InfoRes.Data.File.Url.Short))
			logger.Info.Println("获取表情包-上传文件(第三方文件托管平台)成功！！！")
		} else {
			logger.Info.Println("获取表情包-正在上传文件(Telegram)")
			err = utils.SendFile(&update, zipFilePath)
			if err != nil {
				logger.Error.Println("获取表情包-上传失败:", err)
				utils.EditMsgText(update.CallbackQuery.Message.Chat.ID, msg.MessageID, "上传失败！")
				return
			}
			utils.EditMsgText(update.CallbackQuery.Message.Chat.ID, msg.MessageID, fmt.Sprintf("上传成功！！\n表情包名:%s\n文件大小:%dMB\n", stickerSet.Name, fileStat.Size()>>20))
			logger.Info.Println("获取表情包-上传文件(Telegram)成功！！！")
		}

	} else {
		utils.EditMsgText(update.CallbackQuery.Message.Chat.ID, msg.MessageID, fmt.Sprintf("任务超时!!!"))
	}

	return
}

func downloadWorker(ctx context.Context, queue chan tgbotapi.Sticker, task *downloadTask) {
	var sticker tgbotapi.Sticker
	for {
		select {
		case <-ctx.Done():
			return
		case sticker = <-queue:
			i := task.finished + task.failed
			sum := task.total
			remoteFile, err := bot.GetFile(tgbotapi.FileConfig{
				FileID: sticker.FileID,
			})
			if err != nil {
				logger.Error.Printf("获取表情包[%d/%d]-获取文件失败:%s", i, sum, err.Error())
				atomic.AddInt32(&task.failed, 1)
				continue
			}

			tempFilePath, err := utils.DownloadFile(remoteFile.Link(config.Get().General.BotToken))
			if err != nil {
				logger.Error.Printf("获取表情包[%d/%d]-下载文件失败:%s", i, sum, err.Error())
				atomic.AddInt32(&task.failed, 1)
				continue
			}
			logger.Info.Printf("获取表情包[%d/%d]-已下载临时文件：%s\n", i, sum, tempFilePath)

			outFilePath := fmt.Sprintf("%s/%d.gif", task.folderName, i)
			err = utils.Mp4ToGif(tempFilePath, outFilePath)
			utils.RemoveFile(tempFilePath)
			if err != nil {
				logger.Error.Printf("获取表情包[%d/%d]-转换失败：%s\n", i, sum, err.Error())
				atomic.AddInt32(&task.failed, 1)
				continue
			}
			atomic.AddInt32(&task.finished, 1)
		}
	}
}
