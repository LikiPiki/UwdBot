package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"

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
	Token  string
	ChatId int64
)

func init() {
	if err := godotenv.Load(); err != nil {
		log.Print("No .env file found. Using PRODUCTION enviroment!")
	}
}

func main() {
	botToken, exists := os.LookupEnv("TOKEN")
	// Change bot token to .env token, or use production TOKEN
	if exists {
		Token = botToken
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

	app := app2.NewApp(plugins)
	cntrl := controller.NewController(bot, app, snd, db.UserStorage)

	u := tgbotapi.NewUpdate(0)
	u.Timeout = 60

	updates, err := bot.GetUpdatesChan(u)

	cntrl.Switch(context.Background(), updates)
}
