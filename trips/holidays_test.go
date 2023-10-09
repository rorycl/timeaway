package trips

import (
	"bytes"
	"encoding/json"
	"net/url"
	"strconv"
	"testing"
	"time"
)

func TestBasics(t *testing.T) {

	// test newHoliday
	_, err := newHoliday(time.Now(), time.Now().Add(24*time.Hour))
	if err != nil {
		t.Errorf("unexpected error %v", err)
	}
	_, err = newHoliday(time.Now(), time.Now().Add(-24*time.Hour))
	if err == nil {
		t.Error("error expected")
	}

	// test newHolidayFromStr
	_, err = newHolidayFromStr("2023-01-01", "2023-01-02")
	if err != nil {
		t.Errorf("unexpected error %v", err)
	}
	_, err = newHolidayFromStr("2023-02-01", "2023-01-02")
	if err == nil {
		t.Error("error expected")
	}

	// test jsonHoliday encoding and decoding
	var jh jsonHoliday
	j := []byte(`{"Start":"2023-01-31","End":"2023-02-01"}`)
	err = json.Unmarshal(j, &jh)
	if err != nil {
		t.Fatal(err)
	}
	j2, err := json.Marshal(jh)
	if err != nil {
		t.Fatal(err)
	}
	if !bytes.Equal(j, j2) {
		t.Errorf("j %s != j2 %s", string(j), string(j2))
	}

}

func TestMakeHolidays(t *testing.T) {

	testCases := []struct {
		start, end string
		isErr      bool
	}{
		{"2023-02-01", "2023-02-02", false},
		{"2023-01-31", "2023-01-30", true},
		{"2023-01-15", "", true},
	}

	for i, tc := range testCases {
		t.Run("test-"+strconv.Itoa(i), func(t *testing.T) {
			_, err := newHolidayFromStr(tc.start, tc.end)
			if err != nil && tc.isErr == false {
				t.Errorf("unexpected error %v", err)
			}
			if err == nil && tc.isErr == true {
				t.Errorf("expected error %+v", tc)
			}
		})
	}
}

func TestHolidaysOverlaps(t *testing.T) {

	tp := func(s, e string) *Holiday {
		h, err := newHolidayFromStr(s, e)
		if err != nil {
			t.Fatalf("could not parse time %s", s)
		}
		return h
	}

	// comparitor
	h := tp("2023-01-01", "2023-02-01")

	testCases := []struct {
		holiday *Holiday
		overlap bool
		days    int
	}{
		{tp("2023-02-01", "2023-02-02"), true, 1},
		{tp("2023-01-31", "2023-01-31"), true, 1},
		{tp("2023-01-15", "2023-02-10"), true, 18},
		{tp("2023-02-02", "2023-02-02"), false, 0},
		{tp("2022-12-30", "2022-12-31"), false, 0},
	}

	for i, tc := range testCases {
		t.Run("test-"+strconv.Itoa(i), func(t *testing.T) {
			ph := h.overlaps(tc.holiday.Start, tc.holiday.End)
			if ph == nil && tc.overlap {
				t.Errorf("expected no overlap")
			}
			if ph != nil && !tc.overlap {
				t.Errorf("expected overlap")
			}
			if ph != nil {
				if got, want := ph.Duration, tc.days; got != want {
					t.Errorf("partial days got %d want %d", got, want)
				}
			}

		})
	}
}

func TestHolidaysFromURL(t *testing.T) {

	testCases := []struct {
		input   string
		holsLen int
		isErr   bool
	}{
		{
			input:   `http://test.com/?Start=2022-12-18&End=2023-01-07&Start=2023-02-10&End=2023-02-17&Start=2023-03-26&End=2023-04-14&Start=2023-07-01&End=2023-07-25&Start=2023-08-01&End=2023-09-01`,
			holsLen: 5,
			isErr:   false,
		},
		{
			input:   `http://test.com/?Start=2022-12-18&End=2023-01-07`,
			holsLen: 1,
			isErr:   false,
		},
		{
			input:   `http://test.com/?Start=2022-12-18`,
			holsLen: 0,
			isErr:   true,
		},
		{
			input:   `http://test.com/`, // no holidays
			holsLen: 0,
			isErr:   false,
		},
	}

	for i, tc := range testCases {
		t.Run("test-"+strconv.Itoa(i), func(t *testing.T) {
			u, err := url.ParseRequestURI(tc.input)
			if err != nil {
				t.Fatal(err)
			}

			holidays, err := HolidaysURLDecoder(u.Query())
			if err != nil && !tc.isErr {
				t.Errorf("unexpected error %v", err)
			}
			if err == nil && tc.isErr {
				t.Error("expected error")
			}
			t.Log(holidays, err)
			if got, want := len(holidays), tc.holsLen; got != want {
				t.Errorf("holidays length got %d != want %d", got, want)
			}
		})
	}
}

func TestHolidaysFromJSON(t *testing.T) {

	testCases := []struct {
		input   []byte // json
		holsLen int
		err     bool
	}{
		{
			input:   []byte(`[{"Start":"2022-12-01","End":"2022-12-02"},{"Start":"2023-01-02","End":"2023-03-30"},{"Start":"2023-04-01","End":"2023-04-02"}]`),
			holsLen: 3,
			err:     false,
		},
		{
			input:   []byte(`[{"Start":"2022-12-01","End":"2022-12-02"},{"Start":"2023-01-02","End":"2023-03-30"}]`),
			holsLen: 2,
			err:     false,
		},
		{
			input:   []byte(`[{"Start":"2022-12-01","End":"2022-11-02"}]`),
			holsLen: 0,
			err:     true,
		},
		{
			input:   []byte(`[{}]`),
			holsLen: 0,
			err:     true,
		},
	}

	for i, tc := range testCases {
		t.Run("test-"+strconv.Itoa(i), func(t *testing.T) {
			holidays, err := HolidaysJSONDecoder(tc.input)
			if err != nil && !tc.err {
				t.Errorf("unexpected error %v", err)
			}
			if err == nil && tc.err {
				t.Error("expected error")
			}
			t.Log(holidays, err)
			if !tc.err {
				if got, want := len(holidays), tc.holsLen; got != want {
					t.Errorf("holidays length got %d != want %d", got, want)
				}
			}
		})
	}
}
