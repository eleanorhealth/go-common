package date

import (
	"errors"
	"time"
)

var (
	ErrNoLayoutMatched = errors.New("no layout matched")
)

// EOD returns the end of the day in the provided timezone
func EOD(t time.Time, loc *time.Location) time.Time {
	if loc == nil {
		loc = t.Location()
	}

	return time.Date(t.Year(), t.Month(), t.Day(), 23, 59, 59, 0, loc)
}

// BOD returns the beginning of the day in the provided timezone
func BOD(t time.Time, loc *time.Location) time.Time {
	if loc == nil {
		loc = t.Location()
	}

	return time.Date(t.Year(), t.Month(), t.Day(), 0, 0, 0, 0, loc)
}

// Days returns number of whole days that the duration represents (floored to nearest day)
func Days(duration time.Duration) int {
	return int(duration) / int(time.Hour*24)
}

// WithinDuration tests to see if two time.Time's are within a duration of eachother
func WithinDuration(expected, actual time.Time, delta time.Duration) bool {
	dt := expected.Sub(actual)

	return dt > -delta && dt < delta
}

// ParseAny parses a string into a time.Time according to the first successful layout format
func ParseAny(layouts []string, dateString string) (time.Time, error) {
	for _, layout := range layouts {
		t, err := time.Parse(layout, dateString)
		if err != nil {
			continue
		}

		return t, nil
	}

	return time.Time{}, ErrNoLayoutMatched
}
