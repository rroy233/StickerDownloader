package utils

import (
	"encoding/json"
	"errors"
	"fmt"
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"github.com/rroy233/StickerDownloader/statistics"
	"gopkg.in/rroy233/logger.v2"
	"io/ioutil"
	"log"
	"math/rand"
	"net/http"
	"os"
	"strings"
	"time"
	"unicode/utf16"
)

type ChatAction string

const (
	ChatActionTyping       = ChatAction("typing")
	ChatActionUploadPhoto  = ChatAction("upload_photo")
	ChatActionSendDocument = ChatAction("upload_document")
)

func SendPlainText(update *tgbotapi.Update, text string, entity ...tgbotapi.MessageEntity) {
	if update.Message == nil {
		return
	}
	var msg tgbotapi.MessageConfig
	if update.Message != nil {
		msg = tgbotapi.NewMessage(update.Message.Chat.ID, text)
		msg.ReplyToMessageID = update.Message.MessageID
		if entity != nil {
			msg.Entities = entity
		}
		addToSendQueue(msg)
	} else if update.CallbackQuery != nil || update.CallbackQuery.Message != nil {
		msg = tgbotapi.NewMessage(update.CallbackQuery.Message.Chat.ID, text)
		if entity != nil {
			msg.Entities = entity
		}
		addToSendQueue(msg)
	}
}

func SendImg(update *tgbotapi.Update, fileData []byte) (msgSent tgbotapi.Message) {
	if update.Message == nil {
		return
	}
	var msg tgbotapi.PhotoConfig
	file := tgbotapi.FileBytes{
		Name:  "image.jpg",
		Bytes: fileData,
	}
	if update.Message != nil {
		msg = tgbotapi.NewPhoto(update.Message.Chat.ID, file)
	} else if update.CallbackQuery != nil || update.CallbackQuery.Message != nil {
		msg = tgbotapi.NewPhoto(update.CallbackQuery.Message.Chat.ID, file)
	}
	smsg, _ := BotSend(msg)
	return smsg
}

// SendFileByPath 发送文件
//
// 传入文件本地存储地址
func SendFileByPath(update *tgbotapi.Update, filePath string) (tgbotapi.Message, error) {
	if update.Message == nil && update.CallbackQuery == nil {
		return tgbotapi.Message{}, errors.New("message nil")
	}

	var msg tgbotapi.MediaGroupConfig
	file := tgbotapi.FilePath(filePath)
	if update.Message != nil {
		msg = tgbotapi.NewMediaGroup(update.Message.Chat.ID, []interface{}{tgbotapi.NewInputMediaDocument(file)})
	} else if update.CallbackQuery != nil || update.CallbackQuery.Message != nil {
		msg = tgbotapi.NewMediaGroup(update.CallbackQuery.Message.Chat.ID, []interface{}{tgbotapi.NewInputMediaDocument(file)})
	}

	Limiter.Take()

	//bot.SendMediaGroup(msg)
	//github.com/go-telegram-bot-api/telegram-bot-api/v5这个库有bug
	//运行一段时间之后的http请求会得到500 Internal Server Error的异常响应
	//考虑到该库处于年久失修的状态，且目前暂不考虑使更换此项目的tg机器人库
	//若遇到上述问题，重启一下即可解决
	resp, err := bot.Request(msg)
	if err != nil {
		logger.Error.Println("bot.Request error：", err, JsonEncode(resp))
		return tgbotapi.Message{}, errors.New("network error-1")
	}
	var messages []tgbotapi.Message
	err = json.Unmarshal(resp.Result, &messages)
	if err != nil {
		logger.Error.Println("failed to send file：", err)
		return tgbotapi.Message{}, errors.New("network error-2")
	}

	//记录statistic
	info, err := os.Stat(filePath)
	if err == nil {
		statistics.Statistics.Record("NetworkUpload", int32(info.Size()))
	}

	return messages[0], err
}

// SendFileByFileID 发送文件
//
// 传入文件 file_id
func SendFileByFileID(update *tgbotapi.Update, fileID string) error {
	if update.Message == nil && update.CallbackQuery == nil {
		return errors.New("message nil")
	}

	var msg tgbotapi.MediaGroupConfig
	file := tgbotapi.FileID(fileID)
	if update.Message != nil {
		msg = tgbotapi.NewMediaGroup(update.Message.Chat.ID, []interface{}{tgbotapi.NewInputMediaDocument(file)})
	} else if update.CallbackQuery != nil || update.CallbackQuery.Message != nil {
		msg = tgbotapi.NewMediaGroup(update.CallbackQuery.Message.Chat.ID, []interface{}{tgbotapi.NewInputMediaDocument(file)})
	}

	Limiter.Take()
	_, err := bot.SendMediaGroup(msg)
	if err != nil {
		logger.Error.Println("failed to send file：", err)
		if strings.Contains(err.Error(), "api.telegram.org") == false {
			return err
		}
		return errors.New("network error")
	}

	return err
}

func LogUserInfo(update *tgbotapi.Update) string {
	if update.Message != nil {
		return fmt.Sprintf("[%s(@%s) %d]", update.Message.Chat.FirstName+update.Message.Chat.LastName, update.Message.Chat.UserName, update.Message.Chat.ID)
	}
	if update.CallbackQuery != nil {
		return fmt.Sprintf("[%s(@%s) %d]", update.CallbackQuery.Message.Chat.FirstName+update.CallbackQuery.Message.Chat.LastName, update.CallbackQuery.Message.Chat.UserName, update.CallbackQuery.Message.Chat.ID)
	}
	return ""
}

func SendSticker(update *tgbotapi.Update, fileID string) {
	if update.Message == nil {
		return
	}
	var msg tgbotapi.StickerConfig
	if update.Message != nil {
		msg = tgbotapi.NewSticker(update.Message.Chat.ID, tgbotapi.FileID(fileID))
		msg.ReplyToMessageID = update.Message.MessageID
		addToSendQueue(msg)
	} else if update.CallbackQuery != nil || update.CallbackQuery.Message != nil {
		msg = tgbotapi.NewSticker(update.CallbackQuery.Message.Chat.ID, tgbotapi.FileID(fileID))
		addToSendQueue(msg)
	}
}

func SendPlainTextWithKeyboard(update *tgbotapi.Update, text string, keyboard *tgbotapi.ReplyKeyboardMarkup, entity ...tgbotapi.MessageEntity) {
	if update.Message == nil {
		return
	}
	var msg tgbotapi.MessageConfig
	if update.Message != nil {
		msg = tgbotapi.NewMessage(update.Message.Chat.ID, text)
		msg.ReplyToMessageID = update.Message.MessageID
		msg.ReplyMarkup = *keyboard
		if entity != nil {
			msg.Entities = entity
		}
		addToSendQueue(msg)
	} else if update.CallbackQuery != nil || update.CallbackQuery.Message != nil {
		msg = tgbotapi.NewMessage(update.CallbackQuery.Message.Chat.ID, text)
		msg.ReplyMarkup = *keyboard
		if entity != nil {
			msg.Entities = entity
		}
		addToSendQueue(msg)
	}
}
func BotGetFile(config tgbotapi.FileConfig) (tgbotapi.File, error) {
	Limiter.Take()
	return bot.GetFile(config)
}

func BotGetSelf() tgbotapi.User {
	return bot.Self
}

func BotGetStickerSet(config tgbotapi.GetStickerSetConfig) (tgbotapi.StickerSet, error) {
	Limiter.Take()
	return bot.GetStickerSet(config)
}

func DownloadFile(fileUrl string) (string, error) {
	req, err := http.NewRequest(http.MethodGet, fileUrl, nil)
	if err != nil {
		return "", err
	}

	client := &http.Client{
		Timeout: 60 * time.Second,
	}

	resp, err := client.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	data, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}

	//get file name
	oFileName := "file"
	urls := strings.Split(fileUrl, "/")
	if len(urls) == 0 {
		return "", errors.New("url无效")
	}
	if strings.Contains(urls[len(urls)-1], ".") != false {
		oFileName = urls[len(urls)-1]
	}

	fileName := fmt.Sprintf("./storage/tmp/upload_%s_%s", RandString(), oFileName)
	err = ioutil.WriteFile(fileName, data, 0666)
	if err != nil {
		return "", err
	}

	//记录statistic
	statistics.Statistics.Record("NetworkDownload", int32(len(data)))

	return fileName, nil
}

func SendAction(chaiID int64, action ChatAction) {
	msgQueue <- tgbotapi.NewChatAction(chaiID, string(action))
}

func EditMsgText(chatID int64, msgID int, msg string, entity ...tgbotapi.MessageEntity) {
	newMsg := tgbotapi.NewEditMessageText(chatID, msgID, msg)
	if len(entity) != 0 {
		newMsg.Entities = entity
	}
	addToSendQueue(newMsg)
	return
}

func EditMsgTextAndMarkup(chatID int64, msgID int, msg string, markup tgbotapi.InlineKeyboardMarkup) {
	newMsg := tgbotapi.NewEditMessageTextAndMarkup(chatID, msgID, msg, markup)
	addToSendQueue(newMsg)
	return
}

func DeleteMsg(chatID int64, MsgID int) {
	addToSendQueue(tgbotapi.NewDeleteMessage(chatID, MsgID))
	return
}

func GetUID(update *tgbotapi.Update) int64 {
	if update.Message != nil {
		return update.Message.Chat.ID
	}
	if update.CallbackQuery != nil {
		return update.CallbackQuery.Message.Chat.ID
	}
	return -1
}

func CallBack(callbackQueryID string, text string) {
	callback := tgbotapi.NewCallback(callbackQueryID, text)
	//不能用bot.Send(callback)方法，有bug
	Limiter.Take()
	resp, err := bot.Request(callback)
	if err != nil {
		logger.Error.Println("[CallBack]bot.Request失败:", err)
		return
	}
	if string(resp.Result) != "true" {
		logger.Error.Println("[CallBack]请求不ok:", resp)
		return
	}
	return
}
func CallBackWithAlert(callbackQueryID string, text string) {
	callback := tgbotapi.NewCallbackWithAlert(callbackQueryID, text)
	//不能用bot.Send(callback)方法，有bug
	Limiter.Take()
	resp, err := bot.Request(callback)
	if err != nil {
		logger.Error.Println("[CallBackWithAlert]bot.Request失败:", err)
		return
	}
	if string(resp.Result) != "true" {
		logger.Error.Println("[CallBackWithAlert]请求不ok:", resp)
		return
	}
	return
}

func GetLogPrefixMessage(update *tgbotapi.Update) string {
	return fmt.Sprintf("[Message][User:%d @%s %s][Chat:%s]",
		update.Message.From.ID,
		update.Message.From.UserName,
		update.Message.From.FirstName+update.Message.From.LastName,
		fmt.Sprintf("(%s) %d %s", update.Message.Chat.Type, update.Message.Chat.ID, update.Message.Chat.Title),
	)
}

func GetLogPrefixCallbackQuery(update *tgbotapi.Update) string {
	return fmt.Sprintf("[CallbackQuery][User:%d @%s %s][Chat:%s]",
		update.CallbackQuery.From.ID,
		update.CallbackQuery.From.UserName,
		update.CallbackQuery.From.FirstName+update.CallbackQuery.From.LastName,
		fmt.Sprintf("(%s) %d %s", update.CallbackQuery.Message.Chat.Type, update.CallbackQuery.Message.Chat.ID, update.CallbackQuery.Message.Chat.Title),
	)
}

// 顺序查找UTF-6编码字符串中子串的第一次出现的位置
// 返回offset=-1则为找到
func getPartIndex(text, part string) (offset, length int) {
	textUTF16 := utf16.Encode([]rune(text))
	PartUTF16 := utf16.Encode([]rune(part))
	offset = 0
	i := 0
	j := 0

	for {
		//越界
		if i > len(textUTF16)-1 || j > len(PartUTF16) {
			offset = -1
			break
		}

		//判断
		if textUTF16[i] == PartUTF16[j] {
			i++
			j++
		} else {
			j = 0
			offset++
			i = offset
		}

		//结果
		if j > len(PartUTF16)-1 {
			break
		}
	}
	if offset == -1 {
		if logger.Error != nil {
			logger.FATAL.Printf("%s[util][getPartIndex]无法在【%s】中找到【%s】\n", loggerPrefix, text, part)
		} else {
			log.Printf("%s[util][getPartIndex]无法在【%s】中找到【%s】\n", loggerPrefix, text, part)
		}
	}

	return offset, len(PartUTF16)
}

func RandString() string {
	rand.Seed(time.Now().UnixNano())
	return MD5Short(fmt.Sprintf("%d_%d", time.Now().UnixNano(), rand.Int()))
}
