package utils

import (
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"gopkg.in/rroy233/logger.v2"
	"time"
)

var msgQueue chan tgbotapi.Chattable

func initSender(senderNum int) {
	msgQueue = make(chan tgbotapi.Chattable, senderNum*2)
	for i := 0; i < senderNum; i++ {
		go sender()
	}
}

func addToSendQueue(msg tgbotapi.Chattable) {
	msgQueue <- msg
	return
}

func sender() {
	for {
		msg, _ := <-msgQueue
		Limiter.Take()
		resp, err := bot.Request(msg)
		if err != nil {
			logger.Error.Printf("%s[sender][%s]%s", loggerPrefix, err.Error(), JsonEncode(resp))
			if resp != nil && resp.ErrorCode == 429 {
				//Too Many Requests error
				time.Sleep(10 * time.Second)
				msgQueue <- msg
			}
		}
	}
}

func BotRequest(c tgbotapi.Chattable) error {
	Limiter.Take()
	_, err := bot.Request(c)
	if err != nil {
		logger.Error.Println(loggerPrefix + "[BotRequest]" + err.Error())
	}
	return err
}

func BotSend(c tgbotapi.Chattable) (tgbotapi.Message, error) {
	Limiter.Take()
	msg, err := bot.Send(c)
	if err != nil {
		logger.Error.Println(loggerPrefix + "[BotSend]" + err.Error())
	}
	return msg, err
}
