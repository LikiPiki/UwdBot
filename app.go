package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"math/rand"
	"net/http"
	"strconv"
	"time"

	data "UwdBot/database"

	"github.com/PuerkitoBio/goquery"
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api"
)

const (
	UWDChannelVideos = "https://www.youtube.com/user/uwebdesign/videos"
	poll_api_url     = "https://engine.lifeis.porn/api/millionaire.php"
	LEN              = 20
)

var (
	UserRanks = []Rank{
		{"Король", 1000, 1000},
		{"Депутат от народа", 0, 500},
		{"Зажиточный", 500, 300},
		{"Программист", 300, 300},
		{"Только что сдал ЕГЭ", 150, 50},
		{"Пельмень", 100, 100},
		{"Днарь", 0, 50},
		{"Изгой", 0, 0},
	}
)

// for test only
type Rank struct {
	Rank       string
	Coins      int
	Reputation int
}

func GetRank(user data.User) string {
	for _, rank := range UserRanks {
		if (rank.Coins <= user.Coins) && (rank.Reputation <= user.Reputation) {
			return rank.Rank
		}
	}
	return UserRanks[len(UserRanks)-1].Rank
}

type App struct {
	Polls  []Poll
	Videos []string
	admins map[string]bool
}

func InitApp() *App {
	return &App{
		Polls: []Poll{},
		admins: map[string]bool{
			"likipiki": true,
			"websanya": true,
		},
	}
}

func (a *App) IsAdmin(username string) bool {
	return a.admins[username]
}

func (a *App) ParseVideos() {
	a.Videos = make([]string, 0)
	doc, err := goquery.NewDocument(UWDChannelVideos)
	if err != nil {
		log.Println(err)
	}
	parsed := make([]string, 0)
	doc.Find("a").Each(func(i int, s *goquery.Selection) {
		class, _ := s.Attr("class")
		link, _ := s.Attr("href")
		if class == "yt-uix-sessionlink yt-uix-tile-link  spf-link  yt-ui-ellipsis yt-ui-ellipsis-2" {
			video := fmt.Sprintf("youtube.com%s", link)
			parsed = append(parsed, video)
		}
	})
	a.Videos = parsed
}

func (a *App) UpdatePollMessage(id int, msg *tgbotapi.Message) {
	if len(a.Polls) > id {
		a.Polls[id].Message = msg
	} else {
		log.Println("Invalid id")
	}
}

func (a *App) getLastVideoLink() (string, bool) {
	a.ParseVideos()
	if len(a.Videos) > 0 {
		return a.Videos[0], true
	}
	return "", false
}

func (a *App) LoadPoll() Poll {
	req, err := http.NewRequest("GET", poll_api_url, nil)
	if err != nil {
		log.Println(err)
	}

	poll := Poll{}
	poll.members = make(map[string]int)

	q := req.URL.Query()
	q.Add("count", "1")
	value := rand.Intn(3) + 1
	q.Add("q", strconv.Itoa(value))
	req.URL.RawQuery = q.Encode()

	resp, err := http.Get(req.URL.String())

	if err != nil {
		log.Println(err)
	}

	defer resp.Body.Close()

	jsonCode, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		log.Println(err)
	}

	err = json.Unmarshal(jsonCode, &poll)
	if err != nil {
		log.Println(err)
	}
	poll.Shuffle()
	poll.Data.Time = time.Now()
	poll.Data.Solved = false
	return poll
}

func (a *App) GetPoll() int {
	poll := a.LoadPoll()
	if len(a.Polls) < LEN {
		a.Polls = append(a.Polls, poll)
		return len(a.Polls) - 1
	} else {
		for id, current := range a.Polls {
			if current.Data.Solved == true {
				a.Polls[id] = poll
				return id
			}
		}
	}
	minTime, index := a.Polls[0].Data.Time, 0
	for id, current := range a.Polls {
		if current.Data.Time.Sub(minTime).String()[0] == '-' {
			minTime = current.Data.Time
			index = id
		}
	}
	a.Polls[index] = poll
	return index
}

func (a *App) CheckNumberQuestions(num, ans int) bool {
	if (len(a.Polls) > num) && (ans < 4) {
		return true
	}
	return false
}

func (a *App) SolvePoll(num, ans int) {
	a.Polls[num].Data.Solved = true
}

func (a *App) CheckPoll(num, ans int) (bool, bool) {
	if a.Polls[num].Data.Valid == ans {
		return true, a.Polls[num].Data.Solved
	}
	return false, a.Polls[num].Data.Solved
}

// Registration and Delete account
func (a *App) RegisterNewUser(msg *tgbotapi.Message) string {
	user := data.User{}
	count, err := user.CountUsersWithID(msg.From.ID)
	if err != nil {
		log.Panicln(err)
		return "Что то пошло не так..."
	}
	if count > 0 {
		return "Ты уже зарегистрирован!"
	}

	user.UserID = uint64(msg.From.ID)
	user.Username = msg.From.UserName
	_, err = user.CreateNewUser()

	if err != nil {
		return "Не удалось добавить. Попробуй позже..."
	}

	return "Вы успешно прошли регистрацию. /me"
}

func (a *App) UnregUser(msg *tgbotapi.Message) string {
	user := data.User{}
	user.DeleteUser(msg.From.ID)
	return "Ну заходи как нибудь еще, что делать..."
}

func (a *App) ShowUserInfo(msg *tgbotapi.Message) string {
	var err error
	var user data.User
	user, err = user.FindUserByID(msg.From.ID)
	if err != nil {
		log.Println(err)
	}

	var repStat, coinsStat float32
	repStat, coinsStat, err = user.GetUserStatistics()
	if err != nil {
		log.Println(err)
	}

	rank := GetRank(user)

	return fmt.Sprintf(
		"Привет ***@%s*** - ___%s___\nТвоя репутация: ***%d\n***💰: ***%d***\n\nТы на ***%d***%% круче остальных и на ***%d***%% богаче!",
		user.Username,
		rank,
		user.Reputation,
		user.Coins,
		int(repStat*100),
		int(coinsStat*100),
	)
}

func (a *App) GetShop(msg *tgbotapi.Message) string {
	weap := data.Weapon{}
	weapons, err := weap.GetAllWeapons()
	if err != nil {
		return "Не удалось загрузить магазин..."
	}
	reply := "***Уютный shop 🛒 ***\n\n***Оружие:***\n"
	for _, w := range weapons {
		reply += fmt.Sprintf(
			"%d) ___%s___ %d🗡️, %d💰\n",
			w.ID,
			w.Name,
			w.Power,
			w.Cost,
		)
	}
	reply += "\n___Интересный стафф:___\nПоявится в скором времени...\n\n___Купить товар /buy номер товара___"
	return reply
}
