package world

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	rl "github.com/gen2brain/raylib-go/raylib"

	"thirty.nights/src/npc"
	"thirty.nights/src/objects"
)

var (
	_ Interactable         = (*objects.Door)(nil)
	_ Interactable         = (*objects.Seat)(nil)
	_ Interactable         = (*objects.Trigger)(nil)
	_ DialogueInteractable = (*objects.Trigger)(nil)
	_ DialogueInteractable = (*npc.NPC)(nil)
	_ ObjectType           = (*npc.NPC)(nil)
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

	// Build asset map: resolve asset paths relative to asset/tiles/
	if mapData.Assets != nil {
		mapData.AssetsByID = make(map[string]string, len(mapData.Assets))
		for assetID, relPath := range mapData.Assets {
			mapData.AssetsByID[assetID] = assetPath(relPath)
		}
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

// assetPath returns the full path to an asset relative to the project root.
// The map's Assets map stores paths relative to "asset/tiles/".
func assetPath(relPath string) string {
	return "asset/tiles/" + relPath
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
		width, height := object.Footprint().Pixels(m.TileSize)

		switch object.Type {
		case "door":
			tileDef, ok := m.TileDefAt(object.Type)
			if !ok {
				return fmt.Errorf("failed to find tiledef for door: %s", object.ID)
			}
			color := HexToRlColor(tileDef.Visual.Color)
			door := objects.NewDoor(object.ID, object.X*m.TileSize, object.Y*m.TileSize, width, height, color, object.Properties)
			if tex, ok := m.TileTextures[object.Type]; ok {
				door.Texture = tex
			}
			m.ObjectsByID[object.ID] = door
		case "seat":
			tileDef, ok := m.TileDefAt(object.Type)
			if !ok {
				return fmt.Errorf("failed to find tiledef for seat: %s", object.ID)
			}
			color := HexToRlColor(tileDef.Visual.Color)
			seat := objects.NewSeat(object.ID, object.X*m.TileSize, object.Y*m.TileSize, width, height, color, object.Properties)
			if tex, ok := m.TileTextures[object.Type]; ok {
				seat.Texture = tex
			}
			m.ObjectsByID[object.ID] = seat
		case "trigger":
			trigger, err := objects.NewTrigger(object.ID, object.X*m.TileSize, object.Y*m.TileSize, width, height, object.Properties)
			if err != nil {
				return fmt.Errorf("failed to create trigger %q: %w", object.ID, err)
			}
			m.ObjectsByID[object.ID] = trigger
		case "npc":
			color := npc.DefaultColor
			if raw, ok := object.Properties["color"].(string); ok && IsValidHexColor(raw) {
				color = HexToRlColor(raw)
			}
			n, err := npc.NewNPC(object.ID, object.X*m.TileSize, object.Y*m.TileSize, width, height, m.TileSize, color, object.Properties)
			if err != nil {
				return fmt.Errorf("failed to create npc %q: %w", object.ID, err)
			}
			m.ObjectsByID[object.ID] = n
		default:
			return fmt.Errorf("unknown object type: %s", object.Type)
		}
	}
	return nil
}

// LoadTileTextures attempts to load a texture for each tile whose Visual.Texture
// field names a file. Files are resolved relative to asset/tiles/. Missing or
// unreadable files are skipped so the tile falls back to its solid colour.
func (m *MapData) LoadTileTextures() error {
	m.TileTextures = make(map[string]rl.Texture2D, len(m.Tiles))

	for tileID, tile := range m.Tiles {
		textureName := tile.Visual.Texture
		if textureName == "" {
			continue
		}

		fullPath := "asset/tiles/" + textureName
		if _, statErr := os.Stat(fullPath); statErr != nil {
			fmt.Printf("warning: tile texture not found for %q: %s (skipping)\n", tileID, fullPath)
			continue
		}

		texture := rl.LoadTexture(fullPath)
		if texture.ID == 0 {
			fmt.Printf("warning: failed to load tile texture for %q from %s\n", tileID, fullPath)
			continue
		}

		m.TileTextures[tileID] = texture
	}

	return nil
}

// LoadAssets attempts to load all asset textures referenced in the map's
// asset map. Each entry in MapData.Assets maps an asset ID to a relative
// path under asset/tiles/. Loaded textures are stored in
// MapData.assetTextures for later use. Errors for missing files are logged
// but do not prevent map loading.
func (m *MapData) LoadAssets() error {
	if m.AssetsByID == nil {
		return nil
	}

	m.assetTextures = make(map[string]rl.Texture2D, len(m.AssetsByID))

	for assetID, relPath := range m.AssetsByID {
		fullPath := relPath // already resolved to "asset/tiles/..."

		// Check if file exists before attempting to load
		if _, statErr := os.Stat(fullPath); statErr != nil {
			fmt.Printf("warning: asset not found for %q: %s (skipping)\n", assetID, fullPath)
			continue
		}

		texture := rl.LoadTexture(fullPath)
		if texture.ID == 0 {
			fmt.Printf("warning: failed to load texture for %q from %s\n", assetID, fullPath)
			continue
		}

		m.assetTextures[assetID] = texture
		fmt.Printf("loaded asset %q from %s\n", assetID, fullPath)
	}

	return nil
}
