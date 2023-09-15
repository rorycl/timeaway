package web

import (
	"bytes"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func bodyMaker(l int) string {
	s := "abcdefghijklmnopqrstuvwxyz"
	output := ""
	o := 0
	for o < l {
		output = output + string(s[o%len(s)])
		o++
	}
	return output
}

/* TestBodyLimit tests that a too large body will be rejected by the
* BodyLimit middleware. See Boris Nagaev's email on golang-nuts 30 Aug
* 2023; you can't set HTTP status in response after you start writing
* the body, so writes have to buffered in the test */
func TestBodyLimit(t *testing.T) {

	// override the package BodyLimitSize
	BodyLimitSize = 1 << 3

	// testHandler to report on body errors, if any
	testHandler := http.HandlerFunc(func(w http.ResponseWriter, ri *http.Request) {
		buf := bytes.Buffer{}
		_, err := io.Copy(&buf, ri.Body)
		if err != nil {
			e := new(http.MaxBytesError)
			if errors.As(err, &e) {
				w.WriteHeader(http.StatusRequestEntityTooLarge)
			} else {
				t.Fatal(err)
			}
		}
		_, err = buf.WriteTo(w)
		if err != nil {
			panic(fmt.Sprintf("could not write to w %v", err))
		}
	})

	for _, tt := range []struct {
		name   string
		size   int
		status int
	}{
		{
			name:   "oversized",
			size:   1 << 4,
			status: http.StatusRequestEntityTooLarge,
		},
		{
			name:   "undersized",
			size:   1 << 2,
			status: http.StatusOK,
		},
	} {

		t.Run(tt.name, func(t *testing.T) {

			r := httptest.NewRequest(http.MethodPost, "http://www.example.com/", strings.NewReader(bodyMaker(tt.size)))
			w := httptest.NewRecorder()
			bodyLimitMiddleware(testHandler).ServeHTTP(w, r)

			if got, want := w.Result().StatusCode, tt.status; got != want {
				t.Errorf("bad status for test %s: got %v want %v", tt.name, got, want)
			}
		})
	}
}
