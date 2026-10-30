package game

import (
	"thirty.nights/src/camera"
	"thirty.nights/src/dialogue"
	"thirty.nights/src/player"
	"thirty.nights/src/world"
)

type Mode int

const (
	ModePlaying Mode = iota
	ModeDialogue
)

type Game struct {
	Levels       world.Catalog
	currentLevel string

	Map    *world.MapData
	Player *player.Player
	Cam    *camera.Controller

	Dialogues dialogue.Catalog
	Dialogue  *dialogue.Session

	ActiveInteractable world.Interactable

	mode            Mode
	dialogueReturn  Mode
	pendingDialogue string

	Debug bool
}

type Input struct {
	Interact bool
}
