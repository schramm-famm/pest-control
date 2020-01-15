package handlers

import (
	"encoding/json"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"net/url"
	"pest-control/models"
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

// PostPrefsHandler creates new preferences for a user
func (env *Env) PostPrefsHandler(w http.ResponseWriter, r *http.Request) {
	defer r.Body.Close()

	reqBody := models.NewPreferences()
	if err := parseReqBody(w, r.Body, reqBody); err != nil {
		return
	}

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

	resBody := map[string]string{
		"user_id":         queryValues.Get("user_id"),
		"conversation_id": queryValues.Get("conversation_id"),
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

	resBody := map[string]string{
		"user_id":         queryValues.Get("user_id"),
		"conversation_id": queryValues.Get("conversation_id"),
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
