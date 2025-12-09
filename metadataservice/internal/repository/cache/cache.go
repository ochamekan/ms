package cache

import (
	"context"
	"encoding/json"
	"fmt"
	"os"

	"github.com/ochamekan/ms/metadataservice/pkg/model"
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

func (c *Cache) Get(ctx context.Context, id int) (*model.Metadata, error) {
	val, err := c.client.Get(ctx, fmt.Sprintf("%s:%d", c.name, id)).Bytes()
	if err != nil {
		return nil, err
	}

	var m model.Metadata
	err = json.Unmarshal(val, &m)
	if err != nil {
		return nil, err
	}

	return &m, nil
}

func (c *Cache) Put(ctx context.Context, metadata *model.Metadata) error {
	json, err := json.Marshal(metadata)
	if err != nil {
		return err
	}

	return c.client.Set(ctx, fmt.Sprintf("%s:%d", c.name, metadata.ID), json, 0).Err()
}
