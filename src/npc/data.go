package npc

import rl "github.com/gen2brain/raylib-go/raylib"

const interactionPadding = 6

type NPC struct {
	ID     string
	Bounds rl.Rectangle
	Color  rl.Color
	Name   string
	Facing rl.Vector2

	dialogueID         string
	solid              bool
	interactionPadding int
}
