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

type IdentityCache struct {
	client *redis.Client
	ttl    time.Duration
}

func NewIdentityCache(c *redis.Client, ttl time.Duration) IdentityCache {
	return IdentityCache{client: c, ttl: ttl}
}

func identityCacheKey(tgUID int64) string {
	return fmt.Sprintf("identity:%d", tgUID)
}

func (c IdentityCache) Save(ctx context.Context, id model.Identity) error {
	if err := id.Validate(); err != nil {
		return fmt.Errorf("invalid identity: %w", err)
	}

	bb, err := json.Marshal(id)
	if err != nil {
		return fmt.Errorf("identity not marshaled: %w", err)
	}

	return c.client.Set(ctx, identityCacheKey(id.TGUID), bb, c.ttl).Err()
}

var ErrNotFound = errors.New("entry not found")

func (c IdentityCache) Get(ctx context.Context, tgUID int64) (model.Identity, error) {
	cmd := c.client.Get(ctx, identityCacheKey(tgUID))
	if err := cmd.Err(); err != nil {
		if errors.Is(err, ErrNotFound) {
			return model.Identity{}, ErrNotFound
		}

		return model.Identity{}, fmt.Errorf("redis cmd: %w", err)
	}

	bb, err := cmd.Bytes()
	if err != nil {
		return model.Identity{}, fmt.Errorf("bytes not received from result: %w", err)
	}

	var id model.Identity

	if err := json.Unmarshal(bb, &id); err != nil {
		return model.Identity{}, fmt.Errorf("result not unmarshaled: %w", err)
	}

	return id, nil
}
