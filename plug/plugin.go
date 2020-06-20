package plug

import (
	data "UwdBot/database"
	"UwdBot/sender"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api"
)

// Plugin - interface, which provide independent bot component
type Plugin interface {
	// Plugin initialization
	Init(s *sender.Sender)
	// Not register, simple commands
	HandleCommands(msg *tgbotapi.Message, command string)
	// Commands for registered users only
	HandleRegisterCommands(msg *tgbotapi.Message, command string, user *data.User)
	// Callbacks from keyboard
	HandleCallbackQuery(update *tgbotapi.Update)
	// Commands for admin only
	HandleAdminCommands(msg *tgbotapi.Message)
	// Get all plugin existing commands
	GetRegisteredCommands() []string
}

// Plugins is an array of Plugin interfaces
type Plugins []Plugin
