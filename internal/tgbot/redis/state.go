package redis

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/redis/go-redis/v9"
	"github.com/ysomad/financer/internal/tgbot/model"
)

type StateCache struct {
	client *redis.Client
	ttl    time.Duration
}

func NewStateCache(c *redis.Client, ttl time.Duration) StateCache {
	return StateCache{client: c, ttl: ttl}
}

func stateCacheKey(tgUID int64) string {
	return fmt.Sprintf("state:%d", tgUID)
}

func (c StateCache) Save(ctx context.Context, tgUID int64, st model.State) error {
	return c.client.Set(ctx, stateCacheKey(tgUID), string(st), c.ttl).Err()
}

var ErrNotFound = errors.New("entry not found")

func (c StateCache) Get(ctx context.Context, tgUID int64) (model.State, error) {
	res, err := c.client.Get(ctx, stateCacheKey(tgUID)).Result()
	if err != nil {
		return "", fmt.Errorf("%w:%w", ErrNotFound, err)
	}

	return model.State(res), nil
}

func (c StateCache) Del(ctx context.Context, tgUID int64) error {
	return c.client.Del(ctx, stateCacheKey(tgUID)).Err()
}
