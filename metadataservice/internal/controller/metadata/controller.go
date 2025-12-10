package metadata

import (
	"context"
	"errors"
	"fmt"

	"github.com/ochamekan/ms/metadataservice/internal/repository"
	"github.com/ochamekan/ms/metadataservice/pkg/model"
)

var ErrNotFound = errors.New("not found")

type metadataRepository interface {
	Get(ctx context.Context, id int) (*model.Metadata, error)
	Put(ctx context.Context, metadata *model.Metadata) error
}

type Controller struct {
	repo  metadataRepository
	cache metadataRepository
}

func New(repo metadataRepository, cache metadataRepository) *Controller {
	return &Controller{repo, cache}
}

func (c *Controller) GetMetadata(ctx context.Context, id int) (*model.Metadata, error) {
	cachedRes, err := c.cache.Get(ctx, id)
	if err == nil {
		return cachedRes, nil
	}

	res, err := c.repo.Get(ctx, id)
	if err != nil && errors.Is(err, repository.ErrNotFound) {
		return nil, ErrNotFound
	} else if err != nil {
		return nil, err
	}

	if err := c.cache.Put(ctx, res); err != nil {
		fmt.Println("error updating redis cache: " + err.Error())
	}

	return res, err
}

func (c *Controller) PutMovieData(ctx context.Context, metadata *model.Metadata) error {
	return c.repo.Put(ctx, metadata)
}
