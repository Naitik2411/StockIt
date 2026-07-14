package cache

import "github.com/redis/go-redis/v9"

type Cache struct {
	redis *redis.Client
}

func New(redis *redis.Client) *Cache {
	return &Cache{
		redis: redis,
	}
}
