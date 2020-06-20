package main

import (
	data "UwdBot/database"
	pl "UwdBot/plug"
)

type App struct {
	Plugs pl.Plugins
}

func InitApp(plugs pl.Plugins) *App {
	return &App{
		Plugs: plugs,
	}
}

func (a *App) IsAdmin(ID int) bool {
	var err error
	var user data.User
	user, err = user.FindUserByID(ID)

	if err != nil {
		return false
	}

	return user.IsAdmin
}
