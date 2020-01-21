package handlers

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"net/url"
	"pest-control/models"
	"strconv"

	"github.com/gorilla/mux"
	"go.mongodb.org/mongo-driver/mongo"
)

type Env struct {
	DB models.Datastore
}

func parseReqBody(w http.ResponseWriter, body io.ReadCloser, bodyObj *models.Preferences) error {
	bodyBytes, err := ioutil.ReadAll(body)
	if err != nil {
		errMsg := "failed to read request body: " + err.Error()
		log.Println(errMsg)
		http.Error(w, errMsg, http.StatusBadRequest)
		return err
	}

	if err := json.Unmarshal(bodyBytes, bodyObj); err != nil {
		errMsg := "failed to parse request body: " + err.Error()
		log.Println(errMsg)
		http.Error(w, errMsg, http.StatusBadRequest)
		return err
	}

	if len(bodyObj.Conversation) > 0 {
		// Set the conversation preferences to the default
		for i := 0; i < len(bodyObj.Conversation); i++ {
			bodyObj.Conversation[i] = models.NewConversationPrefs()
		}

		// Overwrite the default preferences with the specified preferences.
		// This needs to be done since there is no way to initialize the
		// conversations array with all the fields set to true.
		if err := json.Unmarshal(bodyBytes, bodyObj); err != nil {
			errMsg := "failed to parse request body: " + err.Error()
			log.Println(errMsg)
			http.Error(w, errMsg, http.StatusBadRequest)
			return err
		}
	}

	return nil
}

func parseStringToInt(strings ...string) ([]int, error) {
	var (
		val int
		err error
	)

	vals := make([]int, 0)
	for _, str := range strings {
		val = 0
		if str != "" {
			val, err = strconv.Atoi(str)
			if err != nil {
				errMsg := "failed to convert string to int: " + err.Error()
				return nil, errors.New(errMsg)
			}
		}
		vals = append(vals, val)
	}

	return vals, nil
}

// PostPrefsHandler creates new preferences for a user
func (env *Env) PostPrefsHandler(w http.ResponseWriter, r *http.Request) {
	defer r.Body.Close()

	reqBody := models.NewPreferences()
	if err := parseReqBody(w, r.Body, reqBody); err != nil {
		return
	}

	prefsID, err := env.DB.CreatePrefs(*reqBody)
	if err != nil {
		errMsg := fmt.Sprintf(
			"failed to create prefs (%+v): %s",
			*reqBody,
			err.Error(),
		)
		log.Println(errMsg)
		responseCode := http.StatusInternalServerError
		if err == models.ErrPrefsExists {
			responseCode = http.StatusConflict
		}
		http.Error(w, errMsg, responseCode)
		return
	}

	reqBody.ID = prefsID
	reqBody.UserID = 0

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(reqBody)
}

// PostPrefsConvHandler creates new conversation preferences for a user
func (env *Env) PostPrefsConvHandler(w http.ResponseWriter, r *http.Request) {
	defer r.Body.Close()

	reqBody := models.NewConversationPrefs()
	bodyBytes, err := ioutil.ReadAll(r.Body)
	if err != nil {
		errMsg := "failed to read request body: " + err.Error()
		log.Println(errMsg)
		http.Error(w, errMsg, http.StatusBadRequest)
		return
	}

	if err = json.Unmarshal(bodyBytes, reqBody); err != nil {
		errMsg := "failed to parse request body: " + err.Error()
		log.Println(errMsg)
		http.Error(w, errMsg, http.StatusBadRequest)
		return
	}

	if err = env.DB.CreatePrefsConv(*reqBody); err != nil {
		errMsg := fmt.Sprintf(
			"failed to create conversation prefs (%+v): %s",
			*reqBody,
			err.Error(),
		)
		log.Println(errMsg)
		responseCode := http.StatusInternalServerError
		if err == models.ErrPrefsExists {
			responseCode = http.StatusConflict
		}
		http.Error(w, errMsg, responseCode)
		return
	}

	reqBody.UserID = 0

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(reqBody)
}

// GetPrefsHandler gets a user's global preferences
func (env *Env) GetPrefsHandler(w http.ResponseWriter, r *http.Request) {
	queryValues, err := url.ParseQuery(r.URL.RawQuery)
	if err != nil {
		errMsg := "failed to parse query: " + err.Error()
		log.Println(errMsg)
		http.Error(w, errMsg, http.StatusBadRequest)
		return
	}

	vals, err := parseStringToInt(queryValues.Get("user_id"))
	if err != nil {
		log.Println(err.Error())
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	prefs, err := env.DB.GetPrefs(vals[0])
	if err != nil {
		errMsg := fmt.Sprintf(
			"unable to get preferences for user: %s",
			err.Error(),
		)
		responseCode := http.StatusInternalServerError
		if err == mongo.ErrNoDocuments {
			errMsg = err.Error()
			responseCode = http.StatusNotFound
		}
		log.Println(errMsg)
		http.Error(w, errMsg, responseCode)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(prefs)
}

// GetPrefsConvHandler gets a user's preferences for a conversation
func (env *Env) GetPrefsConvHandler(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)

	queryValues, err := url.ParseQuery(r.URL.RawQuery)
	if err != nil {
		errMsg := "failed to parse query: " + err.Error()
		log.Println(errMsg)
		http.Error(w, errMsg, http.StatusBadRequest)
		return
	}

	vals, err := parseStringToInt(queryValues.Get("user_id"), vars["conversation"])
	if err != nil {
		log.Println(err.Error())
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	prefs, err := env.DB.GetPrefsConv(vals[0], vals[1])
	if err != nil {
		errMsg := fmt.Sprintf(
			"unable to get conversation preferences for user: %s",
			err.Error(),
		)
		responseCode := http.StatusInternalServerError
		if err == mongo.ErrNoDocuments {
			errMsg = err.Error()
			responseCode = http.StatusNotFound
		}
		log.Println(errMsg)
		http.Error(w, errMsg, responseCode)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(prefs)
}

// DeletePrefsHandler deletes a user's preferences
func (env *Env) DeletePrefsHandler(w http.ResponseWriter, r *http.Request) {
	queryValues, err := url.ParseQuery(r.URL.RawQuery)
	if err != nil {
		errMsg := "Failed to parse query: " + err.Error()
		log.Println(errMsg)
		http.Error(w, errMsg, http.StatusBadRequest)
		return
	}

	vals, err := parseStringToInt(queryValues.Get("user_id"))
	if err != nil {
		log.Println(err.Error())
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	if err = env.DB.DeletePrefs(vals[0]); err != nil {
		errMsg := fmt.Sprintf(
			"unable to delete preferences for user: %s",
			err.Error(),
		)
		responseCode := http.StatusInternalServerError
		if err == models.ErrPrefsDNE {
			errMsg = err.Error()
			responseCode = http.StatusNotFound
		}
		log.Println(errMsg)
		http.Error(w, errMsg, responseCode)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

// DeletePrefsConvHandler deletes a user's preferences for a conversation
func (env *Env) DeletePrefsConvHandler(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)

	queryValues, err := url.ParseQuery(r.URL.RawQuery)
	if err != nil {
		errMsg := "failed to parse query: " + err.Error()
		log.Println(errMsg)
		http.Error(w, errMsg, http.StatusBadRequest)
		return
	}

	vals, err := parseStringToInt(queryValues.Get("user_id"), vars["conversation"])
	if err != nil {
		log.Println(err.Error())
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	if err = env.DB.DeletePrefsConv(vals[0], vals[1]); err != nil {
		errMsg := fmt.Sprintf(
			"unable to delete preferences for user: %s",
			err.Error(),
		)
		responseCode := http.StatusInternalServerError
		if err == models.ErrPrefsDNE {
			errMsg = err.Error()
			responseCode = http.StatusNotFound
		}
		log.Println(errMsg)
		http.Error(w, errMsg, responseCode)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

// PatchPrefsHandler updates a user's preferences
func (env *Env) PatchPrefsHandler(w http.ResponseWriter, r *http.Request) {
	defer r.Body.Close()

	reqBody := models.NewPreferences()
	if err := parseReqBody(w, r.Body, reqBody); err != nil {
		return
	}

	json.NewEncoder(w).Encode(reqBody)
}
