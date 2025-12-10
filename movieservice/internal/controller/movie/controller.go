package movie

import (
	"context"
	"errors"

	metadatamodel "github.com/ochamekan/ms/metadataservice/pkg/model"
	"github.com/ochamekan/ms/movieservice/internal/gateway"
	"github.com/ochamekan/ms/movieservice/pkg/model"
	ratingmodel "github.com/ochamekan/ms/ratingservice/pkg/model"
)

var ErrNotFound = errors.New("movie metadata not found")

type ratingGateway interface {
	GetAggregatedRating(ctx context.Context, movieID ratingmodel.MovieID) (float64, error)
	PutRating(ctx context.Context, movieID ratingmodel.MovieID, rating ratingmodel.RatingValue) error
}

type metadataGateway interface {
	GetMetadata(ctx context.Context, id int) (*metadatamodel.Metadata, error)
	PutMetadata(ctx context.Context, title, description, director string, year int) error
}

type Controller struct {
	ratingGateway   ratingGateway
	metadataGateway metadataGateway
}

func New(rg ratingGateway, mg metadataGateway) *Controller {
	return &Controller{rg, mg}
}

func (c *Controller) Get(ctx context.Context, id int) (*model.MovieDetails, error) {
	metadata, err := c.metadataGateway.GetMetadata(ctx, id)
	if err != nil && errors.Is(err, gateway.ErrNotFound) {
		return nil, ErrNotFound
	} else if err != nil {
		return nil, err
	}

	details := &model.MovieDetails{Metadata: *metadata}

	rating, err := c.ratingGateway.GetAggregatedRating(ctx, ratingmodel.MovieID(metadata.ID))
	if err != nil && !errors.Is(err, gateway.ErrNotFound) {
		placeholder := float64(0)
		details.Rating = &placeholder
	} else if err != nil {
		return nil, err
	} else {
		details.Rating = &rating
	}

	return details, nil
}
