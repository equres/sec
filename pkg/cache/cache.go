package cache

import (
	"context"

	"github.com/equres/sec/pkg/config"
	"github.com/go-redis/redis/v8"
)

type Cache struct {
	Redis redis.Client
}

func NewCache(cfg *config.Config) Cache {
	rdb := redis.NewClient(&redis.Options{
		Addr:     cfg.Redis.GetRedisURL(),
		Password: "",
		DB:       0,
	})

	return Cache{
		Redis: *rdb,
	}
}

func (c *Cache) Set(k, v string) error {
	err := c.Redis.Set(context.Background(), k, v, 0).Err()
	if err != nil {
		return err
	}

	return nil
}

func (c *Cache) Get(k string) (string, error) {
	v, err := c.Redis.Get(context.Background(), k).Result()
	if err != nil {
		return "", err
	}

	return v, err
}

func (c *Cache) MustSet(k, v string) error {
	err := c.Redis.Set(context.Background(), k, v, 0).Err()
	if err != nil {
		return err
	}

	return nil
}

func (c *Cache) MustGet(k string) (string, error) {
	v, err := c.Redis.Get(context.Background(), k).Result()
	if err != nil {
		return "", err
	}

	return v, err
}
