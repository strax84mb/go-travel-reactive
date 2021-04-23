package entity

import (
	"errors"
)

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

var (
	ErrUsernameNotFound = errors.New("username not found")
	ErrUsernameTaken    = errors.New("username is already taken")
)
