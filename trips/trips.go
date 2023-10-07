package trips

import (
	"errors"
	"fmt"
	"strings"
	"time"
)

var (
	// WindowMaxDays maximum window of days to calculate over
	WindowMaxDays int = 180
	// CompoundStayMaxDays is the longest allowed compound trip length
	CompoundStayMaxDays int = 90
)

// Trips describe a set of trips and other metadata
type Trips struct {
	windowSize int       // size of window of days to search over
	maxStay    int       // the maximum length of trips in window
	startFrame time.Time // date at which to start calculating windows
	endFrame   time.Time // date at which to stop calculating windows
	windows    []window  // all windows kept for testing
	/* exported variables */
	Error    error     `json:"error"`
	Window   window    `json:"longestWindow"`
	DaysAway int       `json:"daysAway"`
	Holidays []Holiday `json:"holidays"` // longest days away
	Breach   bool      `json:"breach"`
}

// String returns a simple string representation of trips
func (trips Trips) String() string {
	tpl := `
		breach         : %t
		days away      : %d
		largest window : %+v
	`
	tpl = strings.ReplaceAll(tpl, "\t", "")
	return fmt.Sprintf(
		tpl,
		trips.Breach,
		trips.DaysAway,
		trips.Window,
	)
}

// newTrips makes a new Trips struct after checking the overrideable
// package variables are ok
func newTrips() (*Trips, error) {
	trips := Trips{}
	trips.Breach = false
	if WindowMaxDays < 3 {
		return &trips, errors.New("window size cannot be less than 3 days")
	}
	if CompoundStayMaxDays < 2 {
		return &trips, errors.New("maximum stay cannot be less than 2 days")
	}
	if CompoundStayMaxDays > WindowMaxDays {
		return &trips, errors.New("maximum stay cannot be greater than the window size")
	}
	trips.windowSize = WindowMaxDays
	trips.maxStay = CompoundStayMaxDays
	return &trips, nil
}

// addHoliday adds a holiday to Trips, checking for validity and overlaps
func (trips *Trips) addHoliday(h Holiday) error {

	// check validity of this holiday
	if h.End.Before(h.Start) {
		return fmt.Errorf("start date %s after %s", dayShortFmt(h.Start), dayShortFmt(h.End))
	}
	// check no overlaps
	for _, o := range trips.Holidays {
		if ok := o.overlaps(h.Start, h.End); ok != nil {
			return fmt.Errorf(
				"trip %s to %s overlaps with %s to %s",
				h.Start, h.End, dayShortFmt(o.Start), dayShortFmt(o.End),
			)
		}
	}

	// set window dates
	x := Holiday{}
	if trips.startFrame == x.Start || trips.startFrame.After(h.Start) {
		trips.startFrame = h.Start
	}
	if trips.endFrame.Before(h.End) {
		trips.endFrame = h.End
	}
	h.Duration = h.days()

	trips.Holidays = append(trips.Holidays, h)
	return nil
}

// window stores the results of a calculation window
type window struct {
	Start        time.Time `json:Start`
	End          time.Time `json:End`
	HolidayParts []Holiday `json:partialHolidays` // parts of any overlapping holidays
	DaysAway     int       `json:daysAway`        // days away for this window
}

// String returns a printable version of a window
func (w window) String() string {
	tpl := "%s : %s (%d)\n    components: "
	s := fmt.Sprintf(
		tpl, DayFmt(w.Start), DayFmt(w.End), w.DaysAway,
	)
	if len(w.HolidayParts) == 0 {
		s += "none"
		return s
	}
	for _, h := range w.HolidayParts {
		s = s + fmt.Sprintf(" %s", h)
	}
	return s
}

// calculate performs the window calculation returning the Trips struct
// and error for returning by Calculate.
// The window calculator could be moved to goroutines to speed up
// processing, although it seems sufficiently fast already.
func (trips *Trips) calculate() (*Trips, error) {

	// check trips has been properly initialised and there are holidays
	// to process
	if trips.maxStay == 0 || trips.windowSize == 0 {
		return trips, errors.New("trip not properly initialised")
	}
	if len(trips.Holidays) < 1 {
		return trips, errors.New("no holidays provided")
	}

	// set suitable frame start and end in which to calculate windows
	windowDuration := durationDays(trips.windowSize - 1) // remove last day
	trips.endFrame = trips.endFrame.Add(-windowDuration)
	if trips.endFrame.Before(trips.startFrame) {
		trips.endFrame = trips.startFrame
	}

	// generate a series of windows starting on each day between
	// trips.startFrame and trips.endFrame and store the results in
	// trips.windows. This loop could be moved to a set of goroutines
	// although peformance for very large windows is still very quick,
	// around 0.005s for a 720 day/180 stay use case.
	for d := trips.startFrame; !d.After(trips.endFrame); d = d.Add(durationDays(1)) {
		w := window{}
		w.Start = d
		w.End = d.Add(windowDuration)

		for _, t := range trips.Holidays {
			partialTrip := t.overlaps(w.Start, w.End)
			if partialTrip == nil {
				continue
			}
			partialTrip.Duration = partialTrip.days()
			w.HolidayParts = append(w.HolidayParts, *partialTrip)
			w.DaysAway += partialTrip.days()

			// set longest trip/window if appropriate
			if w.DaysAway > trips.DaysAway {
				trips.DaysAway = w.DaysAway
				trips.Window = w
			}
			if w.DaysAway > trips.maxStay {
				trips.Breach = true
			}
		}
		trips.windows = append(trips.windows, w) // kept for testing
	}
	return trips, nil
}

// Calculate initialises a new Trips struct with the configuration
// windowSize (the number of days over which to do the calculation) and
// maxStay (the length of compound holidays), adds holidays then runs
// the calculation, returning the resulting Trips object, and error if
// any.
func Calculate(hols []Holiday) (*Trips, error) {

	// initialise Trips
	trips, err := newTrips()
	trips.Error = err
	if trips.Error != nil {
		return trips, trips.Error
	}

	// add holidays
	if len(hols) == 0 {
		trips.Error = errors.New("no trips were provided to calculate")
		return trips, trips.Error
	}
	for _, h := range hols {
		trips.Error = trips.addHoliday(h)
		if trips.Error != nil {
			return trips, trips.Error
		}
	}

	// perform the calculation
	return trips.calculate()

}
