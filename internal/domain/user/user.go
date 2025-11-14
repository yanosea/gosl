package user

type User struct {
	ID          string
	Name        string
	DisplayName string
	Color       *string // Slack profile color in hex format (e.g., "#9F69E7"), nil if not set
}

func NewUser(id, name, displayName string, color *string) User {
	return User{
		ID:          id,
		Name:        name,
		DisplayName: displayName,
		Color:       color,
	}
}

func (u *User) GetDisplayName() string {
	if u.DisplayName != "" {
		return u.DisplayName
	}
	return u.Name
}

// GetColor returns Slack profile color if set, nil otherwise
func (u *User) GetColor() *string {
	return u.Color
}
