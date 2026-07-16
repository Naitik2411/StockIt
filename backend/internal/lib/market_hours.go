package lib

import "time"

func IsMarketOpen(now time.Time, timezone string) bool {
	loc, err := time.LoadLocation(timezone)
	if err != nil {
		loc, _ = time.LoadLocation("America/New_York")
	}
	t := now.In(loc)

	if t.Weekday() == time.Saturday || t.Weekday() == time.Sunday {
		return false
	}
	open := time.Date(t.Year(), t.Month(), t.Day(), 9, 30, 0, 0, loc)
	close := time.Date(t.Year(), t.Month(), t.Day(), 16, 0, 0, 0, loc)
	return !t.Before(open) && t.Before(close)
}
