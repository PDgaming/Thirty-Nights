package world

import "fmt"

// Footprint is an object's size in tiles.
type Footprint struct {
	Width  int
	Height int
}

// Footprint returns the object's normalized footprint in tiles.
// Missing or non-positive dimensions collapse to one tile so zero-valued
// object data keeps the legacy 1x1 behavior. Validation reports negative
// values before the loader relies on this normalization.
func (o ObjectData) Footprint() Footprint {
	width := o.Width
	if width < 1 {
		width = 1
	}

	height := o.Height
	if height < 1 {
		height = 1
	}

	return Footprint{Width: width, Height: height}
}

// Pixels converts a footprint to pixel dimensions for the given tile size.
func (f Footprint) Pixels(tileSize int) (width, height int) {
	return f.Width * tileSize, f.Height * tileSize
}

// ValidateFootprint checks the raw width and height fields.
// Zero is valid and means the default one tile; negative values are not.
func (o ObjectData) ValidateFootprint() error {
	if o.Width < 0 {
		return fmt.Errorf("object %q: width must not be negative (got %d)", o.ID, o.Width)
	}
	if o.Height < 0 {
		return fmt.Errorf("object %q: height must not be negative (got %d)", o.ID, o.Height)
	}
	return nil
}
