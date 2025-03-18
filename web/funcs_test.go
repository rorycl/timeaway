package web

import (
	"html/template"
	"strings"
	"testing"
	"time"
)

func TestFuncs(t *testing.T) {

	tpl := `output: {{ yearsAgo . +2 | dateStr }}`

	var err error
	tt := template.New("n")
	tt = tt.Funcs(webFuncMap)
	tt, err = tt.Parse(tpl)
	if err != nil {
		t.Fatal(err)
	}

	s := strings.Builder{}
	d := time.Date(2020, 1, 1, 0, 0, 0, 0, time.Local)
	err = tt.Execute(&s, d)
	if err != nil {
		t.Fatal(err)
	}

	if got, want := s.String(), "output: 2022-01-01"; got != want {
		t.Errorf("got %s != want %s", got, want)
	}

}
