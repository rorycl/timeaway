package web

// https://ieftimov.com/posts/testing-in-go-testing-http-servers/
// https://bignerdranch.com/blog/using-the-httptest-package-in-golang/

import (
	"errors"
	"io"
	"log"
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/rorycl/timeaway/trips"
)

// Test Home page returns a 200
func TestHome(t *testing.T) {

	// home uses templates fs
	DirFS = &fileSystem{}
	DirFS.TplFS = os.DirFS("templates")

	r := httptest.NewRequest(http.MethodGet, "http://example.com/home", nil)
	w := httptest.NewRecorder()

	Home(w, r)

	res := w.Result()
	defer res.Body.Close()
	_, err := io.ReadAll(res.Body)
	if err != nil {
		log.Fatal(err)
	}

	if want, got := 200, res.StatusCode; want != got {
		t.Errorf("expected status %d, got %d", want, got)
	}
}

// Test Health page returns a 200
func TestHealth(t *testing.T) {

	r := httptest.NewRequest(http.MethodGet, "http://example.com/health", nil)
	w := httptest.NewRecorder()

	Health(w, r)

	res := w.Result()
	defer res.Body.Close()
	data, err := io.ReadAll(res.Body)
	if err != nil {
		log.Fatal(err)
	}

	if want, got := 200, res.StatusCode; want != got {
		t.Errorf("expected status %d, got %d", want, got)
	}
	responseBody := string(data)
	if want, got := strings.TrimSpace(`{"status":"up"}`), strings.TrimSpace(responseBody); want != got {
		t.Errorf("expected status %s, got %s", want, got)
	}
}

// Favicon page returns a 200
func TestFavicon(t *testing.T) {

	// favicon uses the static fs
	DirFS = &fileSystem{}
	DirFS.StaticFS = os.DirFS("static")

	h := http.StripPrefix("/static/", http.FileServer(http.FS(DirFS.StaticFS)))
	server := httptest.NewServer(h)
	defer server.Close()

	result, err := http.Get(server.URL + "/static/favicon.svg")
	if err != nil {
		t.Fatal(err)
	}
	if want, got := 200, result.StatusCode; want != got {
		t.Errorf("expected status %d, got %d", want, got)
	}
}

// TestTripsEndpoint tests the JSON endpoint; note that the main
// webserver package level func vars are swapped out.
func TestTripsEndpoint(t *testing.T) {

	// holidayJSONDecoder makes holidays from a POSTED json body
	holidayJSONDecoder = func(b []byte) ([]trips.Holiday, error) {
		trs := []trips.Holiday{}
		if len(b) < 1 {
			return trs, errors.New("no content received")
		}
		tp := func(s string) time.Time {
			ti, err := time.Parse("2006-01-02", s)
			if err != nil {
				t.Fatalf("could not parse %s in holidayJSONDecoder: %v", s, err)
			}
			return ti
		}
		holiday := trips.Holiday{Start: tp("2023-01-01"), End: tp("2023-01-02"), Duration: 2}
		trs = append(trs, holiday)
		return trs, nil
	}

	// calculate is the main method for calculations
	calculate = func([]trips.Holiday) (*trips.Trips, error) {
		hs := trips.Trips{}
		return &hs, nil
	}

	// holidayJSONMarshall returns a json representation of a
	// trips.Trips
	tripsJSONMarshal = func(v any) ([]byte, error) {
		return []byte(`{"result":"ok"}`), nil
	}

	tt := []struct {
		name       string
		method     string
		input      string // json
		statusCode int
	}{
		{
			name:       "succeed post",
			method:     http.MethodPost,
			input:      `[{"Start":"2022-12-01","End":"2022-12-02"}]`,
			statusCode: http.StatusOK,
		},
		{
			name:       "fail no POST body",
			method:     http.MethodPost,
			input:      ``,
			statusCode: http.StatusBadRequest,
		},
		{
			name:       "fail due to GET",
			method:     http.MethodGet,
			input:      ``,
			statusCode: http.StatusBadRequest,
		},
	}

	for _, tc := range tt {
		t.Logf("%+v\n", tc)
		t.Run(tc.name, func(t *testing.T) {

			r := httptest.NewRequest(tc.method, "http://example.com/trips", strings.NewReader(tc.input))
			w := httptest.NewRecorder()

			Trips(w, r)

			res := w.Result()
			defer res.Body.Close()
			_, err := io.ReadAll(res.Body)
			if err != nil {
				t.Fatal(err)
			}

			if tc.statusCode != res.StatusCode {
				t.Errorf("expected status %d, got %d", tc.statusCode, res.StatusCode)
			}

		})
	}
}

// TestPartialEndpoints tests the partials used for htmx partial
// rendering
func TestPartialEndpoints(t *testing.T) {

	// the partials use the templates endpoint (only used for
	// PartialAddTrip)
	DirFS = &fileSystem{}
	DirFS.TplFS = os.DirFS("templates")

	testCases := []struct {
		name       string
		method     string
		fn         func(w http.ResponseWriter, r *http.Request)
		endpoint   string
		statusCode int
	}{
		{"PartialDetailsShow", http.MethodGet, PartialDetailsShow, "/partials/details/show", 200},
		{"PartialDetailsHide", http.MethodGet, PartialDetailsHide, "/partials/details/hide", 200},
		{"PartialNoContent", http.MethodGet, PartialNoContent, "/partials/nocontent", 200},
		{"PartialAddTrip", http.MethodGet, PartialAddTrip, "/partials/addtrip", 200},
		{"PartialReport", http.MethodGet, PartialReport, "/partials/report", http.StatusBadRequest},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			r := httptest.NewRequest(tc.method, "http://example.com/"+tc.endpoint, nil)
			w := httptest.NewRecorder()
			tc.fn(w, r)
			res := w.Result()
			if got, want := tc.statusCode, res.StatusCode; got != want {
				t.Errorf("got %v != want %v", got, want)
			}
		})
	}
}
