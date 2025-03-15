package svg

import (
	"strings"
	"testing"
	"time"

	"github.com/rorycl/timeaway/trips"
)

func makeTrips() *trips.Trips {

	// Calculation results
	//
	// The planned trips breached the 90 days in 180 day rule with 91 days away.
	//
	// The maximum days away were for a 180 day window from Tuesday 01/07/2025 to Saturday 27/12/2025.
	//
	// The trips in this calculation are:
	//
	//  1. Tuesday 17/12/2024 to Saturday 04/01/2025 (19 days)
	//     not covered by the window.
	//  2. Friday 14/02/2025 to Thursday 27/02/2025 (14 days)
	//     not covered by the window.
	//  3. Thursday 03/04/2025 to Wednesday 23/04/2025 (21 days)
	//     not covered by the window.
	//  4. Tuesday 01/07/2025 to Wednesday 03/09/2025 (65 days)
	//     fully covered by the window.
	//  5. Sunday 09/11/2025 to Sunday 16/11/2025 (8 days)
	//     fully covered by the window.
	//  6. Wednesday 10/12/2025 to Tuesday 06/01/2026 (28 days)
	//     parially covered by the window from Wednesday 10/12/2025 for 18 days.

	return &trips.Trips{
		WindowSize: 180,
		MaxStay:    90,
		Start:      time.Date(2024, 12, 17, 0, 0, 0, 0, time.UTC),
		End:        time.Date(2026, 1, 6, 0, 0, 0, 0, time.UTC),
		OriginalHolidays: []trips.Holiday{
			trips.Holiday{
				Start:          time.Date(2024, 12, 17, 0, 0, 0, 0, time.UTC),
				End:            time.Date(2025, 1, 4, 0, 0, 0, 0, time.UTC),
				Duration:       19,
				PartialHoliday: nil,
			},
			trips.Holiday{
				Start:          time.Date(2025, 2, 14, 0, 0, 0, 0, time.UTC),
				End:            time.Date(2025, 2, 27, 0, 0, 0, 0, time.UTC),
				Duration:       14,
				PartialHoliday: nil,
			},
			trips.Holiday{
				Start:          time.Date(2025, 4, 3, 0, 0, 0, 0, time.UTC),
				End:            time.Date(2025, 4, 23, 0, 0, 0, 0, time.UTC),
				Duration:       21,
				PartialHoliday: nil,
			},
			trips.Holiday{
				Start:          time.Date(2025, 7, 1, 0, 0, 0, 0, time.UTC),
				End:            time.Date(2025, 9, 3, 0, 0, 0, 0, time.UTC),
				Duration:       65,
				PartialHoliday: nil,
			},
			trips.Holiday{
				Start:          time.Date(2025, 11, 9, 0, 0, 0, 0, time.UTC),
				End:            time.Date(2025, 11, 16, 0, 0, 0, 0, time.UTC),
				Duration:       8,
				PartialHoliday: nil,
			},
			trips.Holiday{
				Start:          time.Date(2025, 12, 10, 0, 0, 0, 0, time.UTC),
				End:            time.Date(2026, 1, 6, 0, 0, 0, 0, time.UTC),
				Duration:       28,
				PartialHoliday: nil,
			},
		},
		Window: trips.Window{
			Start:    time.Date(2025, 7, 1, 0, 0, 0, 0, time.UTC),
			End:      time.Date(2025, 12, 27, 0, 0, 0, 0, time.UTC),
			DaysAway: 91,
			Overlaps: 3,
			Holidays: []trips.Holiday{
				trips.Holiday{
					Start:          time.Date(2024, 12, 17, 0, 0, 0, 0, time.UTC),
					End:            time.Date(2025, 1, 4, 0, 0, 0, 0, time.UTC),
					Duration:       19,
					PartialHoliday: nil,
				},
				trips.Holiday{
					Start:          time.Date(2025, 2, 14, 0, 0, 0, 0, time.UTC),
					End:            time.Date(2025, 2, 27, 0, 0, 0, 0, time.UTC),
					Duration:       14,
					PartialHoliday: nil,
				},
				trips.Holiday{
					Start:          time.Date(2025, 4, 3, 0, 0, 0, 0, time.UTC),
					End:            time.Date(2025, 4, 23, 0, 0, 0, 0, time.UTC),
					Duration:       21,
					PartialHoliday: nil,
				},
				trips.Holiday{
					Start:    time.Date(2025, 7, 1, 0, 0, 0, 0, time.UTC),
					End:      time.Date(2025, 9, 3, 0, 0, 0, 0, time.UTC),
					Duration: 65,
					PartialHoliday: &trips.Holiday{
						Start:          time.Date(2025, 7, 1, 0, 0, 0, 0, time.UTC),
						End:            time.Date(2025, 9, 3, 0, 0, 0, 0, time.UTC),
						Duration:       65,
						PartialHoliday: nil,
					},
				},
				trips.Holiday{
					Start:    time.Date(2025, 11, 9, 0, 0, 0, 0, time.UTC),
					End:      time.Date(2025, 11, 16, 0, 0, 0, 0, time.UTC),
					Duration: 8,
					PartialHoliday: &trips.Holiday{
						Start:          time.Date(2025, 11, 9, 0, 0, 0, 0, time.UTC),
						End:            time.Date(2025, 11, 16, 0, 0, 0, 0, time.UTC),
						Duration:       8,
						PartialHoliday: nil,
					},
				},
				trips.Holiday{
					Start:    time.Date(2025, 12, 10, 0, 0, 0, 0, time.UTC),
					End:      time.Date(2026, 1, 6, 0, 0, 0, 0, time.UTC),
					Duration: 28,
					PartialHoliday: &trips.Holiday{
						Start:          time.Date(2025, 12, 10, 0, 0, 0, 0, time.UTC),
						End:            time.Date(2025, 12, 27, 0, 0, 0, 0, time.UTC),
						Duration:       18,
						PartialHoliday: nil,
					},
				},
			},
		},
		LongestDaysAway: 91,
		Error:           nil,
		Breach:          true,
	}
}

func TestSVG(t *testing.T) {

	var svgOutput strings.Builder
	trips := makeTrips()

	err := TripsAsSVG(trips, &svgOutput)
	if err != nil {
		t.Fatal(err)
	}

	got, want := svgOutput.String(), "<title>breach (91 days) : 2025-07-01 to 2025-12-27</title>"
	if !strings.Contains(
		got,
		want,
	) {
		t.Errorf("expected breach group title")
	}
}
