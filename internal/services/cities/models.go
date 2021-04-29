package cities

import (
	"context"
)

type repository interface {
	GetCityByNameAndCountry(ctx context.Context, city interface{}) (interface{}, error)
	AddCity(ctx context.Context, city interface{}) (interface{}, error)
	UpdateCity(ctx context.Context, city interface{}) (interface{}, error)
	GetCity(ctx context.Context, id interface{}) (interface{}, error)
	GetAllCities(ctx context.Context, nothing interface{}) (interface{}, error)
	DeleteCity(ctx context.Context, id interface{}) (interface{}, error)
}

type cityDto struct {
	ID       int          `json:"id"`
	Name     string       `json:"name"`
	Country  string       `json:"country"`
	Comments []commentDto `json:"comments"`
}

type commentDto struct {
	PosterID int    `json:"posterId"`
	Poster   string `json:"posterUsername"`
	Text     string `json:"text"`
	Created  string `json:"created"`
	Modified string `json:"modified"`
}

type ctxIndex int

const ctxCommentNumIdx ctxIndex = iota + 1
