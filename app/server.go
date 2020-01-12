package main

import (
	"github.com/gorilla/mux"
	"log"
	"net/http"
	"pest-control/handlers"
	"time"
)

func logging(f http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		log.Printf("path: %s, method: %s", r.URL.Path, r.Method)
		f(w, r)
	}
}

func main() {
	httpMux := mux.NewRouter()
	httpMux.HandleFunc(
		"/api/prefs",
		logging(handlers.PostPrefsHandler),
	).Methods("POST")
	httpMux.HandleFunc(
		"/api/prefs",
		logging(handlers.GetPrefsHandler),
	).Methods("GET")
	httpMux.HandleFunc(
		"/api/prefs",
		logging(handlers.PutPrefsHandler),
	).Methods("PUT")
	httpMux.HandleFunc(
		"/api/prefs",
		logging(handlers.DeletePrefsHandler),
	).Methods("DELETE")

	httpSrv := &http.Server{
		Addr:         ":80",
		ReadTimeout:  5 * time.Second,
		WriteTimeout: 5 * time.Second,
		IdleTimeout:  120 * time.Second,
		Handler:      httpMux,
	}

	log.Fatal(httpSrv.ListenAndServe())
}
