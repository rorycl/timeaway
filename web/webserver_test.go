package web

// https://ieftimov.com/posts/testing-in-go-testing-http-servers/
// https://bignerdranch.com/blog/using-the-httptest-package-in-golang/

import (
	"errors"
	"io"
	"log"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/rorycl/timeaway/trips"
)

// Test Home page returns a 200
func TestHome(t *testing.T) {

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

func TestTripsEndpoint(t *testing.T) {

	// swap out the webserver development/testing package level func vars

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

			/*
				responseBody := string(data)
				fmt.Println(responseBody)
					for _, w := range tc.want {
						if !strings.Contains(responseBody, w) {
							t.Errorf("body %s did not contain %s", responseBody, w)
						}
					}
			*/
		})
	}
}
