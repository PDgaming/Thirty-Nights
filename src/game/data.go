package game

import (
	"thirty.nights/src/camera"
	"thirty.nights/src/cutscene"
	"thirty.nights/src/dialogue"
	"thirty.nights/src/player"
	"thirty.nights/src/world"
)

type Mode int

const (
	ModePlaying Mode = iota
	ModeDialogue
	ModeCutscene
)

type Game struct {
	Levels       world.Catalog
	currentLevel string

	Map    *world.MapData
	Player *player.Player
	Cam    *camera.Controller

	Dialogues dialogue.Catalog
	Cutscenes cutscene.Catalog
	Dialogue  *dialogue.Session
	Cutscene  *cutscene.Session

	ActiveInteractable world.Interactable

	mode            Mode
	dialogueReturn  Mode
	pendingDialogue string
	pendingCutscene string

	transitionPhase   TransitionPhase
	transitionAlpha   float32
	transitionElapsed float32
	transitionTarget  string

	Debug bool
}

type Input struct {
	Interact bool
}
