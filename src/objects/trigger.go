package objects

import (
	"fmt"

	rl "github.com/gen2brain/raylib-go/raylib"
)

const DefaultTriggerPrompt = "E - Look"

func NewTrigger(id string, x, y, width, height int, properties map[string]any) (*Trigger, error) {
	t := &Trigger{
		ID: id,
		Bounds: rl.Rectangle{
			X:      float32(x),
			Y:      float32(y),
			Width:  pixelDimension(width),
			Height: pixelDimension(height),
		},
		PromptText: DefaultTriggerPrompt,
	}

	raw, ok := properties["dialogueId"]
	if !ok || raw == nil {
		return nil, fmt.Errorf("trigger %q: no \"dialogueId\" in properties", id)
	}

	dialogueID, ok := raw.(string)
	if !ok {
		return nil, fmt.Errorf("trigger %q: \"dialogueId\" must be a string, got %T", id, raw)
	}
	if dialogueID == "" {
		return nil, fmt.Errorf("trigger %q: \"dialogueId\" must not be empty", id)
	}

	t.dialogueID = dialogueID

	if prompt, ok := properties["prompt"].(string); ok && prompt != "" {
		t.PromptText = prompt
	}

	return t, nil
}

func (t *Trigger) DialogueID() string {
	return t.dialogueID
}

func (t *Trigger) Update(dt float32) {
}

func (t *Trigger) Draw(debug bool) {
	if !debug {
		return
	}
	rl.DrawRectangleLinesEx(t.Bounds, 1, rl.Magenta)
}

func (t *Trigger) CanInteractRect(rect rl.Rectangle) bool {
	return rl.CheckCollisionRecs(rect, t.InteractionRect())
}

func (t *Trigger) CanActivate(playerRect rl.Rectangle) bool {
	return rl.CheckCollisionRecs(playerRect, t.InteractionRect())
}

func (t *Trigger) Interact() {
}

func (t *Trigger) Prompt() string {
	return t.PromptText
}

func (t *Trigger) PromptFor(playerRect rl.Rectangle) string {
	return t.PromptText
}

func (t *Trigger) PromptAnchor() rl.Vector2 {
	return rl.Vector2{X: float32(t.Bounds.X) + t.Bounds.Width/2, Y: float32(t.Bounds.Y) - 6}
}

func (t *Trigger) CollisionRect() rl.Rectangle {
	return t.Bounds
}

func (t *Trigger) InteractionRect() rl.Rectangle {
	return rl.Rectangle{
		X:      t.Bounds.X - interactionPadding,
		Y:      t.Bounds.Y - interactionPadding,
		Width:  t.Bounds.Width + interactionPadding*2,
		Height: t.Bounds.Height + interactionPadding*2,
	}
}

func (t *Trigger) Solid() bool {
	return false
}
