package world

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"thirty.nights/src/objects"
)

var (
	_ Interactable = (*objects.Door)(nil)
	_ Interactable = (*objects.Seat)(nil)
)

func LoadDir(dir string) (Catalog, error) {
	entries, err := os.ReadDir(dir)
	if err != nil {
		return nil, fmt.Errorf("world: read %s: %w", dir, err)
	}

	catalog := make(Catalog, len(entries))
	sources := make(map[string]string, len(entries))

	for _, entry := range entries {
		if entry.IsDir() || !strings.EqualFold(filepath.Ext(entry.Name()), ".json") {
			continue
		}
		path := filepath.Join(dir, entry.Name())

		m, err := LoadMap(path)
		if err != nil {
			return nil, err
		}

		if previous, exists := sources[m.ID]; exists {
			return nil, fmt.Errorf("world: duplicate map id %q in %s and %s", m.ID, previous, path)
		}
		sources[m.ID] = path
		catalog[m.ID] = m
	}

	return catalog, nil
}

func LoadMap(path string) (*MapData, error) {
	mapFile, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("world: read %s: %w", path, err)
	}

	var mapData MapData
	if err := json.Unmarshal(mapFile, &mapData); err != nil {
		return nil, fmt.Errorf("world: parse %s: %w", path, err)
	}

	for i := range mapData.Layers {
		mode, err := ParseCollisionType(mapData.Layers[i].CollisionName)
		if err != nil {
			return nil, fmt.Errorf("world: invalid collision type in %s: %w", path, err)
		}
		mapData.Layers[i].Collision = mode
	}

	if err := ValidateMap(mapData); err != nil {
		return nil, fmt.Errorf("world: invalid map data in %s: %w", path, err)
	}

	mapData.BuildLayerIndex()
	mapData.ObjectsByID = make(map[string]ObjectType)
	if err := mapData.BuildObjects(); err != nil {
		return nil, fmt.Errorf("world: failed to build objects in %s: %w", path, err)
	}

	return &mapData, nil
}

// Helper functions
func (m *MapData) BuildLayerIndex() {
	m.LayersByID = make(map[string]*LayerData, len(m.Layers))

	for i := range m.Layers {
		m.LayersByID[m.Layers[i].ID] = &m.Layers[i]
	}
}
func (m *MapData) BuildObjects() error {
	for _, object := range m.Objects {
		switch object.Type {
		case "door":
			tileDef, ok := m.TileDefAt(object.Type)
			if !ok {
				return fmt.Errorf("failed to find tiledef for door: %s", object.ID)
			}
			color := HexToRlColor(tileDef.Visual.Color)
			m.ObjectsByID[object.ID] = objects.NewDoor(object.ID, object.X*m.TileSize, object.Y*m.TileSize, m.TileSize, color, object.Properties)
		case "seat":
			tileDef, ok := m.TileDefAt(object.Type)
			if !ok {
				return fmt.Errorf("failed to find tiledef for seat: %s", object.ID)
			}
			color := HexToRlColor(tileDef.Visual.Color)
			m.ObjectsByID[object.ID] = objects.NewSeat(object.ID, object.X*m.TileSize, object.Y*m.TileSize, m.TileSize, color, object.Properties)
		default:
			return fmt.Errorf("unknown object type: %s", object.Type)
		}
	}
	return nil
}
