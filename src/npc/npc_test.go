package npc

import (
	"strings"
	"testing"

	rl "github.com/gen2brain/raylib-go/raylib"
)

func assertRectangle(t *testing.T, name string, got, want rl.Rectangle) {
	t.Helper()
	if got != want {
		t.Fatalf("%s = %+v, want %+v", name, got, want)
	}
}

func newTestNPC(t *testing.T, width, height, tileSize int, properties map[string]any) *NPC {
	t.Helper()
	n, err := NewNPC("npc", 10, 20, width, height, tileSize, rl.White, properties)
	if err != nil {
		t.Fatalf("NewNPC() error = %v", err)
	}
	return n
}

func TestNewNPCBounds(t *testing.T) {
	tests := []struct {
		name          string
		width, height int
	}{
		{name: "one tile", width: 16, height: 16},
		{name: "two by two tiles", width: 32, height: 32},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			n := newTestNPC(t, tt.width, tt.height, 16, map[string]any{"dialogueId": "intro"})

			want := rl.Rectangle{
				X:      10,
				Y:      20,
				Width:  float32(tt.width),
				Height: float32(tt.height),
			}
			assertRectangle(t, "Bounds", n.Bounds, want)
			assertRectangle(t, "CollisionRect()", n.CollisionRect(), want)
		})
	}
}

func TestNewNPCClampsNonPositiveDimensions(t *testing.T) {
	n := newTestNPC(t, 0, -8, 16, map[string]any{"dialogueId": "intro"})

	assertRectangle(t, "Bounds", n.Bounds, rl.Rectangle{X: 10, Y: 20, Width: 1, Height: 1})
}

func TestNPCInteractionPadding(t *testing.T) {
	tests := []struct {
		name       string
		properties map[string]any
		tileSize   int
		wantPad    float32
	}{
		{
			name:       "absent padding uses six pixel default",
			properties: map[string]any{"dialogueId": "intro"},
			tileSize:   16,
			wantPad:    interactionPadding,
		},
		{
			name:       "one tile padding",
			properties: map[string]any{"dialogueId": "intro", "interactionPadding": float64(1)},
			tileSize:   16,
			wantPad:    16,
		},
		{
			name:       "two tile padding",
			properties: map[string]any{"dialogueId": "intro", "interactionPadding": float64(2)},
			tileSize:   16,
			wantPad:    32,
		},
		{
			name:       "padding scales with tile size",
			properties: map[string]any{"dialogueId": "intro", "interactionPadding": float64(1)},
			tileSize:   32,
			wantPad:    32,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			n := newTestNPC(t, 16, 16, tt.tileSize, tt.properties)

			want := rl.Rectangle{
				X:      10 - tt.wantPad,
				Y:      20 - tt.wantPad,
				Width:  16 + tt.wantPad*2,
				Height: 16 + tt.wantPad*2,
			}
			assertRectangle(t, "InteractionRect()", n.InteractionRect(), want)
		})
	}
}

func TestNewNPCRequiresDialogueID(t *testing.T) {
	tests := []struct {
		name       string
		properties map[string]any
		contains   string
	}{
		{
			name:       "missing",
			properties: map[string]any{},
			contains:   "no \"dialogueId\"",
		},
		{
			name:       "nil value",
			properties: map[string]any{"dialogueId": nil},
			contains:   "no \"dialogueId\"",
		},
		{
			name:       "empty",
			properties: map[string]any{"dialogueId": ""},
			contains:   "must not be empty",
		},
		{
			name:       "wrong type",
			properties: map[string]any{"dialogueId": 42},
			contains:   "must be a string",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := NewNPC("npc", 0, 0, 16, 16, 16, rl.White, tt.properties)
			if err == nil {
				t.Fatal("NewNPC() error = nil, want an error")
			}
			if !strings.Contains(err.Error(), tt.contains) {
				t.Fatalf("NewNPC() error = %q, want it to contain %q", err, tt.contains)
			}
		})
	}
}

func TestNewNPCFacing(t *testing.T) {
	tests := []struct {
		name     string
		facing   any
		wantAxis rl.Vector2
	}{
		{name: "absent defaults down", facing: nil, wantAxis: rl.Vector2{X: 0, Y: 1}},
		{name: "up", facing: "up", wantAxis: rl.Vector2{X: 0, Y: -1}},
		{name: "left", facing: "left", wantAxis: rl.Vector2{X: -1, Y: 0}},
		{name: "right", facing: "right", wantAxis: rl.Vector2{X: 1, Y: 0}},
		{name: "invalid defaults down", facing: "diagonal", wantAxis: rl.Vector2{X: 0, Y: 1}},
		{name: "wrong type defaults down", facing: 3, wantAxis: rl.Vector2{X: 0, Y: 1}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			properties := map[string]any{"dialogueId": "intro"}
			if tt.facing != nil {
				properties["facing"] = tt.facing
			}
			n := newTestNPC(t, 16, 16, 16, properties)

			if n.Facing != tt.wantAxis {
				t.Fatalf("Facing = %+v, want %+v", n.Facing, tt.wantAxis)
			}
		})
	}
}

func TestNewNPCSolidDefaultsTrue(t *testing.T) {
	tests := []struct {
		name       string
		properties map[string]any
		want       bool
	}{
		{
			name:       "default",
			properties: map[string]any{"dialogueId": "intro"},
			want:       true,
		},
		{
			name:       "explicit false",
			properties: map[string]any{"dialogueId": "intro", "solid": false},
			want:       false,
		},
		{
			name:       "explicit true",
			properties: map[string]any{"dialogueId": "intro", "solid": true},
			want:       true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			n := newTestNPC(t, 16, 16, 16, tt.properties)
			if got := n.Solid(); got != tt.want {
				t.Fatalf("Solid() = %v, want %v", got, tt.want)
			}
		})
	}
}
