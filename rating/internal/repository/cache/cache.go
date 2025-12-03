package cache

import (
	"context"
	"fmt"
	"os"
	"time"

	"github.com/ochamekan/ms/rating/pkg/model"
	"github.com/redis/go-redis/v9"
)

type Cache struct {
	client *redis.Client
	name   string
}

func New(name string) (*Cache, error) {
	client := redis.NewClient(&redis.Options{
		Addr:     os.Getenv("REDIS_ADDR"),
		Password: os.Getenv("REDIS_PASSWORD"),
		DB:       0,
	})

	err := client.Ping(context.Background()).Err()
	if err != nil {
		return nil, err
	}

	return &Cache{client, name}, nil
}

func (c *Cache) GetAggregatedRating(ctx context.Context, recordID model.RecordID, recordType model.RecordType) (float64, error) {
	val, err := c.client.Get(ctx, fmt.Sprintf("%s:%s", c.name, string(recordID)+string(recordType))).Float64()
	if err != nil {
		return 0, err
	}
	return val, nil
}

func (c *Cache) PutAggregatedRating(ctx context.Context, recordID model.RecordID, recordType model.RecordType, ratingValue float64) error {
	return c.client.Set(ctx, fmt.Sprintf("%s:%s", c.name, string(recordID)+string(recordType)), ratingValue, 1*time.Minute).Err()
}
