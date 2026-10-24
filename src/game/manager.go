package game

import (
	"fmt"
	"log"

	rl "github.com/gen2brain/raylib-go/raylib"
	"thirty.nights/src/camera"
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

func (g *Game) LoadLevel(level int) (err error) {
	if level < 1 {
		return fmt.Errorf("level must be greater than 0")
	}
	if level == g.level {
		return
	}
	g.level = level

	mapData, err := world.LoadMap(fmt.Sprintf("asset/maps/%d.json", level))
	if err != nil {
		return fmt.Errorf("failed to load map: %w", err)
	}
	g.Map = mapData

	// check for valid spawn point
	if g.Map.Spawn.X < 0 || g.Map.Spawn.X > g.Map.Width {
		log.Fatalf("invalid X spawn point: %+v", g.Map.Spawn)
	}
	if g.Map.Spawn.Y < 0 || g.Map.Spawn.Y > g.Map.Height {
		log.Fatalf("invalid Y spawn point: %+v", g.Map.Spawn)
	}
	g.Player = player.New(float32(g.Map.Spawn.X*16), float32(g.Map.Spawn.Y*16))
	g.ActiveInteractable = nil
	g.Cam = camera.New()
	g.Cam.View.Target = rl.Vector2{
		X: float32(g.Map.Width*g.Map.TileSize) / 2,
		Y: float32(g.Map.Height*g.Map.TileSize) / 2,
	}
	return nil
}
