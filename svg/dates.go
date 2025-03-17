package svg

import (
	"errors"
	"fmt"
	"time"
)

// changeDate is a function to either advance or retreat towards a day
// of the week. targetDay is the day reported by date.Weekday
func changeDate(date time.Time, targetDay int, d time.Duration) (time.Time, error) {
	if targetDay < 0 || targetDay > 6 {
		return time.Time{}, fmt.Errorf("targetDay %d out of bounds", targetDay)
	}
	if date.IsZero() {
		return time.Time{}, errors.New("date is empty")
	}

	if int(date.Weekday()) == targetDay {
		return date, nil
	}
	for i := 0; i < 6; i++ {
		date = date.Add(d)
		if int(date.Weekday()) == targetDay {
			return date, nil
		}
	}
	return time.Time{}, fmt.Errorf("targetDay %d fall through error", targetDay)
}

// isoDOW returns an ISO day of week number where numbering starts
// from 0 for Monday to 6 for Sunday. Note that the ISO standard is
// actually 1 indexed rather than 0 indexed.
func isoDOW(date time.Time) int {
	dow := (int(date.Weekday()) - 1) % 7
	if dow == -1 {
		return 6
	}
	return dow
}
