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

	g.Player.Update(dt, g.Map)
	g.Player.Animator.Update(dt)

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

	g.Player.Draw(g.Debug)
}

func (g *Game) LoadLevel(level int) (err error) {
	if level < 1 {
		return fmt.Errorf("level must be greater than 0")
	}
	if level == g.level {
		return
	}
	g.level = level

	mapData, err := world.LoadMap(fmt.Sprintf("asset/maps/deck%d.json", level))
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

	g.Cam = camera.New()
	g.Cam.View.Target = rl.Vector2{
		X: float32(g.Map.Width*g.Map.TileSize) / 2,
		Y: float32(g.Map.Height*g.Map.TileSize) / 2,
	}
	return nil
}
