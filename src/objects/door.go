package objects

import rl "github.com/gen2brain/raylib-go/raylib"

func NewDoor(id string, x, y, width, height int, color rl.Color, properties map[string]any) *Door {
	d := &Door{
		ID: id,
		Bounds: rl.Rectangle{
			X:      float32(x),
			Y:      float32(y),
			Width:  pixelDimension(width),
			Height: pixelDimension(height),
		},
		Color: color,
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
	if d.Texture.ID != 0 {
		src := rl.Rectangle{X: 0, Y: 0, Width: float32(d.Texture.Width), Height: float32(d.Texture.Height)}
		dst := d.Bounds
		rl.DrawTexturePro(d.Texture, src, dst, rl.Vector2{}, 0, rl.White)
	} else {
		rl.DrawRectangleRec(d.Bounds, d.Color)
	}

	if debug {
		rl.DrawRectangleLinesEx(d.Bounds, 1, rl.Black)
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
