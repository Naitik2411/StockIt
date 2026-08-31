package lib_test

import (
	"testing"

	"github.com/Naitik2411/stockit/internal/lib"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestExpectedScore_equalRatings(t *testing.T) {
	got := lib.ExpectedScore(1000, 1000)
	assert.InDelta(t, 0.5, got, 1e-9)
}

func TestExpectedScore_higherRatedFavored(t *testing.T) {
	got := lib.ExpectedScore(1200, 1000)
	assert.Greater(t, got, 0.5)
	assert.Less(t, got, 1.0)
}

func TestNewRating_winAgainstEqual(t *testing.T) {
	got := lib.NewRating(1000, 32, 1.0, 0.5)
	assert.Equal(t, 1016, got)
}

func TestNewRating_lossAgainstEqual(t *testing.T) {
	got := lib.NewRating(1000, 32, 0.0, 0.5)
	assert.Equal(t, 984, got)
}

func TestCalcSeasonELOChanges_soloUnchanged(t *testing.T) {
	a := uuid.New()
	current := map[uuid.UUID]int{a: 1100}
	standings := []lib.Standing{{UserID: a, Rank: 1}}

	got := lib.CalcSeasonELOChanges(current, standings, 32, 1000)
	require.Len(t, got, 1)
	assert.Equal(t, 1100, got[a])
}

func TestCalcSeasonELOChanges_winnerGainsLoserDrops(t *testing.T) {
	a := uuid.New()
	b := uuid.New()
	current := map[uuid.UUID]int{
		a: 1000,
		b: 1000,
	}
	standings := []lib.Standing{
		{UserID: a, Rank: 1},
		{UserID: b, Rank: 2},
	}

	got := lib.CalcSeasonELOChanges(current, standings, 32, 1000)
	require.Len(t, got, 2)
	assert.Greater(t, got[a], 1000)
	assert.Less(t, got[b], 1000)
	assert.Equal(t, 2000, got[a]+got[b])
}

func TestCalcSeasonELOChanges_missingUsesStartingRating(t *testing.T) {
	a := uuid.New()
	b := uuid.New()
	current := map[uuid.UUID]int{a: 1000}
	standings := []lib.Standing{
		{UserID: a, Rank: 1},
		{UserID: b, Rank: 2},
	}

	got := lib.CalcSeasonELOChanges(current, standings, 32, 1000)
	require.Contains(t, got, b)
	assert.Less(t, got[b], 1000)
	assert.Greater(t, got[a], 1000)
}

func TestCalcSeasonELOChanges_tieSplits(t *testing.T) {
	a := uuid.New()
	b := uuid.New()
	current := map[uuid.UUID]int{a: 1000, b: 1000}
	standings := []lib.Standing{
		{UserID: a, Rank: 1},
		{UserID: b, Rank: 1},
	}

	got := lib.CalcSeasonELOChanges(current, standings, 32, 1000)
	assert.Equal(t, 1000, got[a])
	assert.Equal(t, 1000, got[b])
}

func TestCalcSeasonELOChanges_defaultKAndStarting(t *testing.T) {
	a := uuid.New()
	b := uuid.New()
	standings := []lib.Standing{
		{UserID: a, Rank: 1},
		{UserID: b, Rank: 2},
	}

	got := lib.CalcSeasonELOChanges(nil, standings, 0, 0)
	assert.Greater(t, got[a], 1000)
	assert.Less(t, got[b], 1000)
}

func TestCalcSeasonELOChanges_threePlayersOrdering(t *testing.T) {
	first := uuid.New()
	second := uuid.New()
	third := uuid.New()
	current := map[uuid.UUID]int{
		first:  1000,
		second: 1000,
		third:  1000,
	}
	standings := []lib.Standing{
		{UserID: first, Rank: 1},
		{UserID: second, Rank: 2},
		{UserID: third, Rank: 3},
	}

	got := lib.CalcSeasonELOChanges(current, standings, 32, 1000)
	assert.Greater(t, got[first], got[second])
	assert.Greater(t, got[second], got[third])
}
