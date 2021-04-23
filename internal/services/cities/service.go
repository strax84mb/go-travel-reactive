package cities

import (
	"context"
	"errors"
	"fmt"

	"github.com/reactivex/rxgo/v2"
	"github.com/strax84mb/go-travel-reactive/internal/app"
	"github.com/strax84mb/go-travel-reactive/internal/entity"
)

type cityService struct {
	repo   repository
	logger app.Logger
}

func NewCityService(repo repository, logger app.Logger) *cityService {
	return &cityService{
		repo:   repo,
		logger: logger,
	}
}

func (c *cityService) cityToDto(numberOfComments int) func(context.Context, interface{}) (interface{}, error) {
	return func(ctx context.Context, item interface{}) (interface{}, error) {
		city := item.(entity.City)

		dto := cityDto{
			ID:      city.ID,
			Name:    city.Name,
			Country: city.Country,
		}

		return dto, nil
	}
}

func (c *cityService) GetCity(ctx context.Context, id, numberOfComments int) (cityDto, error) {
	ctx = app.ContextWithValue(ctx, "function", "cityService.GetCity")

	item := <-rxgo.JustItem(id, rxgo.WithContext(context.WithValue(ctx, ctxRepoIdx, c.repo))).
		Map(func(ctx context.Context, id interface{}) (interface{}, error) {
			return ctx.Value(ctxRepoIdx).(repository).GetCity(id.(int))
		}).Map(c.cityToDto(numberOfComments)).Observe()
	if item.Error() {
		c.logger.Error(app.ContextWithError(ctx, item.E), "could not get city with ID %d", id)
		return cityDto{}, fmt.Errorf("get city failed: %w", item.E)
	}

	return item.V.(cityDto), nil
}

func (c *cityService) ListAllCities(ctx context.Context, numberOfComments int) ([]cityDto, error) {
	ctx = app.ContextWithValue(ctx, "function", "cityService.ListAllCities")

	obs := rxgo.Just(c.repo)().Map(func(ctx context.Context, item interface{}) (interface{}, error) {
		return item.(repository).GetAllCities()
	}).FlatMap(func(i rxgo.Item) rxgo.Observable {
		cities := i.V.([]entity.City)
		result := make([]interface{}, len(cities))

		for i, v := range cities {
			result[i] = v
		}

		return rxgo.Just(result...)()
	}).Map(c.cityToDto(numberOfComments), rxgo.WithPool(3))

	var list []cityDto

	for dto := range obs.Observe() {
		if dto.Error() {
			c.logger.Error(app.ContextWithError(ctx, dto.E), "could not list all cities")
			return nil, fmt.Errorf("could not list all cities: %w", dto.E)
		}

		list = append(list, dto.V.(cityDto))
	}

	return list, nil
}

func (c *cityService) AddCity(ctx context.Context, name, country string) (int, error) {
	ctx = app.ContextWithValue(ctx, "function", "cityService.AddCity")
	city := entity.City{
		Name:    name,
		Country: country,
	}

	item := <-rxgo.JustItem(city, rxgo.WithContext(context.WithValue(ctx, ctxRepoIdx, c.repo))).
		Map(func(ctx context.Context, item interface{}) (interface{}, error) {
			city := item.(entity.City)

			_, err := ctx.Value(ctxRepoIdx).(repository).GetCityByNameAndCountry(city.Name, city.Country)
			if err != nil {
				if errors.Is(err, entity.ErrCityNotFound) {
					return city, nil
				}

				return city, err
			}

			return city, errors.New("city with same name and country already exists")
		}).
		Map(func(ctx context.Context, item interface{}) (interface{}, error) {
			city := item.(entity.City)
			return ctx.Value(ctxRepoIdx).(repository).AddCity(city.Name, city.Country)
		}).Observe()
	if item.Error() {
		c.logger.Error(app.ContextWithError(ctx, item.E), "could not add city")

		return 0, fmt.Errorf("could not add city: %w", item.E)
	}

	return item.V.(int), nil
}

func (c *cityService) UpdateCity(ctx context.Context, id int, name, country string) error {
	ctx = app.ContextWithValue(ctx, "function", "cityService.UpdateCity")
	city := entity.City{
		ID:      id,
		Name:    name,
		Country: country,
	}

	item := <-rxgo.JustItem(city, rxgo.WithContext(context.WithValue(ctx, ctxRepoIdx, c.repo))).
		Map(func(ctx context.Context, item interface{}) (interface{}, error) {
			return item, ctx.Value(ctxRepoIdx).(repository).UpdateCity(item.(entity.City))
		}).Observe()
	if item.Error() {
		c.logger.Error(app.ContextWithError(ctx, item.E), "could not update city")
		return fmt.Errorf("could not update city: %w", item.E)
	}

	return nil
}
