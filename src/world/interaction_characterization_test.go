package world

import (
	"testing"

	rl "github.com/gen2brain/raylib-go/raylib"
	"thirty.nights/src/npc"
	"thirty.nights/src/objects"
)

func TestFindInteractableReturnsNilWhenNoCandidate(t *testing.T) {
	m := &MapData{ObjectsByID: map[string]ObjectType{}, Objects: nil}
	if got := m.FindInteractable(rl.Rectangle{X: 0, Y: 0, Width: 16, Height: 16}, rl.Vector2{}, rl.Vector2{X: 1, Y: 0}); got != nil {
		t.Fatalf("FindInteractable() = %v, want nil", got)
	}
}

func TestNPCPromptAndActivation(t *testing.T) {
	n, err := npc.NewNPC("guide", 100, 100, 16, 16, 16, npc.DefaultColor, map[string]any{"dialogueId": "intro", "name": "Guide"})
	if err != nil {
		t.Fatalf("NewNPC() error = %v", err)
	}
	if got := n.Prompt(); got != "E - Talk to Guide" {
		t.Fatalf("Prompt() = %q, want %q", got, "E - Talk to Guide")
	}
	if got := n.PromptFor(rl.Rectangle{}); got != "E - Talk to Guide" {
		t.Fatalf("PromptFor() = %q, want %q", got, "E - Talk to Guide")
	}
	if !n.CanActivate(rl.Rectangle{}) {
		t.Fatal("NPC CanActivate() = false, want true")
	}
	if n.DialogueID() != "intro" {
		t.Fatalf("DialogueID() = %q, want %q", n.DialogueID(), "intro")
	}
}

func TestDoorPromptAndActivationCharacterization(t *testing.T) {
	locked := objects.NewDoor("locked", 0, 0, 16, 16, rl.White, map[string]any{"locked": true})
	away := rl.Rectangle{X: 100, Y: 100, Width: 16, Height: 16}
	if locked.Prompt() != "Locked" {
		t.Fatalf("locked Prompt() = %q, want %q", locked.Prompt(), "Locked")
	}
	if locked.PromptFor(away) != "Locked" {
		t.Fatalf("locked PromptFor() = %q, want %q", locked.PromptFor(away), "Locked")
	}
	if locked.CanActivate(away) {
		t.Fatal("locked CanActivate() = true, want false")
	}

	open := objects.NewDoor("open", 0, 0, 16, 16, rl.White, map[string]any{"open": true})
	overlap := rl.Rectangle{X: 0, Y: 0, Width: 16, Height: 16}
	if open.PromptFor(overlap) != "" {
		t.Fatalf("open overlapping PromptFor() = %q, want empty", open.PromptFor(overlap))
	}
	if open.CanActivate(overlap) {
		t.Fatal("open overlapping CanActivate() = true, want false")
	}
	if open.PromptFor(away) != "E - Close" {
		t.Fatalf("open clear PromptFor() = %q, want %q", open.PromptFor(away), "E - Close")
	}
	if !open.CanActivate(away) {
		t.Fatal("open clear CanActivate() = false, want true")
	}
}

func TestTriggerDialogueAndPromptCharacterization(t *testing.T) {
	tr, err := objects.NewTrigger("t", 0, 0, 16, 16, map[string]any{"dialogueId": "intro", "prompt": "E - Inspect"})
	if err != nil {
		t.Fatalf("NewTrigger() error = %v", err)
	}
	if tr.DialogueID() != "intro" {
		t.Fatalf("DialogueID() = %q, want %q", tr.DialogueID(), "intro")
	}
	if tr.Prompt() != "E - Inspect" {
		t.Fatalf("Prompt() = %q, want %q", tr.Prompt(), "E - Inspect")
	}
	if tr.PromptFor(rl.Rectangle{}) != "E - Inspect" {
		t.Fatalf("PromptFor() = %q, want %q", tr.PromptFor(rl.Rectangle{}), "E - Inspect")
	}
	if !tr.CanActivate(rl.Rectangle{X: 0, Y: 0, Width: 16, Height: 16}) {
		t.Fatal("Trigger CanActivate() = false, want true")
	}
}

func TestCutsceneIDCharacterization(t *testing.T) {
	door := objects.NewDoor("door", 0, 0, 16, 16, rl.White, map[string]any{"cutsceneId": "cs1"})
	if door.CutsceneID() != "cs1" {
		t.Fatalf("door CutsceneID() = %q, want %q", door.CutsceneID(), "cs1")
	}
	seat := objects.NewSeat("seat", 0, 0, 16, 16, rl.White, map[string]any{"cutsceneId": "cs2"})
	if seat.CutsceneID() != "cs2" {
		t.Fatalf("seat CutsceneID() = %q, want %q", seat.CutsceneID(), "cs2")
	}
	tr, err := objects.NewTrigger("t", 0, 0, 16, 16, map[string]any{"dialogueId": "d", "cutsceneId": "cs3"})
	if err != nil {
		t.Fatalf("NewTrigger() error = %v", err)
	}
	if tr.CutsceneID() != "cs3" {
		t.Fatalf("trigger CutsceneID() = %q, want %q", tr.CutsceneID(), "cs3")
	}
	n, err := npc.NewNPC("n", 0, 0, 16, 16, 16, npc.DefaultColor, map[string]any{"dialogueId": "d", "cutsceneId": "cs4"})
	if err != nil {
		t.Fatalf("NewNPC() error = %v", err)
	}
	if n.CutsceneID() != "cs4" {
		t.Fatalf("npc CutsceneID() = %q, want %q", n.CutsceneID(), "cs4")
	}
}
