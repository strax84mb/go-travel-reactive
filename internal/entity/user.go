package entity

import "fmt"

type UserRole string

const (
	CommonUserRole UserRole = "USER"
	AdminUserRole  UserRole = "ADMIN"
	AnyUserRole    UserRole = "ANY" // not to be saved to DB
)

type User struct {
	ID       int
	Username string
	Password string
	Salt     []byte
	Role     UserRole
}

type ErrUsernameNotFound struct {
	Username string
}

func (e ErrUsernameNotFound) Error() string {
	return fmt.Sprintf("username %s not found", e.Username)
}

type ErrUsernameTaken struct {
	Username string
}

func (e ErrUsernameTaken) Error() string {
	return fmt.Sprintf("username %s is taken", e.Username)
}
