// Rory 07 October 2019
package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"text/template"
	"time"

	"github.com/braintree/manners"
	"github.com/gorilla/handlers"
	"github.com/gorilla/mux"
	"github.com/gorilla/schema"
	flags "github.com/jessevdk/go-flags"
	"github.com/rorycl/timeaway/trips"
)

var options struct {
	Port string `short:"p" long:"port" description:"port to run on"`
	Addr string `short:"a" long:"address" description:"network address to run on"`
}

func init() {
	log.SetOutput(os.Stderr)
	flags.Parse(&options)
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

	// endpoint routing; gorilla mux is used because "/" in http.NewServeMux
	// is a catch-all pattern
	r := mux.NewRouter()
	r.HandleFunc("/home", Home)
	r.HandleFunc("/trips", Trips)
	r.HandleFunc("/trips-verbose", TripsVerbose)

	// create a handler wrapped in a recovery handler and logging handler
	hdl := handlers.RecoveryHandler()(
		handlers.LoggingHandler(os.Stdout, r))

	// configure server options
	server := &http.Server{
		Addr:         options.Addr + ":" + options.Port,
		ReadTimeout:  1 * time.Second,
		WriteTimeout: 3 * time.Second,
		Handler:      hdl,
	}
	log.Printf("serving on %s:%s", options.Addr, options.Port)

	// wrap server with manners
	manners.ListenAndServe(options.Addr+":"+options.Port, server.Handler)

	// catch signals
	ch := make(chan os.Signal)
	signal.Notify(ch, os.Interrupt, os.Kill)
	go listenForShutdown(ch)
}

// home is the home page
func Home(w http.ResponseWriter, r *http.Request) {
	t := template.Must(template.New("home.html").ParseFiles("home.html"))
	data := struct {
		Title   string
		Address string
		Port    string
	}{
		"trip calculator",
		options.Addr,
		options.Port,
	}
	err := t.Execute(w, data)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		fmt.Fprintf(w, "template writing problem : %s", err.Error())
	}
}

// trip is a POST endpoint returning json
func Trips(w http.ResponseWriter, r *http.Request) {

	// struct to contain results
	type Result struct {
		Error        string   `json:"error"`
		Breach       bool     `json:"breach"`
		StartDate    string   `json:"startdate"`
		EndDate      string   `json:"enddate"`
		DaysAway     int      `json:"daysaway"`
		PartialTrips []string `json:"partialtrips"`
	}
	result := Result{}

	log.Print(r)
	w.Header().Set("Content-Type", "application/json")

	errSender := func(note string, err error) {
		w.WriteHeader(http.StatusBadRequest)
		result.Error = note + " " + err.Error()
		j, _ := json.Marshal(result)
		w.Write(j)
	}

	var decoder = schema.NewDecoder()

	err := r.ParseForm()
	if err != nil {
		errSender("parse form error", err)
		return
	}

	// extract holidays from front end
	var holidays Holidays
	err = decoder.Decode(&holidays, r.PostForm)
	log.Print("postform ", r.PostForm)
	if err != nil {
		errSender("form decode error", err)
		return
	}

	// trips (from module)
	window := 180
	compoundStayMaxLength := 90
	resultsNo := 1

	trs, err := trips.NewTrips(window, compoundStayMaxLength)
	if err != nil {
		errSender("Could not register trip:", err)
		return
	}

	for _, h := range holidays.Holidays {
		err = trs.AddTrip(h.Start, h.End)
		if err != nil {
			errSender("Error adding trip:", err)
			return
		}
	}

	err = trs.Calculate()
	if err != nil {
		errSender("Calculation error: ", err)
		return
	}

	breach, windows := trs.LongestTrips(resultsNo)
	result.Breach = breach
	if len(windows) > 0 {
		window := windows[0]
		result.StartDate = trips.DayFmt(window.Start)
		result.EndDate = trips.DayFmt(window.End)
		result.DaysAway = window.DaysAway
		for _, pt := range window.TripParts {
			result.PartialTrips = append(result.PartialTrips,
				fmt.Sprintf("%s (%d days)", pt, pt.Days()),
			)
		}
	}

	w.WriteHeader(http.StatusOK)
	result.Error = ""
	j, _ := json.Marshal(result)
	w.Write(j)

}

// tripVerbose is a verbose version of trip for non-json endpoints
func TripsVerbose(w http.ResponseWriter, r *http.Request) {

	var decoder = schema.NewDecoder()

	log.Print(r)
	fmt.Fprint(w, "Test http server\n")

	err := r.ParseForm()
	if err != nil {
		http.Error(w, "form error "+err.Error(), 500)
		return
	}

	// extract holidays from front end
	var holidays Holidays
	err = decoder.Decode(&holidays, r.PostForm)
	log.Print(r.PostForm)
	log.Print(holidays, err)
	fmt.Fprintf(w, "holidays: %v\nerror: %v", holidays, err)

	// trips (from module)
	window := 180
	compoundStayMaxLength := 90
	resultsNo := 1

	trips, err := trips.NewTrips(window, compoundStayMaxLength)
	if err != nil {
		log.Printf("could not make new trips %v", err)
		http.Error(w, "new trips error "+err.Error(), 500)
		return
	}

	for i, h := range holidays.Holidays {
		err = trips.AddTrip(h.Start, h.End)
		if err != nil {
			log.Printf("error making holiday %d %v %v", i, h, err)
			http.Error(w, "holiday add error"+err.Error(), 500)
			return
		}
	}

	err = trips.Calculate()
	if err != nil {
		if err != nil {
			log.Printf("calculation error %v", err)
			http.Error(w, "calculation error"+err.Error(), 500)
			return
		}
	}

	breach, windows := trips.LongestTrips(resultsNo)

	/*
		tpl := "breach : %t\nwindow : %s\nlength : %d\n"
		fmt.Fprintf(res, fmt.Sprintf(tpl, breach, windows[0], trips.longestStay))
	*/
	tpl := "breach : %t\nwindow : %s"
	fmt.Fprintf(w, fmt.Sprintf(tpl, breach, windows[0]))

}

// catch shutdown
func listenForShutdown(ch <-chan os.Signal) {
	<-ch
	log.Print("Shutting down the server")
	manners.Close()
}
