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
)

type Env struct {
	DB models.Datastore
}

func parseReqBody(w http.ResponseWriter, body io.ReadCloser, bodyObj *models.Preferences) error {
	bodyBytes, err := ioutil.ReadAll(body)
	if err != nil {
		errMsg := "Failed to read request body: " + err.Error()
		log.Println(errMsg)
		http.Error(w, errMsg, http.StatusBadRequest)
		return err
	}

	if err := json.Unmarshal(bodyBytes, bodyObj); err != nil {
		errMsg := "Failed to parse request body: " + err.Error()
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
			errMsg := "Failed to parse request body: " + err.Error()
			log.Println(errMsg)
			http.Error(w, errMsg, http.StatusBadRequest)
			return err
		}
	}

	return nil
}

func parseIntParams(query url.Values, params ...string) ([]int, error) {
	vals := []int{}
	for _, param := range params {
		var (
			val int
			err error
		)
		str := query.Get(param)
		if str != "" {
			val, err = strconv.Atoi(str)
			if err != nil {
				errMsg := fmt.Sprintf(
					"invalid query format: query parameter '%s' must be an integer",
					param,
				)
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

	json.NewEncoder(w).Encode(reqBody)
}

// GetPrefsHandler gets filtered preferences for a user
func (env *Env) GetPrefsHandler(w http.ResponseWriter, r *http.Request) {
	queryValues, err := url.ParseQuery(r.URL.RawQuery)
	if err != nil {
		errMsg := "Failed to parse query: " + err.Error()
		log.Println(errMsg)
		http.Error(w, errMsg, http.StatusBadRequest)
		return
	}

	vals, err := parseIntParams(queryValues, "user_id", "conversation_id")
	if err != nil {
		log.Println(err.Error())
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	resBody := models.PrefsFilter{
		UserID:         vals[0],
		ConversationID: vals[1],
	}

	json.NewEncoder(w).Encode(resBody)
}

// PutPrefsHandler replaces, or creates if does not exist, a user's preferences
func (env *Env) PutPrefsHandler(w http.ResponseWriter, r *http.Request) {
	defer r.Body.Close()

	reqBody := models.NewPreferences()
	if err := parseReqBody(w, r.Body, reqBody); err != nil {
		return
	}

	json.NewEncoder(w).Encode(reqBody)
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

	vals, err := parseIntParams(queryValues, "user_id", "conversation_id")
	if err != nil {
		log.Println(err.Error())
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	resBody := models.PrefsFilter{
		UserID:         vals[0],
		ConversationID: vals[1],
	}

	json.NewEncoder(w).Encode(resBody)
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
