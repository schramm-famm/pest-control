package handlers

import (
	"encoding/json"
	"io/ioutil"
	"log"
	"net/http"
	"pest-control/models"
)

// PostPrefsHandler creates new preferences for a user
func PostPrefsHandler(w http.ResponseWriter, r *http.Request) {
	defer r.Body.Close()
	reqBody := models.NewPreferences()

	bodyBytes, err := ioutil.ReadAll(r.Body)
	if err != nil {
		errMsg := "Failed to read request body: " + err.Error()
		log.Println(errMsg)
		http.Error(w, errMsg, http.StatusBadRequest)
		return
	}

	if err := json.Unmarshal(bodyBytes, reqBody); err != nil {
		errMsg := "Failed to parse request body: " + err.Error()
		log.Println(errMsg)
		http.Error(w, errMsg, http.StatusBadRequest)
		return
	}

	if len(reqBody.Conversation) > 0 {
		// Set the conversation preferences to the default
		for i := 0; i < len(reqBody.Conversation); i++ {
			reqBody.Conversation[i] = models.NewConversationPrefs()
		}

		// Overwrite the default preferences with the specified preferences.
		// This needs to be done since there is no way to initialize the
		// conversations array with all the fields set to true.
		if err := json.Unmarshal(bodyBytes, reqBody); err != nil {
			errMsg := "Failed to parse request body: " + err.Error()
			log.Println(errMsg)
			http.Error(w, errMsg, http.StatusBadRequest)
			return
		}
	}

	json.NewEncoder(w).Encode(reqBody)
}
