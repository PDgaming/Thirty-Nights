package game

import (
	"testing"

	"thirty.nights/src/cutscene"
	"thirty.nights/src/dialogue"
	"thirty.nights/src/world"
)

func newTestGameForSession() *Game {
	g := New(false)
	g.Levels = world.Catalog{
		"a": {ID: "a", TileSize: 16, Width: 20, Height: 20, Spawn: world.Spawn{X: 1, Y: 1}, Tiles: map[string]world.TileDef{}, Layers: []world.LayerData{}, ObjectsByID: map[string]world.ObjectType{}},
		"b": {ID: "b", TileSize: 16, Width: 20, Height: 20, Spawn: world.Spawn{X: 2, Y: 2}, Tiles: map[string]world.TileDef{}, Layers: []world.LayerData{}, ObjectsByID: map[string]world.ObjectType{}, StartCutscene: "intro"},
	}
	g.Dialogues = dialogue.Catalog{
		"d1": {ID: "d1", Start: "s", Nodes: map[string]dialogue.DialogueNode{"s": {Text: "hi"}}},
	}
	g.Cutscenes = cutscene.Catalog{
		"cs1":   {ID: "cs1", Steps: []cutscene.Step{{Type: cutscene.StepWait, Duration: 1}}},
		"intro": {ID: "intro", Steps: []cutscene.Step{{Type: cutscene.StepWait, Duration: 1}}},
	}
	return g
}

func TestRequestLevelTransitionGuards(t *testing.T) {
	g := newTestGameForSession()
	g.RequestLevelTransition("")
	if g.TransitionActive() {
		t.Fatal("empty target should not start transition")
	}
	g.RequestLevelTransition("b")
	if !g.TransitionActive() || g.transitionPhase != TransitionFadeOut {
		t.Fatalf("phase=%v active=%v, want fadeout active", g.transitionPhase, g.TransitionActive())
	}
	g.RequestLevelTransition("a")
	if g.transitionTarget != "b" {
		t.Fatalf("target=%q, want b (reentrant request ignored)", g.transitionTarget)
	}
}

func TestTransitionProgressionAndStartCutsceneAfterFadeIn(t *testing.T) {
	g := newTestGameForSession()
	if err := g.LoadLevel("a"); err != nil {
		t.Fatalf("LoadLevel(a): %v", err)
	}
	g.RequestLevelTransition("b")

	for i := 0; i < 400 && (g.transitionPhase == TransitionFadeOut || g.transitionPhase == TransitionLoad); i++ {
		g.Step(0.05, Input{})
	}
	if g.transitionPhase != TransitionFadeIn && g.transitionPhase != TransitionNone {
		t.Fatalf("after fadeout/load phase=%v, want fadein or none", g.transitionPhase)
	}
	if g.currentLevel != "b" {
		t.Fatalf("currentLevel=%q, want b", g.currentLevel)
	}
	// During transition, start cutscene should be deferred until fade-in completes.
	if g.transitionPhase == TransitionFadeIn && g.Mode() == ModeCutscene {
		t.Fatal("start cutscene should not run before fade-in completes")
	}
	for i := 0; i < 200 && g.TransitionActive(); i++ {
		g.Step(0.05, Input{})
	}
	if g.TransitionActive() {
		t.Fatal("transition should complete")
	}
	if g.Mode() != ModeCutscene || g.Cutscene == nil {
		t.Fatalf("mode=%v cutscene=%v, want cutscene started after fade-in", g.Mode(), g.Cutscene != nil)
	}
}

func TestLoadLevelStartsStartCutsceneImmediatelyWhenNotTransitioning(t *testing.T) {
	g := newTestGameForSession()
	if err := g.LoadLevel("b"); err != nil {
		t.Fatalf("LoadLevel(b): %v", err)
	}
	if g.Mode() != ModeCutscene || g.Cutscene == nil {
		t.Fatalf("mode=%v cutscene=%v, want immediate start cutscene", g.Mode(), g.Cutscene != nil)
	}
}

func TestPendingCutsceneHasPriorityOverDialogue(t *testing.T) {
	g := newTestGameForSession()
	if err := g.LoadLevel("a"); err != nil {
		t.Fatalf("LoadLevel(a): %v", err)
	}
	g.RequestDialogue("d1")
	g.RequestCutscene("cs1")
	g.Step(0.016, Input{})
	if g.Mode() != ModeCutscene {
		t.Fatalf("mode=%v, want cutscene", g.Mode())
	}
	if g.pendingDialogue != "d1" {
		t.Fatalf("pendingDialogue=%q, want d1 preserved for later", g.pendingDialogue)
	}
}

func TestDialogueReturnModePreserved(t *testing.T) {
	g := newTestGameForSession()
	g.mode = ModeCutscene
	g.dialogueReturn = ModeCutscene
	g.Dialogue = dialogue.NewSession(g.Dialogues["d1"])
	g.mode = ModeDialogue
	g.finishDialogue()
	if g.Mode() != ModeCutscene {
		t.Fatalf("mode=%v, want cutscene return", g.Mode())
	}
}

func TestLoadLevelNoopWhenSameLevelAndNotTransitioning(t *testing.T) {
	g := newTestGameForSession()
	if err := g.LoadLevel("a"); err != nil {
		t.Fatalf("LoadLevel(a): %v", err)
	}
	p := g.Player
	g.RequestDialogue("d1")
	if err := g.LoadLevel("a"); err != nil {
		t.Fatalf("LoadLevel(a) second: %v", err)
	}
	if g.Player != p {
		t.Fatal("same-level load outside transition should be no-op")
	}
	if g.pendingDialogue != "d1" {
		t.Fatalf("pendingDialogue=%q, want preserved by no-op", g.pendingDialogue)
	}
}
