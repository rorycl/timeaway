// Rory 07 October 2019
package main

import (
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"

	//"strings"

	"github.com/braintree/manners"
	"github.com/gorilla/schema"
	flags "github.com/jessevdk/go-flags"
)

var opts struct {
	Port       string `short:"p" long:"port" description:"port to run on"`
	Addr       string `short:"n" long:"address" description:"network address to run on"`
	Identifier string `short:"i" long:"identifier" description:"identification string"`
}

func init() {
	log.SetOutput(os.Stderr)
	flags.Parse(&opts)
}

// trip describes the start and end dates of a trip
type Trip struct {
	Start string `schema:Start`
	End   string `schema:End`
}

// trips is a slice of trip
type Trips struct {
	Trips []Trip
}

func main() {
	handler := newHander()

	ch := make(chan os.Signal)
	signal.Notify(ch, os.Interrupt, os.Kill)
	go listenForShutdown(ch)

	// s := []string{opts.Addr, opts.Port}
	// server := strings.Join(s, ':')
	log.Printf("serving on %s:%s identifier %s", opts.Addr, opts.Port, opts.Identifier)
	manners.ListenAndServe(opts.Addr+":"+opts.Port, handler)
	//manners.ListenAndServe(server, handler)

}

func newHander() *handler {
	return &handler{}
}

type handler struct{}

func (h *handler) ServeHTTP(res http.ResponseWriter, req *http.Request) {

	var decoder = schema.NewDecoder()

	log.Print(req)
	fmt.Fprint(res, "Test http server : "+opts.Identifier+"\n")

	err := req.ParseForm()
	if err != nil {
		http.Error(res, "form error "+err.Error(), 500)
		return
	}
	var trips Trips
	err = decoder.Decode(&trips, req.PostForm)
	log.Print(req.PostForm)
	log.Print(trips, err)
	fmt.Fprintf(res, "trips : %v\nerror: %v", trips, err)
}

func listenForShutdown(ch <-chan os.Signal) {
	<-ch
	log.Print("Closing the Rtest http server")
	manners.Close()
}
