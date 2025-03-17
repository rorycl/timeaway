package trips

import (
	"errors"
	"fmt"
	"log"
	"os"
	"strings"
	"time"

	"github.com/sanity-io/litter"
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
	WindowSize       int       // size of window of days to search over
	MaxStay          int       // the maximum length of trips in window
	Start, End       time.Time // the start and end of the overall holidays
	startFrame       time.Time // date at which to start calculating windows
	endFrame         time.Time // date at which to stop calculating windows
	OriginalHolidays []Holiday // list of holidays under consideration
	Window                     // the window with the longest compound trip length
	LongestDaysAway  int       // used during window calculations
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
		trips.WindowAsStr,
	)
}

// WindowAsStr returns a string representation of the longest window
func (trips *Trips) WindowAsStr() string {
	return fmt.Sprint(trips.Window)
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
	trips.WindowSize = WindowMaxDays
	trips.MaxStay = CompoundStayMaxDays
	return &trips, nil
}

// addHoliday adds a holiday to Trips, checking for validity and overlaps
func (trips *Trips) addHoliday(h Holiday) error {

	// check validity of this holiday
	if h.End.Before(h.Start) {
		return fmt.Errorf("start date %s after %s", dayShortFmt(h.Start), dayShortFmt(h.End))
	}
	// check no overlaps
	for _, o := range trips.OriginalHolidays {
		if ok := o.overlaps(h.Start, h.End); ok != nil {
			return fmt.Errorf(
				"trip %s to %s overlaps with %s to %s",
				dayShortFmt(h.Start), dayShortFmt(h.End), dayShortFmt(o.Start), dayShortFmt(o.End),
			)
		}
	}
	// set window dates; endFrame gets reset during calculation, so use
	// Start and End for overall start/end
	x := Holiday{}
	if trips.startFrame == x.Start || trips.startFrame.After(h.Start) {
		trips.startFrame = h.Start
		trips.Start = trips.startFrame
	}
	if trips.endFrame.Before(h.End) {
		trips.endFrame = h.End
		trips.End = trips.endFrame
	}
	h.Duration = h.days()

	trips.OriginalHolidays = append(trips.OriginalHolidays, h)
	return nil
}

// Window stores the results of a calculation window. The window with
// the longest compound trip length (or `DaysAway`) is copied into the
// Trips struct.
//
// Start and End describe the start and end of the window (typically 180
// days). While the longest overlap between the window in question and
// holidays is reported by OverlapStart and OverlapEnd.
//
// Holidays are copied from Trips.OriginalHolidays to window.Holidays and
// decorated with partial Holidays where these overlap by the calculate
// function. The window with the longest DaysAway is copied to Trips.
type Window struct {
	Start        time.Time `json:"start"`        // start of this window
	End          time.Time `json:"end"`          // end of this window
	DaysAway     int       `json:"daysAway"`     // days away during this window
	Overlaps     int       `json:"overlaps"`     // number of holiday overlaps
	OverlapStart time.Time `json:"overlapStart"` // start of the overlap
	OverlapEnd   time.Time `json:"overlapEnd"`   // end of the overlap
	Holidays     []Holiday `json:"holidays"`     // trips.OriginalHolidays decorated with overlaps
}

// String returns a printable version of a window
func (w Window) String() string {
	tpl := "window %s:%s overlap %s:%s (%d days, %d overlaps)"
	s := fmt.Sprintf(
		tpl, dayFmt(w.Start), dayFmt(w.End), dayFmt(w.OverlapStart), dayFmt(w.OverlapEnd), w.DaysAway, w.Overlaps,
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
	if trips.MaxStay == 0 || trips.WindowSize == 0 {
		return trips, errors.New("trip not properly initialised")
	}
	if len(trips.OriginalHolidays) < 1 {
		return trips, errors.New("no holidays provided")
	}

	// set suitable frame start and end in which to calculate windows
	windowDuration := durationDays(trips.WindowSize - 1) // remove last day
	trips.endFrame = trips.endFrame.Add(-windowDuration)
	if trips.endFrame.Before(trips.startFrame) {
		trips.endFrame = trips.startFrame
	}

	// generate a series of windows starting on each day between
	// trips.startFrame and trips.endFrame.
	//
	// For each window, if the windows.DaysAway > Trips.LongestDaysAway,
	// embed the window in the Trips struct.
	//
	// This loop could be moved to a set of goroutines although
	// peformance for very large windows is still very quick, around
	// 0.005s for a 720 day/180 stay use case.
	for d := trips.startFrame; !d.After(trips.endFrame); d = d.Add(durationDays(1)) {
		w := Window{}
		w.Start = d
		w.End = d.Add(windowDuration)

		w.Holidays = make([]Holiday, len(trips.OriginalHolidays))
		copy(w.Holidays, trips.OriginalHolidays)

		for i, t := range w.Holidays {
			partialHoliday := t.overlaps(w.Start, w.End)
			if partialHoliday == nil {
				continue
			}
			partialHoliday.Duration = partialHoliday.days()
			w.Overlaps++
			w.DaysAway += partialHoliday.Duration
			w.Holidays[i].PartialHoliday = partialHoliday
			if w.OverlapStart.IsZero() {
				w.OverlapStart = partialHoliday.Start
			}
			if w.OverlapEnd.Before(partialHoliday.End) {
				w.OverlapEnd = partialHoliday.End
			}

			// set longest trip/window if appropriate
			if w.DaysAway > trips.LongestDaysAway {
				trips.LongestDaysAway = w.DaysAway
				trips.Window = w
			}
			if w.DaysAway > trips.MaxStay {
				trips.Breach = true
			}
		}
	}
	dumper := func() {
		f, err := os.Create("trips_dump.txt")
		if err != nil {
			log.Printf("dump file open error %v", err)
			return
		}
		defer f.Close()
		litter.Config.FormatTime = true
		litter.Config.DisablePointerReplacement = true
		tDump := litter.Sdump(trips)
		_, err = f.Write([]byte(tDump))
		if err != nil {
			log.Printf("dump write error %v", err)
			return
		}
	}
	dumper()
	return trips, nil
}

// Calculate initialises a new Trips struct with the package variables
// WindowSize (the number of days over which to do the calculation) and
// MaxStay (the length of compound holidays) and then sequentially adds
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
