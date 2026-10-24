package objects

import rl "github.com/gen2brain/raylib-go/raylib"

func NewDoor(id string, x, y, size int, color rl.Color, properties map[string]any) *Door {
	d := &Door{
		ID: id,
		Position: rl.Vector2{
			X: float32(x),
			Y: float32(y),
		},
		Size: rl.Vector2{
			X: float32(size),
			Y: float32(size),
		},
		Color: color,
		Bounds: rl.Rectangle{
			X:      float32(x),
			Y:      float32(y),
			Width:  float32(size),
			Height: float32(size),
		},
	}

	if open, ok := properties["open"].(bool); ok {
		d.Open = open
	}
	if locked, ok := properties["locked"].(bool); ok {
		d.Locked = locked
	}

	return d
}

func (d *Door) Update(dt float32) {
}

func (d *Door) Draw(debug bool) {
	rl.DrawRectangle(int32(d.Position.X), int32(d.Position.Y), int32(d.Size.X), int32(d.Size.Y), d.Color)

	if debug {
		rl.DrawRectangleLines(int32(d.Position.X), int32(d.Position.Y), int32(d.Size.X), int32(d.Size.Y), rl.Black)
	}
}

func (d *Door) CanInteractRect(rect rl.Rectangle) bool {
	return rl.CheckCollisionRecs(rect, d.InteractionRect())
}

func (d *Door) Interact() {
	d.Open = !d.Open
}

func (d *Door) Prompt() string {
	return "E"
}

// Helper functions
func (d *Door) CollisionRect() rl.Rectangle {
	return rl.Rectangle{
		X:      d.Position.X,
		Y:      d.Position.Y,
		Width:  d.Size.X,
		Height: d.Size.Y,
	}
}
func (d *Door) InteractionRect() rl.Rectangle {
	return rl.Rectangle{
		X:      d.Bounds.X - interactionPadding,
		Y:      d.Bounds.Y - interactionPadding,
		Width:  d.Bounds.Width + interactionPadding*2,
		Height: d.Bounds.Height + interactionPadding*2,
	}
}

func (d *Door) Solid() bool {
	return !d.Open
}
