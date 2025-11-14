package usercolor

import (
	"fmt"

	colorful "github.com/lucasb-eyer/go-colorful"
)

// Color represents RGB color values
type Color struct {
	R uint8
	G uint8
	B uint8
}

// AdaptiveColor represents light/dark theme color variants
type AdaptiveColor struct {
	Light Color
	Dark  Color
}

// ToHex converts Color to hex string (e.g., "#FF5733")
func (c Color) ToHex() string {
	return fmt.Sprintf("#%02X%02X%02X", c.R, c.G, c.B)
}

// FromHex parses hex string to Color
func FromHex(hex string) (Color, error) {
	// Add # prefix if not present
	if len(hex) > 0 && hex[0] != '#' {
		hex = "#" + hex
	}

	col, err := colorful.Hex(hex)
	if err != nil {
		return Color{}, fmt.Errorf("invalid hex color: %w", err)
	}

	return Color{
		R: uint8(col.R * 255),
		G: uint8(col.G * 255),
		B: uint8(col.B * 255),
	}, nil
}
