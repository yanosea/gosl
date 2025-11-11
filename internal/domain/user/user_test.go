package user_test

import (
	"testing"

	"github.com/yanosea/gosl/internal/domain/user"
)

func TestNewUser(t *testing.T) {
	tests := []struct {
		name        string
		id          string
		userName    string
		displayName string
		expected    user.User
	}{
		{
			name:        "user with display name",
			id:          "U123456",
			userName:    "john.doe",
			displayName: "John Doe",
			expected: user.User{
				ID:          "U123456",
				Name:        "john.doe",
				DisplayName: "John Doe",
			},
		},
		{
			name:        "user without display name",
			id:          "U789012",
			userName:    "jane.smith",
			displayName: "",
			expected: user.User{
				ID:          "U789012",
				Name:        "jane.smith",
				DisplayName: "",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			u := user.NewUser(tt.id, tt.userName, tt.displayName)

			if u.ID != tt.expected.ID {
				t.Errorf("ID = %v, want %v", u.ID, tt.expected.ID)
			}
			if u.Name != tt.expected.Name {
				t.Errorf("Name = %v, want %v", u.Name, tt.expected.Name)
			}
			if u.DisplayName != tt.expected.DisplayName {
				t.Errorf("DisplayName = %v, want %v", u.DisplayName, tt.expected.DisplayName)
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
