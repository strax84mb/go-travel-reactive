package entity

import "errors"

type City struct {
	ID      int
	Name    string
	Country string
}

var (
	ErrCityNotFound = errors.New("city not found")
	ErrCityExists   = errors.New("city with same name in same country already exists")
)
