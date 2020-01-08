package main

import (
	"github.com/gorilla/mux"
	"log"
	"net/http"
	"pest-control/handlers"
	"time"
)

func main() {
	httpMux := mux.NewRouter()
	httpMux.HandleFunc("/api/prefs", handlers.PostPrefsHandler).Methods("POST")

	httpSrv := &http.Server{
		Addr:         ":80",
		ReadTimeout:  5 * time.Second,
		WriteTimeout: 5 * time.Second,
		IdleTimeout:  120 * time.Second,
		Handler:      httpMux,
	}

	log.Fatal(httpSrv.ListenAndServe())
}
