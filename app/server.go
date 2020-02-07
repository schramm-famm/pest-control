package main

import (
	"fmt"
	"github.com/gorilla/mux"
	"log"
	"net/http"
	"os"
	"pest-control/handlers"
	"pest-control/models"
	"time"
)

func logging(f http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		log.Printf("path: %s, method: %s", r.URL.Path, r.Method)
		f(w, r)
	}
}

func main() {
	connectionString := fmt.Sprintf(
		"mongodb://%s:%s",
		os.Getenv("PESTCONTROL_DB_HOST"),
		os.Getenv("PESTCONTROL_DB_PORT"),
	)
	db, err := models.NewDB(
		connectionString,
		os.Getenv("PESTCONTROL_DB_USER"),
		os.Getenv("PESTCONTROL_DB_PW"),
	)
	if err != nil {
		log.Panic(err)
	}
	env := &handlers.Env{db}

	httpMux := mux.NewRouter()
	httpMux.HandleFunc(
		"/pest-control/v1/prefs",
		logging(env.PostPrefsHandler),
	).Methods("POST")
	httpMux.HandleFunc(
		"/pest-control/v1/prefs/conversations",
		logging(env.PostPrefsConvHandler),
	).Methods("POST")
	httpMux.HandleFunc(
		"/pest-control/v1/prefs",
		logging(env.GetPrefsHandler),
	).Methods("GET")
	httpMux.HandleFunc(
		"/pest-control/v1/prefs/conversations/{conversation:[0-9]+}",
		logging(env.GetPrefsConvHandler),
	).Methods("GET")
	httpMux.HandleFunc(
		"/pest-control/v1/prefs",
		logging(env.DeletePrefsHandler),
	).Methods("DELETE")
	httpMux.HandleFunc(
		"/pest-control/v1/prefs/conversations/{conversation:[0-9]+}",
		logging(env.DeletePrefsConvHandler),
	).Methods("DELETE")
	httpMux.HandleFunc(
		"/pest-control/v1/prefs",
		logging(env.PatchPrefsHandler),
	).Methods("PATCH")
	httpMux.HandleFunc(
		"/pest-control/v1/prefs/conversations/{conversation:[0-9]+}",
		logging(env.PatchPrefsConvHandler),
	).Methods("PATCH")

	httpSrv := &http.Server{
		Addr:         ":80",
		ReadTimeout:  5 * time.Second,
		WriteTimeout: 5 * time.Second,
		IdleTimeout:  120 * time.Second,
		Handler:      httpMux,
	}

	log.Fatal(httpSrv.ListenAndServe())
}
