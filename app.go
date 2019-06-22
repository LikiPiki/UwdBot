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

	"github.com/PuerkitoBio/goquery"
)

const (
	UWDChannelVideos = "https://www.youtube.com/user/uwebdesign/videos"
	poll_api_url     = "https://engine.lifeis.porn/api/millionaire.php"
	LEN              = 20
)

type App struct {
	Polls  []Poll
	Videos []string
	admins []string
}

func InitApp() *App {
	return &App{
		Polls: []Poll{},
		admins: []string{
			"likipiki",
			"websanya",
		},
	}
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
