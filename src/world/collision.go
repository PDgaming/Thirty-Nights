package world

import (
	"fmt"
	"math"

	rl "github.com/gen2brain/raylib-go/raylib"
)

func ParseCollisionType(name string) (CollisionType, error) {
	switch name {
	case "none":
		return CollisionNone, nil
	case "tile":
		return CollisionTile, nil
	case "object":
		return CollisionObject, nil
	default:
		return CollisionNone, fmt.Errorf("invalid collision type: %s", name)
	}
}

func (m *MapData) CanCollide(layerID string, x, y int) bool {
	layer, ok := m.LayersByID[layerID]
	if !ok {
		return false
	}

	if layer.Collision == CollisionNone {
		return false
	}

	tile, ok := m.TileAt(layer.ID, x, y)
	if !ok {
		return false
	}

	tileDef, ok := m.TileDefAt(tile)
	if !ok {
		return false
	}

	return tileDef.Solid
}

func (m *MapData) CanCollidRect(rect rl.Rectangle) bool {
	mapPixelWidth := float32(m.Width * m.TileSize)
	mapPixelHeight := float32(m.Height * m.TileSize)

	if rect.X < 0 || rect.Y < 0 || rect.X+rect.Width > mapPixelWidth || rect.Y+rect.Height > mapPixelHeight {
		return true
	}

	minTileX := int(math.Floor(float64(rect.X) / float64(m.TileSize)))
	maxTileX := int(math.Ceil(float64(rect.X+rect.Width)/float64(m.TileSize)) - 1)
	minTileY := int(math.Floor(float64(rect.Y) / float64(m.TileSize)))
	maxTileY := int(math.Ceil(float64(rect.Y+rect.Height)/float64(m.TileSize)) - 1)

	for y := minTileY; y <= maxTileY; y++ {
		for x := minTileX; x <= maxTileX; x++ {
			for _, layer := range m.Layers {
				if layer.Collision == CollisionNone {
					continue
				}

				collision := m.CanCollide(layer.ID, x, y)
				if collision {
					return true
				}
			}
		}
	}

	return false
}

func (m *MapData) CanCollideObjects(rect rl.Rectangle) bool {
	for _, obj := range m.ObjectsByID {
		collidable, ok := obj.(Collidable)
		if !ok {
			continue
		}

		if !collidable.Solid() {
			continue
		}

		if rl.CheckCollisionRecs(rect, collidable.CollisionRect()) {
			return true
		}
	}
	return false
}
