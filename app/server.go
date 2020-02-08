package main

import (
	"crypto/tls"
	"crypto/x509"
	"errors"
	"fmt"
	"github.com/gorilla/mux"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"pest-control/handlers"
	"pest-control/models"
	"strings"
	"time"
)

func logging(f http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		log.Printf("path: %s, method: %s", r.URL.Path, r.Method)
		f(w, r)
	}
}

func getCustomTLSConfig(caFile string) (*tls.Config, error) {
	tlsConfig := new(tls.Config)
	certs, err := ioutil.ReadFile(caFile)

	if err != nil {
		return tlsConfig, err
	}

	tlsConfig.RootCAs = x509.NewCertPool()
	ok := tlsConfig.RootCAs.AppendCertsFromPEM(certs)

	if !ok {
		return tlsConfig, errors.New("Failed parsing pem file")
	}

	return tlsConfig, nil
}

func main() {
	b := new(strings.Builder)

	fmt.Fprint(b, "mongodb://")
	if user := os.Getenv("PESTCONTROL_DB_USER"); user != "" {
		fmt.Fprintf(b, "%s:%s@", user, os.Getenv("PESTCONTROL_DB_PW"))
	}
	fmt.Fprintf(
		b,
		"%s:%s",
		os.Getenv("PESTCONTROL_DB_HOST"),
		os.Getenv("PESTCONTROL_DB_PORT"),
	)

	var tlsConfig *tls.Config

	if caFilePath := os.Getenv("DOCDB_CERT_PATH"); caFilePath != "" {
		var err error
		fmt.Fprint(b, "/?ssl=true&replicaSet=rs0")
		fmt.Fprint(b, "&readPreference=secondaryPreferred&retryWrites=false")
		tlsConfig, err = getCustomTLSConfig(caFilePath)
		if err != nil {
			log.Fatalf("Failed getting TLS configuration: %v", err)
		}
	}

	db, err := models.NewDB(b.String(), tlsConfig)

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
