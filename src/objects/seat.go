package objects

import rl "github.com/gen2brain/raylib-go/raylib"

func NewSeat(id string, x, y, width, height int, color rl.Color, properties map[string]any) *Seat {
	s := &Seat{
		ID: id,
		Bounds: rl.Rectangle{
			X:      float32(x),
			Y:      float32(y),
			Width:  pixelDimension(width),
			Height: pixelDimension(height),
		},
		Color: color,
	}
	if occupied, ok := properties["occupied"].(bool); ok {
		s.Occupied = occupied
	}

	return s
}

func (s *Seat) Update(dt float32) {}

func (s *Seat) Draw(debug bool) {
	if s.Occupied {
		rl.DrawRectangleRec(s.Bounds, rl.Red)
	} else if s.Texture.ID != 0 {
		src := rl.Rectangle{X: 0, Y: 0, Width: float32(s.Texture.Width), Height: float32(s.Texture.Height)}
		dst := s.Bounds
		rl.DrawTexturePro(s.Texture, src, dst, rl.Vector2{}, 0, rl.White)
	} else {
		rl.DrawRectangleRec(s.Bounds, s.Color)
	}

	if debug {
		rl.DrawRectangleLinesEx(s.Bounds, 1, rl.Black)
	}
}

func (s *Seat) CanInteractRect(rect rl.Rectangle) bool {
	return rl.CheckCollisionRecs(rect, s.InteractionRect())
}

func (s *Seat) CanActivate(playerRect rl.Rectangle) bool {
	return rl.CheckCollisionRecs(playerRect, s.InteractionRect())
}

func (s *Seat) Interact() {
	s.Occupied = !s.Occupied
}

func (s *Seat) PromptFor(playerRect rl.Rectangle) string {
	if rl.CheckCollisionRecs(playerRect, s.Bounds) {
		return "E - Leave"
	}
	if s.Occupied {
		return "Occupied"
	}
	return "E - Sit"
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
