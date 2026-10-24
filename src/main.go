package main

import (
	"log"

	rl "github.com/gen2brain/raylib-go/raylib"
	"thirty.nights/src/game"
	"thirty.nights/src/world"
)

// Constants
const (
	level = 1
	debug = true
)

func main() {
	rl.SetConfigFlags(rl.FlagWindowResizable)
	rl.InitWindow(800, 450, "Thirty Nights")
	defer rl.CloseWindow()
	rl.SetTargetFPS(60)

	g := game.New(debug)
	err := g.LoadLevel(level)
	if err != nil {
		log.Fatalf("error: %+v", err)
	}

	for !rl.WindowShouldClose() {
		dt := rl.GetFrameTime()
		g.Update(dt)

		voidTile, ok := g.Map.TileDefAt("void")
		if !ok {
			log.Fatal("failed to find void tile")
		}
		voidColor := world.HexToRlColor(voidTile.Visual.Color)

		rl.BeginDrawing()
		rl.ClearBackground(voidColor)
		rl.BeginMode2D(g.Cam.View)

		g.Draw()

		rl.EndMode2D()
		g.DrawUI()
		rl.EndDrawing()
	}
}
