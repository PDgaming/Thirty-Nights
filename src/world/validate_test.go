package world

import (
	"strings"
	"testing"
)

func validObject(id string) ObjectData {
	return ObjectData{
		ID:    id,
		Type:  "door",
		Layer: "structure",
	}
}

func TestValidateObjectPropertiesAcceptsValidFootprints(t *testing.T) {
	tests := []struct {
		name      string
		object    ObjectData
		mapWidth  int
		mapHeight int
	}{
		{
			name:      "legacy object without dimensions",
			object:    validObject("legacy"),
			mapWidth:  4,
			mapHeight: 4,
		},
		{
			name:      "explicit one by one",
			object:    ObjectData{ID: "door", Type: "door", Layer: "structure", Width: 1, Height: 1},
			mapWidth:  4,
			mapHeight: 4,
		},
		{
			name:      "wide object",
			object:    ObjectData{ID: "door", Type: "door", Layer: "structure", X: 1, Width: 2, Height: 1},
			mapWidth:  4,
			mapHeight: 4,
		},
		{
			name:      "tall object",
			object:    ObjectData{ID: "door", Type: "door", Layer: "structure", Y: 1, Width: 1, Height: 2},
			mapWidth:  4,
			mapHeight: 4,
		},
		{
			name:      "exact fit at right edge",
			object:    ObjectData{ID: "door", Type: "door", Layer: "structure", X: 2, Width: 2, Height: 1},
			mapWidth:  4,
			mapHeight: 4,
		},
		{
			name:      "exact fit at bottom edge",
			object:    ObjectData{ID: "door", Type: "door", Layer: "structure", Y: 2, Width: 1, Height: 2},
			mapWidth:  4,
			mapHeight: 4,
		},
		{
			name:      "exact fit in a single tile map",
			object:    validObject("door"),
			mapWidth:  1,
			mapHeight: 1,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if err := ValidateObjectProperties([]ObjectData{tt.object}, tt.mapHeight, tt.mapWidth, map[string]bool{}); err != nil {
				t.Fatalf("ValidateObjectProperties() error = %v, want nil", err)
			}
		})
	}
}

func TestValidateObjectPropertiesRejectsInvalidFootprints(t *testing.T) {
	tests := []struct {
		name      string
		object    ObjectData
		mapWidth  int
		mapHeight int
		contains  string
	}{
		{
			name:      "negative width",
			object:    ObjectData{ID: "door", Type: "door", Layer: "structure", Width: -1, Height: 1},
			mapWidth:  4,
			mapHeight: 4,
			contains:  "width must not be negative",
		},
		{
			name:      "negative height",
			object:    ObjectData{ID: "door", Type: "door", Layer: "structure", Width: 1, Height: -1},
			mapWidth:  4,
			mapHeight: 4,
			contains:  "height must not be negative",
		},
		{
			name:      "origin outside left edge",
			object:    ObjectData{ID: "door", Type: "door", Layer: "structure", X: -1, Width: 1, Height: 1},
			mapWidth:  4,
			mapHeight: 4,
			contains:  "x must be greater than or equal to 0",
		},
		{
			name:      "origin outside top edge",
			object:    ObjectData{ID: "door", Type: "door", Layer: "structure", Y: -1, Width: 1, Height: 1},
			mapWidth:  4,
			mapHeight: 4,
			contains:  "y must be greater than or equal to 0",
		},
		{
			name:      "origin at right edge",
			object:    ObjectData{ID: "door", Type: "door", Layer: "structure", X: 4, Width: 1, Height: 1},
			mapWidth:  4,
			mapHeight: 4,
			contains:  "x must be greater than or equal to 0",
		},
		{
			name:      "origin at bottom edge",
			object:    ObjectData{ID: "door", Type: "door", Layer: "structure", Y: 4, Width: 1, Height: 1},
			mapWidth:  4,
			mapHeight: 4,
			contains:  "y must be greater than or equal to 0",
		},
		{
			name:      "wide footprint overflows right edge",
			object:    ObjectData{ID: "door", Type: "door", Layer: "structure", X: 3, Width: 2, Height: 1},
			mapWidth:  4,
			mapHeight: 4,
			contains:  "footprint width",
		},
		{
			name:      "tall footprint overflows bottom edge",
			object:    ObjectData{ID: "door", Type: "door", Layer: "structure", Y: 3, Width: 1, Height: 2},
			mapWidth:  4,
			mapHeight: 4,
			contains:  "footprint height",
		},
		{
			name:      "footprint wider than the map",
			object:    ObjectData{ID: "door", Type: "door", Layer: "structure", Width: 5, Height: 1},
			mapWidth:  4,
			mapHeight: 4,
			contains:  "footprint width",
		},
		{
			name:      "footprint taller than the map",
			object:    ObjectData{ID: "door", Type: "door", Layer: "structure", Width: 1, Height: 5},
			mapWidth:  4,
			mapHeight: 4,
			contains:  "footprint height",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidateObjectProperties([]ObjectData{tt.object}, tt.mapHeight, tt.mapWidth, map[string]bool{})
			if err == nil {
				t.Fatal("ValidateObjectProperties() error = nil, want an error")
			}
			if !strings.Contains(err.Error(), tt.contains) {
				t.Fatalf("ValidateObjectProperties() error = %q, want it to contain %q", err, tt.contains)
			}
		})
	}
}

func TestValidateObjectPropertiesKeepsExistingChecks(t *testing.T) {
	tests := []struct {
		name     string
		objects  []ObjectData
		contains string
	}{
		{
			name:     "missing ID",
			objects:  []ObjectData{{Type: "door", Layer: "structure"}},
			contains: "object must have an ID",
		},
		{
			name: "duplicate ID",
			objects: []ObjectData{
				{ID: "door", Type: "door", Layer: "structure"},
				{ID: "door", Type: "door", Layer: "structure"},
			},
			contains: "duplicate object ID",
		},
		{
			name:     "missing type",
			objects:  []ObjectData{{ID: "door", Layer: "structure"}},
			contains: "object must have a type",
		},
		{
			name:     "missing layer",
			objects:  []ObjectData{{ID: "door", Type: "door"}},
			contains: "object must have a layer",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidateObjectProperties(tt.objects, 4, 4, map[string]bool{})
			if err == nil {
				t.Fatal("ValidateObjectProperties() error = nil, want an error")
			}
			if !strings.Contains(err.Error(), tt.contains) {
				t.Fatalf("ValidateObjectProperties() error = %q, want it to contain %q", err, tt.contains)
			}
		})
	}
}
