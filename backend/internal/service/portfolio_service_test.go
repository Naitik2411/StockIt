package service_test

import (
	"context"
	"io"
	"testing"

	"github.com/Naitik2411/stockit/internal/cache"
	"github.com/Naitik2411/stockit/internal/config"
	errorss "github.com/Naitik2411/stockit/internal/errors"
	"github.com/Naitik2411/stockit/internal/server"
	"github.com/Naitik2411/stockit/internal/service"
	"github.com/alicebob/miniredis/v2"
	"github.com/google/uuid"
	"github.com/redis/go-redis/v9"
	"github.com/rs/zerolog"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestBuy_InvalidShares(t *testing.T) {
	svc := service.NewPortfolioService(nil, nil, nil, nil)
	userID := uuid.New()

	tests := []struct {
		name   string
		shares string
	}{
		{name: "zero", shares: "0"},
		{name: "negative", shares: "-1"},
		{name: "not a number", shares: "abc"},
		{name: "empty", shares: ""},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := svc.Buy(context.Background(), userID, "AAPL", tt.shares)
			require.Error(t, err)

			var httpErr *errorss.HTTPError
			require.ErrorAs(t, err, &httpErr)
			assert.Equal(t, "INVALID_SHARES", httpErr.Code)
		})
	}
}

func TestSell_InvalidShares(t *testing.T) {
	svc := service.NewPortfolioService(nil, nil, nil, nil)
	userID := uuid.New()

	err := svc.Sell(context.Background(), userID, "AAPL", "0")
	require.Error(t, err)

	var httpErr *errorss.HTTPError
	require.ErrorAs(t, err, &httpErr)
	assert.Equal(t, "INVALID_SHARES", httpErr.Code)
}

func TestBuy_TickerNotFound(t *testing.T) {
	mr := miniredis.RunT(t)
	client := redis.NewClient(&redis.Options{Addr: mr.Addr()})
	t.Cleanup(func() { _ = client.Close() })

	nop := zerolog.New(io.Discard)
	srv := &server.Server{
		Config: &config.Config{
			Integration: config.IntegrationConfig{
				MarketTimezone: "America/New_York",
			},
		},
		Logger: &nop,
		Cache:  cache.New(client),
	}
	svc := service.NewPortfolioService(srv, nil, nil, nil)

	err := svc.Buy(context.Background(), uuid.New(), "NOSUCHTICKERXYZ", "1")
	require.Error(t, err)

	var httpErr *errorss.HTTPError
	require.ErrorAs(t, err, &httpErr)
	assert.Equal(t, "TICKER_NOT_FOUND", httpErr.Code)
}
