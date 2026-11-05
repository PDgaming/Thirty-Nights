package world

import (
	"testing"

	rl "github.com/gen2brain/raylib-go/raylib"
	"thirty.nights/src/objects"
)

func playerRectAt(x, y float32) rl.Rectangle {
	return rl.Rectangle{X: x - 8, Y: y - 8, Width: 16, Height: 16}
}

func interactableMap(byID map[string]ObjectType, order ...string) *MapData {
	m := &MapData{ObjectsByID: byID}
	for _, id := range order {
		m.Objects = append(m.Objects, ObjectData{ID: id})
	}
	return m
}

func requireTarget(t *testing.T, m *MapData, playerRect rl.Rectangle, center, facing rl.Vector2) Interactable {
	t.Helper()
	target := m.FindInteractable(playerRect, center, facing)
	if target == nil {
		t.Fatal("FindInteractable() = nil, want a target")
	}
	return target
}

func TestFindInteractableWideDoorFromEveryHalf(t *testing.T) {
	door := objects.NewDoor("door", 100, 100, 64, 16, rl.White, nil)
	m := interactableMap(map[string]ObjectType{"door": door}, "door")

	tests := []struct {
		name   string
		center rl.Vector2
		facing rl.Vector2
	}{
		{name: "left half", center: rl.Vector2{X: 88, Y: 108}, facing: rl.Vector2{X: 1, Y: 0}},
		{name: "right half", center: rl.Vector2{X: 176, Y: 108}, facing: rl.Vector2{X: -1, Y: 0}},
		{name: "below", center: rl.Vector2{X: 132, Y: 88}, facing: rl.Vector2{X: 0, Y: 1}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := requireTarget(t, m, playerRectAt(tt.center.X, tt.center.Y), tt.center, tt.facing)
			if got != door {
				t.Fatalf("FindInteractable() = %T, want the wide door", got)
			}
		})
	}
}

func TestFindInteractablePlayerInsideNonSolidTrigger(t *testing.T) {
	trigger, err := objects.NewTrigger("trigger", 100, 100, 32, 32, map[string]any{"dialogueId": "intro"})
	if err != nil {
		t.Fatalf("NewTrigger() error = %v", err)
	}

	m := interactableMap(map[string]ObjectType{"trigger": trigger}, "trigger")
	center := rl.Vector2{X: 116, Y: 116}
	facingAway := rl.Vector2{X: 0, Y: -1}

	got := requireTarget(t, m, playerRectAt(center.X, center.Y), center, facingAway)
	if got != trigger {
		t.Fatalf("FindInteractable() = %T, want the trigger despite facing away", got)
	}
}

func TestFindInteractableRanksByClosestPointNotCenter(t *testing.T) {
	small := objects.NewDoor("small", 96, 100, 16, 16, rl.White, nil)
	large := objects.NewDoor("large", 94, 100, 200, 16, rl.White, nil)

	// The small object's center is much closer, but the large object's edge is
	// closer to the player.
	m := interactableMap(
		map[string]ObjectType{"small": small, "large": large},
		"small", "large",
	)

	center := rl.Vector2{X: 84, Y: 108}
	facingRight := rl.Vector2{X: 1, Y: 0}

	got := requireTarget(t, m, playerRectAt(center.X, center.Y), center, facingRight)
	if got != large {
		t.Fatalf("FindInteractable() = %T, want the large object whose edge is closest", got)
	}
}

func TestFindInteractableZeroFacingSkipsDirectionCheck(t *testing.T) {
	door := objects.NewDoor("door", 100, 100, 64, 16, rl.White, nil)
	m := interactableMap(map[string]ObjectType{"door": door}, "door")

	center := rl.Vector2{X: 132, Y: 88}
	got := requireTarget(t, m, playerRectAt(center.X, center.Y), center, rl.Vector2{})
	if got != door {
		t.Fatalf("FindInteractable() = %T, want the door with zero facing", got)
	}
}

func TestFindInteractableOneTileObjectKeepsWorking(t *testing.T) {
	door := objects.NewDoor("door", 100, 100, 16, 16, rl.White, nil)
	m := interactableMap(map[string]ObjectType{"door": door}, "door")

	center := rl.Vector2{X: 108, Y: 88}
	facingDown := rl.Vector2{X: 0, Y: 1}

	got := requireTarget(t, m, playerRectAt(center.X, center.Y), center, facingDown)
	if got != door {
		t.Fatalf("FindInteractable() = %T, want the one-tile door", got)
	}
}

func TestFindInteractablePlayerOnInteractionEdge(t *testing.T) {
	door := objects.NewDoor("door", 100, 100, 64, 16, rl.White, nil)
	m := interactableMap(map[string]ObjectType{"door": door}, "door")

	// The player center sits exactly on the interaction rectangle's left edge
	// while facing away.
	center := rl.Vector2{X: 94, Y: 108}
	got := requireTarget(t, m, playerRectAt(center.X, center.Y), center, rl.Vector2{X: -1, Y: 0})
	if got != door {
		t.Fatalf("FindInteractable() = %T, want the door from its edge", got)
	}
}

func TestFindInteractableReturnsLockedObject(t *testing.T) {
	door := objects.NewDoor("door", 100, 100, 64, 16, rl.White, map[string]any{"locked": true})
	m := interactableMap(map[string]ObjectType{"door": door}, "door")

	playerRect := playerRectAt(110, 88)
	center := rl.Vector2{X: 110, Y: 88}
	facingDown := rl.Vector2{X: 0, Y: 1}

	got := requireTarget(t, m, playerRect, center, facingDown)
	if got != door {
		t.Fatalf("FindInteractable() = %T, want the locked door", got)
	}
	if door.CanActivate(playerRect) {
		t.Fatal("locked door should still be filtered by CanActivate in the game manager")
	}
}
