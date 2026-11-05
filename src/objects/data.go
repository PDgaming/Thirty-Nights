package objects

import rl "github.com/gen2brain/raylib-go/raylib"

const interactionPadding = 6

// pixelDimension guards object constructors against empty or inverted
// rectangles. The loader and map validation are the real guards; this keeps
// direct callers from producing degenerate geometry.
func pixelDimension(value int) float32 {
	if value <= 0 {
		return 1
	}
	return float32(value)
}

type Door struct {
	ID           string
	Bounds       rl.Rectangle
	Color        rl.Color
	Texture      rl.Texture2D
	OpenTexture  rl.Texture2D
	Open         bool
	Locked       bool
}

type Seat struct {
	ID       string
	Bounds   rl.Rectangle
	Color    rl.Color
	Texture  rl.Texture2D
	Occupied bool
}

type Trigger struct {
	ID         string
	Bounds     rl.Rectangle
	PromptText string

	dialogueID string
}
