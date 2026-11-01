package npc

import rl "github.com/gen2brain/raylib-go/raylib"

type NPC struct {
	ID       string
	Position rl.Vector2
	Size     rl.Vector2
	Bounds   rl.Rectangle
	Color    rl.Color
	Name     string
	Facing   rl.Vector2

	dialogueID string
	solid      bool
}
