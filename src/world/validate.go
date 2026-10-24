package world

import (
	"fmt"
	"regexp"
)

var hexColorRegex = regexp.MustCompile(`^#(?:[0-9a-fA-F]{3}|[0-9a-fA-F]{6}|[0-9a-fA-F]{8})$`)

func ValidateMap(m MapData) error {
	err := ValidateMapProperties(m)
	if err != nil {
		return err
	}

	err = ValidateTileDefProperties(m.Tiles)
	if err != nil {
		return err
	}

	var layerIDs = make(map[string]bool)
	err = ValidateLayerProperties(m.Layers, m.Height, m.Width, m.Tiles, layerIDs)
	if err != nil {
		return err
	}

	var roomIDs = make(map[string]bool)
	err = ValidateRoomProperties(m.Rooms, m.Width, m.Height, roomIDs)
	if err != nil {
		return err
	}

	var objectIDs = make(map[string]bool)
	err = ValidateObjectProperties(m.Objects, m.Height, m.Width, roomIDs, objectIDs)
	if err != nil {
		return err
	}

	return nil
}

// Helper functions
func IsValidHexColor(s string) bool {
	return hexColorRegex.MatchString(s)
}
func ValidateMapProperties(m MapData) error {
	if m.FormatVersion != 1 {
		return fmt.Errorf("unsupported map format version: %d", m.FormatVersion)
	}
	if m.ID == "" || m.Name == "" {
		return fmt.Errorf("map must have an ID and a name")
	}
	if m.TileSize <= 0 || m.Width <= 0 || m.Height <= 0 {
		return fmt.Errorf("tile size and dimensions must be greater than 0")
	}
	if len(m.Tiles) == 0 {
		return fmt.Errorf("map must have at least one tile")
	}
	if len(m.Layers) == 0 {
		return fmt.Errorf("map must have at least one layer")
	}
	return nil
}
func ValidateTileDefProperties(tiles map[string]TileDef) error {
	for name, tile := range tiles {
		if name == "" {
			return fmt.Errorf("tile must have a name")
		}
		if !IsValidHexColor(tile.Visual.Color) {
			return fmt.Errorf("invalid tile color: %s", tile.Visual.Color)
		}
	}
	return nil
}
func ValidateLayerProperties(layers []LayerData, height, width int, tiles map[string]TileDef, layerIDs map[string]bool) error {
	for _, layer := range layers {
		if layer.ID == "" {
			return fmt.Errorf("layer must have an ID")
		}
		if layerIDs[layer.ID] {
			return fmt.Errorf("duplicate layer ID: %s", layer.ID)
		}
		layerIDs[layer.ID] = true

		if layer.Type != "tile" {
			return fmt.Errorf("invalid layer type: %s", layer.Type)
		}
		if layer.ZIndex < 0 {
			return fmt.Errorf("layer z-index must be greater than or equal to 0")
		}

		if layer.CollisionName == "" {
			return fmt.Errorf("layer must have a collision type")
		}

		if len(layer.Grid) != height {
			return fmt.Errorf("layer grid height must match map height")
		}
		for _, row := range layer.Grid {
			if len(row) != width {
				return fmt.Errorf("layer grid row width must match map width")
			}
			for _, tileID := range row {
				if tileID == "" {
					continue
				}
				if _, ok := tiles[tileID]; !ok {
					return fmt.Errorf("invalid tile ID: %s", tileID)
				}
			}
		}
	}
	return nil
}
func ValidateRoomProperties(rooms []RoomData, width, height int, roomIDs map[string]bool) error {
	for _, room := range rooms {
		if room.ID == "" {
			return fmt.Errorf("room must have an ID")
		}
		if roomIDs[room.ID] {
			return fmt.Errorf("duplicate room ID: %s", room.ID)
		}
		roomIDs[room.ID] = true

		if room.Name == "" {
			return fmt.Errorf("room must have a name")
		}

		if room.Bounds.X <= 0 || room.Bounds.Y <= 0 {
			return fmt.Errorf("room bounds x and y must be less than 0")
		}
		if room.Bounds.Width <= 0 || room.Bounds.Height <= 0 {
			return fmt.Errorf("room bounds width and height must be greater than 0")
		}
		if room.Bounds.X+room.Bounds.Width > width || room.Bounds.Y+room.Bounds.Height > height {
			return fmt.Errorf("room bounds x and y must be less than the map width and height")
		}
	}
	return nil
}
func ValidateObjectProperties(objects []ObjectData, height, width int, roomIDs map[string]bool, objectIDs map[string]bool) error {
	for _, object := range objects {
		if object.ID == "" {
			return fmt.Errorf("object must have an ID")
		}
		if objectIDs[object.ID] {
			return fmt.Errorf("duplicate object ID: %s", object.ID)
		}
		objectIDs[object.ID] = true

		if object.Type == "" {
			return fmt.Errorf("object must have a type")
		}

		if object.Layer == "" {
			return fmt.Errorf("object must have a layer")
		}

		if object.X < 0 || object.X >= width {
			return fmt.Errorf("object x must be greater than 0 and less than or equal to map width")
		}
		if object.Y < 0 || object.Y >= height {
			return fmt.Errorf("object y must be greater than 0 and less than or equal to map height")
		}

		if _, ok := object.Properties["room_id"]; !ok {
			continue
		}
		if _, ok := object.Properties["room_id"].(string); !ok {
			return fmt.Errorf("object room id must be a string")
		}

		if _, ok := roomIDs[object.Properties["room_id"].(string)]; !ok {
			return fmt.Errorf("object must have a room ID")
		}
	}
	return nil
}
