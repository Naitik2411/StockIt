package cache

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"time"

	"github.com/Naitik2411/stockit/internal/model"
	"github.com/redis/go-redis/v9"
)

const leaderboardKey = "leaderboard:global"

func (c *Cache) SetLeaderboard(ctx context.Context, entries []model.LeaderboardEntry, ttlSeconds int) error {
	data, err := json.Marshal(entries)
	if err != nil {
		return fmt.Errorf("marshal leaderboard entries : %w", err)
	}

	if ttlSeconds <= 0 {
		ttlSeconds = 300
	}

	if err := c.redis.Set(ctx, leaderboardKey, data, time.Duration(ttlSeconds)*time.Second).Err(); err != nil {
		return fmt.Errorf("redis set leaderboard : %w", err)
	}

	return nil
}

func (c *Cache) GetLeaderboard(ctx context.Context) ([]model.LeaderboardEntry, error) {
	data, err := c.redis.Get(ctx, leaderboardKey).Bytes()
	if err != nil {
		if errors.Is(err, redis.Nil) {
			return nil, nil // cache miss — not an error
		}
		return nil, fmt.Errorf("redis get leaderboard: %w", err)
	}

	var entries []model.LeaderboardEntry

	if err := json.Unmarshal(data, &entries); err != nil {
		return nil, fmt.Errorf("json unmarshal leaderboard : %w", err)
	}

	return entries, nil

}
