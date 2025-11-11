package user

type User struct {
	ID          string
	Name        string
	DisplayName string
}

func NewUser(id, name, displayName string) User {
	return User{
		ID:          id,
		Name:        name,
		DisplayName: displayName,
	}
}

func (u *User) GetDisplayName() string {
	if u.DisplayName != "" {
		return u.DisplayName
	}
	return u.Name
}
