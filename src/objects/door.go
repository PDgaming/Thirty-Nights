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

func (d *Door) CanActivate(playerRect rl.Rectangle) bool {
	if d.Locked {
		return false
	}

	if d.Open && rl.CheckCollisionRecs(playerRect, d.CollisionRect()) {
		return false
	}

	return true
}

func (d *Door) Interact() {
	if d.Locked {
		return
	}
	d.Open = !d.Open
}

func (d *Door) PromptFor(playerRect rl.Rectangle) string {
	if d.Locked {
		return "Locked"
	}

	if d.Open {
		if rl.CheckCollisionRecs(playerRect, d.CollisionRect()) {
			return ""
		}
		return "E - Close"
	}

	return "E - Open"
}

func (d *Door) Prompt() string {
	if d.Locked {
		return "Locked"
	}
	if d.Open {
		return "E - Close"
	}
	return "E - Open"
}

func (d *Door) PromptAnchor() rl.Vector2 {
	return rl.Vector2{X: float32(d.Bounds.X) + d.Bounds.Width/2, Y: float32(d.Bounds.Y) - 6}
}

func (d *Door) CollisionRect() rl.Rectangle {
	return d.Bounds
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
