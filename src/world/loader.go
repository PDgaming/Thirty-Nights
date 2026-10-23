package world

import (
	"encoding/json"
	"fmt"
	"os"
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

	return &mapData, nil
}

// Helper functions
func (m *MapData) BuildLayerIndex() {
	m.LayersByID = make(map[string]*LayerData, len(m.Layers))

	for i := range m.Layers {
		m.LayersByID[m.Layers[i].ID] = &m.Layers[i]
	}
}
