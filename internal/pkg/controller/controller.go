package controller

import (
	"context"
	"fmt"
	"log"
	"math/rand"
	"sort"

	"github.com/LikiPiki/UwdBot/cmd/uwdbot/app"
	"github.com/LikiPiki/UwdBot/internal/pkg/database"
	"github.com/LikiPiki/UwdBot/internal/pkg/sender"
	"github.com/pkg/errors"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api"
)

type Controller struct {
	bot                *tgbotapi.BotAPI
	app                *app.App
	sender             *sender.Sender
	registeredCommands []string

	userStorage *database.UserStorage
}

func NewController(bot *tgbotapi.BotAPI, app *app.App, sender *sender.Sender, userStorage *database.UserStorage) *Controller {
	controller := Controller{
		bot:         bot,
		app:         app,
		sender:      sender,
		userStorage: userStorage,
	}

	for _, plug := range app.Plugs {
		controller.registeredCommands = append(
			controller.registeredCommands,
			plug.GetRegisteredCommands()...,
		)
	}
	sort.Strings(controller.registeredCommands)

	return &controller
}

func (c *Controller) Switch(ctx context.Context, updates tgbotapi.UpdatesChannel) {
	for update := range updates {
		msg := update.Message

		if update.CallbackQuery != nil {
			c.handleCallbackQuery(&update)
			continue
		}

		if (update.Message == nil) && (update.InlineQuery != nil) {
			c.handleInlineQuery(&update)
			continue
		}

		if update.Message.From.UserName == "devstorm" {
			num := rand.Intn(100)
			if num <= 2 {
				c.sender.SendSticker(msg, "CAACAgIAAxkBAAECNsNggSsNF51N1F4KmHQ62d1pLubR-gACPQ0AAiUjIEqkeMbZBUuUVx8E")
			}
		}

		if update.Message == nil { // ignore any non-Message Updates
			continue
		}

		if msg.IsCommand() {
			c.handleCommand(msg)
			if err := c.handleRegisterUserCommand(ctx, msg); err != nil {
				log.Println(err)
			}
		} else {
			if c.userStorage.IsAdmin(ctx, msg.From.ID) {
				if err := c.handleAdminCommands(ctx, msg); err != nil {
					log.Print("cannot handle admin command")
				}
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
				if err := c.handleLeftMembers(msg); err != nil {
					log.Println(err)
				}
			}
		}

	}
}

func (c *Controller) handleAdminCommands(ctx context.Context, msg *tgbotapi.Message) error {
	user, err := c.userStorage.FindUserByID(ctx, msg.From.ID)
	if err != nil {
		if err := c.sender.SendReply(msg, "Вы не зарегистрированы"); err != nil {
			return errors.Wrap(err, "cannot handle admin commands")
		}

		return nil
	}

	if !user.IsAdmin {
		if err := c.sender.SendReply(msg, "Вы не являетесь администратором! Ухадите..."); err != nil {
			return errors.Wrap(err, "cannot handle admin commands")
		}

		return nil
	}

	for _, plug := range c.app.Plugs {
		plug.HandleAdminCommands(msg)
	}

	return nil
}

func (c *Controller) handleLeftMembers(msg *tgbotapi.Message) error {
	if len(msg.LeftChatMember.UserName) > 0 {
		_ = c.sender.SendReply(
			msg,
			fmt.Sprintf("Пошёл в жопу @%s!", msg.LeftChatMember.UserName),
		)
	} else {
		if err := c.sender.SendReply(msg, "Пошёл в жопу!"); err != nil {
			return errors.Wrap(err, "cannot handle left members")
		}
	}

	return nil
}

func (c *Controller) handleJoinMembers(msg *tgbotapi.Message) {
	text := getJoin((*msg.NewChatMembers)[0].UserName)

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

func (c *Controller) handleRegisterUserCommand(ctx context.Context, msg *tgbotapi.Message) error {
	command := msg.Command()
	var user database.User

	// Binary search command
	i := sort.Search(
		len(c.registeredCommands),
		func(i int) bool {
			return command <= c.registeredCommands[i]
		},
	)

	if i >= len(c.registeredCommands) || c.registeredCommands[i] != command {
		return nil
	}

	if c.sender.UWDChatID != msg.Chat.ID {
		if err := c.sender.SendReplyToMessage(msg, "Этот функционал не работет в этом чате"); err != nil {
			return errors.Wrap(err, "cannot send reply")
		}

		return nil
	}

	// check user exits
	var err error
	user, err = c.userStorage.FindUserByID(ctx, msg.From.ID)
	if err != nil {
		if err := c.sender.SendReplyToMessage(msg, "Ты не зарегистрирован, сначала /reg"); err != nil {
			return errors.Wrap(err, "cannot send reply")
		}

		return nil
	}

	if user.Blacklist {
		if err := c.sender.SendReplyToMessage(msg, "Ты заблокирован!"); err != nil {
			return errors.Wrap(err, "cannot send reply")
		}

		return nil
	}

	for _, plug := range c.app.Plugs {
		plug.HandleRegisterCommands(msg, command, &user)
	}

	return nil
}

func (c *Controller) handleCallbackQuery(update *tgbotapi.Update) {
	for _, plug := range c.app.Plugs {
		plug.HandleCallbackQuery(update)
	}
}

func (c *Controller) handleInlineQuery(update *tgbotapi.Update) {
	for _, plug := range c.app.Plugs {
		plug.HandleInlineCommands(update)
	}
}

func getJoin(username string) string {
	if len(username) == 0 {
		return "Привет, представься, иначе я выкину тебя из чата!"
	}
	return fmt.Sprintf(
		"Привет @%s, представься, иначе я выкину тебя из чата!",
		username,
	)
}
