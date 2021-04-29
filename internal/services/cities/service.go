package cities

import (
	"context"
	"errors"
	"fmt"
	"time"

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

func cityToDto(ctx context.Context, item interface{}) (interface{}, error) {
	input := item.(entity.GetCityCommentsOutput)
	city := cityDto{
		ID:       input.City.ID,
		Name:     input.City.Name,
		Country:  input.City.Country,
		Comments: make([]commentDto, len(input.Comments)),
	}

	for i, v := range input.Comments {
		city.Comments[i] = commentDto{
			PosterID: v.Comment.PosterID,
			Poster:   v.PosterName,
			Text:     v.Comment.Text,
			Created:  v.Comment.Created.Format(time.RFC3339),
			Modified: v.Comment.Modified.Format(time.RFC3339),
		}
	}

	return city, nil
}

// TODO replace this with call to repository
func addCommentsToCity(_ context.Context, item interface{}) (interface{}, error) {
	return entity.GetCityCommentsOutput{
		City:     item.(entity.GetCityCommentsInput).City,
		Comments: []entity.CommentWithPosterName{},
	}, nil
}

func (c *cityService) GetCity(ctx context.Context, id, numberOfComments int) (cityDto, error) {
	ctx = app.ContextWithValue(ctx, "function", "cityService.GetCity")

	item := <-rxgo.JustItem(id).
		Map(c.repo.GetCity).
		Map(func(ctx context.Context, city interface{}) (interface{}, error) {
			return entity.GetCityCommentsInput{
				City:             city.(entity.City),
				NumberOfComments: ctx.Value(ctxCommentNumIdx).(int),
			}, nil
		}, rxgo.WithContext(context.WithValue(ctx, ctxCommentNumIdx, numberOfComments))).
		Map(addCommentsToCity).
		Map(cityToDto).
		Observe()
	if item.Error() {
		c.logger.Error(app.ContextWithError(ctx, item.E), "could not get city with ID %d", id)
		return cityDto{}, fmt.Errorf("get city failed: %w", item.E)
	}

	return item.V.(cityDto), nil
}

func (c *cityService) ListAllCities(ctx context.Context, numberOfComments int) ([]cityDto, error) {
	ctx = app.ContextWithValue(ctx, "function", "cityService.ListAllCities")

	obs := rxgo.Just(true)().
		Map(c.repo.GetAllCities).
		FlatMap(func(i rxgo.Item) rxgo.Observable {
			cities := i.V.([]entity.City)
			result := make([]interface{}, len(cities))

			for i, v := range cities {
				result[i] = v
			}

			return rxgo.Just(result...)()
		}).
		Map(func(ctx context.Context, city interface{}) (interface{}, error) {
			return entity.GetCityCommentsInput{
				City:             city.(entity.City),
				NumberOfComments: ctx.Value(ctxCommentNumIdx).(int),
			}, nil
		}, rxgo.WithContext(context.WithValue(ctx, ctxCommentNumIdx, numberOfComments))).
		Map(addCommentsToCity).
		Map(cityToDto)

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

func checkIfCityExists(_ context.Context, item interface{}, city interface{}) (interface{}, error) {
	if err, ok := item.(error); ok {
		if errors.Is(err, entity.ErrCityNotFound) {
			return city, nil
		}

		return city, err
	}

	return city, entity.ErrCityExists
}

func currentTime(_ interface{}) time.Time {
	return time.Now()
}

func (c *cityService) AddCity(ctx context.Context, name, country string) (int, error) {
	ctx = app.ContextWithValue(ctx, "function", "cityService.AddCity")
	city := entity.City{
		Name:    name,
		Country: country,
	}

	item := <-rxgo.Just(city)().
		Map(c.repo.GetCityByNameAndCountry).
		OnErrorReturn(func(err error) interface{} {
			return err
		}).
		Join(checkIfCityExists, rxgo.Just(city)(), currentTime, rxgo.WithDuration(5*time.Second)).
		Map(c.repo.AddCity).
		Observe()
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

	item := <-rxgo.JustItem(city).Map(c.repo.UpdateCity).Observe()
	if item.Error() {
		c.logger.Error(app.ContextWithError(ctx, item.E), "could not update city")
		return fmt.Errorf("could not update city: %w", item.E)
	}

	return nil
}

func (c *cityService) DeleteCity(ctx context.Context, id int) error {
	ctx = app.ContextWithValue(ctx, "function", "cityService.DeleteCity")

	item := <-rxgo.JustItem(id).Map(c.repo.DeleteCity).Observe()
	if item.Error() {
		c.logger.Error(app.ContextWithError(ctx, item.E), "could not delete city")
		return fmt.Errorf("could not delete city: %w", item.E)
	}

	return nil
}
