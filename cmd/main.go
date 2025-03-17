package main

import (
	_ "embed"
	"fmt"
	"log"
	"net"
	"os"
	"strconv"

	flags "github.com/jessevdk/go-flags"
	"github.com/rorycl/timeaway/web"
)

var options struct {
	Port    string `short:"p" long:"port" description:"port to run on" default:"8000"`
	Addr    string `short:"a" long:"address" description:"network address to run on" default:"127.0.0.1"`
	BaseURL string `short:"b" long:"baseurl" description:"web server base URL" default:""`
}

var serve func(string, string, string) = web.Serve
var exit func(int) = os.Exit

func getOptions() (string, string, string) {
	log.SetOutput(os.Stderr)
	_, err := flags.Parse(&options)
	if err != nil {
		fmt.Printf("flag parsing error: %v\n", err)
		exit(1)
	}

	// verify options
	port, err := strconv.Atoi(options.Port)
	if err != nil || port == 0 {
		fmt.Printf("port %s invalid; exiting\n", options.Port)
		exit(1)
	}
	if net.ParseIP(options.Addr) == nil {
		fmt.Printf("address %s invalid; exiting\n", options.Addr)
		exit(1)
	}
	return options.Addr, options.Port, options.BaseURL
}

func main() {
	// run the server
	serve(getOptions())
}
