package main

import (
	"fmt"
	"log"
	"net/http"
	"os"
	"strconv"
	"strings"

	"github.com/go-telegram-bot-api/telegram-bot-api"
	"golang.org/x/net/proxy"
)

const (
	TOKEN = "861382625:AAH0kDDXzb1ZVlOVoVDB3O1wZw00U_YfVME"
	// DEBUG token
	// TOKEN = "427558135:AAEnSxpTD_wOMxhoWjVzrNO5YQa3vZHbEMM"
	PROXY = "195.201.103.36:1080"
)

func main() {
	dialer, err := proxy.SOCKS5("tcp", PROXY, &proxy.Auth{
		User:     "kirillq",
		Password: "pahnaaaale",
	}, proxy.Direct)

	if err != nil {
		fmt.Fprintln(os.Stderr, "can't connect to the proxy:", err)
		os.Exit(1)
	} else {
		fmt.Println("Success connecting to proxy!")
	}

	httpTransport := &http.Transport{}
	httpClient := &http.Client{Transport: httpTransport}
	httpTransport.Dial = dialer.Dial

	bot, err := tgbotapi.NewBotAPIWithClient(TOKEN, httpClient)
	if err != nil {
		log.Panic(err)
	}

	bot.Debug = false

	log.Printf("Authorized on account %s", bot.Self.UserName)
	var sender Sender
	sender.Init(bot)

	app := InitApp()

	u := tgbotapi.NewUpdate(0)
	u.Timeout = 60

	updates, err := bot.GetUpdatesChan(u)

	for update := range updates {
		msg := update.Message

		if update.CallbackQuery != nil {
			callbackQuery := update.CallbackQuery
			text := update.CallbackQuery.Data
			words := strings.Split(text, "|")
			if len(words) > 0 {
				if words[0] == "poll" {
					num, ans := words[1], words[2]
					questionNumber, err := strconv.Atoi(num)
					if err != nil {
						log.Println(err)
					}
					ansNumber, err := strconv.Atoi(ans)
					if err != nil {
						log.Println(err)
					}
					ok := app.CheckNumberQuestions(questionNumber, ansNumber)

					if ok {
						check, solved := app.CheckPoll(questionNumber, ansNumber)
						if !solved {
							if check {
								app.SolvePoll(questionNumber, ansNumber)
								sender.SendInlineKeyboardReply(
									callbackQuery,
									generateUserSolve(callbackQuery.From.UserName),
								)
								currentPoll := app.Polls[questionNumber]
								sender.EditMessageMarkup(
									currentPoll.Message,
									nil,
								)
								sender.EditMessageText(
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
								sender.SendInlineKeyboardReply(
									callbackQuery,
									generateWrong(callbackQuery.From.UserName),
								)
							}
						} else {
							sender.SendInlineKeyboardReply(
								callbackQuery,
								generateSolved(callbackQuery.From.UserName),
							)
						}
					} else {
						sender.SendInlineKeyboardReply(callbackQuery, "Данный пол устарел! /poll")
					}
				}
			}
		}

		if update.Message == nil { // ignore any non-Message Updates
			continue
		}

		if msg.IsCommand() {
			command := msg.Command()
			switch command {
			case "last":
				link, fl := app.getLastVideoLink()
				if fl {
					sender.SendReply(msg,
						fmt.Sprintf("Последнее видео: %s", link),
					)
				}
			case "kek":
				go sender.SendReply(
					msg,
					generateKek(),
				)
			case "poll":
				id := app.GetPoll()
				msg := sender.SendPoll(
					msg,
					&app.Polls[id],
					id,
				)
				app.UpdatePollMessage(id, &msg)
			default:
				go sender.SendUnknown(msg)
			}
		}

	}
}
