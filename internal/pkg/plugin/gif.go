package plugin

import (
	"context"
	"math/rand"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api"
	"github.com/pkg/errors"
)

const (
	gifLimit = 500
)

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
	ctx := context.Background()

	gifCount, err := g.db.GifsStorage.CountAllGifs(ctx)
	if err != nil {
		g.errors <- errors.Wrap(err, "cannot count gifs")
		return
	}

	randGif := rand.Intn(gifCount)
	gifToReplace, err := g.db.GifsStorage.GetGifWithOffset(context.Background(), randGif)
	if err != nil {
		g.errors <- errors.Wrap(err, "cannot get gif to replace")
		return
	}

	if gifCount >= gifLimit {
		err := g.db.GifsStorage.ReplaceGifByID(ctx, gifToReplace.ID, msg.Animation.FileID)
		if err != nil {
			g.errors <- errors.Wrap(err, "cannot replace gif to another")
		}
		return
	}

	err = g.db.GifsStorage.InsertGif(ctx, msg.Animation.FileID)
	if err != nil {
		g.errors <- errors.Wrap(err, "cannot add gif to database")
		return
	}
}

func (g *Gif) DeleteGif(msg *tgbotapi.Message, fileID string) {
	// Deleting gif from postgre
	err := g.db.GifsStorage.DeleteGifByFileID(context.Background(), fileID)
	if err != nil {
		g.errors <- errors.Wrap(err, "cannot delete this GIF")
		return
	}

	err = g.c.SendMarkdownReply(msg, "Капитан, гифка удалена!")
	if err != nil {
		g.errors <- errors.Wrap(err, "cannot send GIF deleted msg")
		return
	}

	//Deleting gif from chat
	err = g.c.DeleteMessage(msg.ReplyToMessage)
	if err != nil {
		g.errors <- errors.Wrap(err, "cannot delete gif message")
		return
	}
}
