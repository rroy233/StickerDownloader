package utils

import tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"

func entityBold(text, boldPart string) tgbotapi.MessageEntity {
	offset, length := getPartIndex(text, boldPart)
	return tgbotapi.MessageEntity{
		Type:   "bold",
		Offset: offset,
		Length: length,
	}
}

func entityUnderline(text, boldPart string) tgbotapi.MessageEntity {
	offset, length := getPartIndex(text, boldPart)
	return tgbotapi.MessageEntity{
		Type:   "underline",
		Offset: offset,
		Length: length,
	}
}

func entityLink(text, part, url string) tgbotapi.MessageEntity {
	offset, length := getPartIndex(text, part)
	return tgbotapi.MessageEntity{
		Type:   "url",
		Offset: offset,
		Length: length,
		URL:    url,
	}
}

func entityMention(text, part string) tgbotapi.MessageEntity {
	offset, length := getPartIndex(text, part)
	return tgbotapi.MessageEntity{
		Type:   "mention",
		Offset: offset,
		Length: length,
	}
}

func entityTag(text, part string) tgbotapi.MessageEntity {
	offset, length := getPartIndex(text, part)
	return tgbotapi.MessageEntity{
		Type:   "hashtag",
		Offset: offset,
		Length: length,
	}
}

func entityCode(text, part string) tgbotapi.MessageEntity {
	offset, length := getPartIndex(text, part)
	return tgbotapi.MessageEntity{
		Type:   "code",
		Offset: offset,
		Length: length,
	}
}

func entityTextLink(text, part, url string) tgbotapi.MessageEntity {
	offset, length := getPartIndex(text, part)
	return tgbotapi.MessageEntity{
		Type:   "text_link",
		Offset: offset,
		Length: length,
		URL:    url,
	}
}
