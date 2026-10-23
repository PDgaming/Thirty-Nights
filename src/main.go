package main

import (
	"log"

	rl "github.com/gen2brain/raylib-go/raylib"
	"thirty.nights/src/camera"
	"thirty.nights/src/player"
	"thirty.nights/src/render"
	"thirty.nights/src/world"
)

func main() {
	rl.SetConfigFlags(rl.FlagWindowResizable)
	rl.InitWindow(800, 450, "raylib [core] example - basic window")
	defer rl.CloseWindow()
	rl.SetTargetFPS(60)

	mapData, err := world.LoadMap("asset/maps/deck01.json")
	if err != nil {
		log.Fatal(err)
	}

	player := player.New(0, 0)

	Renderer := render.Renderer{}
	cam := camera.New()
	cam.View.Target = rl.Vector2{
		X: float32(mapData.Width*mapData.TileSize) / 2,
		Y: float32(mapData.Height*mapData.TileSize) / 2,
	}

	for !rl.WindowShouldClose() {
		dt := rl.GetFrameTime()

		cam.UpdateViewport()

		player.Update(dt)
		player.Animator.Update(dt)

		rl.BeginDrawing()
		rl.ClearBackground(rl.RayWhite)
		rl.BeginMode2D(cam.View)

		Renderer.DrawMap(mapData)

		player.Draw()

		rl.EndMode2D()
		rl.EndDrawing()
	}
}
