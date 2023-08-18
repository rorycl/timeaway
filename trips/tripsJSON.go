package trips

import (
	"encoding/json"
	"time"
)

// TripsJSON provides a json representation of a Trips structure for the
// window with the longest compound trips. Where several windows have
// the same trip length the earliest window is used.
type TripsJSON struct {
	Error        string    `json:"error"`
	Breach       bool      `json:"breach"`
	StartDate    time.Time `json:"windowStart"`
	EndDate      time.Time `json:"windowEnd"`
	DaysAway     int       `json:"windowDaysAway"`
	PartialTrips []trip    `json:"partialTrips"`
	Holidays     []trip    `json:"holidays"`
}

// UnmarshalTripsJSON decodes a byte sequence to a TripsJSON struct
func UnmarshalTripsJSON(b []byte) (TripsJSON, error) {
	tj := TripsJSON{}
	err := json.Unmarshal(b, &tj)
	return tj, err
}

// AsJSON returns a json summary of trips
func (trips *Trips) AsJSON() ([]byte, error) {

	tj := TripsJSON{}

	breach, windows := trips.LongestTrips(1) // get longest window only
	if len(windows) == 0 {
		tj.Error = "no results found"
		return json.Marshal(tj)
	}

	tj.Breach = breach
	window := windows[0] // only top (longest & earliest) window of interest
	tj.StartDate = window.Start
	tj.EndDate = window.End
	tj.DaysAway = window.DaysAway
	for _, pt := range window.TripParts {
		tj.PartialTrips = append(tj.PartialTrips,
			trip{pt.Start, pt.End, pt.Days()},
		)
	}
	for _, t := range trips.trips {
		tj.Holidays = append(tj.Holidays,
			trip{t.Start, t.End, t.Days()},
		)
	}
	return json.Marshal(tj)
}
