package trips

import (
	"testing"
)

// generate trips for tests
func generateTrips(addhols int) (*Trips, error) {

	window := 5
	compoundStayMaxLength := 4

	trips, err := NewTrips(window, compoundStayMaxLength)
	if err != nil {
		return trips, err
	}

	for i, h := range []holiday{
		holiday{"2023-01-01", "2023-01-01"},
		holiday{"2023-01-06", "2023-01-07"},
		holiday{"2023-01-11", "2023-01-12"},
		holiday{"2023-01-15", "2023-01-15"},
		holiday{"2023-01-21", "2023-01-22"},
		holiday{"2023-01-24", "2023-01-25"},
		holiday{"2023-01-24", "2023-01-25"}, // error at 7th add
	} {
		if i > addhols-1 {
			break
		}
		err = trips.AddTrip(h.start, h.end)
		if err != nil {
			return trips, err
		}
	}
	err = trips.Calculate()
	return trips, err
}

func TestJSONSuccess(t *testing.T) {

	trs, err := generateTrips(3)
	if err != nil {
		t.Fatalf("generate trs error %v\n", err)
	}

	// get json result
	jsonResult, err := trs.AsJSON()
	if err != nil {
		t.Fatalf("json reading error %v\n", err)
	}

	// unmarshal back to struct
	check, err := UnmarshalTripsJSON(jsonResult)
	if err != nil {
		t.Fatalf("unmarshal error %v", err)
	}

	if check.Breach != trs.breach {
		t.Errorf("breach should be %v, got %v", trs.breach, check.Breach)
	}
	if check.DaysAway != trs.longestStay {
		t.Errorf("window days away should be %v got %v", trs.longestStay, check.DaysAway)
	}

	if len(check.PartialTrips) != len(trs.windows[0].TripParts) {
		t.Errorf("partial trs should be %d, got %d", len(trs.windows[0].TripParts), len(check.PartialTrips))
	}

	if len(check.Holidays) != len(trs.trips) {
		t.Errorf("number of trip want %d, got %d", len(trs.windows[0].TripParts), len(check.PartialTrips))
	}

}

func TestJSONError(t *testing.T) {

	trs, err := generateTrips(7)
	if err == nil {
		t.Fatal("expected generate trips error")
	}

	t.Logf("trs %+v\n", trs)

	// get json result
	jsonResult, err := trs.AsJSON()
	if err != nil {
		t.Fatalf("json reading error %v\n", err)
	}

	// unmarshal back to struct
	check, err := UnmarshalTripsJSON(jsonResult)
	if err != nil {
		t.Fatalf("unmarshal error %v", err)
	}

	if check.Error != "no results found" {
		t.Errorf("expected error %v, got %v", "no results found", check.Error)
	}

}
