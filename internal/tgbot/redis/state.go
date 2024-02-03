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

func stateCacheKey(tguid int64) string {
	return fmt.Sprintf("state:%d", tguid)
}

func (c StateCache) Save(ctx context.Context, tguid int64, st model.State) error {
	return c.client.Set(ctx, stateCacheKey(tguid), string(st), c.ttl).Err()
}

func (c StateCache) Get(ctx context.Context, tguid int64) (model.State, error) {
	res, err := c.client.Get(ctx, stateCacheKey(tguid)).Result()
	if err != nil {
		if errors.Is(err, redis.Nil) {
			return model.StateUnknown, nil
		}

		return "", err
	}

	return model.State(res), nil
}

func (c StateCache) Del(ctx context.Context, tguid int64) error {
	return c.client.Del(ctx, stateCacheKey(tguid)).Err()
}
