package main

import (
	"errors"
	"fmt"
	"os"
	"time"
)

// holiday describes a period of holiday with start and end dates
type holiday struct {
	Start time.Time
	End   time.Time
}

// String provides a string representation of a holiday
func (h holiday) String() string {
	tpl := `
  start: %s
  end  : %s
`
	return fmt.Sprintf(
		tpl,
		h.Start.Format("2006-01-02"),
		h.End.Format("2006-01-02"),
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
	return h, err

}

// holidays are a slice of holiday, which requires holidays to be
// entered sequentially in time with no overlaps.
type holidays []holiday

// addHoliday adds a holiday to a holidays slice, checking that it does
// not overlap other holiday items already in the slice.
func (hdays *holidays) addHoliday(h holiday) error {
	// check if there is an overlap
	for i, o := range *hdays {
		if h.Start.Before(o.End) {
			return fmt.Errorf("item %d (%v) overlaps with %v", i, o, h)
		}
	}
	*hdays = append(*hdays, h)
	return nil
}

// isHoliday determines if a provided date falls on a holiday, returning
// a boolean and an integer offset from the holidays slice recording
// where the holiday was found
func (hdays holidays) isHoliday(d time.Time) (found bool, offset int) {
	// day := durationHours(24)
	for i, h := range hdays {
		if (h.Start == d || h.Start.Before(d)) && (h.End == d || h.End.After(d)) {
			return true, i
		}
	}
	return false, -1
}

// window describes a window of time starting on start over which a
// count of holidays is made
type window struct {
	start time.Time
	count int // count of holidays in the window
}

// windows is a slice of window
type windows []window

func durationHours(h int) time.Duration {
	return time.Duration(h) * time.Hour
}

func dateFmt(d time.Time) string {
	return d.Format("2006-01-02")
}

func windowChecker(hols holidays, windowSizeDays int) (windows, error) {
	ws := windows{}
	if len(hols) == 0 {
		return ws, errors.New("holidays slice is empty")
	}

	// calculate start and end dates for windows calculator
	windowSizeDuration := durationHours(-24 * windowSizeDays)
	startDate := hols[0].Start
	endDate := hols[len(hols)-1].End

	if endDate.Add(windowSizeDuration).Before(startDate) {
		fmt.Println("before!")
		endDate = startDate
	}

	// run through the windows of time each offset by a day, starting
	// from startDate and ending on endDate, each of windowSizeDays
	day := durationHours(24)
	for d := startDate; d.Before(endDate.Add(day)); d = d.Add(day) {
		w := window{}
		w.start = d
		fmt.Printf("\n%s\n", dateFmt(d))
		for i := 0; i < windowSizeDays; i++ {
			offset := durationHours(24 * i)
			innerWindowDay := d.Add(offset)
			isHol, _ := hols.isHoliday(innerWindowDay)
			if isHol {
				w.count++
			}
		}
		ws = append(ws, w)
	}

	return ws, nil
}

func main() {

	windowSize := 40
	limit := 35

	h1, err := makeHoliday("2023-02-01", "2023-02-28")
	fmt.Println(h1, err)
	h2, err := makeHoliday("2023-04-02", "2023-05-10")
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

	fmt.Println("-----")
	ws, err := windowChecker(hols, windowSize)
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
	for i, w := range ws {
		if w.count > limit {
			fmt.Println(i, w.start.Format("2006-01-02"), w.count)
		}
	}

}
