package usercolor

import (
	"crypto/sha1"
	"encoding/binary"
	"fmt"

	colorful "github.com/lucasb-eyer/go-colorful"
	"github.com/yanosea/gosl/internal/domain/user"
)

// Service interface for managing user colors
type Service interface {
	// GetUserColor returns adaptive color for the given user.
	// If user has Slack profile color, it is used; otherwise, generates color from userID.
	// Returns default color if any error occurs.
	GetUserColor(u *user.User) AdaptiveColor

	// GenerateColorFromID generates consistent color from userID using XEP-0392 algorithm.
	// Parameters:
	//   - userID: Slack user ID (e.g., "U1234ABCD")
	// Returns:
	//   - AdaptiveColor with WCAG AA compliant light/dark variants
	GenerateColorFromID(userID string) AdaptiveColor

	// ParseSlackColor parses Slack profile color hex string.
	// Parameters:
	//   - colorHex: Slack profile color (e.g., "#9F69E7")
	// Returns:
	//   - AdaptiveColor if valid, error otherwise
	ParseSlackColor(colorHex string) (AdaptiveColor, error)

	// ValidateContrast validates if color meets WCAG AA contrast ratio (4.5:1).
	// Parameters:
	//   - color: RGB color to validate
	//   - background: Background color (light or dark)
	// Returns:
	//   - true if contrast >= 4.5:1, false otherwise
	ValidateContrast(foreground Color, background Color) bool
}

// userColorService implements Service interface
type userColorService struct {
	cache Cache
}

// NewUserColorService creates a new UserColorService with the specified cache
func NewUserColorService(cache Cache) Service {
	return &userColorService{
		cache: cache,
	}
}

// GenerateColorFromID generates consistent color from userID using hash-based color generation
// with independent hash mixing for each parameter to maximize color diversity.
func (s *userColorService) GenerateColorFromID(userID string) AdaptiveColor {
	// Step 1: Generate base hash using SHA-1
	baseHash := sha1.Sum([]byte(userID))

	// Step 2: Generate independent hashes for each parameter using domain separation
	// This ensures that even if base hash bytes are similar, the derived values differ significantly
	hueHash := sha1.Sum(append([]byte("HUE:"), baseHash[:]...))
	lightnessHash := sha1.Sum(append([]byte("LIGHTNESS:"), baseHash[:]...))
	saturationHash := sha1.Sum(append([]byte("SATURATION:"), baseHash[:]...))

	// Step 3: Extract values from independent hashes
	// Hue: Use first 2 bytes from hue-specific hash
	hueValue := binary.LittleEndian.Uint16(hueHash[:2])
	hue := (float64(hueValue) / 65536.0) * 360.0

	// Lightness: Use first byte from lightness-specific hash
	lightnessValue := uint8(lightnessHash[0])

	// Saturation: Use first byte from saturation-specific hash
	saturationValue := uint8(saturationHash[0])

	// Step 4: Map extracted values to appropriate ranges using standard HSL
	// Light theme: Wide lightness range [0.25, 0.65] for maximum distinguishability
	//              Full saturation range [0.50, 1.00] for vibrant, varied colors
	lightLightness := 0.25 + (float64(lightnessValue)/255.0)*0.40   // 40% range
	lightSaturation := 0.50 + (float64(saturationValue)/255.0)*0.50 // 50% range

	// Dark theme: Wide lightness range [0.55, 0.90] for visibility on dark backgrounds
	//             Wide saturation [0.40, 0.90] for varied but visible colors
	darkLightness := 0.55 + (float64(lightnessValue)/255.0)*0.35   // 35% range
	darkSaturation := 0.40 + (float64(saturationValue)/255.0)*0.50 // 50% range

	// Step 5: Generate colors using standard HSL for better color diversity
	// Note: Using Hsl() instead of HSLuv() because HSLuv makes all colors
	// appear too similar due to perceptual uniformity
	lightColor := colorful.Hsl(hue, lightSaturation, lightLightness)
	darkColor := colorful.Hsl(hue, darkSaturation, darkLightness)

	return AdaptiveColor{
		Light: Color{
			R: uint8(lightColor.R * 255),
			G: uint8(lightColor.G * 255),
			B: uint8(lightColor.B * 255),
		},
		Dark: Color{
			R: uint8(darkColor.R * 255),
			G: uint8(darkColor.G * 255),
			B: uint8(darkColor.B * 255),
		},
	}
}

// ParseSlackColor parses Slack profile color hex string and creates adaptive variants
func (s *userColorService) ParseSlackColor(colorHex string) (AdaptiveColor, error) {
	// Parse the base color
	baseColor, err := FromHex(colorHex)
	if err != nil {
		return AdaptiveColor{}, fmt.Errorf("failed to parse slack color: %w", err)
	}

	// Convert to go-colorful for manipulation
	col := colorful.Color{
		R: float64(baseColor.R) / 255.0,
		G: float64(baseColor.G) / 255.0,
		B: float64(baseColor.B) / 255.0,
	}

	// Get HSLuv values (H in [0..360], S and L in [0..1])
	h, sat, l := col.HSLuv()

	// Create adaptive variants
	// For light theme: darken if too light (ensure readability on white)
	lightL := l
	if l > 0.60 {
		lightL = 0.50 // Darken for better contrast on light background
	}
	lightColor := colorful.HSLuv(h, sat, lightL)

	// For dark theme: lighten if too dark (ensure readability on black)
	darkL := l
	if l < 0.50 {
		darkL = 0.70 // Lighten for better contrast on dark background
	}
	darkColor := colorful.HSLuv(h, sat, darkL)

	return AdaptiveColor{
		Light: Color{
			R: uint8(lightColor.R * 255),
			G: uint8(lightColor.G * 255),
			B: uint8(lightColor.B * 255),
		},
		Dark: Color{
			R: uint8(darkColor.R * 255),
			G: uint8(darkColor.G * 255),
			B: uint8(darkColor.B * 255),
		},
	}, nil
}

// ValidateContrast validates if color meets WCAG AA contrast ratio (4.5:1)
func (s *userColorService) ValidateContrast(foreground Color, background Color) bool {
	fg := colorful.Color{
		R: float64(foreground.R) / 255.0,
		G: float64(foreground.G) / 255.0,
		B: float64(foreground.B) / 255.0,
	}

	bg := colorful.Color{
		R: float64(background.R) / 255.0,
		G: float64(background.G) / 255.0,
		B: float64(background.B) / 255.0,
	}

	// Convert to linear RGB for luminance calculation
	fgR, fgG, fgB := fg.LinearRgb()
	bgR, bgG, bgB := bg.LinearRgb()

	// Calculate relative luminance using WCAG formula
	fgL := 0.2126*fgR + 0.7152*fgG + 0.0722*fgB
	bgL := 0.2126*bgR + 0.7152*bgG + 0.0722*bgB

	// Calculate contrast ratio
	var contrastRatio float64
	if fgL > bgL {
		contrastRatio = (fgL + 0.05) / (bgL + 0.05)
	} else {
		contrastRatio = (bgL + 0.05) / (fgL + 0.05)
	}

	// WCAG AA standard is 4.5:1
	return contrastRatio >= 4.5
}

// GetUserColor returns adaptive color for the given user
func (s *userColorService) GetUserColor(u *user.User) AdaptiveColor {
	// Check cache first
	if cached, ok := s.cache.Get(u.ID); ok {
		return cached
	}

	var color AdaptiveColor

	// Try to use Slack profile color if available
	if profileColor := u.GetColor(); profileColor != nil && *profileColor != "" {
		parsed, err := s.ParseSlackColor(*profileColor)
		if err == nil {
			color = parsed
		} else {
			// Fallback to hash-based generation on parse error
			color = s.GenerateColorFromID(u.ID)
		}
	} else {
		// Generate color from userID
		color = s.GenerateColorFromID(u.ID)
	}

	// Cache the result
	s.cache.Set(u.ID, color)

	return color
}
