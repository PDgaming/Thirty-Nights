package world

import (
	"testing"

	rl "github.com/gen2brain/raylib-go/raylib"
	"thirty.nights/src/objects"
)

func TestActivateInteractionDialoguePrecedence(t *testing.T) {
	tr, err := objects.NewTrigger("t", 0, 0, 16, 16, map[string]any{"dialogueId": "intro", "cutsceneId": "cs"})
	if err != nil {
		t.Fatalf("NewTrigger() error = %v", err)
	}
	out := ActivateInteraction(tr, rl.Rectangle{}, false)
	if out.Kind != ActivationDialogue || out.DialogueID != "intro" {
		t.Fatalf("outcome = %+v, want dialogue intro", out)
	}
}

func TestActivateInteractionSeatSitStand(t *testing.T) {
	seat := objects.NewSeat("seat", 0, 0, 16, 16, rl.White, nil)
	player := rl.Rectangle{X: 0, Y: 0, Width: 16, Height: 16}

	out := ActivateInteraction(seat, player, false)
	if out.Kind != ActivationSit {
		t.Fatalf("sit outcome = %+v, want sit", out)
	}
	if !seat.Occupied {
		t.Fatal("seat should be occupied after interact")
	}

	out = ActivateInteraction(seat, player, true)
	if out.Kind != ActivationStand {
		t.Fatalf("stand outcome = %+v, want stand", out)
	}
	if seat.Occupied {
		t.Fatal("seat should be unoccupied after stand interact")
	}
}

func TestActivateInteractionCutsceneAfterSideEffect(t *testing.T) {
	door := objects.NewDoor("door", 0, 0, 16, 16, rl.White, map[string]any{"cutsceneId": "open_cs"})
	player := rl.Rectangle{X: 100, Y: 100, Width: 16, Height: 16}
	out := ActivateInteraction(door, player, false)
	if out.Kind != ActivationCutscene || out.CutsceneID != "open_cs" {
		t.Fatalf("outcome = %+v, want cutscene open_cs", out)
	}
	if !door.Open {
		t.Fatal("door side effect should run before cutscene outcome")
	}
}

func TestActivateInteractionBlockedReturnsNone(t *testing.T) {
	locked := objects.NewDoor("locked", 0, 0, 16, 16, rl.White, map[string]any{"locked": true})
	out := ActivateInteraction(locked, rl.Rectangle{X: 100, Y: 100, Width: 16, Height: 16}, false)
	if out.Kind != ActivationNone {
		t.Fatalf("outcome = %+v, want none", out)
	}
	if locked.Open {
		t.Fatal("locked door must not toggle")
	}
}

func TestResolveInteractionUsesPlayerAwarePrompt(t *testing.T) {
	door := objects.NewDoor("door", 0, 0, 16, 16, rl.White, map[string]any{"locked": true})
	m := &MapData{ObjectsByID: map[string]ObjectType{"door": door}, Objects: []ObjectData{{ID: "door"}}}
	center := rl.Vector2{X: 8, Y: -4}
	player := rl.Rectangle{X: 0, Y: -12, Width: 16, Height: 16}
	res := ResolveInteraction(m, player, center, rl.Vector2{X: 0, Y: 1})
	if res.Target != door {
		t.Fatalf("target = %v, want door", res.Target)
	}
	if res.Prompt != "Locked" {
		t.Fatalf("prompt = %q, want Locked", res.Prompt)
	}
	if res.CanActivate {
		t.Fatal("CanActivate = true, want false")
	}
}
