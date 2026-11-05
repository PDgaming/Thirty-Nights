package objects

import (
	"testing"

	rl "github.com/gen2brain/raylib-go/raylib"
)

func assertRectangle(t *testing.T, name string, got, want rl.Rectangle) {
	t.Helper()
	if got != want {
		t.Fatalf("%s = %+v, want %+v", name, got, want)
	}
}

func TestDoorGeometry(t *testing.T) {
	tests := []struct {
		name          string
		width, height int
	}{
		{name: "square", width: 16, height: 16},
		{name: "wide", width: 32, height: 16},
		{name: "tall", width: 16, height: 32},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			d := NewDoor("door", 10, 20, tt.width, tt.height, rl.White, nil)

			wantBounds := rl.Rectangle{
				X:      10,
				Y:      20,
				Width:  float32(tt.width),
				Height: float32(tt.height),
			}
			assertRectangle(t, "Bounds", d.Bounds, wantBounds)
			assertRectangle(t, "CollisionRect()", d.CollisionRect(), wantBounds)

			wantInteraction := rl.Rectangle{
				X:      10 - interactionPadding,
				Y:      20 - interactionPadding,
				Width:  float32(tt.width) + interactionPadding*2,
				Height: float32(tt.height) + interactionPadding*2,
			}
			assertRectangle(t, "InteractionRect()", d.InteractionRect(), wantInteraction)
		})
	}
}

func TestDoorGeometryClampsNonPositiveDimensions(t *testing.T) {
	tests := []struct {
		name          string
		width, height int
	}{
		{name: "zero", width: 0, height: 0},
		{name: "negative", width: -3, height: -7},
		{name: "mixed", width: 0, height: 16},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			d := NewDoor("door", 4, 6, tt.width, tt.height, rl.White, nil)

			wantWidth := float32(tt.width)
			if tt.width <= 0 {
				wantWidth = 1
			}
			wantHeight := float32(tt.height)
			if tt.height <= 0 {
				wantHeight = 1
			}

			assertRectangle(t, "Bounds", d.Bounds, rl.Rectangle{
				X:      4,
				Y:      6,
				Width:  wantWidth,
				Height: wantHeight,
			})
		})
	}
}

func TestDoorOpenCloseChangesSolidity(t *testing.T) {
	d := NewDoor("door", 0, 0, 32, 16, rl.White, nil)

	if !d.Solid() {
		t.Fatal("new door should be solid")
	}

	d.Interact()
	if !d.Open {
		t.Fatal("Interact() should open the door")
	}
	if d.Solid() {
		t.Fatal("open door should not be solid")
	}

	d.Interact()
	if d.Open {
		t.Fatal("Interact() should close the door again")
	}
	if !d.Solid() {
		t.Fatal("closed door should be solid")
	}
}

func TestDoorCannotCloseWhilePlayerOverlaps(t *testing.T) {
	d := NewDoor("door", 10, 20, 32, 16, rl.White, map[string]any{"open": true})

	player := rl.Rectangle{X: 10, Y: 20, Width: 16, Height: 16}
	if d.CanActivate(player) {
		t.Fatal("open door should not close while the player overlaps it")
	}
	if got := d.PromptFor(player); got != "" {
		t.Fatalf("PromptFor(overlapping player) = %q, want empty", got)
	}

	away := rl.Rectangle{X: 100, Y: 100, Width: 16, Height: 16}
	if !d.CanActivate(away) {
		t.Fatal("open door should close when the player is clear of it")
	}
	if got := d.PromptFor(away); got != "E - Close" {
		t.Fatalf("PromptFor(clear player) = %q, want %q", got, "E - Close")
	}
}

func TestDoorLockedDoesNotInteract(t *testing.T) {
	d := NewDoor("door", 0, 0, 16, 16, rl.White, map[string]any{"locked": true})

	if d.CanActivate(rl.Rectangle{X: 100, Y: 100, Width: 16, Height: 16}) {
		t.Fatal("locked door should not activate")
	}
	d.Interact()
	if d.Open {
		t.Fatal("locked door should not open")
	}
	if got := d.Prompt(); got != "Locked" {
		t.Fatalf("Prompt() = %q, want %q", got, "Locked")
	}
}

func TestSeatGeometry(t *testing.T) {
	tests := []struct {
		name          string
		width, height int
	}{
		{name: "square", width: 16, height: 16},
		{name: "wide", width: 32, height: 16},
		{name: "tall", width: 16, height: 32},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := NewSeat("seat", 10, 20, tt.width, tt.height, rl.White, nil)

			wantBounds := rl.Rectangle{
				X:      10,
				Y:      20,
				Width:  float32(tt.width),
				Height: float32(tt.height),
			}
			assertRectangle(t, "Bounds", s.Bounds, wantBounds)
			assertRectangle(t, "CollisionRect()", s.CollisionRect(), wantBounds)

			wantInteraction := rl.Rectangle{
				X:      10 - interactionPadding,
				Y:      20 - interactionPadding,
				Width:  float32(tt.width) + interactionPadding*2,
				Height: float32(tt.height) + interactionPadding*2,
			}
			assertRectangle(t, "InteractionRect()", s.InteractionRect(), wantInteraction)
		})
	}
}

func TestSeatGeometryClampsNonPositiveDimensions(t *testing.T) {
	s := NewSeat("seat", 0, 0, 0, -4, rl.White, nil)

	assertRectangle(t, "Bounds", s.Bounds, rl.Rectangle{X: 0, Y: 0, Width: 1, Height: 1})
}

func TestSeatPrompts(t *testing.T) {
	const (
		awayX = 100
		awayY = 100
	)

	away := rl.Rectangle{X: awayX, Y: awayY, Width: 16, Height: 16}
	overlap := rl.Rectangle{X: 10, Y: 20, Width: 16, Height: 16}

	s := NewSeat("seat", 10, 20, 32, 16, rl.White, nil)

	if got := s.Prompt(); got != "E - Sit" {
		t.Fatalf("Prompt() = %q, want %q", got, "E - Sit")
	}
	if got := s.PromptFor(away); got != "E - Sit" {
		t.Fatalf("PromptFor(clear player) = %q, want %q", got, "E - Sit")
	}

	s.Interact()
	if !s.Occupied {
		t.Fatal("Interact() should occupy the seat")
	}
	if got := s.Prompt(); got != "E - Leave" {
		t.Fatalf("Prompt() = %q, want %q", got, "E - Leave")
	}
	if got := s.PromptFor(away); got != "Occupied" {
		t.Fatalf("PromptFor(clear player, occupied) = %q, want %q", got, "Occupied")
	}
	if got := s.PromptFor(overlap); got != "E - Leave" {
		t.Fatalf("PromptFor(overlapping player) = %q, want %q", got, "E - Leave")
	}
}
