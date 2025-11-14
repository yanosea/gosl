package usercolor

import (
	"testing"

	"github.com/yanosea/gosl/internal/domain/user"
)

func TestGenerateColorFromID_Deterministic(t *testing.T) {
	cache := NewUserColorCache(10)
	service := NewUserColorService(cache)

	userID := "U1234ABCD"

	// Generate color multiple times for the same userID
	color1 := service.GenerateColorFromID(userID)
	color2 := service.GenerateColorFromID(userID)
	color3 := service.GenerateColorFromID(userID)

	// All should be identical (deterministic)
	if color1.Light != color2.Light || color1.Dark != color2.Dark {
		t.Error("GenerateColorFromID() not deterministic: color1 != color2")
	}
	if color1.Light != color3.Light || color1.Dark != color3.Dark {
		t.Error("GenerateColorFromID() not deterministic: color1 != color3")
	}
}

func TestGenerateColorFromID_DifferentUsersDifferentColors(t *testing.T) {
	cache := NewUserColorCache(10)
	service := NewUserColorService(cache)

	user1 := "U1111111"
	user2 := "U2222222"
	user3 := "U3333333"

	color1 := service.GenerateColorFromID(user1)
	color2 := service.GenerateColorFromID(user2)
	color3 := service.GenerateColorFromID(user3)

	// Different users should (very likely) have different colors
	// Check at least one component is different in either Light or Dark variant
	colorsIdentical12 := (color1.Light.R == color2.Light.R &&
		color1.Light.G == color2.Light.G &&
		color1.Light.B == color2.Light.B &&
		color1.Dark.R == color2.Dark.R &&
		color1.Dark.G == color2.Dark.G &&
		color1.Dark.B == color2.Dark.B)

	colorsIdentical13 := (color1.Light.R == color3.Light.R &&
		color1.Light.G == color3.Light.G &&
		color1.Light.B == color3.Light.B &&
		color1.Dark.R == color3.Dark.R &&
		color1.Dark.G == color3.Dark.G &&
		color1.Dark.B == color3.Dark.B)

	colorsIdentical23 := (color2.Light.R == color3.Light.R &&
		color2.Light.G == color3.Light.G &&
		color2.Light.B == color3.Light.B &&
		color2.Dark.R == color3.Dark.R &&
		color2.Dark.G == color3.Dark.G &&
		color2.Dark.B == color3.Dark.B)

	if colorsIdentical12 {
		t.Errorf("user1 and user2 generated identical colors: %+v", color1)
	}
	if colorsIdentical13 {
		t.Errorf("user1 and user3 generated identical colors: %+v", color1)
	}
	if colorsIdentical23 {
		t.Errorf("user2 and user3 generated identical colors: %+v", color2)
	}
}

func TestGenerateColorFromID_HasLightAndDarkVariants(t *testing.T) {
	cache := NewUserColorCache(10)
	service := NewUserColorService(cache)

	userID := "U1234ABCD"
	color := service.GenerateColorFromID(userID)

	// Both variants should have valid RGB values (not zero)
	lightIsZero := color.Light.R == 0 && color.Light.G == 0 && color.Light.B == 0
	darkIsZero := color.Dark.R == 0 && color.Dark.G == 0 && color.Dark.B == 0

	if lightIsZero {
		t.Error("Light variant should not be pure black (likely uninitialized)")
	}
	if darkIsZero {
		t.Error("Dark variant should not be pure black (likely uninitialized)")
	}

	// The variants should be different in at least one component (R, G, or B)
	// Note: Due to HSLuv with different lightness/chroma, they should differ
	sameR := color.Light.R == color.Dark.R
	sameG := color.Light.G == color.Dark.G
	sameB := color.Light.B == color.Dark.B

	if sameR && sameG && sameB {
		t.Errorf("Light and Dark variants are identical: Light=%+v, Dark=%+v", color.Light, color.Dark)
	}
}

func TestParseSlackColor_ValidHex(t *testing.T) {
	cache := NewUserColorCache(10)
	service := NewUserColorService(cache)

	tests := []struct {
		name     string
		colorHex string
		wantErr  bool
	}{
		{
			name:     "valid purple color",
			colorHex: "#9F69E7",
			wantErr:  false,
		},
		{
			name:     "valid red color",
			colorHex: "#FF0000",
			wantErr:  false,
		},
		{
			name:     "valid color without hash",
			colorHex: "00FF00",
			wantErr:  false,
		},
		{
			name:     "invalid hex - too short",
			colorHex: "#FF",
			wantErr:  true,
		},
		{
			name:     "invalid hex - non-hex characters",
			colorHex: "#GGHHII",
			wantErr:  true,
		},
		{
			name:     "empty string",
			colorHex: "",
			wantErr:  true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			color, err := service.ParseSlackColor(tt.colorHex)

			if (err != nil) != tt.wantErr {
				t.Errorf("ParseSlackColor() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if !tt.wantErr {
				// Valid colors should have light and dark variants
				if color.Light.R == 0 && color.Light.G == 0 && color.Light.B == 0 {
					t.Error("Light variant should not be pure black")
				}
			}
		})
	}
}

func TestValidateContrast_WCAGCompliant(t *testing.T) {
	cache := NewUserColorCache(10)
	service := NewUserColorService(cache)

	white := Color{R: 255, G: 255, B: 255}
	black := Color{R: 0, G: 0, B: 0}
	gray := Color{R: 128, G: 128, B: 128}
	lightGray := Color{R: 200, G: 200, B: 200}

	tests := []struct {
		name       string
		foreground Color
		background Color
		want       bool
	}{
		{
			name:       "white on black - high contrast",
			foreground: white,
			background: black,
			want:       true, // Contrast ratio is 21:1
		},
		{
			name:       "black on white - high contrast",
			foreground: black,
			background: white,
			want:       true, // Contrast ratio is 21:1
		},
		{
			name:       "gray on white - low contrast",
			foreground: gray,
			background: white,
			want:       false, // Contrast ratio is ~3.9:1 (below 4.5:1)
		},
		{
			name:       "light gray on white - very low contrast",
			foreground: lightGray,
			background: white,
			want:       false, // Contrast ratio is ~1.6:1
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := service.ValidateContrast(tt.foreground, tt.background)
			if result != tt.want {
				t.Errorf("ValidateContrast() = %v, want %v", result, tt.want)
			}
		})
	}
}

func TestGetUserColor_WithSlackProfileColor(t *testing.T) {
	cache := NewUserColorCache(10)
	service := NewUserColorService(cache)

	purple := "#9F69E7"
	u := &user.User{
		ID:          "U12345",
		Name:        "test.user",
		DisplayName: "Test User",
		Color:       &purple,
	}

	color := service.GetUserColor(u)

	// Should return a valid color (not zero values)
	if color.Light.R == 0 && color.Light.G == 0 && color.Light.B == 0 {
		t.Error("GetUserColor() returned zero color for user with profile color")
	}

	// Verify it's cached
	if cache.Len() != 1 {
		t.Errorf("cache.Len() = %v, want 1 (should be cached)", cache.Len())
	}
}

func TestGetUserColor_WithoutSlackProfileColor(t *testing.T) {
	cache := NewUserColorCache(10)
	service := NewUserColorService(cache)

	u := &user.User{
		ID:          "U12345",
		Name:        "test.user",
		DisplayName: "Test User",
		Color:       nil,
	}

	color := service.GetUserColor(u)

	// Should generate a color from userID
	if color.Light.R == 0 && color.Light.G == 0 && color.Light.B == 0 {
		t.Error("GetUserColor() returned zero color for user without profile color")
	}

	// Verify it's cached
	if cache.Len() != 1 {
		t.Errorf("cache.Len() = %v, want 1 (should be cached)", cache.Len())
	}

	// Second call should use cache
	color2 := service.GetUserColor(u)
	if color.Light != color2.Light || color.Dark != color2.Dark {
		t.Error("GetUserColor() should return cached color on second call")
	}

	// Cache length should still be 1
	if cache.Len() != 1 {
		t.Errorf("cache.Len() = %v, want 1 (should not duplicate)", cache.Len())
	}
}

func TestGetUserColor_UsesCache(t *testing.T) {
	cache := NewUserColorCache(10)
	service := NewUserColorService(cache)

	u := &user.User{
		ID:          "U12345",
		Name:        "test.user",
		DisplayName: "Test User",
		Color:       nil,
	}

	// First call - miss cache
	color1 := service.GetUserColor(u)

	// Second call - should hit cache
	color2 := service.GetUserColor(u)

	// Colors should be identical
	if color1.Light != color2.Light || color1.Dark != color2.Dark {
		t.Error("GetUserColor() should return same color from cache")
	}

	// Cache should only have one entry
	if cache.Len() != 1 {
		t.Errorf("cache.Len() = %v, want 1", cache.Len())
	}
}
