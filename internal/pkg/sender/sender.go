package sender

import (
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api"
	"github.com/pkg/errors"
)

const (
	Message = iota
	Sticker
)

type Sender struct {
	UWDChatID int64
	bot       *tgbotapi.BotAPI
}

func NewSender(bot *tgbotapi.BotAPI, UWDChatID int64) *Sender {
	return &Sender{
		UWDChatID: UWDChatID,
		bot:       bot,
	}
}

func (s *Sender) SendMessageToUWDChat(message string) error {
	reply := tgbotapi.NewMessage(
		s.UWDChatID,
		message,
	)

	_, err := s.bot.Send(reply)
	if err != nil {
		return errors.Wrap(err, "cannot send message")
	}

	return nil
}

func (s *Sender) SendSticker(msg *tgbotapi.Message, stickerID string) error {
	sticker := tgbotapi.NewStickerShare(
		msg.Chat.ID,
		stickerID,
	)

	_, err := s.bot.Send(sticker)
	if err != nil {
		return errors.Wrap(err, "cannot send sticker")
	}

	return nil
}

func (s *Sender) SendReply(msg *tgbotapi.Message, text string) error {
	reply := tgbotapi.NewMessage(
		msg.Chat.ID,
		text,
	)

	_, err := s.bot.Send(reply)
	if err != nil {
		return errors.Wrap(err, "cannot send reply")
	}

	return nil
}

func (s *Sender) Send(msgConfig *tgbotapi.MessageConfig) (*tgbotapi.Message, error) {
	msg, err := s.bot.Send(msgConfig)
	if err != nil {
		return nil, errors.Wrap(err, "cannot send message")
	}

	return &msg, nil
}

func (s *Sender) SendReplyToMessage(msg *tgbotapi.Message, text string) error {
	reply := tgbotapi.NewMessage(
		msg.Chat.ID,
		text,
	)
	reply.ReplyToMessageID = msg.MessageID

	_, err := s.bot.Send(reply)
	if err != nil {
		return errors.Wrap(err, "cannot send reply to message")
	}

	return nil
}

func (s *Sender) SendMarkdownReply(msg *tgbotapi.Message, text string) error {
	reply := tgbotapi.NewMessage(
		msg.Chat.ID,
		text,
	)

	reply.ParseMode = "markdown"
	reply.ReplyToMessageID = msg.MessageID

	_, err := s.bot.Send(reply)
	if err != nil {
		return errors.Wrap(err, "cannot send MD reply")
	}

	return nil
}

func (s *Sender) SendInlineKeyboardReply(CallbackQuery *tgbotapi.CallbackQuery, text string) error {
	_, err := s.bot.AnswerCallbackQuery(tgbotapi.NewCallback(CallbackQuery.ID, text))
	if err != nil {
		return errors.Wrap(err, "cannot answer callback query")
	}

	return nil
}

func (s *Sender) EditMessageMarkup(msg *tgbotapi.Message, markup *tgbotapi.InlineKeyboardMarkup) (tgbotapi.Message, error) {
	edit := tgbotapi.EditMessageReplyMarkupConfig{
		BaseEdit: tgbotapi.BaseEdit{
			ChatID:      msg.Chat.ID,
			MessageID:   msg.MessageID,
			ReplyMarkup: markup,
		},
	}

	message, err := s.bot.Send(edit)
	if err != nil {
		return tgbotapi.Message{}, errors.Wrap(err, "cannot edit MD message")
	}

	return message, nil
}

func (s *Sender) DeleteMessage(msg *tgbotapi.Message) error {
	deleteMsg := tgbotapi.DeleteMessageConfig{
		ChatID:    msg.Chat.ID,
		MessageID: msg.MessageID,
	}

	_, err := s.bot.DeleteMessage(deleteMsg)
	if err != nil {
		return errors.Wrap(err, "cannot delete message")
	}

	return nil
}

func (s *Sender) EditMessageText(msg *tgbotapi.Message, text string, parsemode string) (tgbotapi.Message, error) {
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
		return tgbotapi.Message{}, errors.Wrap(err, "cannot edit message text")
	}

	return message, nil
}

func (s *Sender) AnswerInlineQuery(config *tgbotapi.InlineConfig) error {
	_, err := s.bot.AnswerInlineQuery(*config)
	if err != nil {
		return errors.Wrap(err, "cannot answer to inline query")
	}

	return nil
}

func (s *Sender) SendStickerOrText(msg *tgbotapi.Message, chance int, sending string) error {
	switch chance {
	case Sticker:
		if err := s.SendSticker(msg, sending); err != nil {
			return errors.Wrap(err, "cannot send sticker")
		}
	case Message:
		if err := s.SendReply(msg, sending); err != nil {
			return errors.Wrap(err, "cannot send reply")
		}
	}

	return nil
}
