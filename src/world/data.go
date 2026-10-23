package world

type CollisionType int

const (
	CollisionNone CollisionType = iota
	CollisionTile
	CollisionObject
)

type VisualData struct {
	Type  string `json:"type"`
	Color string `json:"color"`
}

type TileDef struct {
	Solid  bool       `json:"solid"`
	Visual VisualData `json:"visual"`
}

type LayerData struct {
	ID            string        `json:"id"`
	Type          string        `json:"type"`
	ZIndex        int           `json:"z_index"`
	CollisionName string        `json:"collision"`
	Collision     CollisionType `json:"-"`
	Grid          [][]string    `json:"grid"`
}

type Rectangle struct {
	X      int `json:"x"`
	Y      int `json:"y"`
	Width  int `json:"width"`
	Height int `json:"height"`
}

type RoomData struct {
	ID     string    `json:"id"`
	Name   string    `json:"name"`
	Bounds Rectangle `json:"bounds"`
}

type ObjectData struct {
	ID         string         `json:"id"`
	Type       string         `json:"type"`
	X          int            `json:"x"`
	Y          int            `json:"y"`
	Properties map[string]any `json:"properties"`
}

type Spawn struct {
	X int `json:"x"`
	Y int `json:"y"`
}

type MapData struct {
	FormatVersion int                   `json:"format_version"`
	ID            string                `json:"id"`
	Name          string                `json:"name"`
	TileSize      int                   `json:"tile_size"`
	Width         int                   `json:"width"`
	Height        int                   `json:"height"`
	Tiles         map[string]TileDef    `json:"tiles"`
	Layers        []LayerData           `json:"layers"`
	LayersByID    map[string]*LayerData `json:"-"`
	Rooms         []RoomData            `json:"rooms"`
	Objects       []ObjectData          `json:"objects"`
	Spawn         Spawn                 `json:"spawn"`
}
