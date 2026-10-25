package objects

import rl "github.com/gen2brain/raylib-go/raylib"

const interactionPadding = 6

type Door struct {
	ID       string
	Position rl.Vector2
	Size     rl.Vector2
	Bounds   rl.Rectangle
	Color    rl.Color
	Open     bool
	Locked   bool
}

type Seat struct {
	ID       string
	Position rl.Vector2
	Size     rl.Vector2
	Bounds   rl.Rectangle
	Color    rl.Color
	Occupied bool
}
