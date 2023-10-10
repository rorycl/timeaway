package trips

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/url"
	"time"

	"github.com/go-playground/form"
)

// Holiday describes a trip to a EU state with Start and End date. A
// Holiday describes a period of with a duration of at least one day
// (when Start and End are the same date). A holiday End date may not be
// before its Start.
// The Holiday struct is also used to describe partial holidays for
// `window.HolidayParts`.
type Holiday struct {
	Start    time.Time `json:"Start"`                        // start date
	End      time.Time `json:"End"`                          // end date
	Duration int       `json:",omitempty" form:",omitempty"` // duration in days
}

// newHoliday returns a new Holiday from two dates (time.Time values)
func newHoliday(s, e time.Time) (*Holiday, error) {
	h := new(Holiday)
	if s.After(e) {
		return h, fmt.Errorf("start date %s after %s", dayShortFmt(s), dayShortFmt(e))
	}
	empty := time.Time{}
	if s == empty {
		return h, errors.New("start date not set")
	}
	if e == empty {
		return h, errors.New("end date not set")
	}
	h.Start = s
	h.End = e
	h.Duration = h.days()
	return h, nil
}

// newHoliday returns a new Holiday from two date strings
func newHolidayFromStr(s, e string) (*Holiday, error) {
	h := new(Holiday)
	c := func(s string) (time.Time, error) {
		ti, err := time.Parse("2006-01-02", s)
		if err != nil {
			return ti, err
		}
		return ti, nil
	}
	st, err := c(s)
	if err != nil {
		return h, err
	}
	et, err := c(e)
	if err != nil {
		return h, err
	}
	return newHoliday(st, et)
}

// HolidaysURLDecoder decodes a set of holidays provided as a URL.Query
func HolidaysURLDecoder(input url.Values) ([]Holiday, error) {

	// holidaysFromURL is a struct suitable for decoding parameters provided
	// in a url eg `?Start=2022-12-18&End=2023-01-07&Start=2023-02-10&End=2023-02-15`
	type holidaysFromURL struct {
		Start []time.Time
		End   []time.Time
	}

	holsByURL := holidaysFromURL{}
	hols := []Holiday{}

	decoder := form.NewDecoder()
	decoder.RegisterCustomTypeFunc(func(vals []string) (interface{}, error) {
		return time.Parse("2006-01-02", vals[0])
	}, time.Time{})

	err := decoder.Decode(&holsByURL, input)
	if err != nil {
		return hols, err
	}
	if len(holsByURL.Start) < 1 {
		return hols, nil
	}
	if len(holsByURL.Start) != len(holsByURL.End) {
		return hols, errors.New("incorrect number of url arguments")
	}
	for i := 0; i < len(holsByURL.Start); i++ {
		h, err := newHoliday(holsByURL.Start[i], holsByURL.End[i])
		if err != nil {
			return hols, err
		}
		hols = append(hols, *h)
	}
	return hols, err
}

// HolidaysJSONDecoder decodes a set of holidays provided as JSON
func HolidaysJSONDecoder(input []byte) ([]Holiday, error) {

	var hols []Holiday
	// internal struct to convert from 2006-01-02 values by first
	// converting to string
	type jsonHoliday struct {
		Start string
		End   string
	}
	var jsonHols []jsonHoliday
	err := json.Unmarshal(input, &jsonHols)
	if err != nil {
		return hols, err
	}
	if len(jsonHols) < 1 {
		return hols, nil
	}
	// make Holiday objects from each jsonHoliday in the slice
	for _, j := range jsonHols {

		hol, err := newHolidayFromStr(j.Start, j.End)
		if err != nil {
			return hols, err
		}
		hols = append(hols, *hol)
	}
	return hols, nil
}

// String returns a string representation of a holiday
func (h Holiday) String() string {
	return fmt.Sprintf(
		"%s to %s (%d)", dayFmt(h.Start), dayFmt(h.End), h.Duration,
	)
}

// days returns the number of inclusive days between the start and end
// dates of a holiday
func (h Holiday) days() int {
	days := 0
	for d := h.Start; !d.After(h.End); d = d.Add(durationDays(1)) {
		days++
	}
	return days
}

// overlaps returns a pointer to a partial or full holiday if there is
// an overlap with the provided dates, else a nil pointer
func (h Holiday) overlaps(start, end time.Time) *Holiday {
	partialTrip := new(Holiday)
	// no overlap
	if h.Start.After(end) || h.End.Before(start) {
		return nil
	}
	// contained
	if h.Start.After(start) && h.End.Before(end) {
		partialTrip.Start = h.Start
		partialTrip.End = h.End
		return partialTrip
	}
	// partial overlap
	if h.Start.Before(start) || h.Start == start {
		partialTrip.Start = start
	} else {
		partialTrip.Start = h.Start
	}
	if h.End.After(end) || h.End == end {
		partialTrip.End = end
	} else {
		partialTrip.End = h.End
	}
	partialTrip.Duration = partialTrip.days()
	return partialTrip
}

// durationDays returns a duration for the number of days specified
func durationDays(d int) time.Duration {
	return time.Duration(d) * time.Hour * 24
}

// dayFmt returns a custom string representation of a date
func dayFmt(d time.Time) string {
	return d.Format("Monday 2 January 2006")
}

// dayShortFmt returns a short custom string representation of a date
func dayShortFmt(d time.Time) string {
	return d.Format("02/01/2006")
}
