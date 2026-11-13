package world

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"

	rl "github.com/gen2brain/raylib-go/raylib"
	"thirty.nights/src/objects"
)

func assertRect(t *testing.T, name string, got, want rl.Rectangle) {
	t.Helper()
	if got != want {
		t.Fatalf("%s = %+v, want %+v", name, got, want)
	}
}

const testInteractionPadding = 6

func writeTestMap(t *testing.T, tileSize int, objects []ObjectData) string {
	t.Helper()

	grid := make([][]string, 6)
	for y := range grid {
		grid[y] = make([]string, 8)
	}

	m := MapData{
		FormatVersion: 1,
		ID:            "test_map",
		Name:          "Test Map",
		TileSize:      tileSize,
		Width:         8,
		Height:        6,
		Tiles: map[string]TileDef{
			"floor": {Visual: VisualData{Color: "#000000"}},
			"door":  {Solid: true, Visual: VisualData{Color: "#9a7047"}},
			"seat":  {Solid: true, Visual: VisualData{Color: "#60a5fa"}},
		},
		Layers: []LayerData{
			{
				ID:            "floor",
				Type:          "tile",
				ZIndex:        0,
				CollisionName: "none",
				Grid:          grid,
			},
		},
		Objects: objects,
		Spawn:   Spawn{X: 0, Y: 0},
	}

	data, err := json.MarshalIndent(m, "", "  ")
	if err != nil {
		t.Fatalf("marshal test map: %v", err)
	}

	path := filepath.Join(t.TempDir(), "map.json")
	if err := os.WriteFile(path, data, 0o600); err != nil {
		t.Fatalf("write test map: %v", err)
	}
	return path
}

func testMapObjects() []ObjectData {
	return []ObjectData{
		{ID: "door", Type: "door", Layer: "floor", X: 1, Y: 1, Width: 2, Height: 1},
		{ID: "seat", Type: "seat", Layer: "floor", X: 3, Y: 1, Width: 1, Height: 2},
		{
			ID:         "trigger",
			Type:       "trigger",
			Layer:      "floor",
			X:          0,
			Y:          3,
			Width:      3,
			Height:     1,
			Properties: map[string]any{"dialogueId": "intro"},
		},
		{
			ID:         "npc",
			Type:       "npc",
			Layer:      "floor",
			X:          4,
			Y:          3,
			Width:      2,
			Height:     2,
			Properties: map[string]any{"dialogueId": "intro"},
		},
	}
}

func objectRects(t *testing.T, m *MapData, id string) (rl.Rectangle, rl.Rectangle) {
	t.Helper()

	object, ok := m.ObjectsByID[id]
	if !ok {
		t.Fatalf("object %q was not built", id)
	}
	interactable, ok := object.(Interactable)
	if !ok {
		t.Fatalf("object %q is not interactable", id)
	}

	return object.CollisionRect(), interactable.InteractionRect()
}

func TestLoadMapConvertsObjectFootprintsToPixels(t *testing.T) {
	wants := []struct {
		id                         string
		tileX, tileY, tileW, tileH int
	}{
		{id: "door", tileX: 1, tileY: 1, tileW: 2, tileH: 1},
		{id: "seat", tileX: 3, tileY: 1, tileW: 1, tileH: 2},
		{id: "trigger", tileX: 0, tileY: 3, tileW: 3, tileH: 1},
		{id: "npc", tileX: 4, tileY: 3, tileW: 2, tileH: 2},
	}

	for _, tileSize := range []int{16, 32} {
		t.Run(fmt.Sprintf("tile size %d", tileSize), func(t *testing.T) {
			m, err := LoadMap(writeTestMap(t, tileSize, testMapObjects()))
			if err != nil {
				t.Fatalf("LoadMap() error = %v", err)
			}

			for _, want := range wants {
				bounds, interaction := objectRects(t, m, want.id)

				wantBounds := rl.Rectangle{
					X:      float32(want.tileX * tileSize),
					Y:      float32(want.tileY * tileSize),
					Width:  float32(want.tileW * tileSize),
					Height: float32(want.tileH * tileSize),
				}
				assertRect(t, want.id+" bounds", bounds, wantBounds)

				wantInteraction := rl.Rectangle{
					X:      wantBounds.X - testInteractionPadding,
					Y:      wantBounds.Y - testInteractionPadding,
					Width:  wantBounds.Width + testInteractionPadding*2,
					Height: wantBounds.Height + testInteractionPadding*2,
				}
				assertRect(t, want.id+" interaction", interaction, wantInteraction)
			}
		})
	}
}

func TestLoadMapDefaultsMissingDimensionsToOneTile(t *testing.T) {
	objects := []ObjectData{
		{ID: "legacy", Type: "door", Layer: "floor", X: 6, Y: 0},
	}

	m, err := LoadMap(writeTestMap(t, 16, objects))
	if err != nil {
		t.Fatalf("LoadMap() error = %v", err)
	}

	bounds, interaction := objectRects(t, m, "legacy")
	assertRect(t, "legacy bounds", bounds, rl.Rectangle{X: 96, Y: 0, Width: 16, Height: 16})
	assertRect(t, "legacy interaction", interaction, rl.Rectangle{X: 90, Y: -6, Width: 28, Height: 28})
}

func TestLoadMapRejectsObjectFootprintOverflow(t *testing.T) {
	objects := []ObjectData{
		{ID: "door", Type: "door", Layer: "floor", X: 7, Y: 0, Width: 2, Height: 1},
	}

	_, err := LoadMap(writeTestMap(t, 16, objects))
	if err == nil {
		t.Fatal("LoadMap() error = nil, want an error")
	}
	if !strings.Contains(err.Error(), "footprint width") {
		t.Fatalf("LoadMap() error = %q, want it to contain %q", err, "footprint width")
	}
}

func writeTestMapWithFields(t *testing.T, extra map[string]any) string {
	t.Helper()

	grid := make([][]string, 2)
	for y := range grid {
		grid[y] = make([]string, 2)
	}

	m := map[string]any{
		"format_version": 1,
		"id":             "test_map",
		"name":           "Test Map",
		"tile_size":      16,
		"width":          2,
		"height":         2,
		"tiles": map[string]any{
			"floor": map[string]any{"solid": false, "visual": map[string]any{"color": "#000000"}},
			"door":  map[string]any{"solid": true, "visual": map[string]any{"color": "#9a7047"}},
			"seat":  map[string]any{"solid": true, "visual": map[string]any{"color": "#60a5fa"}},
		},
		"layers": []any{
			map[string]any{
				"id":        "floor",
				"type":      "tile",
				"z_index":   0,
				"collision": "none",
				"grid":      grid,
			},
		},
		"objects": []any{},
		"spawn":   map[string]any{"x": 0, "y": 0},
	}
	for k, v := range extra {
		m[k] = v
	}

	data, err := json.MarshalIndent(m, "", "  ")
	if err != nil {
		t.Fatalf("marshal test map: %v", err)
	}
	path := filepath.Join(t.TempDir(), "map.json")
	if err := os.WriteFile(path, data, 0o600); err != nil {
		t.Fatalf("write test map: %v", err)
	}
	return path
}

func TestLoadMapNormalizesMissingMapsDependencies(t *testing.T) {
	m, err := LoadMap(writeTestMapWithFields(t, nil))
	if err != nil {
		t.Fatalf("LoadMap() error = %v", err)
	}
	if m.Maps == nil {
		t.Fatal("Maps = nil, want empty slice")
	}
	if len(m.Maps) != 0 {
		t.Fatalf("len(Maps) = %d, want 0", len(m.Maps))
	}
}

func TestLoadMapPreservesMapsDependencies(t *testing.T) {
	m, err := LoadMap(writeTestMapWithFields(t, map[string]any{
		"maps": []any{"bar", "terminal_office"},
	}))
	if err != nil {
		t.Fatalf("LoadMap() error = %v", err)
	}
	if len(m.Maps) != 2 || m.Maps[0] != "bar" || m.Maps[1] != "terminal_office" {
		t.Fatalf("Maps = %#v, want [bar terminal_office]", m.Maps)
	}
}

func TestLoadMapBuildsExitObject(t *testing.T) {
	grid := make([][]string, 2)
	for y := range grid {
		grid[y] = make([]string, 2)
	}
	m := MapData{
		FormatVersion: 1,
		ID:            "test_map",
		Name:          "Test Map",
		TileSize:      16,
		Width:         2,
		Height:        2,
		Tiles:         map[string]TileDef{"floor": {Visual: VisualData{Color: "#000000"}}},
		Layers:        []LayerData{{ID: "floor", Type: "tile", ZIndex: 0, CollisionName: "none", Grid: grid}},
		Objects: []ObjectData{{
			ID: "exit1", Type: "exit", Layer: "floor", X: 1, Y: 0, Width: 1, Height: 1,
			Properties: map[string]any{"nextLevel": "bar"},
		}},
		Spawn: Spawn{X: 0, Y: 0},
	}
	data, err := json.MarshalIndent(m, "", "  ")
	if err != nil {
		t.Fatalf("marshal: %v", err)
	}
	path := filepath.Join(t.TempDir(), "map.json")
	if err := os.WriteFile(path, data, 0o600); err != nil {
		t.Fatalf("write: %v", err)
	}
	loaded, err := LoadMap(path)
	if err != nil {
		t.Fatalf("LoadMap() error = %v", err)
	}
	obj, ok := loaded.ObjectsByID["exit1"]
	if !ok {
		t.Fatal("exit1 not built")
	}
	exit, ok := obj.(*objects.Exit)
	if !ok {
		t.Fatalf("exit1 type = %T, want *objects.Exit", obj)
	}
	if exit.NextLevel() != "bar" {
		t.Fatalf("NextLevel() = %q, want bar", exit.NextLevel())
	}
}
