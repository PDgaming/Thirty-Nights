package game

import (
	"thirty.nights/src/camera"
	"thirty.nights/src/player"
	"thirty.nights/src/world"
)

type Game struct {
	level  int
	Map    *world.MapData
	Player *player.Player
	Cam    *camera.Controller
	Debug  bool
}
