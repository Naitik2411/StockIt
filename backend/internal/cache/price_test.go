package cache_test

import (
	"context"
	"testing"
	"time"

	"github.com/Naitik2411/stockit/internal/cache"
	errorss "github.com/Naitik2411/stockit/internal/errors"
	"github.com/redis/go-redis/v9"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestSetPriceGetPriceRoundTrip(t *testing.T) {
	client := redis.NewClient(&redis.Options{Addr: "localhost:6379"})
	defer client.Close()

	ctx := context.Background()
	require.NoError(t, client.Ping(ctx).Err())

	c := cache.New(client)

	defer client.Del(ctx, "price:AAPL")

	input := cache.StockPrice{
		Price:     "178.50",
		ChangePct: "1.25",
		SyncedAt:  time.Now().UTC(),
	}

	err := c.SetPrice(ctx, "AAPL", input)
	require.NoError(t, err)

	got, err := c.GetPrice(ctx, "AAPL")

	assert.Equal(t, "AAPL", got.Ticker)
	assert.Equal(t, "178.50", got.Price)
	assert.Equal(t, "1.25", got.ChangePct)
}

func TestGetPriceNotFound(t *testing.T) {
	client := redis.NewClient(&redis.Options{Addr: "localhost:6379"})
	defer client.Close()
	c := cache.New(client)
	_, err := c.GetPrice(context.Background(), "DOESNOTEXIST")
	assert.ErrorIs(t, err, errorss.ErrTickerNotFound)
}
