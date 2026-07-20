package service_test

import (
	"context"
	"io"
	"testing"

	"github.com/Naitik2411/stockit/internal/cache"
	"github.com/Naitik2411/stockit/internal/config"
	"github.com/Naitik2411/stockit/internal/model"
	"github.com/Naitik2411/stockit/internal/server"
	"github.com/Naitik2411/stockit/internal/service"
	"github.com/alicebob/miniredis/v2"
	"github.com/google/uuid"
	"github.com/redis/go-redis/v9"
	"github.com/rs/zerolog"
	"github.com/shopspring/decimal"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func newLeaderboardFixture(t *testing.T) (*service.LeaderboardService, *cache.Cache) {
	t.Helper()

	mr := miniredis.RunT(t)
	client := redis.NewClient(&redis.Options{Addr: mr.Addr()})
	t.Cleanup(func() { _ = client.Close() })

	c := cache.New(client)
	nop := zerolog.New(io.Discard)
	srv := &server.Server{
		Config: &config.Config{
			Integration: config.IntegrationConfig{
				LeaderboardCacheTTL: 60,
				StartingBalance:     100000,
			},
		},
		Logger: &nop,
		Cache:  c,
	}
	return service.NewLeaderboardService(srv, nil, nil, nil), c
}

func TestLeaderboardGlobal_Pagination(t *testing.T) {
	svc, c := newLeaderboardFixture(t)

	entries := make([]model.LeaderboardEntry, 0, 120)
	for i := 0; i < 120; i++ {
		entries = append(entries, model.LeaderboardEntry{
			Rank:       i + 1,
			UserID:     uuid.New(),
			TotalValue: decimal.NewFromInt(100000 + int64(120-i)),
			ReturnPct:  decimal.NewFromInt(int64(120 - i)),
		})
	}
	require.NoError(t, c.SetLeaderboard(context.Background(), entries, 60))

	page1, err := svc.Global(context.Background(), 1, 50)
	require.NoError(t, err)
	assert.Equal(t, 1, page1.Page)
	assert.Equal(t, 50, page1.Limit)
	assert.Equal(t, 120, page1.Total)
	require.Len(t, page1.Entries, 50)
	assert.Equal(t, 1, page1.Entries[0].Rank)

	page3, err := svc.Global(context.Background(), 3, 50)
	require.NoError(t, err)
	assert.Equal(t, 120, page3.Total)
	require.Len(t, page3.Entries, 20)
	assert.Equal(t, 101, page3.Entries[0].Rank)
}

func TestLeaderboardGlobal_Defaults(t *testing.T) {
	svc, c := newLeaderboardFixture(t)

	entries := []model.LeaderboardEntry{
		{Rank: 1, UserID: uuid.New(), ReturnPct: decimal.NewFromInt(5)},
	}
	require.NoError(t, c.SetLeaderboard(context.Background(), entries, 60))

	page, err := svc.Global(context.Background(), 0, 0)
	require.NoError(t, err)
	assert.Equal(t, 1, page.Page)
	assert.Equal(t, 50, page.Limit)
	require.Len(t, page.Entries, 1)
}

func TestLeaderboardMyRank_Found(t *testing.T) {
	svc, c := newLeaderboardFixture(t)

	userID := uuid.New()
	username := "alice"
	entries := []model.LeaderboardEntry{
		{
			Rank:       1,
			UserID:     userID,
			Username:   &username,
			ReturnPct:  decimal.NewFromInt(20),
			TotalValue: decimal.NewFromInt(120000),
		},
		{
			Rank:       2,
			UserID:     uuid.New(),
			ReturnPct:  decimal.NewFromInt(10),
			TotalValue: decimal.NewFromInt(110000),
		},
	}
	require.NoError(t, c.SetLeaderboard(context.Background(), entries, 60))

	rank, err := svc.MyRank(context.Background(), userID)
	require.NoError(t, err)
	assert.Equal(t, 1, rank.Rank)
	assert.Equal(t, 2, rank.Total)
	assert.True(t, rank.Percentile.Equal(decimal.NewFromInt(100)))
	assert.Equal(t, userID, rank.Entry.UserID)
	require.NotNil(t, rank.Entry.Username)
	assert.Equal(t, "alice", *rank.Entry.Username)
}

func TestLeaderboardMyRank_NotFound(t *testing.T) {
	svc, c := newLeaderboardFixture(t)

	entries := []model.LeaderboardEntry{
		{Rank: 1, UserID: uuid.New(), ReturnPct: decimal.NewFromInt(5)},
	}
	require.NoError(t, c.SetLeaderboard(context.Background(), entries, 60))

	missing := uuid.New()
	rank, err := svc.MyRank(context.Background(), missing)
	require.NoError(t, err)
	assert.Equal(t, 0, rank.Rank)
	assert.Equal(t, 1, rank.Total)
	assert.True(t, rank.Percentile.IsZero())
	assert.Equal(t, missing, rank.Entry.UserID)
}
