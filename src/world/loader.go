package world

import (
	"encoding/json"
	"fmt"
	"os"

	"thirty.nights/src/objects"
)

func LoadMap(path string) (*MapData, error) {
	// Read the map file
	mapFile, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("failed to read map file: %w", err)
	}

	// Parse the map file
	var mapData MapData
	err = json.Unmarshal(mapFile, &mapData)
	if err != nil {
		return nil, fmt.Errorf("failed to parse map file: %w", err)
	}

	for i := range mapData.Layers {
		mode, err := ParseCollisionType(mapData.Layers[i].CollisionName)
		if err != nil {
			return nil, fmt.Errorf("invalid collision type: %w", err)
		}
		mapData.Layers[i].Collision = mode
	}

	// Validate the map data
	err = ValidateMap(mapData)
	if err != nil {
		return nil, fmt.Errorf("invalid map data: %w", err)
	}

	mapData.BuildLayerIndex()
	mapData.ObjectsByID = make(map[string]ObjectType)
	err = mapData.BuildObjects()
	if err != nil {
		return nil, fmt.Errorf("failed to build objects: %w", err)
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
