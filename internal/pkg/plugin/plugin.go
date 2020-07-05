package plugin

import (
	"github.com/LikiPiki/UwdBot/internal/pkg/database"
	"github.com/LikiPiki/UwdBot/internal/pkg/sender"
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api"
)

// Plugin - interface, which provide independent bot component
type Plugin interface {
	// Plugin initialization
	Init(s *sender.Sender, db *database.Database)
	// Handle messages (not commands, like regex queries)
	HandleMessages(msg *tgbotapi.Message)
	// Not register, simple commands
	HandleCommands(msg *tgbotapi.Message, command string)
	// Commands for registered users only
	HandleRegisterCommands(msg *tgbotapi.Message, command string, user *database.User)
	// Callbacks from keyboard
	HandleCallbackQuery(update *tgbotapi.Update)
	// Commands for admin only
	HandleAdminCommands(msg *tgbotapi.Message)
	// Get all plugin existing commands
	GetRegisteredCommands() []string
	// Returns errors from plugin
	Errors() <-chan error
}

// Plugins is an array of Plugin interfaces
type Plugins []Plugin
