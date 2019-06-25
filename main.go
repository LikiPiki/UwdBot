package main

import (
	"fmt"
	"log"
	"net/http"
	"os"

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

	controller := InitController(bot, app, &sender)

	u := tgbotapi.NewUpdate(0)
	u.Timeout = 60

	updates, err := bot.GetUpdatesChan(u)

	controller.Switch(updates)

}
