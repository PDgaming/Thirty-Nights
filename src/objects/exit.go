package objects

import (
	"fmt"

	rl "github.com/gen2brain/raylib-go/raylib"
)

func NewExit(id string, x, y, width, height int, properties map[string]any) (*Exit, error) {
	e := &Exit{
		ID: id,
		Bounds: rl.Rectangle{
			X:      float32(x),
			Y:      float32(y),
			Width:  pixelDimension(width),
			Height: pixelDimension(height),
		},
	}

	raw, ok := properties["nextLevel"]
	if !ok || raw == nil {
		return nil, fmt.Errorf("exit %q: no \"nextLevel\" in properties", id)
	}
	nextLevel, ok := raw.(string)
	if !ok || nextLevel == "" {
		return nil, fmt.Errorf("exit %q: \"nextLevel\" must be a non-empty string", id)
	}
	e.nextLevel = nextLevel
	return e, nil
}

func (e *Exit) NextLevel() string {
	return e.nextLevel
}

func (e *Exit) Update(dt float32) {}

func (e *Exit) Draw(debug bool) {
	if !debug {
		return
	}
	rl.DrawRectangleLinesEx(e.Bounds, 1, rl.SkyBlue)
}

func (e *Exit) CollisionRect() rl.Rectangle {
	return e.Bounds
}

func (e *Exit) Solid() bool {
	return false
}
