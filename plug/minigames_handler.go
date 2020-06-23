package plug

import (
	data "UwdBot/database"
	"UwdBot/sender"
	"fmt"
	"log"
	"strconv"
	"strings"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api"
)

type Minigames struct {
	c                 *sender.Sender
	Polls             []Poll
	LastCasinoMessage *tgbotapi.Message
}

func (m *Minigames) Init(s *sender.Sender) {
	m.c = s
	m.Polls = []Poll{}
	m.LastCasinoMessage = &tgbotapi.Message{}
}

func (m *Minigames) HandleMessages(msg *tgbotapi.Message) {}

func (m *Minigames) HandleCommands(msg *tgbotapi.Message, command string) {}

func (m *Minigames) HandleRegisterCommands(msg *tgbotapi.Message, command string, user *data.User) {
	switch command {
	case "poll":
		id := m.GetPoll()
		msg := m.SendPoll(
			msg,
			&m.Polls[id],
			id,
		)
		m.UpdatePollMessage(id, &msg)
	case "casino":
		go m.sendCasinoMiniGame(
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
			m.handlePollCallback(callbackQuery, words)
		}
	}
}

func (m *Minigames) HandleAdminCommands(msg *tgbotapi.Message) {}

func (m *Minigames) GetRegisteredCommands() []string {
	return []string{
		"casino",
		"poll",
	}
}

// --- Poll logic ---
func (m *Minigames) SendPoll(msg *tgbotapi.Message, poll *Poll, id int) tgbotapi.Message {
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

	message := m.c.Send(&reply)

	return *message
}

func (m *Minigames) handlePollCallback(callbackQuery *tgbotapi.CallbackQuery, words []string) {
	username := callbackQuery.From.UserName
	userID := callbackQuery.From.ID
	num, ans := words[1], words[2]
	questionNumber, err := strconv.Atoi(num)
	if err != nil {
		log.Println(err)
		return
	}
	ansNumber, err := strconv.Atoi(ans)
	if err != nil {
		log.Println(err)
	}
	ok := m.CheckNumberQuestions(questionNumber, ansNumber)
	if len(m.Polls) <= questionNumber {
		m.c.SendInlineKeyboardReply(
			callbackQuery,
			"Эта викторина уже устарела!",
		)
		return
	}
	currentPoll := m.Polls[questionNumber]
	memberClicked := currentPoll.HaveMember(username)
	currentPoll.AddMember(username)
	if memberClicked >= 1 {
		m.c.SendInlineKeyboardReply(
			callbackQuery,
			"Чувак, у тебя только одна попытка!",
		)
		return
	}

	if ok {
		check, solved := m.CheckPoll(questionNumber, ansNumber)
		if !solved {
			if check {
				m.SolvePoll(questionNumber, ansNumber)
				m.c.SendInlineKeyboardReply(
					callbackQuery,
					generateSolved(),
				)
				m.c.EditMessageMarkup(
					currentPoll.Message,
					nil,
				)
				m.c.EditMessageText(
					currentPoll.Message,
					currentPoll.GetPollResults(username, userID),
					"markdown",
				)
			} else {
				m.c.SendInlineKeyboardReply(
					callbackQuery,
					generateWrong(),
				)
			}
		} else {
			m.c.SendInlineKeyboardReply(
				callbackQuery,
				generateSolved(),
			)
		}
	} else {
		m.c.SendInlineKeyboardReply(callbackQuery, "Данный пол устарел! /poll")
	}
}
