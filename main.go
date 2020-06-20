package main

import (
	"fmt"
	"log"
	"net/http"
	"os"
	"strconv"

	"UwdBot/database"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api"
	"github.com/joho/godotenv"
	"golang.org/x/net/proxy"
)

const (
	PROXY = "195.201.103.36:1080"
)

var (
	TOKEN   = "861382625:AAH0kDDXzb1ZVlOVoVDB3O1wZw00U_YfVME"
	CHAT_ID int64
)

func init() {
	log.Print("No .env file found. Using PRODUCTION enviroment!")
	if err := godotenv.Load(); err != nil {
	}
}

func main() {
	botToken, exists := os.LookupEnv("TOKEN")

	// Change bot token to .env token, or use production TOKEN
	if exists {
		TOKEN = botToken
	}

	var chatID string
	chatID, exists = os.LookupEnv("CHAT_ID")
	if exists {
		var err error
		CHAT_ID, err = strconv.ParseInt(chatID, 10, 32)
		if err != nil {
			panic(err)
		}
	} else {
		// UWD CHAT ID
		CHAT_ID = -1001094145433
	}

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

	database.InitDB()
	defer database.CloseDatabase()

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
