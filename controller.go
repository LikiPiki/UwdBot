package main

import (
	"fmt"
	"log"
	"strconv"
	"strings"

	"github.com/go-telegram-bot-api/telegram-bot-api"
)

type Controller struct {
	bot    *tgbotapi.BotAPI
	app    *App
	sender *Sender
}

func InitController(bot *tgbotapi.BotAPI, app *App, sender *Sender) *Controller {
	controller := Controller{
		bot:    bot,
		app:    app,
		sender: sender,
	}
	return &controller
}

func (c Controller) Switch(updates tgbotapi.UpdatesChannel) {
	for update := range updates {
		msg := update.Message

		if update.CallbackQuery != nil {
			c.handleCallbackQuery(update)
		}

		if update.Message == nil { // ignore any non-Message Updates
			continue
		}

		if msg.IsCommand() {
			c.handleCommand(msg)
		} else {
			c.handleJoinMember(
				msg,
			)
		}

	}
}

func (c Controller) handleJoinMember(msg *tgbotapi.Message) {
	text := GetJoin()

	go c.sender.SendMarkdownReply(
		msg,
		text,
	)
}

func (c Controller) handleCommand(msg *tgbotapi.Message) {
	command := msg.Command()
	switch command {
	case "last":
		link, fl := c.app.getLastVideoLink()
		if fl {
			c.sender.SendReply(msg,
				fmt.Sprintf("Последнее видео: %s", link),
			)
		}
	case "kek":
		go c.sender.SendReply(
			msg,
			generateKek(),
		)
	case "riot":
		messageType, sending := GenerateRiot()
		go c.sender.SendStickerOrText(
			msg,
			messageType,
			sending,
		)
	case "poll":
		id := c.app.GetPoll()
		msg := c.sender.SendPoll(
			msg,
			&c.app.Polls[id],
			id,
		)
		c.app.UpdatePollMessage(id, &msg)
	default:
		go c.sender.SendUnknown(msg)
	}
}

func (c Controller) handleCallbackQuery(update tgbotapi.Update) {
	callbackQuery := update.CallbackQuery
	text := update.CallbackQuery.Data
	words := strings.Split(text, "|")
	if len(words) > 0 {
		if words[0] == "poll" {
			c.handlePollCallback(callbackQuery, words)
		}
	}
}

func (c Controller) handlePollCallback(callbackQuery *tgbotapi.CallbackQuery, words []string) {
	num, ans := words[1], words[2]
	questionNumber, err := strconv.Atoi(num)
	if err != nil {
		log.Println(err)
	}
	ansNumber, err := strconv.Atoi(ans)
	if err != nil {
		log.Println(err)
	}
	ok := c.app.CheckNumberQuestions(questionNumber, ansNumber)

	if ok {
		check, solved := c.app.CheckPoll(questionNumber, ansNumber)
		if !solved {
			if check {
				c.app.SolvePoll(questionNumber, ansNumber)
				c.sender.SendInlineKeyboardReply(
					callbackQuery,
					generateUserSolve(callbackQuery.From.UserName),
				)
				currentPoll := c.app.Polls[questionNumber]
				c.sender.EditMessageMarkup(
					currentPoll.Message,
					nil,
				)
				c.sender.EditMessageText(
					currentPoll.Message,
					fmt.Sprintf(
						"`%s`\nПравильный ответ - ___%s___.\nОтветил - @%s",
						currentPoll.Message.Text,
						currentPoll.GetSuccess(),
						callbackQuery.From.UserName,
					),
					"markdown",
				)
			} else {
				c.sender.SendInlineKeyboardReply(
					callbackQuery,
					generateWrong(callbackQuery.From.UserName),
				)
			}
		} else {
			c.sender.SendInlineKeyboardReply(
				callbackQuery,
				generateSolved(callbackQuery.From.UserName),
			)
		}
	} else {
		c.sender.SendInlineKeyboardReply(callbackQuery, "Данный пол устарел! /poll")
	}
}