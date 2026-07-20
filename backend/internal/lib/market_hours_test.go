package lib_test

import (
	"testing"
	"time"

	"github.com/Naitik2411/stockit/internal/lib"
	"github.com/stretchr/testify/assert"
)

func TestIsMarketOpen(t *testing.T) {
	loc, _ := time.LoadLocation("America/New_York")

	tests := []struct {
		name     string
		at       time.Time
		expected bool
	}{
		{
			name:     "weekday during market hours",
			at:       time.Date(2026, 7, 8, 10, 0, 0, 0, loc), // Tue 10am ET
			expected: true,
		},
		{
			name:     "exactly at open",
			at:       time.Date(2026, 7, 8, 9, 30, 0, 0, loc),
			expected: true,
		},
		{
			name:     "one minute before close",
			at:       time.Date(2026, 7, 8, 15, 59, 0, 0, loc),
			expected: true,
		},
		{
			name:     "exactly at close",
			at:       time.Date(2026, 7, 8, 16, 0, 0, 0, loc),
			expected: false,
		},
		{
			name:     "weekday before open",
			at:       time.Date(2026, 7, 8, 9, 0, 0, 0, loc),
			expected: false,
		},
		{
			name:     "weekday after close",
			at:       time.Date(2026, 7, 8, 17, 0, 0, 0, loc),
			expected: false,
		},
		{
			name:     "saturday",
			at:       time.Date(2026, 7, 11, 12, 0, 0, 0, loc),
			expected: false,
		},
		{
			name:     "sunday",
			at:       time.Date(2026, 7, 12, 12, 0, 0, 0, loc),
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := lib.IsMarketOpen(tt.at, "America/New_York")
			assert.Equal(t, tt.expected, got)
		})
	}
}
