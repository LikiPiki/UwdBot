package main

import (
	"log"
	"math/rand"

	"github.com/go-telegram-bot-api/telegram-bot-api"
)

const (
	UWDChatID = -1001094145433
	Message   = 0
	Sticker   = 1
)

type Sender struct {
	bot *tgbotapi.BotAPI
}

func (s *Sender) SendMessageToUWDChat(message string) {
	var reply tgbotapi.MessageConfig
	reply = tgbotapi.NewMessage(
		UWDChatID,
		message,
	)

	_, err := s.bot.Send(reply)

	if err != nil {
		log.Println(err)
	}
}

func (s *Sender) Init(bot *tgbotapi.BotAPI) {
	s.bot = bot
}

func (s *Sender) SendSticker(msg *tgbotapi.Message, stickerID string) {
	sticker := tgbotapi.NewStickerShare(
		msg.Chat.ID,
		stickerID,
	)

	_, err := s.bot.Send(sticker)

	if err != nil {
		log.Println(err)
	}
}

func (s Sender) SendReply(msg *tgbotapi.Message, text string) {
	var reply tgbotapi.MessageConfig
	reply = tgbotapi.NewMessage(
		msg.Chat.ID,
		text,
	)

	_, err := s.bot.Send(reply)

	if err != nil {
		log.Println(err)
	}
}

func (s Sender) SendUnknown(msg *tgbotapi.Message) {
	messages := []string{
		"Я не понимаю эту команду",
		"Возможно банан все сломал...",
		"Почитай /help",
	}
	stickers := []string{
		"CAADAgAD6gEAAsE8ngaA44zCtd3nBAI",
		"CAADAgADngADk8vUCDsXkJ5Ka6VsAg",
		"CAADAgADRwYAAkxb1gn9h6PpAyEkggI",
	}

	chance := rand.Intn(2)
	switch chance {
	case Sticker:
		s.SendSticker(
			msg,
			stickers[rand.Intn(len(stickers))],
		)
	case Message:
		s.SendReply(
			msg,
			messages[rand.Intn(len(messages))],
		)
	}
}
