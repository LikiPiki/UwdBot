package app

import (
	pl "github.com/LikiPiki/UwdBot/internal/pkg/plugin"
)

type App struct {
	Plugs pl.Plugins
}

func NewApp(pl pl.Plugins) *App {
	return &App{Plugs: pl}
}
