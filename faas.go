// for the main package contents, please see the trips package
// the faas top-level package is a GCP cloud function entry point.

// faas is an entry point for GCP serverless function, which cheekily
// supports other urls too
package faas

import (
	"net/http"

	"github.com/gorilla/mux"
	"github.com/rorycl/timeaway/web"
)

// the BaseURL for GCP is not "/" but the project name
func init() {
	web.BaseURL = "/timeaway"
}

// GCPServer is the entry point for a FAAS application as set out in
// section 4.5 of Joel Holmes' "Shipping Go". The idea for multiple
// endpoints is discussed at
// https://medium.com/google-cloud/hack-use-cloud-functions-as-a-webserver-with-golang-42edc7935247
func GCPServer(w http.ResponseWriter, r *http.Request) {

	m := mux.NewRouter()

	// static
	m.PathPrefix("/static/").Handler(
		http.StripPrefix("/static/",
			http.FileServer(http.FS(web.DirFS.StaticFS))),
	)

	// partials
	m.HandleFunc("/partials/details/show", web.PartialDetailsShow)
	m.HandleFunc("/partials/details/hide", web.PartialDetailsHide)
	m.HandleFunc("/partials/report", web.PartialReport)
	m.HandleFunc("/partials/nocontent", web.PartialNoContent)
	m.HandleFunc("/partials/addtrip", web.PartialAddTrip)

	// main routes
	m.HandleFunc("/", web.Home)
	m.HandleFunc("/home", web.Home)
	m.HandleFunc("/trips", web.Trips)
	m.HandleFunc("/health", web.Health)

	m.ServeHTTP(w, r)
}
