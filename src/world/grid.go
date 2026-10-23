package world

func (m *MapData) InBounds(x, y int) bool {
	if x < 0 || y < 0 {
		return false
	}
	if x >= m.Width || y >= m.Height {
		return false
	}

	return true
}

func (m *MapData) TileAt(layerID string, x, y int) (string, bool) {
	layer, ok := m.LayersByID[layerID]
	if !ok {
		return "", false
	}

	if !m.InBounds(x, y) {
		return "", false
	}

	if layer.Grid[y][x] == "" {
		return "", false
	}

	return layer.Grid[y][x], true

}

func (m *MapData) TileDefAt(tileID string) (*TileDef, bool) {
	tile, ok := m.Tiles[tileID]
	return &tile, ok
}
