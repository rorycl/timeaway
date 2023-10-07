package trips

import (
	"encoding/json"
	"testing"
)

// overlap detection checks
func TestTripAdditions(t *testing.T) {

	WindowMaxDays = 5
	CompoundStayMaxDays = 3

	trips, err := newTrips()
	if err != nil {
		t.Fatalf("could not make trips %v", err)
	}

	tp := func(s, e string) Holiday {
		h, err := newHolidayFromStr(s, e)
		if err != nil {
			t.Fatal(err)
		}
		return *h
	}

	type holidayTest struct {
		hol Holiday
		err bool
		msg string
	}
	for i, h := range []holidayTest{
		{tp("2023-01-01", "2023-01-01"), false, "simple add"},
		{tp("2023-01-01", "2023-01-07"), true, "overlap already registered date"},
		{tp("2023-01-02", "2023-01-04"), false, "simple add again"},
		{tp("2023-01-03", "2023-01-05"), true, "overlap two days"},
		{tp("2023-01-04", "2023-01-05"), true, "overlap one day"},
		{tp("2022-12-31", "2023-01-01"), true, "overlap first day"},
		{tp("2022-11-01", "2022-11-02"), false, "add before first should be ok"},
	} {
		err = trips.addHoliday(h.hol)
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

	WindowMaxDays = 5
	CompoundStayMaxDays = 4

	trips, err := newTrips()
	if err != nil {
		t.Fatalf("could not make trips %v", err)
	}

	tp := func(s, e string) Holiday {
		h, err := newHolidayFromStr(s, e)
		if err != nil {
			t.Fatal(err)
		}
		return *h
	}

	hols := []Holiday{
		tp("2023-01-01", "2023-01-01"),
		tp("2023-01-06", "2023-01-07"),
		tp("2023-01-11", "2023-01-12"),
		tp("2023-01-15", "2023-01-15"),
		tp("2023-01-21", "2023-01-22"),
		tp("2023-01-24", "2023-01-25"),
	}

	trips, err = Calculate(hols)
	if err != nil {
		t.Fatalf("calculation error %v", err)
	}
	t.Log(trips)

	if trips.DaysAway != 4 {
		t.Errorf("Expected longest stay to be 4, got %d", trips.DaysAway)
	}

	if trips.Breach != false {
		t.Error("Expected breach to be false, got true")
	}

	if got, want := len(trips.Window.HolidayParts), 2; got != want {
		t.Errorf("partial trips should be %d, got %d", got, want)
	}

	jsonResult, err := json.Marshal(trips)
	if err != nil {
		t.Fatalf("json reading error %v\n", err)
	}

	// unmarshal back to struct
	var checkTrips Trips
	err = json.Unmarshal(jsonResult, &checkTrips)
	if err != nil {
		t.Fatalf("unmarshal error %v", err)
	}

	if checkTrips.Breach != false {
		t.Errorf("breach should be false, got %v", checkTrips.Breach)
	}

	if got, want := checkTrips.DaysAway, trips.DaysAway; got != want {
		t.Errorf("window days away should be %v got %v", got, want)
	}

	if len(checkTrips.Window.HolidayParts) != 2 {
		t.Errorf("partial trips should be 2, got %d", len(checkTrips.Window.HolidayParts))
	}

	if len(checkTrips.Holidays) != 6 {
		t.Errorf("holiday trips should be 6, got %d", len(checkTrips.Holidays))
	}

}

// the same test as above, but with a lower CompoundStayMaxDays to
// breach
func TestTripsToBreach(t *testing.T) {

	WindowMaxDays = 5
	CompoundStayMaxDays = 3

	tp := func(s, e string) Holiday {
		h, err := newHolidayFromStr(s, e)
		if err != nil {
			t.Fatal(err)
		}
		return *h
	}

	hols := []Holiday{
		tp("2023-01-01", "2023-01-01"),
		tp("2023-01-06", "2023-01-07"),
		tp("2023-01-11", "2023-01-12"),
		tp("2023-01-15", "2023-01-15"),
		tp("2023-01-21", "2023-01-22"),
		tp("2023-01-24", "2023-01-25"),
	}

	trips, err := Calculate(hols)
	if err != nil {
		t.Fatalf("calculation error %v", err)
	}
	t.Log(trips)

	if trips.DaysAway != 4 {
		t.Errorf("Expected longest stay to be 4, got %d", trips.DaysAway)
	}
	if trips.Breach != true {
		t.Error("Expected breach to be true, got false")
	}

	jsonResult, err := json.Marshal(trips)
	if err != nil {
		t.Fatalf("json reading error %v\n", err)
	}

	// unmarshal back to struct
	var checkTrips Trips
	err = json.Unmarshal(jsonResult, &checkTrips)
	if err != nil {
		t.Fatalf("unmarshal error %v", err)
	}

	if checkTrips.Breach != true {
		t.Errorf("breach should be true , got %v", checkTrips.Breach)
	}

	if checkTrips.DaysAway != trips.DaysAway {
		t.Errorf("window days away should be %v got %v", trips.DaysAway, checkTrips.DaysAway)
	}

	if len(checkTrips.Window.HolidayParts) != 2 {
		t.Errorf("partial trips should be 2, got %d", len(checkTrips.Window.HolidayParts))
	}

	if len(checkTrips.Holidays) != 6 {
		t.Errorf("holiday trips should be 6, got %d", len(checkTrips.Holidays))
	}

}

func TestTripsLonger(t *testing.T) {

	WindowMaxDays = 40
	CompoundStayMaxDays = 35

	trips, err := newTrips()
	if err != nil {
		t.Fatalf("could not make trips %v", err)
	}

	adder := func(s, e string) error {
		h, err := newHolidayFromStr(s, e)
		if err != nil {
			t.Fatal(s, e, err)
		}
		return trips.addHoliday(*h)
	}

	err = adder("2023-02-01", "2023-02-28")
	if err != nil {
		t.Fatalf("1. unexpected error making holiday %v", err)
	}

	err = adder("2023-04-02", "2023-05-10")
	if err != nil {
		t.Fatalf("2. unexpected error making holiday %v", err)
	}

	// call to package internal calculate
	tps, err := trips.calculate()
	if err != nil {
		t.Fatalf("calculation error %v", err)
	}

	// start 2023-04-02 end 2023-05-10 is 39 days inclusive
	if tps.DaysAway != 39 {
		t.Errorf("Expected longest stay to be 39, got %d", trips.DaysAway)
	}

	if tps.Breach != true {
		t.Error("Expected breach to be true, got false")
	}
}

// test performance over a much larger window and larger stay size
func TestTripsLong(t *testing.T) {

	WindowMaxDays = 720
	CompoundStayMaxDays = 180

	trips, err := newTrips()
	if err != nil {
		t.Fatalf("could not make trips %v", err)
	}

	tp := func(s, e string) Holiday {
		h, err := newHolidayFromStr(s, e)
		if err != nil {
			t.Fatal(err)
		}
		return *h
	}

	for i, h := range []Holiday{
		tp("2020-01-01", "2020-02-01"),
		tp("2021-03-01", "2021-05-01"),
		tp("2023-07-11", "2023-07-12"),
		tp("2023-09-01", "2023-11-01"),
		tp("2024-01-01", "2024-03-25"),
		tp("2024-05-01", "2024-06-01"),
		tp("2028-01-01", "2028-01-01"),
	} {
		err = trips.addHoliday(h)
		if err != nil {
			t.Fatalf("error making holiday %d %v %v", i, h, err)
		}
	}

	_, err = trips.calculate()
	if err != nil {
		t.Fatalf("calculation error %v", err)
	}

	if trips.Breach != true {
		t.Error("expected breach to be true")
	}

	if trips.DaysAway != 181 {
		t.Errorf("Expected longest stay to be 181, got %d", trips.DaysAway)
	}

	if len(trips.Holidays) != 7 {
		t.Errorf("holiday trips should be 7, got %d", len(trips.Holidays))
	}

}
