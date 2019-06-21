package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"

	"github.com/PuerkitoBio/goquery"
)

const (
	UWDChannelVideos = "https://www.youtube.com/user/uwebdesign/videos"
	poll_api_url     = "https://engine.lifeis.porn/api/millionaire.php"
)

type App struct {
	Videos []string
	admins []string
}

func InitApp() *App {
	return &App{
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

func (a *App) GetPoll() Poll {
	req, err := http.NewRequest("GET", poll_api_url, nil)
	if err != nil {
		log.Println(err)
	}

	q := req.URL.Query()
	q.Add("count", "1")
	q.Add("q", "3")
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

	poll := Poll{}
	err = json.Unmarshal(jsonCode, &poll)
	if err != nil {
		log.Println(err)
	}

	poll.Shuffle()
	poll.Numerate()
	return poll
}
