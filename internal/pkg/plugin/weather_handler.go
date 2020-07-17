package plugin

import (
	"os"
	"time"

	"github.com/LikiPiki/UwdBot/internal/pkg/database"
	"github.com/LikiPiki/UwdBot/internal/pkg/sender"
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api"
	"github.com/pkg/errors"
)

// Weather inline plugin to show current weather
type Weather struct {
	c              *sender.Sender
	db             *database.Database
	errors         chan error
	weatherTOKEN   string
	replies        map[string]WeatherInlineReply
	lastUpdateTime time.Time
}

// Plugin initialization
func (w *Weather) Init(s *sender.Sender, db *database.Database) {
	w.c = s
	w.db = db

	token, exists := os.LookupEnv("WEATHER_TOKEN")
	if !exists {
		panic("cannot read weather TOKEN, add TOKEN or disable this plugin")
	}
	w.weatherTOKEN = token
	w.replies = make(map[string]WeatherInlineReply)

	weatherResp := WeatherAPIResponse{}
	err := w.updateWeather(&weatherResp)
	if err != nil {
		w.errors <- errors.Wrap(err, "cannot update weather")
	}
}

// Handle messages (not commands, like regex queries)
func (w *Weather) HandleMessages(msg *tgbotapi.Message) {}

// Not register, simple commands
func (w *Weather) HandleCommands(msg *tgbotapi.Message, command string) {}

// Commands for registered users only
func (w *Weather) HandleRegisterCommands(msg *tgbotapi.Message, command string, user *database.User) {
}

// Callbacks from keyboard
func (w *Weather) HandleCallbackQuery(update *tgbotapi.Update) {}

// Commands for admin only
func (w *Weather) HandleAdminCommands(msg *tgbotapi.Message) {}

// Handle inline commands
func (w *Weather) HandleInlineCommands(update *tgbotapi.Update) {
	w.HandleWeatherInline(update)
}

// Get all plugin existing commands
func (w *Weather) GetRegisteredCommands() []string {
	return []string{}
}

// Returns errors from plugin
func (w *Weather) Errors() <-chan error {
	return w.errors
}
