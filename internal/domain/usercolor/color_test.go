package usercolor

import (
	"testing"
)

func TestColor_ToHex(t *testing.T) {
	tests := []struct {
		name  string
		color Color
		want  string
	}{
		{
			name:  "Red color",
			color: Color{R: 255, G: 0, B: 0},
			want:  "#FF0000",
		},
		{
			name:  "Green color",
			color: Color{R: 0, G: 255, B: 0},
			want:  "#00FF00",
		},
		{
			name:  "Blue color",
			color: Color{R: 0, G: 0, B: 255},
			want:  "#0000FF",
		},
		{
			name:  "White color",
			color: Color{R: 255, G: 255, B: 255},
			want:  "#FFFFFF",
		},
		{
			name:  "Black color",
			color: Color{R: 0, G: 0, B: 0},
			want:  "#000000",
		},
		{
			name:  "Purple color",
			color: Color{R: 159, G: 105, B: 231},
			want:  "#9F69E7",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := tt.color.ToHex()
			if got != tt.want {
				t.Errorf("Color.ToHex() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestFromHex(t *testing.T) {
	tests := []struct {
		name    string
		hex     string
		want    Color
		wantErr bool
	}{
		{
			name: "Valid red color",
			hex:  "#FF0000",
			want: Color{R: 255, G: 0, B: 0},
		},
		{
			name: "Valid green color",
			hex:  "#00FF00",
			want: Color{R: 0, G: 255, B: 0},
		},
		{
			name: "Valid blue color",
			hex:  "#0000FF",
			want: Color{R: 0, G: 0, B: 255},
		},
		{
			name: "Valid purple color",
			hex:  "#9F69E7",
			want: Color{R: 159, G: 105, B: 231},
		},
		{
			name: "Valid color without hash",
			hex:  "FF5733",
			want: Color{R: 255, G: 87, B: 51},
		},
		{
			name:    "Invalid hex format - too short",
			hex:     "#FF",
			wantErr: true,
		},
		{
			name:    "Invalid hex format - non-hex characters",
			hex:     "#GGHHII",
			wantErr: true,
		},
		{
			name:    "Invalid hex format - empty string",
			hex:     "",
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := FromHex(tt.hex)
			if (err != nil) != tt.wantErr {
				t.Errorf("FromHex() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !tt.wantErr {
				if got.R != tt.want.R || got.G != tt.want.G || got.B != tt.want.B {
					t.Errorf("FromHex() = {R:%d, G:%d, B:%d}, want {R:%d, G:%d, B:%d}",
						got.R, got.G, got.B, tt.want.R, tt.want.G, tt.want.B)
				}
			}
		})
	}
}

func TestToHex_FromHex_RoundTrip(t *testing.T) {
	// Test that ToHex and FromHex are inverse operations
	tests := []Color{
		{R: 255, G: 0, B: 0},
		{R: 0, G: 255, B: 0},
		{R: 0, G: 0, B: 255},
		{R: 159, G: 105, B: 231},
		{R: 128, G: 128, B: 128},
	}

	for _, original := range tests {
		t.Run(original.ToHex(), func(t *testing.T) {
			// Convert to hex and back
			hexStr := original.ToHex()
			roundTrip, err := FromHex(hexStr)
			if err != nil {
				t.Fatalf("FromHex(%s) unexpected error: %v", hexStr, err)
			}

			// Check if we got the same color back
			if roundTrip.R != original.R || roundTrip.G != original.G || roundTrip.B != original.B {
				t.Errorf("Round trip failed: original {R:%d, G:%d, B:%d}, got {R:%d, G:%d, B:%d}",
					original.R, original.G, original.B, roundTrip.R, roundTrip.G, roundTrip.B)
			}
		})
	}
}
