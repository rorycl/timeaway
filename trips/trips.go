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

// Trips describe a set of holidays and their calculation results.
//
// Details following calculation are largely held in the `window`
// embedded struct which reports the window calculated to have longest
// number of days away. Where more than one window has the same number
// of days away, the window with the earliest date is used.
type Trips struct {
	windowSize       int       // size of window of days to search over
	maxStay          int       // the maximum length of trips in window
	startFrame       time.Time // date at which to start calculating windows
	endFrame         time.Time // date at which to stop calculating windows
	originalHolidays []Holiday // list of holidays under consideration
	window                     // the window with the longest compound trip length
	longestDaysAway  int       // used during window calculations
	Error            error     `json:"error"`  // calculation errors
	Breach           bool      `json:"breach"` // if CompoundStayMaxDays is breached
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
		trips.window,
	)
}

// Window returns a string representation of the longest window
func (trips *Trips) Window() string {
	return fmt.Sprint(trips.window)
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
	for _, o := range trips.originalHolidays {
		if ok := o.overlaps(h.Start, h.End); ok != nil {
			return fmt.Errorf(
				"trip %s to %s overlaps with %s to %s",
				dayShortFmt(h.Start), dayShortFmt(h.End), dayShortFmt(o.Start), dayShortFmt(o.End),
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

	trips.originalHolidays = append(trips.originalHolidays, h)
	return nil
}

// window stores the results of a calculation window. The window with
// the longest compound trip length (or `DaysAway`) is copied into the
// Trips struct.
//
// Holidays are copied from Trips.originalHolidays to window.Holidays and
// decorated with partial Holidays where these overlap by the calculate
// function. The window with the longest DaysAway is copied to Trips.
type window struct {
	Start    time.Time `json:"start"`    // start of this window
	End      time.Time `json:"end"`      // end of this window
	DaysAway int       `json:"daysAway"` // days away during this window
	Overlaps int       `json:"overlaps"` // number of holiday overlaps
	Holidays []Holiday `json:"holidays"` // trips.originalHolidays decorated with overlaps
}

// String returns a printable version of a window
func (w window) String() string {
	tpl := "%s : %s (%d days, %d overlaps)"
	s := fmt.Sprintf(
		tpl, dayFmt(w.Start), dayFmt(w.End), w.DaysAway, w.Overlaps,
	)
	return s
}

// calculate performs the window calculation returning the Trips struct
// and error for returning by Calculate.
//
// The window with the longest trip (`window.DaysAway`) are embedded in
// the Trips struct.
//
// The window calculator could be moved to goroutines to speed up
// processing, although it seems sufficiently fast already.
func (trips *Trips) calculate() (*Trips, error) {

	// check trips has been properly initialised and there are holidays
	// to process
	if trips.maxStay == 0 || trips.windowSize == 0 {
		return trips, errors.New("trip not properly initialised")
	}
	if len(trips.originalHolidays) < 1 {
		return trips, errors.New("no holidays provided")
	}

	// set suitable frame start and end in which to calculate windows
	windowDuration := durationDays(trips.windowSize - 1) // remove last day
	trips.endFrame = trips.endFrame.Add(-windowDuration)
	if trips.endFrame.Before(trips.startFrame) {
		trips.endFrame = trips.startFrame
	}

	// generate a series of windows starting on each day between
	// trips.startFrame and trips.endFrame.
	//
	// For each window, if the windows.DaysAway > Trips.longestDaysAway,
	// embed the window in the Trips struct.
	//
	// This loop could be moved to a set of goroutines although
	// peformance for very large windows is still very quick, around
	// 0.005s for a 720 day/180 stay use case.
	for d := trips.startFrame; !d.After(trips.endFrame); d = d.Add(durationDays(1)) {
		w := window{}
		w.Start = d
		w.End = d.Add(windowDuration)

		w.Holidays = make([]Holiday, len(trips.originalHolidays))
		copy(w.Holidays, trips.originalHolidays)

		for i, t := range w.Holidays {
			partialHoliday := t.overlaps(w.Start, w.End)
			if partialHoliday == nil {
				continue
			}
			partialHoliday.Duration = partialHoliday.days()
			w.Overlaps++
			w.DaysAway += partialHoliday.Duration
			w.Holidays[i].PartialHoliday = partialHoliday

			// set longest trip/window if appropriate
			if w.DaysAway > trips.longestDaysAway {
				trips.longestDaysAway = w.DaysAway
				trips.window = w
			}
			if w.DaysAway > trips.maxStay {
				trips.Breach = true
			}
		}
	}
	return trips, nil
}

// Calculate initialises a new Trips struct with the package variables
// windowSize (the number of days over which to do the calculation) and
// maxStay (the length of compound holidays) and then sequentially adds
// holidays, then runs the calculation, returning the resulting Trips
// object and embedded window (with the longest DaysAway), and error if
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
