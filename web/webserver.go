package web

import (
	_ "embed"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"io/fs"
	"log"
	"net/http"
	"net/url"
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
)

// development flags and static and template directory locations
var (
	// production is default; set inDevelopment to true with build tag
	inDevelopment bool   = false
	staticDirDev  string = "web/static"
	tplDirDev     string = "web/templates"
	staticDir     string = "static"
	tplDir        string = "templates"
	DirFS         *fileSystem
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

	// setup the filesystem for templates or static files, depending on
	// development (filesystem) or not (embedded)
	var err error
	if inDevelopment {
		DirFS, err = NewFileSystem(inDevelopment, tplDirDev, staticDirDev)
	} else {
		DirFS, err = NewFileSystem(inDevelopment, tplDir, staticDir)
	}
	if err != nil {
		log.Fatal(err)
	}

	// endpoint routing; gorilla mux is used because "/" in http.NewServeMux
	// is a catch-all pattern
	r := mux.NewRouter()

	// attach static dynamic file system to the http.FileServer
	// https://pkg.go.dev/github.com/gorilla/mux#section-readme :Static Files
	r.PathPrefix("/static/").Handler(
		http.StripPrefix("/static/",
			http.FileServer(http.FS(DirFS.StaticFS))),
	)

	// partials
	r.HandleFunc("/partials/details/show", PartialDetailsShow)
	r.HandleFunc("/partials/details/hide", PartialDetailsHide)
	r.HandleFunc("/partials/report", PartialReport)
	r.HandleFunc("/partials/nocontent", PartialNoContent)
	r.HandleFunc("/partials/addtrip", PartialAddTrip)

	// main routes
	r.HandleFunc("/", Home)
	r.HandleFunc("/home", Home)
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
		Addr:    addr + ":" + port,
		Handler: r,
		// timeouts and limits
		MaxHeaderBytes:    WebMaxHeaderBytes,
		ReadTimeout:       1 * time.Second,
		WriteTimeout:      2 * time.Second,
		IdleTimeout:       30 * time.Second,
		ReadHeaderTimeout: 2 * time.Second,
	}
	log.Printf("serving on %s:%s", addr, port)

	err = server.ListenAndServe()
	if err != nil {
		log.Printf("fatal server error: %v", err)
	}
}

// Home is the home page
func Home(w http.ResponseWriter, r *http.Request) {

	t, err := template.ParseFS(DirFS.TplFS, "home.html")
	if err != nil {
		log.Fatal(err)
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

// Trips is a POST endpoint for JSON queries, receiving json dates,
// turning this data into Holidays and then performing a calculation on
// the data, finally returning the json result.
func Trips(w http.ResponseWriter, r *http.Request) {

	w.Header().Set("Content-Type", "application/json")

	// short cut error retuener
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
		log.Print("health error: unable to encode response")
	}
}

// partialWriter writes the partial file at path/fp (via an fs.FS) for
// inclusion to the http.ResponseWriter or errors for partials not
// needing a template
func partialWriter(w http.ResponseWriter, path, fp string) error {
	var m fs.FS
	switch path {
	case "templates":
		m = DirFS.TplFS
	case "static":
		m = DirFS.StaticFS
	default:
		return fmt.Errorf("template dir %s not known", path)
	}
	f, err := m.Open(fp)
	if err != nil {
		log.Printf("partial open error for %v", err)
		return err
	}
	_, err = io.Copy(w, f)
	if err != nil {
		log.Printf("partial write error for %s: %v", fp, err)
		return err
	}
	return nil
}

// PartialDetailsShow shows an information details partial
func PartialDetailsShow(w http.ResponseWriter, r *http.Request) {
	_ = partialWriter(w, "templates", "partial-details-show.html")
}

// PartialDetailsHide shows the concise information details partial
func PartialDetailsHide(w http.ResponseWriter, r *http.Request) {
	_ = partialWriter(w, "templates", "partial-details-hide.html")
}

// PartialNoContent returns no content
func PartialNoContent(w http.ResponseWriter, r *http.Request) {
	_, _ = w.Write([]byte(""))
}

// PartialAddTrip adds a trip button row
func PartialAddTrip(w http.ResponseWriter, r *http.Request) {
	_ = partialWriter(w, "templates", "partial-addtrip.html")
}

// PartialReport shows the results of a form submission in html
func PartialReport(w http.ResponseWriter, r *http.Request) {

	if r.Method != "POST" {
		w.WriteHeader(http.StatusBadRequest)
		log.Print("endpoint only accepts POST requests, got", r.Method)
		return
	}

	// read body
	body, err := io.ReadAll(r.Body)
	defer r.Body.Close()
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		log.Print("body reading error", err)
		return
	}
	if inDevelopment {
		log.Println("body content:", string(body))
	}

	// extract Holidays from POSTed htmx form
	urlVals, err := url.ParseQuery(string(body))
	if err != nil {
		log.Fatal(err)
	}
	holidays, err := trips.HolidaysURLDecoder(urlVals)
	if inDevelopment {
		log.Printf("holidays GET : %+v err : %v", holidays, err)
	}
	if err != nil {
		_, _ = w.Write([]byte(err.Error()))
		log.Print("form data decoding error", err)
		return
	}
	if len(holidays) < 1 {
		_, _ = w.Write([]byte("no holidays were found"))
		log.Print("no holidays were found")
		return
	}
	if inDevelopment {
		log.Println("holidays", holidays)
	}

	// perform the calculation
	var trs *trips.Trips

	// push htmx browser url to client's browser history
	w.Header().Set("HX-Push-Url", BaseURL+"/?"+trips.HolidaysURLEncode(holidays))

	// error captured in trs.Error
	trs, _ = calculate(holidays)
	log.Println("Error ", trs.Error)

	t := template.Must(template.ParseFS(DirFS.TplFS, "partial-report.html"))
	// t := template.Must(template.New("partial-report.html").ParseFiles(tplBasePath + "partial-report.html"))
	err = t.Execute(w, trs)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		fmt.Fprintf(w, "template writing problem : %s", err.Error())
	}
}
