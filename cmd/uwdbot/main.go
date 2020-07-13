package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"strconv"

	app2 "github.com/LikiPiki/UwdBot/cmd/uwdbot/app"
	"github.com/LikiPiki/UwdBot/internal/pkg/controller"
	"github.com/LikiPiki/UwdBot/internal/pkg/database"
	pl "github.com/LikiPiki/UwdBot/internal/pkg/plugin"
	"github.com/LikiPiki/UwdBot/internal/pkg/sender"
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api"
	"github.com/joho/godotenv"
	"golang.org/x/net/proxy"
)

const (
	PROXY = "195.201.103.36:1080"
)

var (
	Token   string
	ChatId  int64
	AdminId int64
)

func init() {
	if err := godotenv.Load(); err != nil {
		log.Println(err, "No .env file found. Using PRODUCTION enviroment!")
	}
}

func main() {
	botToken, exists := os.LookupEnv("TOKEN")
	// Change bot token to .env token, or use production TOKEN
	if exists {
		Token = botToken
	}

	uwdChatID, exists := os.LookupEnv("CHAT_ID")
	if exists {
		var err error
		ChatId, err = strconv.ParseInt(uwdChatID, 10, 64)
		if err != nil {
			panic(err)
		}
	}
	adminChatID, exists := os.LookupEnv("ADMIN_ID")
	if exists {
		var err error
		AdminId, err = strconv.ParseInt(adminChatID, 10, 64)
		if err != nil {
			panic(err)
		}
	}

	dialer, err := proxy.SOCKS5("tcp", PROXY, &proxy.Auth{
		User:     "kirillq",
		Password: "pahnaaaale",
	}, proxy.Direct)

	if err != nil {
		log.Fatalln("can't connect to the proxy:", err)
	}

	fmt.Println("Success connecting to proxy!")

	httpTransport := &http.Transport{}
	httpClient := &http.Client{Transport: httpTransport}
	httpTransport.Dial = dialer.Dial

	bot, err := tgbotapi.NewBotAPIWithClient(Token, httpClient)
	if err != nil {
		log.Fatalln(err)
	}

	bot.Debug = false
	log.Printf("Authorized on account %s", bot.Self.UserName)

	db, err := database.NewDatabase(context.Background())
	defer db.Close()
	if err != nil {
		log.Fatalln("cannot connect to DB", err)
	}

	snd := sender.NewSender(bot, ChatId)
	// Register plugins here
	plugins := pl.Plugins{
		&pl.Base{},
		&pl.Wars{},
		&pl.Minigames{},
		&pl.Profiler{},
	}

	for _, plug := range plugins {
		plug.Init(snd, db)
	}

	// Listen and log errors from plugins
	errorsAgregate := make(chan error)
	for _, plug := range plugins {
		go func(c <-chan error) {
			errorsAgregate <- (<-c)
		}(plug.Errors())
	}

	go func() {
		select {
		case err := <-errorsAgregate:
			log.Println(err)
			msg := tgbotapi.NewMessage(
				AdminId, fmt.Sprintf(
					"***Error:*** ```%s```",
					err.Error(),
				),
			)
			msg.ParseMode = "markdown"

			_, err = snd.Send(&msg)
			if err != nil {
				log.Println(err)
			}
		}
	}()

	app := app2.NewApp(plugins)
	cntrl := controller.NewController(bot, app, snd, db.UserStorage)

	u := tgbotapi.NewUpdate(0)
	u.Timeout = 60

	updates, err := bot.GetUpdatesChan(u)
	if err != nil {
		log.Println(err)
	}

	cntrl.Switch(context.Background(), updates)
}
