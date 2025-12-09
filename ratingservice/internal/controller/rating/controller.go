package rating

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/ochamekan/ms/ratingservice/internal/repository"
	"github.com/ochamekan/ms/ratingservice/pkg/model"
)

var ErrNotFound = errors.New("ratings not found for a record")

type ratingRepository interface {
	Get(ctx context.Context, movieID model.MovieID) ([]model.Rating, error)
	Put(ctx context.Context, movieID model.MovieID, rating model.RatingValue) error
}

type ratingCache interface {
	GetAggregatedRating(ctx context.Context, movieID model.MovieID) (float64, error)
	PutAggregatedRating(ctx context.Context, movieID model.MovieID, rating float64) error
}

type Controller struct {
	repo  ratingRepository
	cache ratingCache
}

func New(repo ratingRepository, cache ratingCache) *Controller {
	return &Controller{repo, cache}
}

func (c *Controller) GetAggregatedRating(ctx context.Context, movieID model.MovieID) (float64, error) {
	cachedRes, err := c.cache.GetAggregatedRating(ctx, movieID)
	if err == nil {
		return cachedRes, nil
	}

	ratings, err := c.repo.Get(ctx, movieID)
	if err != nil && err == repository.ErrNotFound {
		return 0, ErrNotFound
	} else if err != nil {
		return 0, err
	}

	sum := float64(0)
	for _, r := range ratings {
		sum += float64(r.Rating)
	}

	// Emulating hard work
	time.Sleep(1 * time.Second)

	res := sum / float64(len(ratings))

	if err := c.cache.PutAggregatedRating(ctx, movieID, res); err != nil {
		fmt.Println("error updating redis cache: " + err.Error())
	}

	return res, nil
}

func (c *Controller) PutRating(ctx context.Context, movieID model.MovieID, rating model.RatingValue) error {
	return c.repo.Put(ctx, movieID, rating)
}
