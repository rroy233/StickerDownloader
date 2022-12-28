package utils

import (
	"errors"
	"fmt"
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"github.com/rroy233/logger"
	"io/ioutil"
	"log"
	"net/http"
	"strings"
	"time"
	"unicode/utf16"
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
	smsg, _ := bot.Send(msg)
	return smsg
}

func SendFile(update *tgbotapi.Update, filePath string) error {
	if update.Message == nil && update.CallbackQuery == nil {
		return errors.New("message nil")
	}

	var msg tgbotapi.MediaGroupConfig
	file := tgbotapi.FilePath(filePath)
	if update.Message != nil {
		msg = tgbotapi.NewMediaGroup(update.Message.Chat.ID, []interface{}{tgbotapi.NewInputMediaDocument(file)})
	} else if update.CallbackQuery != nil || update.CallbackQuery.Message != nil {
		msg = tgbotapi.NewMediaGroup(update.CallbackQuery.Message.Chat.ID, []interface{}{tgbotapi.NewInputMediaDocument(file)})
	}

	_, err := bot.SendMediaGroup(msg)
	if err != nil {
		logger.Error.Println("上传文件失败：", err)
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

	fileName := fmt.Sprintf("./storage/tmp/upload_%d_%s", time.Now().UnixMicro(), oFileName)
	err = ioutil.WriteFile(fileName, data, 0666)
	if err != nil {
		return "", err
	}

	return fileName, nil
}

func EditMsgText(chatID int64, msgID int, msg string) {
	_, err := bot.Send(tgbotapi.NewEditMessageText(chatID, msgID, msg))
	if err != nil {
		logger.Error.Println("[EditMsgText]bot.Send失败:", err)
	}
	return
}

func CallBack(callbackQueryID string, text string) {
	callback := tgbotapi.NewCallback(callbackQueryID, text)
	//不能用bot.Send(callback)方法，有bug
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

func getLogPrefixMessage(update *tgbotapi.Update) string {
	return fmt.Sprintf("[Message][User:%d @%s %s][Chat:%s]",
		update.Message.From.ID,
		update.Message.From.UserName,
		update.Message.From.FirstName+update.Message.From.LastName,
		fmt.Sprintf("(%s) %d %s", update.Message.Chat.Type, update.Message.Chat.ID, update.Message.Chat.Title),
	)
}

func getLogPrefixCallbackQuery(update *tgbotapi.Update) string {
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

	//debug
	//log.Println("textUTF16=", textUTF16)
	//log.Println("PartUTF16=", PartUTF16)

	for {
		//越界
		if i > len(textUTF16)-1 || j > len(PartUTF16) {
			offset = -1
			break
		}
		//debug
		//log.Printf("offset=%d,textUTF16[%d]=%d,PartUTF16[%d]=%d\n", offset, i, textUTF16[i], j, PartUTF16[j])

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
