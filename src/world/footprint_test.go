package world

import (
	"strings"
	"testing"
)

func TestObjectFootprint(t *testing.T) {
	tests := []struct {
		name   string
		object ObjectData
		want   Footprint
	}{
		{
			name:   "missing dimensions default to one tile",
			object: ObjectData{},
			want:   Footprint{Width: 1, Height: 1},
		},
		{
			name:   "zero dimensions default to one tile",
			object: ObjectData{Width: 0, Height: 0},
			want:   Footprint{Width: 1, Height: 1},
		},
		{
			name:   "one by one",
			object: ObjectData{Width: 1, Height: 1},
			want:   Footprint{Width: 1, Height: 1},
		},
		{
			name:   "wide",
			object: ObjectData{Width: 2, Height: 1},
			want:   Footprint{Width: 2, Height: 1},
		},
		{
			name:   "tall",
			object: ObjectData{Width: 1, Height: 2},
			want:   Footprint{Width: 1, Height: 2},
		},
		{
			name:   "rectangle",
			object: ObjectData{Width: 3, Height: 4},
			want:   Footprint{Width: 3, Height: 4},
		},
		{
			name:   "negative width clamps to one",
			object: ObjectData{Width: -2, Height: 2},
			want:   Footprint{Width: 1, Height: 2},
		},
		{
			name:   "negative height clamps to one",
			object: ObjectData{Width: 2, Height: -2},
			want:   Footprint{Width: 2, Height: 1},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.object.Footprint(); got != tt.want {
				t.Fatalf("Footprint() = %+v, want %+v", got, tt.want)
			}
		})
	}
}

func TestFootprintPixels(t *testing.T) {
	tests := []struct {
		name                  string
		footprint             Footprint
		tileSize              int
		wantWidth, wantHeight int
	}{
		{name: "one tile at 16", footprint: Footprint{Width: 1, Height: 1}, tileSize: 16, wantWidth: 16, wantHeight: 16},
		{name: "wide at 16", footprint: Footprint{Width: 2, Height: 1}, tileSize: 16, wantWidth: 32, wantHeight: 16},
		{name: "tall at 8", footprint: Footprint{Width: 1, Height: 2}, tileSize: 8, wantWidth: 8, wantHeight: 16},
		{name: "rectangle at 32", footprint: Footprint{Width: 3, Height: 4}, tileSize: 32, wantWidth: 96, wantHeight: 128},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotWidth, gotHeight := tt.footprint.Pixels(tt.tileSize)
			if gotWidth != tt.wantWidth || gotHeight != tt.wantHeight {
				t.Fatalf("Pixels(%d) = %dx%d, want %dx%d", tt.tileSize, gotWidth, gotHeight, tt.wantWidth, tt.wantHeight)
			}
		})
	}
}

func TestObjectValidateFootprint(t *testing.T) {
	tests := []struct {
		name     string
		object   ObjectData
		wantErr  bool
		contains string
	}{
		{name: "missing dimensions are valid", object: ObjectData{ID: "door"}},
		{name: "zero dimensions are valid", object: ObjectData{ID: "door", Width: 0, Height: 0}},
		{name: "one by one is valid", object: ObjectData{ID: "door", Width: 1, Height: 1}},
		{name: "rectangle is valid", object: ObjectData{ID: "door", Width: 3, Height: 2}},
		{name: "negative width is invalid", object: ObjectData{ID: "door", Width: -1, Height: 1}, wantErr: true, contains: "width"},
		{name: "negative height is invalid", object: ObjectData{ID: "door", Width: 1, Height: -1}, wantErr: true, contains: "height"},
		{name: "both negative reports width first", object: ObjectData{ID: "door", Width: -1, Height: -1}, wantErr: true, contains: "width"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.object.ValidateFootprint()
			if !tt.wantErr {
				if err != nil {
					t.Fatalf("ValidateFootprint() error = %v, want nil", err)
				}
				return
			}
			if err == nil {
				t.Fatal("ValidateFootprint() error = nil, want an error")
			}
			if !strings.Contains(err.Error(), tt.contains) {
				t.Fatalf("ValidateFootprint() error = %q, want it to contain %q", err, tt.contains)
			}
		})
	}
}
