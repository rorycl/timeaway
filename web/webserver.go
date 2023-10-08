package web

import (
	_ "embed"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"text/template"
	"time"

	"github.com/gorilla/handlers"
	"github.com/gorilla/mux"
	"github.com/rorycl/timeaway/trips"
)

var (
	// WebMaxHeaderBytes is the largest number of header bytes accepted by
	// the webserver
	WebMaxHeaderBytes int = 1 << 17 // ~125k

	// ServerAddress is the default Server network address
	ServerAddress string = "127.0.0.1"

	// ServerPort is the default Server network port
	ServerPort string = "8000"

	// BaseURL is the base url for redirects, etc.
	BaseURL string = ""
)

// development/testing vars
var (

	// holidayJSONDecoder sets the holiday POST decoder
	holidayJSONDecoder func([]byte) ([]trips.Holiday, error) = trips.HolidaysJSONDecoder

	// calculate sets the calculation method in use to allow swapping
	// out for testing
	calculate func([]trips.Holiday) (*trips.Trips, error) = trips.Calculate

	// tripsJSONMarshall sets the holiday marshaller
	tripsJSONMarshal func(v any) ([]byte, error) = json.Marshal

	// production is default; set inDevelopment to true with build tag
	inDevelopment bool = false
)

// Serve runs the web server on the specified address and port
func Serve(addr string, port string) {

	if addr == "" {
		addr = ServerAddress
	} else {
		ServerAddress = addr
	}

	if port == "" {
		port = ServerPort
	} else {
		ServerPort = port
	}

	// endpoint routing; gorilla mux is used because "/" in http.NewServeMux
	// is a catch-all pattern
	r := mux.NewRouter()
	r.HandleFunc("/", Home)
	r.HandleFunc("/home", Home)
	r.HandleFunc("/favicon.ico", Favicon)
	r.HandleFunc("/trips", Trips)
	r.HandleFunc("/health", Health)

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
		Addr:           addr + ":" + port,
		WriteTimeout:   3 * time.Second,
		MaxHeaderBytes: WebMaxHeaderBytes,
		Handler:        r,
	}
	log.Printf("serving on %s:%s", addr, port)

	err := server.ListenAndServe()
	if err != nil {
		log.Printf("fatal server error: %v", err)
	}
}

//go:embed tpl/home.html
var homeTemplate string

// Home is the home page
func Home(w http.ResponseWriter, r *http.Request) {

	t := template.Must(template.New("home.html").Parse(homeTemplate))
	if inDevelopment {
		t = template.Must(template.New("home.html").ParseFiles("web/tpl/home.html"))
	}

	// retrieve holidays, if any, ignoring errors
	holidays, err := trips.HolidaysURLDecoder(r.URL.Query())
	if inDevelopment {
		log.Printf("holidays GET : %+v err : %v", holidays, err)
	}

	data := struct {
		Title      string
		Address    string
		Port       string
		PostURL    string
		InputDates []trips.Holiday
	}{
		"trip calculator",
		ServerAddress,
		ServerPort,
		BaseURL + "/trips",
		holidays,
	}
	err = t.Execute(w, data)
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
		_, err = w.Write(j)
		if err != nil {
			log.Printf("could not write trips json error %v", err)
		}
	}

	if r.Method != "POST" {
		err := errors.New(r.Method)
		errSender("endpoint only accepts POST requests, got", err)
		return
	}

	// read body
	body, err := io.ReadAll(r.Body)
	defer r.Body.Close()
	if err != nil {
		errSender("body reading error", err)
		return
	}
	if inDevelopment {
		log.Println("body content:", string(body))
	}

	// extract holidays from POSTed json
	holidays, err := holidayJSONDecoder(body)
	if err != nil {
		errSender("form json decoding error", err)
		return
	}
	if len(holidays) < 1 {
		errSender("no holidays were found", nil)
		return
	}
	if inDevelopment {
		log.Println("holidays", holidays)
	}

	// perform the calculation
	trs, err := calculate(holidays)
	if err != nil {
		errSender("calculation error: ", err)
		return
	}

	// convert to json
	jBytes, err := tripsJSONMarshal(trs)
	if err != nil {
		errSender("json encoding error: ", err)
		return
	}

	if inDevelopment {
		log.Println("decoded json:", string(jBytes))
	}

	w.WriteHeader(http.StatusOK)
	_, err = w.Write(jBytes)
	if err != nil {
		log.Printf("could not write trips error %v", err)
	}

}

// HealthCheck shows if the service is up
func Health(w http.ResponseWriter, r *http.Request) {
	enc := json.NewEncoder(w)
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	resp := map[string]string{"status": "up"}
	if err := enc.Encode(resp); err != nil {
		log.Fatal("unable to encode response")
	}
}
