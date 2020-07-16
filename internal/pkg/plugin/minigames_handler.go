package plugin

import (
	"context"
	"fmt"
	"github.com/LikiPiki/UwdBot/internal/pkg/database"
	"github.com/LikiPiki/UwdBot/internal/pkg/sender"
	"github.com/pkg/errors"
	"strconv"
	"strings"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api"
)

type Minigames struct {
	c                 *sender.Sender
	Polls             []Poll
	LastCasinoMessage *tgbotapi.Message
	errors            chan error
	db                *database.Database
}

func (m *Minigames) Errors() <-chan error {
	return m.errors
}

func (m *Minigames) Init(s *sender.Sender, db *database.Database) {
	m.c = s
	m.Polls = []Poll{}
	m.LastCasinoMessage = &tgbotapi.Message{}
	m.errors = make(chan error)
	m.db = db
}

func (m *Minigames) HandleMessages(*tgbotapi.Message) {}

func (m *Minigames) HandleCommands(*tgbotapi.Message, string) {}

func (m *Minigames) HandleRegisterCommands(msg *tgbotapi.Message, command string, user *database.User) {
	switch command {
	case "poll":
		var err error
		defer func() {
			if err != nil {
				m.errors <- errors.Wrap(err, "cannot send poll")
				return
			}
		}()

		id, err := m.GetPoll()
		msg, err := m.SendPoll(msg, &m.Polls[id], id)
		err = m.UpdatePollMessage(id, &msg)
	case "casino":
		m.sendCasinoMiniGame(
			context.Background(),
			msg,
			user,
		)
	}
}

func (m *Minigames) HandleCallbackQuery(update *tgbotapi.Update) {
	callbackQuery := update.CallbackQuery
	text := update.CallbackQuery.Data
	words := strings.Split(text, "|")
	if len(words) > 0 {
		if words[0] == "poll" {
			if err := m.handlePollCallback(context.Background(), callbackQuery, words); err != nil {
				m.errors <- errors.Wrap(err, "cannot handle poll callback")
			}
		}
	}
}

func (m *Minigames) HandleInlineCommands(update *tgbotapi.Update) {}

func (m *Minigames) HandleAdminCommands(*tgbotapi.Message) {}

func (m *Minigames) GetRegisteredCommands() []string {
	return []string{
		"casino",
		"poll",
	}
}

// --- Poll logic ---
func (m *Minigames) SendPoll(msg *tgbotapi.Message, poll *Poll, id int) (tgbotapi.Message, error) {
	reply := tgbotapi.NewMessage(
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

	message, err := m.c.Send(&reply)
	if err != nil {
		return tgbotapi.Message{}, errors.Wrap(err, "cannot send poll")
	}

	return *message, nil
}

func (m *Minigames) handlePollCallback(ctx context.Context, callbackQuery *tgbotapi.CallbackQuery, words []string) error {
	var err error
	username := callbackQuery.From.UserName
	userID := callbackQuery.From.ID
	num, ans := words[1], words[2]
	questionNumber, err := strconv.Atoi(num)
	if err != nil {
		return errors.Wrap(err, "cannot handle poll callback")
	}
	ansNumber, err := strconv.Atoi(ans)
	if err != nil {
		return errors.Wrapf(err, "cannot convert %s to int", ans)
	}
	ok := m.CheckNumberQuestions(questionNumber, ansNumber)
	if len(m.Polls) <= questionNumber {
		if err := m.c.SendInlineKeyboardReply(callbackQuery, "Эта викторина уже устарела!"); err != nil {
			return errors.Wrap(err, "cannot send inline keyboard reply")
		}

		return nil
	}

	currentPoll := m.Polls[questionNumber]
	memberClicked := currentPoll.HaveMember(username)
	m.AddMember(username, currentPoll)
	if memberClicked >= 1 {
		if err := m.c.SendInlineKeyboardReply(callbackQuery, "Чувак, у тебя только одна попытка!"); err != nil {
			return errors.Wrap(err, "cannot send inline keyboard reply")
		}

		return nil
	}

	if ok {
		check, solved := m.CheckPoll(questionNumber, ansNumber)
		if !solved {
			if check {
				m.SolvePoll(questionNumber, ansNumber)
				if err = m.c.SendInlineKeyboardReply(callbackQuery, generateSolved()); err != nil {
					return errors.Wrap(err, "cannot send inline keyboard reply")
				}
				if _, err = m.c.EditMessageMarkup(currentPoll.Message, nil); err != nil {
					return errors.Wrap(err, "cannot edit message markup")
				}

				text, err := m.GetPollResults(ctx, username, userID, currentPoll)
				if err != nil {
					return errors.Wrap(err, "cannot get poll results")
				}

				if _, err = m.c.EditMessageText(currentPoll.Message, text, "markdown"); err != nil {
					return errors.Wrap(err, "cannot edit message text")
				}
			} else {
				if err = m.c.SendInlineKeyboardReply(callbackQuery, generateWrong()); err != nil {
					return errors.Wrap(err, "cannot send inline keyboard reply")
				}
			}
		} else {
			if err = m.c.SendInlineKeyboardReply(callbackQuery, generateSolved()); err != nil {
				return errors.Wrap(err, "cannot send inline keyboard reply")
			}
		}
	} else {
		if err = m.c.SendInlineKeyboardReply(callbackQuery, "Данный пол устарел! /poll"); err != nil {
			return errors.Wrap(err, "cannot send inline keyboard reply")
		}
	}

	return nil
}
