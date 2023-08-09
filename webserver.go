// Rory 07 October 2019
package main

import (
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"

	"github.com/braintree/manners"
	"github.com/gorilla/schema"
	flags "github.com/jessevdk/go-flags"
	"github.com/rorycl/timeaway/trips"
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

// holiday describes the start and end dates of a trip
type Holiday struct {
	Start string `schema:Start`
	End   string `schema:End`
}

// holidays is a slice of holiday
type Holidays struct {
	Holidays []Holiday
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

	// extract holidays from front end
	var holidays Holidays
	err = decoder.Decode(&holidays, req.PostForm)
	log.Print(req.PostForm)
	log.Print(holidays, err)
	fmt.Fprintf(res, "holidays: %v\nerror: %v", holidays, err)

	// trips (from module)
	window := 180
	compoundStayMaxLength := 90
	resultsNo := 1

	trips, err := trips.NewTrips(window, compoundStayMaxLength)
	if err != nil {
		log.Printf("could not make new trips %v", err)
		http.Error(res, "new trips error "+err.Error(), 500)
		return
	}

	for i, h := range holidays.Holidays {
		err = trips.AddTrip(h.Start, h.End)
		if err != nil {
			log.Printf("error making holiday %d %v %v", i, h, err)
			http.Error(res, "holiday add error"+err.Error(), 500)
			return
		}
	}

	err = trips.Calculate()
	if err != nil {
		if err != nil {
			log.Printf("calculation error %v", err)
			http.Error(res, "calculation error"+err.Error(), 500)
			return
		}
	}

	breach, windows := trips.LongestTrips(resultsNo)

	/*
		tpl := "breach : %t\nwindow : %s\nlength : %d\n"
		fmt.Fprintf(res, fmt.Sprintf(tpl, breach, windows[0], trips.longestStay))
	*/
	tpl := "breach : %t\nwindow : %s"
	fmt.Fprintf(res, fmt.Sprintf(tpl, breach, windows[0]))

}

func listenForShutdown(ch <-chan os.Signal) {
	<-ch
	log.Print("Closing the Rtest http server")
	manners.Close()
}
