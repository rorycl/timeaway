package main

import (
	_ "embed"
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"log"
	"net"
	"net/http"
	"os"
	"os/signal"
	"strconv"
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
	Port string `short:"p" long:"port" description:"port to run on" default:"8000"`
	Addr string `short:"a" long:"address" description:"network address to run on" default:"127.0.0.1"`
}

// production is default; set InDevelopment to true with build tag
var InDevelopment bool = false

func main() {

	log.SetOutput(os.Stderr)
	flags.Parse(&options)

	// verify options
	port, err := strconv.Atoi(options.Port)
	if err != nil || port == 0 {
		fmt.Printf("port %s invalid; exiting\n", options.Port)
		os.Exit(1)
	}
	if net.ParseIP(options.Addr) == nil {
		fmt.Printf("address %s invalid; exiting\n", options.Addr)
		os.Exit(1)
	}

	// endpoint routing; gorilla mux is used because "/" in http.NewServeMux
	// is a catch-all pattern
	r := mux.NewRouter()
	r.HandleFunc("/home", Home)
	r.HandleFunc("/favicon.ico", Favicon)
	r.HandleFunc("/trips", Trips)

	// logging converts gorilla's handlers.CombinedLoggingHandler to a
	// func(http.Handler) http.Handler to satisfy type MiddlewareFunc
	logging := func(handler http.Handler) http.Handler {
		return handlers.CombinedLoggingHandler(os.Stdout, handler)
	}

	// recovery converts gorilla's handlers.RecoveryHandler to a
	// func(http.Handler) http.Handler to satisfy type MiddlewareFunc
	recovery := func(handler http.Handler) http.Handler {
		return handlers.RecoveryHandler()(handler)
	}

	// attach middleware
	r.Use(bodyLimitMiddleware)
	r.Use(logging)
	r.Use(recovery)

	// configure server options
	server := &http.Server{
		Addr:           options.Addr + ":" + options.Port,
		ReadTimeout:    1 * time.Second,
		WriteTimeout:   3 * time.Second,
		MaxHeaderBytes: 1 << 17, // ~125k
	}
	log.Printf("serving on %s:%s", options.Addr, options.Port)

	// wrap server with manners
	err = manners.ListenAndServe(options.Addr+":"+options.Port, server.Handler)
	if err != nil {
		log.Printf("fatal server error: %v", err)
	}

	// catch signals
	ch := make(chan os.Signal, 1)
	signal.Notify(ch, os.Interrupt, os.Kill)
	go listenForShutdown(ch)
}

//go:embed tpl/home.html
var homeTemplate string

// holiday structure
type holiday struct {
	Start string `json:"Start"`
	End   string `json:"End"`
}

// holidayByURL
type holidaysByURL struct {
	Start []string `schema:"Start"`
	End   []string `schema:"End"`
}

// holidays are a slice of holiday
var holidays []holiday

// Home is the home page
func Home(w http.ResponseWriter, r *http.Request) {

	t := template.Must(template.New("home.html").Parse(homeTemplate))
	if InDevelopment {
		t = template.Must(template.New("home.html").ParseFiles("tpl/home.html"))
	}

	// grab dates from url, if any
	inputDates := func() []holiday {
		hols := []holiday{}
		var decoder = schema.NewDecoder()
		var hbu holidaysByURL
		_ = decoder.Decode(&hbu, r.URL.Query()) // ignore errors
		for i, s := range hbu.Start {
			_, err := time.Parse("2006-01-02", s)
			if err != nil {
				continue
			}
			if i > len(hbu.End)-1 {
				continue
			}
			_, err = time.Parse("2006-01-02", hbu.End[i])
			if err != nil {
				continue
			}
			hols = append(hols, holiday{s, hbu.End[i]})
		}
		return hols
	}()

	data := struct {
		Title      string
		Address    string
		Port       string
		PostURL    string
		InputDates []holiday
	}{
		"trip calculator",
		options.Addr,
		options.Port,
		"/trips",
		inputDates,
	}
	err := t.Execute(w, data)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		fmt.Fprintf(w, "template writing problem : %s", err.Error())
	}
}

//go:embed tpl/favicon.svg
var favicon string

// Favicon serves an svg favicon
func Favicon(w http.ResponseWriter, r *http.Request) {
	// http.ServeFile(w, r, "favicon.svg")
	fmt.Fprint(w, favicon)
}

// trip is a POST endpoint returning json
func Trips(w http.ResponseWriter, r *http.Request) {

	w.Header().Set("Content-Type", "application/json")

	errSender := func(note string, err error) {
		w.WriteHeader(http.StatusBadRequest)
		j, _ := json.Marshal(struct {
			Error string
		}{
			Error: note + " " + err.Error(),
		})
		w.Write(j)
	}

	if r.Method != "POST" {
		err := errors.New(r.Method)
		errSender("endpoint only accepts POST requests, got", err)
		return
	}

	// read body
	body, err := ioutil.ReadAll(r.Body)
	defer r.Body.Close()
	if err != nil {
		errSender("body reading error", err)
		return
	}
	// log.Printf("body%+v\n", string(body))

	err = json.Unmarshal(body, &holidays)
	if err != nil {
		errSender("form json decoding error", err)
		return
	}
	if len(holidays) < 1 {
		errSender("no holidays were found", nil)
		return
	}

	// trips (from module)
	window := 180
	compoundStayMaxLength := 90

	trs, err := trips.NewTrips(window, compoundStayMaxLength)
	if err != nil {
		errSender("Could not register trip:", err)
		return
	}

	for _, h := range holidays {
		err = trs.AddTrip(h.Start, h.End)
		if err != nil {
			errSender("could not add trip:", err)
			return
		}
	}

	err = trs.Calculate()
	if err != nil {
		errSender("calculation error: ", err)
		return
	}

	json, err := trs.AsJSON()
	if err != nil {
		errSender("json encoding error: ", err)
		return
	}

	w.WriteHeader(http.StatusOK)
	w.Write(json)

}

// catch shutdown
func listenForShutdown(ch <-chan os.Signal) {
	<-ch
	log.Print("Shutting down the server")
	manners.Close()
}
