package objects

import rl "github.com/gen2brain/raylib-go/raylib"

func NewSeat(id string, x, y, size int, color rl.Color, properties map[string]any) *Seat {
	s := &Seat{
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
	if occupied, ok := properties["occupied"].(bool); ok {
		s.Occupied = occupied
	}

	return s
}

func (s *Seat) Update(dt float32) {}

func (s *Seat) Draw(debug bool) {
	rl.DrawRectangle(int32(s.Position.X), int32(s.Position.Y), int32(s.Size.X), int32(s.Size.Y), s.Color)

	if debug {
		rl.DrawRectangleLines(int32(s.Position.X), int32(s.Position.Y), int32(s.Size.X), int32(s.Size.Y), rl.Black)
	}
}

func (s *Seat) CanInteractRect(rect rl.Rectangle) bool {
	return rl.CheckCollisionRecs(rect, s.InteractionRect())
}

func (s *Seat) Interact() {
	s.Occupied = !s.Occupied
}

func (s *Seat) Prompt() string {
	if s.Occupied {
		return "E - Leave"
	}
	return "E - Sit"
}

func (s *Seat) PromptAnchor() rl.Vector2 {
	return rl.Vector2{X: float32(s.Bounds.X) + s.Bounds.Width/2, Y: float32(s.Bounds.Y) - 6}
}

func (s *Seat) CollisionRect() rl.Rectangle {
	return s.Bounds
}

func (s *Seat) InteractionRect() rl.Rectangle {
	return rl.Rectangle{
		X:      s.Bounds.X - interactionPadding,
		Y:      s.Bounds.Y - interactionPadding,
		Width:  s.Bounds.Width + interactionPadding*2,
		Height: s.Bounds.Height + interactionPadding*2,
	}
}

func (s *Seat) Solid() bool {
	return true
}
