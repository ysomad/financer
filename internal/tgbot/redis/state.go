package redis

import (
	"context"
	"encoding/json"
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
	bb, err := json.Marshal(st)
	if err != nil {
		return fmt.Errorf("state not marshaled: %w", err)
	}

	return c.client.Set(ctx, stateCacheKey(tguid), bb, c.ttl).Err()
}

func (c StateCache) Get(ctx context.Context, tguid int64) (model.State, error) {
	bb, err := c.client.Get(ctx, stateCacheKey(tguid)).Bytes()
	if err != nil {
		if errors.Is(err, redis.Nil) {
			return model.State{}, nil
		}

		return model.State{}, err
	}

	var st model.State

	if err := json.Unmarshal(bb, &st); err != nil {
		return model.State{}, fmt.Errorf("result not unmarshaled: %w", err)
	}

	return st, nil
}

func (c StateCache) Del(ctx context.Context, tguid int64) error {
	return c.client.Del(ctx, stateCacheKey(tguid)).Err()
}
