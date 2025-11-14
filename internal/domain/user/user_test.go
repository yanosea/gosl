package user_test

import (
	"testing"

	"github.com/yanosea/gosl/internal/domain/user"
)

func TestNewUser(t *testing.T) {
	purple := "#9F69E7"

	tests := []struct {
		name        string
		id          string
		userName    string
		displayName string
		color       *string
		expected    user.User
	}{
		{
			name:        "user with display name and color",
			id:          "U123456",
			userName:    "john.doe",
			displayName: "John Doe",
			color:       &purple,
			expected: user.User{
				ID:          "U123456",
				Name:        "john.doe",
				DisplayName: "John Doe",
				Color:       &purple,
			},
		},
		{
			name:        "user without display name but with color",
			id:          "U789012",
			userName:    "jane.smith",
			displayName: "",
			color:       &purple,
			expected: user.User{
				ID:          "U789012",
				Name:        "jane.smith",
				DisplayName: "",
				Color:       &purple,
			},
		},
		{
			name:        "user without color (nil)",
			id:          "U345678",
			userName:    "bob.jones",
			displayName: "Bob Jones",
			color:       nil,
			expected: user.User{
				ID:          "U345678",
				Name:        "bob.jones",
				DisplayName: "Bob Jones",
				Color:       nil,
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			u := user.NewUser(tt.id, tt.userName, tt.displayName, tt.color)

			if u.ID != tt.expected.ID {
				t.Errorf("ID = %v, want %v", u.ID, tt.expected.ID)
			}
			if u.Name != tt.expected.Name {
				t.Errorf("Name = %v, want %v", u.Name, tt.expected.Name)
			}
			if u.DisplayName != tt.expected.DisplayName {
				t.Errorf("DisplayName = %v, want %v", u.DisplayName, tt.expected.DisplayName)
			}

			// Check Color field
			if (u.Color == nil) != (tt.expected.Color == nil) {
				t.Errorf("Color nil status mismatch: got %v, want %v", u.Color == nil, tt.expected.Color == nil)
			}
			if u.Color != nil && tt.expected.Color != nil && *u.Color != *tt.expected.Color {
				t.Errorf("Color = %v, want %v", *u.Color, *tt.expected.Color)
			}
		})
	}
}

func TestUser_GetDisplayName(t *testing.T) {
	tests := []struct {
		name        string
		userName    string
		displayName string
		expected    string
	}{
		{
			name:        "prefer display name",
			userName:    "john.doe",
			displayName: "John Doe",
			expected:    "John Doe",
		},
		{
			name:        "fallback to username",
			userName:    "jane.smith",
			displayName: "",
			expected:    "jane.smith",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			u := user.User{
				ID:          "U123456",
				Name:        tt.userName,
				DisplayName: tt.displayName,
			}

			result := u.GetDisplayName()
			if result != tt.expected {
				t.Errorf("GetDisplayName() = %v, want %v", result, tt.expected)
			}
		})
	}
}

func TestUser_GetColor(t *testing.T) {
	purple := "#9F69E7"
	blue := "#0000FF"

	tests := []struct {
		name     string
		color    *string
		expected *string
	}{
		{
			name:     "user with Slack profile color",
			color:    &purple,
			expected: &purple,
		},
		{
			name:     "user with different color",
			color:    &blue,
			expected: &blue,
		},
		{
			name:     "user without color (nil)",
			color:    nil,
			expected: nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			u := user.User{
				ID:    "U123456",
				Name:  "test.user",
				Color: tt.color,
			}

			result := u.GetColor()

			// Check if both are nil or both are non-nil
			if (result == nil) != (tt.expected == nil) {
				t.Errorf("GetColor() nil status mismatch: got %v, want %v", result == nil, tt.expected == nil)
			}

			// If both are non-nil, check values
			if result != nil && tt.expected != nil && *result != *tt.expected {
				t.Errorf("GetColor() = %v, want %v", *result, *tt.expected)
			}
		})
	}
}
