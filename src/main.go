package main

import (
	"log"

	rl "github.com/gen2brain/raylib-go/raylib"
	"thirty.nights/src/game"
	"thirty.nights/src/ui"
	"thirty.nights/src/world"
)

// Constants
const (
	mapDir      = "asset/maps"
	dialogueDir = "asset/dialogues"
	cutsceneDir = "asset/cutscenes"
	fontPath    = "asset/Pixelify_Sans/static/PixelifySans-Regular.ttf"
	level       = "terminal_office"
	debug       = true
)

func main() {
	rl.SetConfigFlags(rl.FlagWindowResizable)
	rl.InitWindow(800, 450, "Thirty Nights")
	defer rl.CloseWindow()
	rl.SetTargetFPS(60)

	uiFont := rl.LoadFontEx(fontPath, 48, nil, 0)
	rl.SetTextureFilter(uiFont.Texture, rl.FilterPoint)
	ui.SetFont(uiFont)
	defer rl.UnloadFont(uiFont)

	g := game.New(debug)
	if err := g.LoadLevels(mapDir); err != nil {
		log.Fatalf("error: %+v", err)
	}
	if err := g.LoadDialogues(dialogueDir); err != nil {
		log.Fatalf("error: %+v", err)
	}
	if err := g.LoadCutscenes(cutsceneDir); err != nil {
		log.Fatalf("error: %+v", err)
	}
	if err := g.LoadLevel(level); err != nil {
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
		g.DrawTransitionOverlay()
		rl.EndDrawing()
	}
}
