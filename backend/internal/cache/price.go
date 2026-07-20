package cache

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"strings"
	"time"

	errorss "github.com/Naitik2411/stockit/internal/errors"
	"github.com/redis/go-redis/v9"
)

const priceKeyPrefix = "price:"

type StockPrice struct {
	Ticker    string    `json:"ticker"`
	Price     string    `json:"price"`
	ChangePct string    `json:"change_pct"`
	SyncedAt  time.Time `json:"synced_at"`
}

func priceKey(ticker string) string {
	return priceKeyPrefix + strings.ToUpper(ticker)
}

func (c *Cache) SetPrice(ctx context.Context, ticker string, price StockPrice) error {
	price.Ticker = strings.ToUpper(ticker)

	data, err := json.Marshal(price)
	if err != nil {
		return fmt.Errorf("MARSHAL PRICE : %w", err)
	}
	key := priceKey(ticker)
	if err := c.redis.Set(ctx, key, data, 0).Err(); err != nil {
		return fmt.Errorf("redis set: %w", err)
	}
	return nil
}

func (c *Cache) GetPrice(ctx context.Context, ticker string) (StockPrice, error) {
	key := priceKey(ticker)

	data, err := c.redis.Get(ctx, key).Bytes()
	if err != nil {
		if errors.Is(err, redis.Nil) {
			return StockPrice{}, errorss.ErrTickerNotFound
		}
		return StockPrice{}, fmt.Errorf("redis get: %w", err)
	}

	var price StockPrice
	if err := json.Unmarshal(data, &price); err != nil {
		return StockPrice{}, fmt.Errorf("unmarshal price : %w", err)
	}
	return price, nil

}
