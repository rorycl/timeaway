package trips

import (
	"fmt"
	"testing"
)

// holiday describes the start and end dates for a holiday in string
// format (using 2006-01-02 format)
type holiday struct {
	start string
	end   string
}

// overlap detection checks
func TestTripAdditions(t *testing.T) {

	window := 5
	compoundStayMaxLength := 3

	trips, err := NewTrips(window, compoundStayMaxLength)
	if err != nil {
		t.Fatalf("could not make trips %v", err)
	}

	type holidayTest struct {
		start string
		end   string
		err   bool
		msg   string
	}
	for i, h := range []holidayTest{
		holidayTest{"2023-01-01", "2023-01-01", false, "simple add"},
		holidayTest{"2023-01-01", "2023-01-07", true, "overlap already registered date"},
		holidayTest{"2023-01-02", "2023-01-04", false, "simple add again"},
		holidayTest{"2023-01-03", "2023-01-05", true, "overlap two days"},
		holidayTest{"2023-01-04", "2023-01-05", true, "overlap one day"},
		holidayTest{"2022-12-31", "2023-01-01", true, "overlap first day"},
		holidayTest{"2022-12-31", "2022-12-30", true, "end before start"},
		holidayTest{"2022-11-01", "2022-11-02", false, "add before first should be ok"},
	} {
		err = trips.AddTrip(h.start, h.end)
		if (err != nil && !h.err) || (err == nil && h.err) {
			t.Errorf("addtrip error for test %d : %v", i, h)
		}
	}
}

// TestTrips checks for a 5 day window over a maximum stay length of 4
// days starting on 1 January
//
//	x....xx...xx..x.....xx.xx
//	1    2    3         4       <- max stay
func TestTrips(t *testing.T) {

	window := 5
	compoundStayMaxLength := 4
	resultsNo := 4

	trips, err := NewTrips(window, compoundStayMaxLength)
	if err != nil {
		t.Fatalf("could not make trips %v", err)
	}

	for i, h := range []holiday{
		holiday{"2023-01-01", "2023-01-01"},
		holiday{"2023-01-06", "2023-01-07"},
		holiday{"2023-01-11", "2023-01-12"},
		holiday{"2023-01-15", "2023-01-15"},
		holiday{"2023-01-21", "2023-01-22"},
		holiday{"2023-01-24", "2023-01-25"},
	} {
		err = trips.AddTrip(h.start, h.end)
		if err != nil {
			t.Fatalf("error making holiday %d %v %v", i, h, err)
		}
	}

	err = trips.Calculate()
	if err != nil {
		t.Fatalf("calculation error %v", err)
	}

	t.Log(trips)

	breach, windows := trips.LongestTrips(resultsNo)
	for i, w := range windows {
		fmt.Printf("%d : %+v\n", i, w)
	}

	if trips.longestStay != 4 {
		t.Errorf("Expected longest stay to be 4, got %d", trips.longestStay)
	}

	if breach != false {
		t.Error("Expected breach to be false, got true")
	}

	// get json result
	jsonResult, err := trips.AsJSON()
	if err != nil {
		t.Fatalf("json reading error %v\n", err)
	}

	// unmarshal back to struct
	check, err := UnmarshalTripsJSON(jsonResult)
	if err != nil {
		t.Fatalf("unmarshal error %v", err)
	}

	// t.Logf("tripsjson %+v\n", check)

	if check.Breach != false {
		t.Errorf("breach should be false, got %v", check.Breach)
	}
	if check.DaysAway != trips.longestStay {
		t.Errorf("window days away should be %v got %v", trips.longestStay, check.DaysAway)
	}

	if len(check.PartialTrips) != 2 {
		t.Errorf("partial trips should be 2, got %d", len(check.PartialTrips))
	}

	if len(check.Holidays) != 6 {
		t.Errorf("holiday trips should be 6, got %d", len(check.Holidays))
	}

}

// the same test as above, but with a lower compoundStayMaxLength to
// breach
func TestTripsToBreach(t *testing.T) {

	window := 5
	compoundStayMaxLength := 3
	resultsNo := 4

	trips, err := NewTrips(window, compoundStayMaxLength)
	if err != nil {
		t.Fatalf("could not make trips %v", err)
	}

	for i, h := range []holiday{
		holiday{"2023-01-01", "2023-01-01"},
		holiday{"2023-01-06", "2023-01-07"},
		holiday{"2023-01-11", "2023-01-12"},
		holiday{"2023-01-15", "2023-01-15"},
		holiday{"2023-01-21", "2023-01-22"},
		holiday{"2023-01-24", "2023-01-25"},
	} {
		err = trips.AddTrip(h.start, h.end)
		if err != nil {
			t.Fatalf("error making holiday %d %v %v", i, h, err)
		}
	}

	err = trips.Calculate()
	if err != nil {
		t.Fatalf("calculation error %v", err)
	}

	fmt.Println(trips)

	breach, windows := trips.LongestTrips(resultsNo)
	fmt.Printf("breach : %t\n", breach)
	for i, w := range windows {
		fmt.Printf("%d : %+v\n", i, w)
	}

	if trips.longestStay != 4 {
		t.Errorf("Expected longest stay to be 4, got %d", trips.longestStay)
	}

	if breach != true {
		t.Error("Expected breach to be true, got false")
	}

	// jsoncheck
	jsonResult, err := trips.AsJSON()
	if err != nil {
		t.Errorf("json reading error %v\n", err)
	}

	// unmarshal back to struct
	check, err := UnmarshalTripsJSON(jsonResult)
	if err != nil {
		t.Fatalf("unmarshal error %v", err)
	}

	// t.Logf("tripsjson %+v\n", check)

	if check.Breach != true {
		t.Errorf("breach should be true , got %v", check.Breach)
	}
	if check.DaysAway != trips.longestStay {
		t.Errorf("window days away should be %v got %v", trips.longestStay, check.DaysAway)
	}

	if len(check.PartialTrips) != 2 {
		t.Errorf("partial trips should be 2, got %d", len(check.PartialTrips))
	}

	if len(check.Holidays) != 6 {
		t.Errorf("holiday trips should be 6, got %d", len(check.Holidays))
	}

}

func TestTripsLonger(t *testing.T) {

	trips, err := NewTrips(40, 35)
	if err != nil {
		t.Fatalf("could not make trips %v", err)
	}

	err = trips.AddTrip("2023-02-01", "2023-02-28")
	if err != nil {
		t.Fatalf("1. unexpected error making holiday %v", err)
	}

	err = trips.AddTrip("2023-04-02", "2023-05-10")
	if err != nil {
		t.Fatalf("2. unexpected error making holiday %v", err)
	}

	err = trips.AddTrip("2022-04-02", "2022-04-01")
	if err == nil {
		t.Fatal("should error with overlapping dates")
	}

	err = trips.Calculate()
	if err != nil {
		t.Fatalf("calculation error %v", err)
	}

	fmt.Println(trips)

	breach, windows := trips.LongestTrips(5)
	fmt.Printf("breach : %t\n", breach)
	for i, w := range windows {
		fmt.Printf("%d : %+v\n", i, w)
	}

	// start 2023-04-02 end 2023-05-10 is 39 days inclusive
	if trips.longestStay != 39 {
		t.Errorf("Expected longest stay to be 39, got %d", trips.longestStay)
	}

	if breach != true {
		t.Error("Expected breach to be true, got false")
	}
}

// test performance over a much larger window and larger stay size
func TestTripsLong(t *testing.T) {

	window := 720
	compoundStayMaxLength := 180
	resultsNo := 3

	trips, err := NewTrips(window, compoundStayMaxLength)
	if err != nil {
		t.Fatalf("could not make trips %v", err)
	}

	for i, h := range []holiday{
		holiday{"2020-01-01", "2020-02-01"},
		holiday{"2021-03-01", "2021-05-01"},
		holiday{"2023-07-11", "2023-07-12"},
		holiday{"2023-09-01", "2023-11-01"},
		holiday{"2024-01-01", "2024-03-25"},
		holiday{"2024-05-01", "2024-06-01"},
		holiday{"2028-01-01", "2028-01-01"},
	} {
		err = trips.AddTrip(h.start, h.end)
		if err != nil {
			t.Fatalf("error making holiday %d %v %v", i, h, err)
		}
	}

	err = trips.Calculate()
	if err != nil {
		t.Fatalf("calculation error %v", err)
	}

	fmt.Println(trips)

	breach, windows := trips.LongestTrips(resultsNo)
	fmt.Printf("breach : %t\n", breach)
	for i, w := range windows {
		fmt.Printf("%d : %+v\n", i, w)
	}

	if breach != true {
		t.Error("expected breach to be true")
	}

	if trips.longestStay != 181 {
		t.Errorf("Expected longest stay to be 181, got %d", trips.longestStay)
	}

	if len(trips.trips) != 7 {
		t.Errorf("holiday trips should be 7, got %d", len(trips.trips))
	}

}
