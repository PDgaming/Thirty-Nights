package game

import (
	"fmt"

	rl "github.com/gen2brain/raylib-go/raylib"
	"thirty.nights/src/camera"
	"thirty.nights/src/objects"
	"thirty.nights/src/player"
	"thirty.nights/src/world"
)

func New(debug bool) *Game {
	g := &Game{
		Debug: debug,
	}
	return g
}

func (g *Game) Update(dt float32) {
	g.Cam.UpdateViewport()

	for _, obj := range g.Map.ObjectsByID {
		obj.Update(dt)
	}

	g.Player.Update(dt, g.Map)
	g.Player.Animator.Update(dt)

	g.ActiveInteractable = g.Map.FindInteractable(g.Player.Rect(), g.Player.Center(), g.Player.Facing)

	if g.ActiveInteractable != nil && rl.IsKeyPressed(rl.KeyE) {
		canActivate := true

		if aware, ok := g.ActiveInteractable.(world.PlayerAwareInteractable); ok {
			canActivate = aware.CanActivate(g.Player.Rect())
		}
		if canActivate {
			g.ActiveInteractable.Interact()
			if seat, ok := g.ActiveInteractable.(*objects.Seat); ok {
				switch {
				case seat.Occupied && !g.Player.Sitting:
					g.Player.Sit(seat.Bounds)
				case !seat.Occupied && g.Player.Sitting:
					g.Player.Stand()
				}
			}
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
	if g.ActiveInteractable != nil {
		prompt := g.ActiveInteractable.Prompt()
		if aware, ok := g.ActiveInteractable.(world.PlayerAwareInteractable); ok {
			prompt = aware.PromptFor(g.Player.Rect())
		}
		anchor := g.ActiveInteractable.PromptAnchor()
		anchor = rl.GetWorldToScreen2D(anchor, g.Cam.View)
		rl.DrawText(prompt, int32(anchor.X), int32(anchor.Y), 20, rl.White)
	}
}

// LoadLevels reads every map in dir into memory, keyed by map ID.
func (g *Game) LoadLevels(dir string) error {
	levels, err := world.LoadDir(dir)
	if err != nil {
		return fmt.Errorf("failed to load levels: %w", err)
	}
	g.Levels = levels
	return nil
}

func (g *Game) LoadLevel(id string) (err error) {
	if g.Levels == nil {
		return fmt.Errorf("no levels loaded")
	}
	if id == g.currentLevel {
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

	g.currentLevel = id
	g.Map = mapData
	g.Player = player.New(float32(g.Map.Spawn.X*16), float32(g.Map.Spawn.Y*16))
	g.ActiveInteractable = nil
	g.Cam = camera.New()
	g.Cam.View.Target = rl.Vector2{
		X: float32(g.Map.Width*g.Map.TileSize) / 2,
		Y: float32(g.Map.Height*g.Map.TileSize) / 2,
	}
	return nil
}
