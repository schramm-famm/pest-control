package main

import (
	"github.com/gorilla/mux"
	"log"
	"net/http"
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
	db, err := models.NewDB("mongodb://localhost:27017")
	if err != nil {
		log.Panic(err)
	}
	env := &handlers.Env{db}

	httpMux := mux.NewRouter()
	httpMux.HandleFunc(
		"/api/prefs",
		logging(env.PostPrefsHandler),
	).Methods("POST")
	httpMux.HandleFunc(
		"/api/prefs/conversations",
		logging(env.PostPrefsConvHandler),
	).Methods("POST")
	httpMux.HandleFunc(
		"/api/prefs",
		logging(env.GetPrefsHandler),
	).Methods("GET")
	httpMux.HandleFunc(
		"/api/prefs/conversations/{conversation:[0-9]+}",
		logging(env.GetPrefsConvHandler),
	).Methods("GET")
	httpMux.HandleFunc(
		"/api/prefs",
		logging(env.DeletePrefsHandler),
	).Methods("DELETE")
	httpMux.HandleFunc(
		"/api/prefs/conversations/{conversation:[0-9]+}",
		logging(env.DeletePrefsConvHandler),
	).Methods("DELETE")
	httpMux.HandleFunc(
		"/api/prefs",
		logging(env.PatchPrefsHandler),
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
