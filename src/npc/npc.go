package npc

import (
	"fmt"

	rl "github.com/gen2brain/raylib-go/raylib"
)

var DefaultColor = rl.NewColor(232, 168, 124, 255)

const defaultPrompt = "E - Talk"

// pixelDimension guards the constructor against empty or inverted rectangles.
// The loader and map validation are the real guards; this keeps direct callers
// from producing degenerate geometry.
func pixelDimension(value int) float32 {
	if value <= 0 {
		return 1
	}
	return float32(value)
}

func NewNPC(id string, x, y, width, height, tileSize int, color rl.Color, properties map[string]any) (*NPC, error) {
	n := &NPC{
		ID: id,
		Bounds: rl.Rectangle{
			X:      float32(x),
			Y:      float32(y),
			Width:  pixelDimension(width),
			Height: pixelDimension(height),
		},
		Color:  color,
		Facing: rl.Vector2{X: 0, Y: 1}, // down
		Name:   id,
		solid:  true,
	}

	raw, ok := properties["dialogueId"]
	if !ok || raw == nil {
		return nil, fmt.Errorf("npc %q: no \"dialogueId\" in properties", id)
	}

	dialogueID, ok := raw.(string)
	if !ok {
		return nil, fmt.Errorf("npc %q: \"dialogueId\" must be a string, got %T", id, raw)
	}
	if dialogueID == "" {
		return nil, fmt.Errorf("npc %q: \"dialogueId\" must not be empty", id)
	}
	n.dialogueID = dialogueID

	if name, ok := properties["name"].(string); ok && name != "" {
		n.Name = name
	}

	if facing, ok := properties["facing"].(string); ok {
		n.Facing = parseFacing(facing)
	}

	if solid, ok := properties["solid"].(bool); ok {
		n.solid = solid
	}

	if ip, ok := properties["interactionPadding"].(float64); ok {
		n.interactionPadding = int(ip) * tileSize
	}

	return n, nil
}

func parseFacing(name string) rl.Vector2 {
	switch name {
	case "up":
		return rl.Vector2{X: 0, Y: -1}
	case "left":
		return rl.Vector2{X: -1, Y: 0}
	case "right":
		return rl.Vector2{X: 1, Y: 0}
	default:
		return rl.Vector2{X: 0, Y: 1}
	}
}

func (n *NPC) Update(dt float32) {
}

func (n *NPC) Draw(debug bool) {
	rl.DrawRectangleRec(n.Bounds, n.Color)
	n.drawFacingMarker()

	if debug {
		rl.DrawRectangleLinesEx(n.Bounds, 1, rl.Magenta)
	}
}

func (n *NPC) drawFacingMarker() {
	const size = float32(4)
	const inset = float32(2)

	cx := n.Bounds.X + n.Bounds.Width/2 - size/2
	cy := n.Bounds.Y + n.Bounds.Height/2 - size/2

	var marker rl.Rectangle
	switch {
	case n.Facing.Y < 0: // up
		marker = rl.Rectangle{X: cx, Y: n.Bounds.Y + inset, Width: size, Height: size}
	case n.Facing.X < 0: // left
		marker = rl.Rectangle{X: n.Bounds.X + inset, Y: cy, Width: size, Height: size}
	case n.Facing.X > 0: // right
		marker = rl.Rectangle{X: n.Bounds.X + n.Bounds.Width - size - inset, Y: cy, Width: size, Height: size}
	default: // down
		marker = rl.Rectangle{X: cx, Y: n.Bounds.Y + n.Bounds.Height - size - inset, Width: size, Height: size}
	}

	rl.DrawRectangleRec(marker, rl.ColorAlpha(rl.Black, 0.35))
}

func (n *NPC) CollisionRect() rl.Rectangle {
	return n.Bounds
}

func (n *NPC) Solid() bool {
	return n.solid
}

func (n *NPC) CanInteractRect(rect rl.Rectangle) bool {
	return rl.CheckCollisionRecs(rect, n.InteractionRect())
}

func (n *NPC) CanActivate(_ rl.Rectangle) bool {
	return true
}

func (n *NPC) Interact() {}

func (n *NPC) Prompt() string {
	return n.promptText()
}

func (n *NPC) PromptFor(_ rl.Rectangle) string {
	return n.promptText()
}

func (n *NPC) promptText() string {
	if n.Name == "" {
		return defaultPrompt
	}
	return "E - Talk to " + n.Name
}

func (n *NPC) PromptAnchor() rl.Vector2 {
	return rl.Vector2{X: n.Bounds.X + n.Bounds.Width/2, Y: n.Bounds.Y - 6}
}

func (n *NPC) InteractionRect() rl.Rectangle {
	pad := float32(n.interactionPadding)
	if pad <= 0 {
		pad = float32(interactionPadding)
	}
	return rl.Rectangle{
		X:      n.Bounds.X - pad,
		Y:      n.Bounds.Y - pad,
		Width:  n.Bounds.Width + pad*2,
		Height: n.Bounds.Height + pad*2,
	}
}

func (n *NPC) DialogueID() string {
	return n.dialogueID
}
