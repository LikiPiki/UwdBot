package main

import (
	"fmt"
	"log"

	data "UwdBot/database"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api"
)

const (
	UWDChatID = -1001094145433
	Message   = 0
	Sticker   = 1
)

type Sender struct {
	bot *tgbotapi.BotAPI
}

func (s *Sender) Init(bot *tgbotapi.BotAPI) {
	s.bot = bot
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

func (s Sender) SendReplyToMessage(msg *tgbotapi.Message, text string) {
	var reply tgbotapi.MessageConfig
	reply = tgbotapi.NewMessage(
		msg.Chat.ID,
		text,
	)
	reply.ReplyToMessageID = msg.MessageID

	_, err := s.bot.Send(reply)

	if err != nil {
		log.Println(err)
	}
}

func (s Sender) SendMarkdownReply(msg *tgbotapi.Message, text string) {
	var reply tgbotapi.MessageConfig
	reply = tgbotapi.NewMessage(
		msg.Chat.ID,
		text,
	)

	reply.ParseMode = "markdown"
	reply.ReplyToMessageID = msg.MessageID

	_, err := s.bot.Send(reply)

	if err != nil {
		log.Println(err)
	}
}

func (s Sender) SendInlineKeyboardReply(CallbackQuery *tgbotapi.CallbackQuery, text string) {
	s.bot.AnswerCallbackQuery(tgbotapi.NewCallback(CallbackQuery.ID, text))
}

func (s Sender) SendPoll(msg *tgbotapi.Message, poll *Poll, id int) tgbotapi.Message {
	var reply tgbotapi.MessageConfig
	reply = tgbotapi.NewMessage(
		msg.Chat.ID,
		poll.Data.Question,
	)
	keyboard := tgbotapi.InlineKeyboardMarkup{}
	for k, class := range poll.Data.Answers {
		var row []tgbotapi.InlineKeyboardButton
		btn := tgbotapi.NewInlineKeyboardButtonData(class, fmt.Sprintf("poll|%d|%d", id, k))
		row = append(row, btn)
		keyboard.InlineKeyboard = append(keyboard.InlineKeyboard, row)
	}
	reply.ReplyMarkup = keyboard

	message, err := s.bot.Send(reply)

	if err != nil {
		log.Println(err)
	}

	return message
}

func (s Sender) SendCasinoMiniGame(msg *tgbotapi.Message, user *data.User) {
	if user.Coins < 10 {
		s.SendReplyToMessage(msg, "Слишком мало денег, накопи еще и приходи потом!")
		return
	}
	user.DecreaseMoney(10)
	miniGame, status := generateCasino()
	s.SendReply(msg, miniGame)
	if status {
		user.AddMoney(50)
		s.SendReplyToMessage(msg, "Уважаемый, вы победили... + 50💰")
	}
}

func (s Sender) EditMessageMarkup(msg *tgbotapi.Message, markup *tgbotapi.InlineKeyboardMarkup) tgbotapi.Message {
	edit := tgbotapi.EditMessageReplyMarkupConfig{
		BaseEdit: tgbotapi.BaseEdit{
			ChatID:      msg.Chat.ID,
			MessageID:   msg.MessageID,
			ReplyMarkup: markup,
		},
	}

	message, err := s.bot.Send(edit)

	if err != nil {
		log.Println(err)
	}
	return message
}

func (s Sender) EditMessageText(msg *tgbotapi.Message, text string, parsemode string) tgbotapi.Message {
	edit := tgbotapi.EditMessageTextConfig{
		BaseEdit: tgbotapi.BaseEdit{
			ChatID:    msg.Chat.ID,
			MessageID: msg.MessageID,
		},
		Text:      text,
		ParseMode: parsemode,
	}

	message, err := s.bot.Send(edit)

	if err != nil {
		log.Println(err)
	}
	return message
}

func (s Sender) SendStickerOrText(msg *tgbotapi.Message, chance int, sending string) {
	switch chance {
	case Sticker:
		s.SendSticker(
			msg,
			sending,
		)
	case Message:
		s.SendReply(
			msg,
			sending,
		)
	}
}
