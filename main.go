package main

import (
	"errors"
	"fmt"
	"os"
	"time"
)

// holiday describes a period of holiday with start and end dates
type holiday struct {
	Start          time.Time
	End            time.Time
	Days           int
	DaysToPrevious int // non-inclusive days from the end of the last holiday
}

// String provides a string representation of a holiday
func (h holiday) String() string {
	tpl := `
  start: %s
  end  : %s
  days : %d
  lag  : %d
`
	return fmt.Sprintf(
		tpl,
		h.Start.Format("2006-01-02"),
		h.End.Format("2006-01-02"),
		h.Days,
		h.DaysToPrevious,
	)
}

// makeHoliday makes a holiday struct from start and end string
// representations of dates. The end date may not be before the start
// date.
func makeHoliday(start, end string) (holiday, error) {

	f := func(s string) (time.Time, error) {
		return time.Parse("2006-01-02", s)
	}
	var h holiday
	var err error
	h.Start, err = f(start)
	if err != nil {
		return h, err
	}
	h.End, err = f(end)
	if h.End.Before(h.Start) {
		return h, errors.New("end date before start date")
	}
	// add a day to be a date inclusive range
	h.Days = int(h.End.Sub(h.Start).Hours()/24) + 1
	return h, err

}

// holidays are a slice of holiday, which requires holidays to be
// entered sequentially in time with no overlaps.
type holidays []holiday

// addHoliday adds a holiday to a holidays slice, checking that it does
// not overlap other holiday items already in the slice.
func (hdays *holidays) addHoliday(h holiday) error {
	var last holiday
	// check if there is an overlap
	for i, o := range *hdays {
		if h.Start.Before(o.End) {
			return errors.New(fmt.Sprintf("item %d (%v) overlaps with %v", i, o, h))
		}
		last = o
	}
	if len(*hdays) > 0 {
		h.DaysToPrevious = int(h.Start.Sub(last.End).Hours() / 24)
	}
	*hdays = append(*hdays, h)
	return nil
}

func main() {

	h1, err := makeHoliday("2023-02-01", "2023-02-28")
	fmt.Println(h1, err)
	h2, err := makeHoliday("2023-05-02", "2022-05-10")
	fmt.Println(h2, err)

	hols := holidays{}
	err = hols.addHoliday(h1)
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
	err = hols.addHoliday(h2)
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
	fmt.Printf("%+v\n", hols)

}
