package movie

import (
	"context"
	"errors"

	metadatamodel "github.com/ochamekan/ms/metadata/pkg/model"
	"github.com/ochamekan/ms/movie/internal/gateway"
	"github.com/ochamekan/ms/movie/pkg/model"
	ratingmodel "github.com/ochamekan/ms/rating/pkg/model"
)

// ErrNotFound is returned when movie metadata is not found.
var ErrNotFound = errors.New("movie metadata not found")

type ratingGateway interface {
	GetAggregatedRating(ctx context.Context, recordID ratingmodel.RecordID, recordType ratingmodel.RecordType) (float64, error)
	PutRating(ctx context.Context, recordID ratingmodel.RecordID, recordType ratingmodel.RecordType, rating *ratingmodel.Rating) error
}

type metadataGateway interface {
	GetMetadata(ctx context.Context, id string) (*metadatamodel.Metadata, error)
	// PutMetadata(ctx context.Context, metadata *metadatamodel.Metadata) error
}

type Controller struct {
	ratingGateway   ratingGateway
	metadataGateway metadataGateway
}

func New(rg ratingGateway, mg metadataGateway) *Controller {
	return &Controller{rg, mg}
}

// Get returns the movie details including the aggreagated rating and movie metadata.
func (c *Controller) Get(ctx context.Context, id string) (*model.MovieDetails, error) {
	metadata, err := c.metadataGateway.GetMetadata(ctx, id)
	if err != nil && errors.Is(err, gateway.ErrNotFound) {
		return nil, ErrNotFound
	} else if err != nil {
		return nil, err
	}

	details := &model.MovieDetails{Metadata: *metadata}

	rating, err := c.ratingGateway.GetAggregatedRating(ctx, ratingmodel.RecordID(id), ratingmodel.RecordTypeMovie)
	if err != nil && !errors.Is(err, gateway.ErrNotFound) {
	} else if err != nil {
		return nil, err
	} else {
		details.Rating = &rating
	}

	return details, nil
}

// func (c *Controller) Put(ctx context.Context)
