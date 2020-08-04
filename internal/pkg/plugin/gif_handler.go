package plugin

import (
	"context"
	"regexp"
	"strconv"
	"sync"

	"github.com/LikiPiki/UwdBot/internal/pkg/database"
	"github.com/LikiPiki/UwdBot/internal/pkg/sender"
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api"
	"github.com/pkg/errors"
)

// Gif plugin to work with Giphy API
type Gif struct {
	c          *sender.Sender
	errors     chan error
	db         *database.Database
	tenorToken string
	// For gif caching
	gifs       []string
	currentGIF int
	currentMux sync.Mutex
	gifLoading bool
}

// Plugin initialization
func (g *Gif) Init(s *sender.Sender, db *database.Database) {
	g.c = s
	g.db = db
	g.errors = make(chan error)
}

// Handle messages (not commands, like regex queries)
func (g *Gif) HandleMessages(msg *tgbotapi.Message) {
	if msg.Animation != nil {
		g.AddGifIfNeed(context.Background(), msg)
		return
	}

	// send a batch of gifs
	re := regexp.MustCompile(`[G|g]if (\d+)`)
	matches := re.FindStringSubmatch(msg.Text)
	if len(matches) == 2 {
		count, err := strconv.Atoi(matches[1])

		if err != nil {
			g.errors <- errors.Wrap(err, "cannot parse many gifs regexp")
			return
		}

		go g.SendManyGifs(context.Background(), msg, count)
	}
}

// Not register, simple commands
func (g *Gif) HandleCommands(msg *tgbotapi.Message, command string) {
	// send single gif
	if command == "gif" {
		go g.SendExistingGif(context.Background(), msg)
		return
	}
}

// Commands for registered users only
func (g *Gif) HandleRegisterCommands(msg *tgbotapi.Message, command string, user *database.User) {}

// Callbacks from keyboard
func (g *Gif) HandleCallbackQuery(update *tgbotapi.Update) {}

// Commands for admin only
func (g *Gif) HandleAdminCommands(msg *tgbotapi.Message) {
	if (msg.Text == "del") && (msg.ReplyToMessage.Animation != nil) {
		go g.DeleteGif(context.Background(), msg, msg.ReplyToMessage.Animation.FileID)
	}
}

// Handle inline commands
func (g *Gif) HandleInlineCommands(update *tgbotapi.Update) {
}

// Get all plugin existing commands
func (g *Gif) GetRegisteredCommands() []string {
	return []string{}
}

// Returns errors from plugin
func (g *Gif) Errors() <-chan error {
	return g.errors
}
