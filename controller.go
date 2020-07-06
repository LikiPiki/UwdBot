package main

import (
	"fmt"
	"sort"

	data "UwdBot/database"
	"UwdBot/sender"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api"
)

type Controller struct {
	bot               *tgbotapi.BotAPI
	app               *App
	sender            *sender.Sender
	registredCommands []string
}

func InitController(bot *tgbotapi.BotAPI, app *App, sender *sender.Sender) *Controller {
	controller := Controller{
		bot:    bot,
		app:    app,
		sender: sender,
	}

	for _, plug := range app.Plugs {
		controller.registredCommands = append(
			controller.registredCommands,
			plug.GetRegisteredCommands()...,
		)
	}
	sort.Strings(controller.registredCommands)

	return &controller
}

func (c *Controller) Switch(updates tgbotapi.UpdatesChannel) {
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
			if c.app.IsAdmin(msg.From.ID) {
				c.handleAdminCommands(msg)
			}
			for _, plug := range c.app.Plugs {
				plug.HandleMessages(msg)
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

func (c *Controller) handleAdminCommands(msg *tgbotapi.Message) {
	user := data.User{}
	var err error
	user, err = user.FindUserByID(msg.From.ID)
	if err != nil {
		c.sender.SendReply(
			msg, "Вы не зарегистрированы",
		)
		return
	}
	if !user.IsAdmin {
		c.sender.SendReply(
			msg, "Вы не являетесь администратором! Ухадите...",
		)
		return
	}
	for _, plug := range c.app.Plugs {
		plug.HandleAdminCommands(msg)
	}
}

func (c *Controller) handleLeftMembers(msg *tgbotapi.Message) {
	if len(msg.LeftChatMember.UserName) > 0 {
		c.sender.SendReply(
			msg,
			fmt.Sprintf("Пошёл в жопу @%s!", msg.LeftChatMember.UserName),
		)
	} else {
		c.sender.SendReply(msg, "Пошёл в жопу!")
	}
}

func (c *Controller) handleJoinMembers(msg *tgbotapi.Message) {
	text := GetJoin((*msg.NewChatMembers)[0].UserName)

	go c.sender.SendMarkdownReply(
		msg,
		text,
	)
}

func (c *Controller) handleCommand(msg *tgbotapi.Message) {
	command := msg.Command()

	for _, plug := range c.app.Plugs {
		plug.HandleCommands(msg, command)
	}
}

func (c *Controller) handleRegisterUserCommand(msg *tgbotapi.Message) {
	command := msg.Command()
	user := data.User{}

	// Binary search command
	i := sort.Search(
		len(c.registredCommands),
		func(i int) bool {
			return command <= c.registredCommands[i]
		},
	)
	if i < len(c.registredCommands) && c.registredCommands[i] == command {
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

	if err != nil {
		c.sender.SendReplyToMessage(msg, "Ты не зарегистрирован, сначала /reg")
		return
	}
	if user.Blacklist {
		c.sender.SendReplyToMessage(msg, "Вы заблокированы за нечестную игру")
		return
	}

	for _, plug := range c.app.Plugs {
		plug.HandleRegisterCommands(msg, command, &user)
	}
}

func (c *Controller) handleCallbackQuery(update tgbotapi.Update) {
	for _, plug := range c.app.Plugs {
		plug.HandleCallbackQuery(&update)
	}
}
