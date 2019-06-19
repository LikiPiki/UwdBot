package main

import (
	"fmt"
	"log"

	"github.com/PuerkitoBio/goquery"
)

const (
	UWDChannelVideos = "https://www.youtube.com/user/uwebdesign/videos"
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
			video := fmt.Sprintf("youtube.com/%s", link)
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
