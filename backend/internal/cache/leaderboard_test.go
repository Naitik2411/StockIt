package cache_test

import (
	"context"
	"testing"

	"github.com/Naitik2411/stockit/internal/model"
	"github.com/google/uuid"
	"github.com/shopspring/decimal"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestLeaderboardRoundTrip(t *testing.T) {
	c, _ := newTestCache(t)
	ctx := context.Background()

	username := "Naitik"
	userID := uuid.New()
	entries := []model.LeaderboardEntry{
		{
			Rank:        1,
			UserID:      userID,
			Username:    &username,
			TotalValue:  decimal.NewFromInt(110000),
			ReturnPct:   decimal.NewFromInt(10),
			CashBalance: decimal.NewFromInt(50000),
		},
	}
	require.NoError(t, c.SetLeaderboard(ctx, entries, 60))

	got, err := c.GetLeaderboard(ctx)
	require.NoError(t, err)
	require.Len(t, got, 1)
	assert.Equal(t, 1, got[0].Rank)
	assert.Equal(t, userID, got[0].UserID)
	require.NotNil(t, got[0].Username)
	assert.Equal(t, "Naitik", *got[0].Username)
	assert.True(t, got[0].TotalValue.Equal(decimal.NewFromInt(110000)))
	assert.True(t, got[0].ReturnPct.Equal(decimal.NewFromInt(10)))
	assert.True(t, got[0].CashBalance.Equal(decimal.NewFromInt(50000)))
}

func TestGetLeaderboardMiss(t *testing.T) {
	c, _ := newTestCache(t)
	got, err := c.GetLeaderboard(context.Background())
	require.NoError(t, err)
	assert.Nil(t, got)
}
