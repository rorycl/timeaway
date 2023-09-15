package web

// https://ieftimov.com/posts/testing-in-go-testing-http-servers/
// https://bignerdranch.com/blog/using-the-httptest-package-in-golang/

import (
	"io"
	"log"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestTripsEndpoint(t *testing.T) {

	// http.HandleFunc("/trips", Trips)
	// log.Fatal(http.ListenAndServe(":8989", nil))

	tt := []struct {
		name       string
		method     string
		input      string   // json
		want       []string // json
		statusCode int
	}{
		{
			name:       "succeed with breach",
			method:     http.MethodPost,
			input:      `[{"Start":"2022-12-01","End":"2022-12-02"},{"Start":"2023-01-02","End":"2023-03-30"},{"Start":"2023-04-01","End":"2023-04-02"}]`,
			want:       []string{`"error":"","breach":true`, `"holidays":[{"Start":"2022-12-01T00:00:00Z","End":"2022-12-02T00:00:00Z","Duration":2},{"Start":"2023-01-02T00:00:00Z","End":"2023-03-30T00:00:00Z","Duration":88},{"Start":"2023-04-01T00:00:00Z","End":"2023-04-02T00:00:00Z","Duration":2}]`},
			statusCode: http.StatusOK,
		},
		{
			name:       "succeed without breach",
			method:     http.MethodPost,
			input:      `[{"Start":"2022-12-01","End":"2022-12-02"},{"Start":"2023-01-02","End":"2023-03-28"},{"Start":"2023-04-01","End":"2023-04-02"}]`,
			want:       []string{`"error":"","breach":false,`},
			statusCode: http.StatusOK,
		},
		{
			name:       "fail due to overlap",
			method:     http.MethodPost,
			input:      `[{"Start":"2022-12-01","End":"2022-12-02"},{"Start":"2023-01-02","End":"2023-03-30"},{"Start":"2023-03-29","End":"2023-04-02"}]`,
			want:       []string{`"Error":"could not add trip: trip 2023-03-29 to 2023-04-02 overlaps with 2023-01-02 to 2023-03-30"`},
			statusCode: http.StatusBadRequest,
		},
		{
			name:       "fail due to end date before start date",
			method:     http.MethodPost,
			input:      `[{"Start":"2022-12-01","End":"2022-11-01"}]`,
			want:       []string{`"Error":"could not add trip: start date 2022-12-01 after 2022-11-01"`},
			statusCode: http.StatusBadRequest,
		},
		{
			name:       "fail due to GET",
			method:     http.MethodGet,
			input:      `[{"Start":"2022-12-01","End":"2022-12-02"}]`,
			want:       []string{`"Error":"endpoint only accepts POST requests, got GET"`},
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
			data, err := io.ReadAll(res.Body)
			if err != nil {
				log.Fatal(err)
			}

			if tc.statusCode != res.StatusCode {
				t.Errorf("expected status %d, got %d", tc.statusCode, res.StatusCode)
			}

			responseBody := string(data)
			for _, w := range tc.want {
				if !strings.Contains(responseBody, w) {
					t.Errorf("body %s did not contain %s", responseBody, w)
				}
			}
		})
	}
}
