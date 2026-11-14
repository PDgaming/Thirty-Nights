package game

import (
	"fmt"
	"log"

	rl "github.com/gen2brain/raylib-go/raylib"
	"thirty.nights/src/camera"
	"thirty.nights/src/cutscene"
	"thirty.nights/src/dialogue"
	"thirty.nights/src/objects"
	"thirty.nights/src/player"
	"thirty.nights/src/ui"
	"thirty.nights/src/world"
)

func New(debug bool) *Game {
	g := &Game{
		Debug: debug,
	}
	return g
}

func (g *Game) Update(dt float32) {
	g.Step(dt, Input{Interact: rl.IsKeyPressed(rl.KeyE)})
}

func (g *Game) Step(dt float32, in Input) {
	if g.Cam != nil {
		g.Cam.UpdateViewport()
	}

	// Drive fade/load state machine every tick.
	if g.TransitionActive() || g.transitionPhase == TransitionLoad {
		g.updateTransition(dt)
	}

	switch g.mode {
	case ModePlaying:
		g.updatePlaying(dt, in)
	case ModeDialogue:
		g.updateDialogue(dt, in)
	case ModeCutscene:
		g.updateCutscene(dt, in)
	}
}

func (g *Game) updatePlaying(dt float32, in Input) {
	if g.TransitionActive() || g.transitionPhase == TransitionLoad {
		return
	}

	if g.pendingCutscene != "" {
		id := g.pendingCutscene
		g.pendingCutscene = ""
		g.startCutscene(id)
		return
	}
	if g.pendingDialogue != "" {
		id := g.pendingDialogue
		g.pendingDialogue = ""
		g.startDialogue(id)
		return
	}

	for _, obj := range g.Map.ObjectsByID {
		obj.Update(dt)
	}

	g.Player.Update(dt, g.Map)
	g.Player.Animator.Update(dt)

	g.checkExitTransitions()
	if g.TransitionActive() || g.transitionPhase == TransitionLoad {
		return
	}

	interaction := world.ResolveInteraction(g.Map, g.Player.Rect(), g.Player.Center(), g.Player.Facing)
	g.ActiveInteractable = interaction.Target

	if interaction.Target != nil && rl.IsKeyPressed(rl.KeyE) {
		outcome := world.ActivateInteraction(interaction.Target, g.Player.Rect(), g.Player.Sitting)
		switch outcome.Kind {
		case world.ActivationDialogue:
			g.RequestDialogue(outcome.DialogueID)
			return
		case world.ActivationSit:
			g.Player.Sit(outcome.SeatBounds)
		case world.ActivationStand:
			g.Player.Stand()
		case world.ActivationCutscene:
			g.RequestCutscene(outcome.CutsceneID)
			return
		}
	}

	playerCenter := rl.Vector2{
		X: g.Player.Position.X + g.Player.Size.X/2,
		Y: g.Player.Position.Y + g.Player.Size.Y/2,
	}
	g.Cam.Follow(playerCenter, dt)
	mapWidthPixels := g.Map.Width * g.Map.TileSize
	mapHeightPixels := g.Map.Height * g.Map.TileSize
	g.Cam.ClampToBounds(mapWidthPixels, mapHeightPixels)
}

func (g *Game) checkExitTransitions() {
	if g.TransitionActive() || g.Player == nil || g.Map == nil {
		return
	}
	playerRect := g.Player.Rect()
	for _, obj := range g.Map.ObjectsByID {
		exit, ok := obj.(*objects.Exit)
		if !ok || exit.NextLevel() == "" {
			continue
		}
		if rl.CheckCollisionRecs(playerRect, exit.CollisionRect()) {
			g.RequestLevelTransition(exit.NextLevel())
			return
		}
	}
}

func (g *Game) updateDialogue(dt float32, in Input) {
	if g.Dialogue == nil {
		g.mode = ModePlaying
		return
	}
	g.Dialogue.Update(dt)

	if in.Interact {
		g.Dialogue.Advance()
	}

	if g.Dialogue.Snapshot().Finished {
		g.finishDialogue()
	}
}

func (g *Game) updateCutscene(dt float32, in Input) {
	if g.Cutscene == nil {
		g.mode = ModePlaying
		return
	}
	g.Cutscene.Update(dt, g.Dialogues)
	if in.Interact {
		g.Cutscene.Advance(g.Dialogues)
	}
	if g.Cutscene.Err() != nil {
		log.Printf("cutscene: %v", g.Cutscene.Err())
	}
	if g.Cutscene.Finished() {
		g.finishCutscene()
	}
}

func (g *Game) startDialogue(id string) {
	def, ok := g.Dialogues.Get(id)
	if !ok {
		log.Printf("dialogue: unknown id %q", id)
		return
	}

	g.Dialogue = dialogue.NewSession(def)
	g.dialogueReturn = g.mode
	g.mode = ModeDialogue
	g.ActiveInteractable = nil
}

func (g *Game) startCutscene(id string) {
	def, ok := g.Cutscenes.Get(id)
	if !ok {
		log.Printf("cutscene: unknown id %q", id)
		return
	}

	g.Cutscene = cutscene.NewSession(def)
	g.mode = ModeCutscene
	g.ActiveInteractable = nil
}

func (g *Game) finishDialogue() {
	g.Dialogue = nil
	g.mode = g.dialogueReturn
	g.ActiveInteractable = nil
}

func (g *Game) finishCutscene() {
	g.Cutscene = nil
	g.mode = ModePlaying
	g.ActiveInteractable = nil
	// Let updatePlaying process any queued triggers on the next tick.
}

func (g *Game) Draw() {
	world.DrawMap(g.Map, g.Debug)
	for _, obj := range g.Map.ObjectsByID {
		obj.Draw(g.Debug)
	}

	g.Player.Draw(g.Debug)

	if g.Debug {
		for _, obj := range g.Map.ObjectsByID {
			obj, ok := obj.(world.Interactable)
			if !ok {
				continue
			}

			if obj == g.ActiveInteractable {
				rl.DrawRectangleLinesEx(obj.InteractionRect(), 2, rl.Green)
			} else {
				rl.DrawRectangleLinesEx(obj.InteractionRect(), 2, rl.Blue)
			}
		}
	}
}

func (g *Game) DrawUI() {
	if g.Debug {
		rl.DrawFPS(0, 0)
	}

	if g.mode == ModeDialogue && g.Dialogue != nil {
		g.drawDialogueSession(g.Dialogue)
		return
	}
	if g.mode == ModeCutscene && g.Cutscene != nil {
		if dlg := g.Cutscene.Dialogue(); dlg != nil {
			g.drawDialogueSession(dlg)
			return
		}
	}

	if g.ActiveInteractable != nil && g.Player != nil && g.Map != nil {
		interaction := world.ResolveInteraction(g.Map, g.Player.Rect(), g.Player.Center(), g.Player.Facing)
		if interaction.Target != g.ActiveInteractable {
			return
		}
		if interaction.Prompt == "" {
			return
		}
		anchor := g.ActiveInteractable.PromptAnchor()
		anchor = rl.GetWorldToScreen2D(anchor, g.Cam.View)
		rl.DrawTextEx(ui.Font(), interaction.Prompt, rl.Vector2{X: anchor.X, Y: anchor.Y}, 20, 1, rl.White)
	}
}

func (g *Game) drawDialogueSession(s *dialogue.Session) {
	snap := s.Snapshot()
	l := dialogue.Compute(float32(rl.GetScreenWidth()), float32(rl.GetScreenHeight()), ui.Measure, snap.Speaker, snap.Text)

	s.SetPages(l.Ranges())
	snap = s.Snapshot()

	ui.DrawDialogue(snap, l)
}

func (g *Game) LoadLevels(dir string) error {
	levels, err := world.LoadDir(dir)
	if err != nil {
		return fmt.Errorf("failed to load levels: %w", err)
	}
	g.Levels = levels
	if err := CheckMapDependencies(g.Levels); err != nil {
		return fmt.Errorf("failed to validate level dependencies: %w", err)
	}
	if err := CheckExitTransitions(g.Levels); err != nil {
		return fmt.Errorf("failed to validate exit transitions: %w", err)
	}
	return nil
}

func (g *Game) LoadLevel(id string) (err error) {
	if g.Levels == nil {
		return fmt.Errorf("no levels loaded")
	}
	if !g.TransitionActive() && id == g.currentLevel {
		return nil
	}

	mapData, ok := g.Levels[id]
	if !ok {
		return fmt.Errorf("unknown level %q", id)
	}

	// check for valid spawn point
	if mapData.Spawn.X < 0 || mapData.Spawn.X > mapData.Width {
		return fmt.Errorf("invalid X spawn point in level %q: %+v", id, mapData.Spawn)
	}
	if mapData.Spawn.Y < 0 || mapData.Spawn.Y > mapData.Height {
		return fmt.Errorf("invalid Y spawn point in level %q: %+v", id, mapData.Spawn)
	}

	// check for valid dialogues and cutscenes
	if err := CheckDialogues(mapData, g.Dialogues); err != nil {
		return fmt.Errorf("invalid dialogue references in level %q: %w", id, err)
	}
	if err := CheckCutscenes(mapData, g.Cutscenes, g.Dialogues); err != nil {
		return fmt.Errorf("invalid cutscene references in level %q: %w", id, err)
	}

	if err := mapData.LoadTileTextures(); err != nil {
		return fmt.Errorf("world: failed to load tile textures in %q: %w", id, err)
	}
	if err := mapData.LoadTextures(); err != nil {
		return fmt.Errorf("world: failed to load textures in %q: %w", id, err)
	}

	// Assign loaded textures to objects that use tile textures.
	for _, obj := range mapData.ObjectsByID {
		if door, ok := obj.(*objects.Door); ok {
			if tex, ok := mapData.TileTextures["door"]; ok {
				door.Texture = tex
			}
			if tex, ok := mapData.TexturesByID["door_closed"]; ok {
				door.Texture = tex
			}
			if tex, ok := mapData.TexturesByID["door_open"]; ok {
				door.OpenTexture = tex
			}
		}
		if seat, ok := obj.(*objects.Seat); ok {
			if tex, ok := mapData.TileTextures["seat"]; ok {
				seat.Texture = tex
			}
		}
	}

	g.currentLevel = id
	g.Map = mapData
	g.Player = player.New(float32(g.Map.Spawn.X*g.Map.TileSize), float32(g.Map.Spawn.Y*g.Map.TileSize))
	g.ActiveInteractable = nil
	g.Cam = camera.New()
	g.Cam.View.Target = rl.Vector2{
		X: float32(g.Map.Width*g.Map.TileSize) / 2,
		Y: float32(g.Map.Height*g.Map.TileSize) / 2,
	}

	// Reset transient runtime state for the new level.
	g.Dialogue = nil
	g.Cutscene = nil
	g.pendingDialogue = ""
	g.pendingCutscene = ""
	g.mode = ModePlaying

	// During transitions start destination cutscene after fade-in.
	if !g.TransitionActive() && mapData.StartCutscene != "" {
		g.startCutscene(mapData.StartCutscene)
	}
	return nil
}

func (g *Game) LoadDialogues(dir string) error {
	catalog, err := dialogue.LoadDir(dir)
	if err != nil {
		return fmt.Errorf("failed to load dialogues: %w", err)
	}
	g.Dialogues = catalog
	return nil
}

func (g *Game) LoadCutscenes(dir string) error {
	catalog, err := cutscene.LoadDir(dir)
	if err != nil {
		return fmt.Errorf("failed to load cutscenes: %w", err)
	}
	g.Cutscenes = catalog
	return nil
}

func (g *Game) Mode() Mode {
	return g.mode
}

func (g *Game) RequestDialogue(id string) {
	g.pendingDialogue = id
}

func (g *Game) RequestCutscene(id string) {
	g.pendingCutscene = id
}

// Helper Functions
func (m Mode) String() string {
	switch m {
	case ModePlaying:
		return "Playing"
	case ModeDialogue:
		return "Dialogue"
	case ModeCutscene:
		return "Cutscene"
	}
	return "Unknown"
}
