package cache_test

import (
	"context"
	"testing"
	"time"

	"github.com/Naitik2411/stockit/internal/cache"
	errorss "github.com/Naitik2411/stockit/internal/errors"
	"github.com/alicebob/miniredis/v2"
	"github.com/redis/go-redis/v9"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func newTestCache(t *testing.T) (*cache.Cache, *miniredis.Miniredis) {
	t.Helper()
	mr := miniredis.RunT(t)
	client := redis.NewClient(&redis.Options{Addr: mr.Addr()})
	t.Cleanup(func() { _ = client.Close() })
	return cache.New(client), mr
}

func TestSetPriceGetPriceRoundTrip(t *testing.T) {
	c, _ := newTestCache(t)
	ctx := context.Background()

	input := cache.StockPrice{
		Price:     "178.50",
		ChangePct: "1.25",
		SyncedAt:  time.Now().UTC(),
	}

	err := c.SetPrice(ctx, "AAPL", input)
	require.NoError(t, err)

	got, err := c.GetPrice(ctx, "AAPL")
	require.NoError(t, err)
	assert.Equal(t, "AAPL", got.Ticker)
	assert.Equal(t, "178.50", got.Price)
	assert.Equal(t, "1.25", got.ChangePct)
}

func TestGetPriceNotFound(t *testing.T) {
	c, _ := newTestCache(t)
	_, err := c.GetPrice(context.Background(), "DOESNOTEXIST")
	assert.ErrorIs(t, err, errorss.ErrTickerNotFound)
}
