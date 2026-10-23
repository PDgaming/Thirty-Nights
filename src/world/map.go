package world

import (
	"fmt"
	"log"
	"sort"
	"strconv"
	"strings"

	rl "github.com/gen2brain/raylib-go/raylib"
)

func DrawMap(m *MapData, debug bool) {
	layers := make([]LayerData, len(m.Layers))
	copy(layers, m.Layers)

	sort.Slice(layers, func(i, j int) bool {
		return layers[i].ZIndex < layers[j].ZIndex
	})

	for _, layer := range layers {
		for y := 0; y < len(layer.Grid); y++ {
			for x := 0; x < len(layer.Grid[y]); x++ {
				tileID := layer.Grid[y][x]
				if tileID == "" {
					continue
				}
				tileDef, ok := m.Tiles[tileID]
				if !ok {
					fmt.Printf("Failed to find tile with ID %q\n", tileID)
					continue
				}

				pixelX := x * m.TileSize
				pixelY := y * m.TileSize
				color := HexToRlColor(tileDef.Visual.Color)

				rl.DrawRectangle(int32(pixelX), int32(pixelY), int32(m.TileSize), int32(m.TileSize), color)

				if debug {
					rl.DrawRectangleLines(int32(pixelX), int32(pixelY), int32(m.TileSize), int32(m.TileSize), rl.Black)
				}
			}
		}
	}
}

// Helper functions
func HexToRlColor(hex string) rl.Color {
	s := strings.TrimPrefix(hex, "#")

	switch len(s) {
	case 3:
		s = string([]byte{s[0], s[0], s[1], s[1], s[2], s[2]})
	case 6:
		s += "ff"
	case 8:
	default:
		log.Printf("Failed to parse tile color %q: expected 3, 6 or 8 hex digits", hex)
		return rl.Magenta
	}

	value, err := strconv.ParseUint(s, 16, 32)
	if err != nil {
		log.Printf("Failed to parse tile color %q: %s", hex, err)
		return rl.Magenta
	}

	// GetColor expects 0xRRGGBBAA, which we normalized to above.
	return rl.GetColor(uint(value))
}
