package game

import (
	"strings"
	"testing"

	rl "github.com/gen2brain/raylib-go/raylib"
	"thirty.nights/src/cutscene"
	"thirty.nights/src/dialogue"
	"thirty.nights/src/objects"
	"thirty.nights/src/world"
)

func TestCheckCutscenesAcceptsValidReferences(t *testing.T) {
	door := objects.NewDoor("door1", 0, 0, 16, 16, rl.White, map[string]any{"cutsceneId": "exit"})
	m := &world.MapData{
		ID:            "map1",
		StartCutscene: "intro",
		ObjectsByID:   map[string]world.ObjectType{"door1": door},
		Objects:       []world.ObjectData{{ID: "door1"}},
	}
	cs := cutscene.Catalog{
		"intro": {ID: "intro", Steps: []cutscene.Step{{Type: cutscene.StepDialogue, ID: "hello"}}},
		"exit":  {ID: "exit", Steps: []cutscene.Step{{Type: cutscene.StepWait, Duration: 0.5}}},
	}
	dl := dialogue.Catalog{"hello": {ID: "hello", Start: "s", Nodes: map[string]dialogue.DialogueNode{"s": {Text: "hi"}}}}

	if err := CheckCutscenes(m, cs, dl); err != nil {
		t.Fatalf("CheckCutscenes() error = %v", err)
	}
}

func TestCheckCutscenesRejectsUnknownMapStartCutscene(t *testing.T) {
	m := &world.MapData{ID: "map1", StartCutscene: "missing"}
	err := CheckCutscenes(m, cutscene.Catalog{}, dialogue.Catalog{})
	if err == nil || !strings.Contains(err.Error(), "unknown start_cutscene") {
		t.Fatalf("err = %v, want unknown start_cutscene", err)
	}
}

func TestCheckCutscenesRejectsUnknownObjectCutscene(t *testing.T) {
	door := objects.NewDoor("door1", 0, 0, 16, 16, rl.White, map[string]any{"cutsceneId": "missing"})
	m := &world.MapData{ID: "map1", ObjectsByID: map[string]world.ObjectType{"door1": door}, Objects: []world.ObjectData{{ID: "door1"}}}
	err := CheckCutscenes(m, cutscene.Catalog{}, dialogue.Catalog{})
	if err == nil || !strings.Contains(err.Error(), `object "door1"`) || !strings.Contains(err.Error(), "unknown cutscene") {
		t.Fatalf("err = %v, want object unknown cutscene", err)
	}
}

func TestCheckCutscenesRejectsUnknownDialogueInCutsceneStep(t *testing.T) {
	m := &world.MapData{ID: "map1", StartCutscene: "intro"}
	cs := cutscene.Catalog{"intro": {ID: "intro", Steps: []cutscene.Step{{Type: cutscene.StepDialogue, ID: "nope"}}}}
	err := CheckCutscenes(m, cs, dialogue.Catalog{})
	if err == nil || !strings.Contains(err.Error(), "unknown dialogue") {
		t.Fatalf("err = %v, want unknown dialogue", err)
	}
}

func TestCheckCutscenesTreatsEmptyAsNoTrigger(t *testing.T) {
	door := objects.NewDoor("door1", 0, 0, 16, 16, rl.White, map[string]any{"cutsceneId": ""})
	m := &world.MapData{ID: "map1", StartCutscene: "", ObjectsByID: map[string]world.ObjectType{"door1": door}, Objects: []world.ObjectData{{ID: "door1"}}}
	if err := CheckCutscenes(m, cutscene.Catalog{}, dialogue.Catalog{}); err != nil {
		t.Fatalf("CheckCutscenes() error = %v, want nil", err)
	}
}

func TestCheckMapDependenciesAcceptsKnownDependencies(t *testing.T) {
	catalog := world.Catalog{
		"a": {ID: "a", Maps: []string{"b"}},
		"b": {ID: "b", Maps: []string{}},
	}
	if err := CheckMapDependencies(catalog); err != nil {
		t.Fatalf("CheckMapDependencies() error = %v", err)
	}
}

func TestCheckMapDependenciesRejectsUnknownDependency(t *testing.T) {
	catalog := world.Catalog{
		"a": {ID: "a", Maps: []string{"missing"}},
	}
	err := CheckMapDependencies(catalog)
	if err == nil || !strings.Contains(err.Error(), `map "a"`) || !strings.Contains(err.Error(), "missing") {
		t.Fatalf("err = %v, want map/context and missing id", err)
	}
}

func TestCheckExitTransitionsAcceptsKnownTarget(t *testing.T) {
	catalog := world.Catalog{
		"a": {
			ID: "a",
			Objects: []world.ObjectData{{
				ID: "exit1", Type: "exit", Layer: "floor", X: 0, Y: 0,
				Properties: map[string]any{"nextLevel": "b"},
			}},
		},
		"b": {ID: "b"},
	}
	if err := CheckExitTransitions(catalog); err != nil {
		t.Fatalf("CheckExitTransitions() error = %v", err)
	}
}

func TestCheckExitTransitionsRejectsUnknownTarget(t *testing.T) {
	catalog := world.Catalog{
		"a": {
			ID: "a",
			Objects: []world.ObjectData{{
				ID: "exit1", Type: "exit", Layer: "floor", X: 0, Y: 0,
				Properties: map[string]any{"nextLevel": "missing"},
			}},
		},
	}
	err := CheckExitTransitions(catalog)
	if err == nil || !strings.Contains(err.Error(), `exit object "exit1"`) || !strings.Contains(err.Error(), "missing") {
		t.Fatalf("err = %v, want exit context and missing target", err)
	}
}

func TestCheckExitTransitionsRejectsMissingOrInvalidNextLevel(t *testing.T) {
	tests := []struct {
		name  string
		props map[string]any
	}{
		{name: "missing", props: map[string]any{}},
		{name: "nil", props: map[string]any{"nextLevel": nil}},
		{name: "empty", props: map[string]any{"nextLevel": ""}},
		{name: "wrong type", props: map[string]any{"nextLevel": 2}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			catalog := world.Catalog{
				"a": {
					ID: "a",
					Objects: []world.ObjectData{{
						ID: "exit1", Type: "exit", Layer: "floor", X: 0, Y: 0,
						Properties: tt.props,
					}},
				},
				"b": {ID: "b"},
			}
			err := CheckExitTransitions(catalog)
			if err == nil {
				t.Fatal("expected error")
			}
		})
	}
}
