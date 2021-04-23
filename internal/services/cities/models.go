package cities

import "github.com/strax84mb/go-travel-reactive/internal/entity"

type repository interface {
	GetCityByNameAndCountry(name, country string) (entity.City, error)
	AddCity(name, country string) (int, error)
	UpdateCity(city entity.City) error
	GetCity(id int) (entity.City, error)
	GetAllCities() ([]entity.City, error)
}

type cityDto struct {
	ID      int    `json:"id"`
	Name    string `json:"name"`
	Country string `json:"country"`
	// TODO add comments
}

type ctxIndex int

const ctxRepoIdx ctxIndex = iota + 1
