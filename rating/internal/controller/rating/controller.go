package rating

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/ochamekan/ms/rating/internal/repository"
	"github.com/ochamekan/ms/rating/pkg/model"
)

var ErrNotFound = errors.New("ratings not found for a record")

type ratingRepository interface {
	Get(ctx context.Context, recordID model.RecordID, recordType model.RecordType) ([]model.Rating, error)
	Put(ctx context.Context, recordID model.RecordID, recordType model.RecordType, rating *model.Rating) error
}

type ratingCache interface {
	GetAggregatedRating(ctx context.Context, recordID model.RecordID, recordType model.RecordType) (float64, error)
	PutAggregatedRating(ctx context.Context, recordID model.RecordID, recordType model.RecordType, ratingValue float64) error
}

// type ratingIngester interface {
// 	Ingest(ctx context.Context) (chan model.RatingEvent, error)
// }

type Controller struct {
	repo  ratingRepository
	cache ratingCache
	// ingester ratingIngester
}

func New(repo ratingRepository, cache ratingCache) *Controller {
	return &Controller{repo, cache}
}

func (c *Controller) GetAggregatedRating(ctx context.Context, recordID model.RecordID, recordType model.RecordType) (float64, error) {
	cachedRes, err := c.cache.GetAggregatedRating(ctx, recordID, recordType)
	if err == nil {
		return cachedRes, nil
	}

	ratings, err := c.repo.Get(ctx, recordID, recordType)
	if err != nil && err == repository.ErrNotFound {
		return 0, ErrNotFound
	} else if err != nil {
		return 0, err
	}

	sum := float64(0)
	for _, r := range ratings {
		sum += float64(r.Value)
	}
	// Emulating hard calculations
	time.Sleep(1 * time.Second)

	res := sum / float64(len(ratings))

	if err := c.cache.PutAggregatedRating(ctx, recordID, recordType, res); err != nil {
		fmt.Println("error updating redis cache: " + err.Error())
	}

	return res, nil
}

func (c *Controller) PutRating(ctx context.Context, recordID model.RecordID, recordType model.RecordType, rating *model.Rating) error {
	return c.repo.Put(ctx, recordID, recordType, rating)
}

// func (c *Controller) StartIngestion(ctx context.Context) error {
// 	ch, err := c.ingester.Ingest(ctx)
// 	if err != nil {
// 		return err
// 	}
//
// 	for e := range ch {
// 		fmt.Printf("Consumed a message: %v\n", e)
// 		if err := c.PutRating(ctx, e.RecordID, e.RecordType, &model.Rating{UserID: e.UserID, Value: e.Value, RecordID: e.RecordID, RecordType: e.RecordType}); err != nil {
// 			fmt.Printf("StartIngestion: failed to put rating: %v\n", err)
// 			continue
// 		}
// 	}
// 	return nil
// }
