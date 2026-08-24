package lib

import (
	"math"

	"github.com/google/uuid"
)

const DefaultELOK = 32

// Standing is one player's finishing position in a season.
type Standing struct {
	UserID uuid.UUID
	Rank   int
}

// ExpectedScore returns the probability that player A beats player B.
func ExpectedScore(ratingA, ratingB int) float64 {
	return 1.0 / (1.0 + math.Pow(10, float64(ratingB-ratingA)/400.0))
}

// NewRating applies one match result.
// actual: 1.0 = win, 0.5 = draw, 0.0 = loss.
func NewRating(rating, k int, actual, expected float64) int {
	return rating + int(math.Round(float64(k)*(actual-expected)))
}

// CalcSeasonELOChanges returns the new rating for every player in standings,
// using pairwise comparisons averaged across all opponents.
//
// current holds each player's pre-season rating. Players missing from current
// are treated as having startingRating.
func CalcSeasonELOChanges(
	current map[uuid.UUID]int,
	standings []Standing,
	k int,
	startingRating int,
) map[uuid.UUID]int {
	if k <= 0 {
		k = DefaultELOK
	}
	if startingRating <= 0 {
		startingRating = 1000
	}

	ratingOf := func(id uuid.UUID) int {
		if r, ok := current[id]; ok {
			return r
		}
		return startingRating
	}

	newRatings := make(map[uuid.UUID]int, len(standings))

	// A solo season has no opponents, so nothing can be earned or lost.
	if len(standings) < 2 {
		for _, s := range standings {
			newRatings[s.UserID] = ratingOf(s.UserID)
		}
		return newRatings
	}

	for i, a := range standings {
		delta := 0.0
		for j, b := range standings {
			if i == j {
				continue
			}

			expected := ExpectedScore(ratingOf(a.UserID), ratingOf(b.UserID))

			actual := 0.0
			switch {
			case a.Rank < b.Rank:
				actual = 1.0
			case a.Rank == b.Rank:
				actual = 0.5
			}

			delta += float64(k) * (actual - expected)
		}

		avg := delta / float64(len(standings)-1)
		newRatings[a.UserID] = ratingOf(a.UserID) + int(math.Round(avg))
	}

	return newRatings
}
