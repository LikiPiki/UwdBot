package main

import (
	"fmt"
	"log"
	"regexp"
	"sort"
	"strconv"
	"strings"

	data "UwdBot/database"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api"
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
			c.handleRegisterUserCommand(msg)
		} else {
			if c.app.IsAdmin(msg.From.UserName) {
				c.handleAdminCommands(msg)
			}

			switch {
			case msg.NewChatMembers != nil && len(*msg.NewChatMembers) > 0:
				c.handleJoinMembers(
					msg,
				)
			case msg.LeftChatMember != nil:
				c.handleLeftMembers(
					msg,
				)
			}
		}

	}
}

func (c Controller) handleAdminCommands(msg *tgbotapi.Message) {
	messageTextBytes := []byte(msg.Text)
	regexSay := regexp.MustCompile(`@say ([^\n]*)`)
	indexes := regexSay.FindSubmatchIndex(messageTextBytes)

	if len(indexes) == 4 {
		go c.sender.SendMessageToUWDChat(msg.Text[indexes[2]:indexes[3]])
	}
}

func (c Controller) handleLeftMembers(msg *tgbotapi.Message) {
	if len(msg.LeftChatMember.UserName) > 0 {
		c.sender.SendReply(
			msg,
			fmt.Sprintf("Пошёл в жопу @%s!", msg.LeftChatMember.UserName),
		)
	} else {
		c.sender.SendReply(msg, "Пошёл в жопу!")
	}
}

func (c Controller) handleJoinMembers(msg *tgbotapi.Message) {
	text := GetJoin((*msg.NewChatMembers)[0].UserName)

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
		fmt.Println("chat id is ", msg.Chat.ID)
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
	case "reg":
		if CHAT_ID != msg.Chat.ID {
			c.sender.SendReplyToMessage(msg, "Этот функционал не работет в этом чате")
		}
		go c.sender.SendReplyToMessage(
			msg,
			c.app.RegisterNewUser(msg),
		)
	}

}

func (c Controller) handleRegisterUserCommand(msg *tgbotapi.Message) {
	// For binary search put commands in alphabetic
	commandList := []string{"casino", "me", "shop", "unreg"}
	command := msg.Command()
	user := data.User{}

	// Binary search command
	i := sort.Search(
		len(commandList),
		func(i int) bool {
			return command <= commandList[i]
		},
	)
	if i < len(commandList) && commandList[i] == command {
	} else {
		return
	}

	if CHAT_ID != msg.Chat.ID {
		c.sender.SendReplyToMessage(msg, "Этот функционал не работет в этом чате")
		return
	}

	// check user exits
	var err error
	user, err = user.FindUserByID(msg.From.ID)

	if err != nil || user.ID == 0 {
		c.sender.SendReplyToMessage(msg, "Ты не зарегистрирован, сначала /reg")
		return
	}

	switch command {
	case "unreg":
		go c.sender.SendReplyToMessage(
			msg,
			c.app.UnregUser(msg),
		)
	case "me":
		go c.sender.SendMarkdownReply(
			msg,
			c.app.ShowUserInfo(msg),
		)
	// Wars commands
	case "shop":
		go c.sender.SendMarkdownReply(
			msg,
			c.app.GetShop(msg),
		)
	case "casino":
		go c.sender.SendCasinoMiniGame(
			msg,
			&user,
		)
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
	ok := c.app.CheckNumberQuestions(questionNumber, ansNumber)
	if len(c.app.Polls) <= questionNumber {
		c.sender.SendInlineKeyboardReply(
			callbackQuery,
			"Эта викторина уже устарела!",
		)
		return
	}
	currentPoll := c.app.Polls[questionNumber]
	memberClicked := currentPoll.HaveMember(username)
	currentPoll.AddMember(username)
	if memberClicked >= 1 {
		c.sender.SendInlineKeyboardReply(
			callbackQuery,
			"Чувак, у тебя только одна попытка!",
		)
		return
	}

	if ok {
		check, solved := c.app.CheckPoll(questionNumber, ansNumber)
		if !solved {
			if check {
				c.app.SolvePoll(questionNumber, ansNumber)
				c.sender.SendInlineKeyboardReply(
					callbackQuery,
					generateSolved(),
				)
				c.sender.EditMessageMarkup(
					currentPoll.Message,
					nil,
				)
				c.sender.EditMessageText(
					currentPoll.Message,
					currentPoll.GetPollResults(username, userID),
					"markdown",
				)
			} else {
				c.sender.SendInlineKeyboardReply(
					callbackQuery,
					generateWrong(),
				)
			}
		} else {
			c.sender.SendInlineKeyboardReply(
				callbackQuery,
				generateSolved(),
			)
		}
	} else {
		c.sender.SendInlineKeyboardReply(callbackQuery, "Данный пол устарел! /poll")
	}
}
