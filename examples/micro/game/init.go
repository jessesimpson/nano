package game

import "github.com/lonng/nano/component"

var Services = &component.Components{}

var gameService = newGameService()

func init() {
	Services.Register(gameService)
}
