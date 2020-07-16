package plugin

import (
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"math/rand"
	"net/http"
	"strconv"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api"
	"github.com/pkg/errors"
)

const (
	tenorSearchUrl = "https://api.tenor.com/v1/random"
	gifLimit       = 50
)

type GifQuery struct {
	Results []struct {
		URL   string `json:"url"`
		Media []struct {
			Mediumgif struct {
				URL     string `json:"url"`
				Dims    []int  `json:"dims"`
				Preview string `json:"preview"`
				Size    int    `json:"size"`
			} `json:"mediumgif"`
		} `json:"media"`
		Title string `json:"title"`
		ID    string `json:"id"`
	} `json:"results"`
	Next string `json:"next"`
}

func (g *Gif) getGifs(searchString string, gifs *GifQuery) error {
	req, err := http.NewRequest("GET", tenorSearchUrl, nil)
	if err != nil {
		return errors.Wrap(err, "cannot create new requst to Giphy")
	}
	q := req.URL.Query()
	q.Add("key", g.tenorToken)
	q.Add("q", searchString)
	q.Add("limit", strconv.Itoa(gifLimit))
	q.Add("locale", "ru_RU")

	req.URL.RawQuery = q.Encode()

	fmt.Println(req.URL.String())
	resp, err := http.Get(req.URL.String())
	if err != nil {
		return errors.Wrap(err, "cannot perform HTTP GET request")
	}

	defer resp.Body.Close()

	jsonGif, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return errors.Wrap(err, "cannot read HTTP body")
	}

	err = json.Unmarshal(jsonGif, &gifs)

	if err != nil {
		return errors.Wrap(err, "cannot unmarshal json")
	}

	return nil
}
func (g *Gif) loadGifs() {
	g.gifLoading = true
	g.currentMux.Lock()
	defer g.currentMux.Unlock()

	gifs := GifQuery{}
	err := g.getGifs("random", &gifs)

	if err != nil {
		g.errors <- errors.Wrap(err, "cannot load a batch of gifs")
		g.gifLoading = false
		return
	}

	for i, gif := range gifs.Results {
		g.gifs[i] = gif.Media[0].Mediumgif.URL
	}

	g.currentGIF = 0
	g.gifLoading = false
}

func (g *Gif) getCachedGif() string {
	// wait if gifs now loading
	for g.gifLoading {
	}

	g.currentMux.Lock()

	newGif := g.gifs[g.currentGIF]
	g.currentGIF++

	// If gifs unaliable load new batch of gifs
	if g.currentGIF == gifLimit-1 {
		go g.loadGifs()
	}

	g.currentMux.Unlock()
	return newGif
}

// SendGif - Sending random tenor gif to chat
func (g *Gif) SendGif(msg *tgbotapi.Message) {
	newGif := g.getCachedGif()

	sended, err := g.c.SendGif(msg, newGif, "random")
	if err != nil {
		log.Println(err)
	}

	fmt.Println(sended)
}

func (g *Gif) SendExistingGif(msg *tgbotapi.Message) {
	gifCount, err := g.db.GifsStorage.CountAllGifs(context.Background())
	if err != nil {
		g.errors <- errors.Wrap(err, "cannot get gifs count")
		return
	}
	randGif := rand.Intn(gifCount)

	gifToSend, err := g.db.GifsStorage.GetGifWithOffset(context.Background(), randGif)

	if err != nil {
		g.errors <- errors.Wrap(err, "cannot get gif with offset")
		return
	}

	_, err = g.c.SendExistingGif(msg, gifToSend.Gif)
	if err != nil {
		g.errors <- errors.Wrap(err, "cannot get gif with offset")
		return
	}
}

func (g *Gif) AddGifIfNeed(msg *tgbotapi.Message) {
	err := g.db.GifsStorage.InsertGif(context.Background(), msg.Animation.FileID)
	if err != nil {
		g.errors <- errors.Wrap(err, "cannot add gif to database")
		return
	}
}
