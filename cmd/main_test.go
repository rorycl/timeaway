package main

import (
	"fmt"
	"os"
	"testing"
)

func TestGetOptions(t *testing.T) {

	tests := []struct {
		args []string
		ok   int
	}{
		{
			args: []string{"prog", "-n", "8000"},
			ok:   1,
		},
		{
			args: []string{"prog"},
			ok:   0,
		},
		{
			args: []string{"prog", "-a", "127.0.0.1", "-p", "8000", "-b", "/baseurl"},
			ok:   0,
		},
	}

	var exitCode int
	// override package exit
	exit = func(i int) {
		exitCode = i
	}

	for i, tt := range tests {
		exitCode = 0
		t.Run(fmt.Sprintf("test_%d", i), func(t *testing.T) {
			os.Args = tt.args
			_, _, _ = getOptions()
			if got, want := exitCode, tt.ok; got != want {
				t.Errorf("got %d want %d", got, want)
			}
		})
	}
}

func TestMain(t *testing.T) {
	os.Args = []string{"prog"}
	// override package serve func
	var ok bool
	serve = func(address, port, baseUrl string) {
		ok = true
	}
	exitCode := 0
	// override package exit
	exit = func(i int) {
		exitCode = i
	}
	main()
	if exitCode != 0 {
		t.Fatal("did not get a 0 exit code")
	}
	if !ok {
		t.Fatal("main failed to return ok")
	}
}
