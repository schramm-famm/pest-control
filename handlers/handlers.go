package handlers

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"pest-control/models"
	"strconv"

	"github.com/gorilla/mux"
)

type Env struct {
	DB models.Datastore
}

const (
	ApplicationJSON        = "application/json"
	InternalServerErrorStr = "Internal Server Error"
)

func parseReqBody(w http.ResponseWriter, body io.ReadCloser, bodyObj interface{}) error {
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

	if prefs, ok := bodyObj.(*models.Preferences); ok && len(prefs.Conversation) > 0 {
		// Set the conversation preferences to the default
		for i := 0; i < len(prefs.Conversation); i++ {
			prefs.Conversation[i] = models.NewConversationPrefs()
		}

		// Overwrite the default preferences with the specified preferences.
		// This needs to be done since there is no way to initialize the
		// conversations array with all the fields set to true.
		if err := json.Unmarshal(bodyBytes, prefs); err != nil {
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

	vals, err := parseStringToInt(r.Header.Get("User-ID"))
	if err != nil {
		errMsg := "Invalid user ID"
		log.Println(errMsg + ": " + err.Error())
		http.Error(w, errMsg, http.StatusBadRequest)
		return
	}

	reqBody.UserID = vals[0]

	if err = env.DB.CreatePrefs(reqBody); err != nil {
		log.Printf(
			"failed to create prefs (%+v): %s",
			*reqBody,
			err.Error(),
		)
		errMsg := InternalServerErrorStr
		responseCode := http.StatusInternalServerError
		if err == models.ErrPrefsExists {
			errMsg = err.Error()
			responseCode = http.StatusConflict
		}
		http.Error(w, errMsg, responseCode)
		return
	}

	reqBody.UserID = 0

	w.Header().Set("Content-Type", ApplicationJSON)
	w.Header().Set("Location", r.URL.Path)
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(reqBody)
}

// PostPrefsConvHandler creates new conversation preferences for a user
func (env *Env) PostPrefsConvHandler(w http.ResponseWriter, r *http.Request) {
	defer r.Body.Close()

	reqBody := models.NewConversationPrefs()
	if err := parseReqBody(w, r.Body, reqBody); err != nil {
		return
	}

	vals, err := parseStringToInt(r.Header.Get("User-ID"))
	if err != nil {
		errMsg := "Invalid user ID"
		log.Println(errMsg + ": " + err.Error())
		http.Error(w, errMsg, http.StatusBadRequest)
		return
	}

	if err := env.DB.CreatePrefsConv(vals[0], reqBody); err != nil {
		log.Printf(
			"failed to create conversation prefs (%+v): %s",
			*reqBody,
			err.Error(),
		)
		errMsg := InternalServerErrorStr
		responseCode := http.StatusInternalServerError
		if err == models.ErrPrefsDNE {
			errMsg = err.Error()
			responseCode = http.StatusNotFound
		} else if err == models.ErrPrefsConvExists {
			errMsg = err.Error()
			responseCode = http.StatusConflict
		}
		http.Error(w, errMsg, responseCode)
		return
	}

	w.Header().Set("Content-Type", ApplicationJSON)
	w.Header().Set(
		"Location",
		fmt.Sprintf("%s/%d", r.URL.Path, reqBody.ConversationID),
	)
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(reqBody)
}

// GetPrefsHandler gets a user's global preferences
func (env *Env) GetPrefsHandler(w http.ResponseWriter, r *http.Request) {
	vals, err := parseStringToInt(r.Header.Get("User-ID"))
	if err != nil {
		errMsg := "Invalid user ID"
		log.Println(errMsg + ": " + err.Error())
		http.Error(w, errMsg, http.StatusBadRequest)
		return
	}

	prefs, err := env.DB.GetPrefs(vals[0])
	if err != nil {
		log.Printf(
			"unable to get preferences for user: %s",
			err.Error(),
		)
		errMsg := InternalServerErrorStr
		responseCode := http.StatusInternalServerError
		if err == models.ErrPrefsDNE {
			errMsg = err.Error()
			responseCode = http.StatusNotFound
		}
		http.Error(w, errMsg, responseCode)
		return
	}

	w.Header().Set("Content-Type", ApplicationJSON)
	json.NewEncoder(w).Encode(prefs)
}

// GetPrefsConvHandler gets a user's preferences for a conversation
func (env *Env) GetPrefsConvHandler(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)

	vals, err := parseStringToInt(r.Header.Get("User-ID"), vars["conversation"])
	if err != nil {
		errMsg := "Invalid user ID or conversation ID"
		log.Println(errMsg + ": " + err.Error())
		http.Error(w, errMsg, http.StatusBadRequest)
		return
	}

	prefs, err := env.DB.GetPrefsConv(vals[0], vals[1])
	if err != nil {
		log.Printf(
			"unable to get conversation preferences for user: %s",
			err.Error(),
		)
		errMsg := InternalServerErrorStr
		responseCode := http.StatusInternalServerError
		if err == models.ErrPrefsConvDNE {
			errMsg = err.Error()
			responseCode = http.StatusNotFound
		}
		http.Error(w, errMsg, responseCode)
		return
	}

	w.Header().Set("Content-Type", ApplicationJSON)
	json.NewEncoder(w).Encode(prefs)
}

// DeletePrefsHandler deletes a user's preferences
func (env *Env) DeletePrefsHandler(w http.ResponseWriter, r *http.Request) {
	vals, err := parseStringToInt(r.Header.Get("User-ID"))
	if err != nil {
		errMsg := "Invalid user ID"
		log.Println(errMsg + ": " + err.Error())
		http.Error(w, errMsg, http.StatusBadRequest)
		return
	}

	if err = env.DB.DeletePrefs(vals[0]); err != nil {
		log.Printf(
			"unable to delete preferences for user: %s",
			err.Error(),
		)
		errMsg := InternalServerErrorStr
		responseCode := http.StatusInternalServerError
		if err == models.ErrPrefsDNE {
			errMsg = err.Error()
			responseCode = http.StatusNotFound
		}
		http.Error(w, errMsg, responseCode)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

// DeletePrefsConvHandler deletes a user's preferences for a conversation
func (env *Env) DeletePrefsConvHandler(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)

	vals, err := parseStringToInt(r.Header.Get("User-ID"), vars["conversation"])
	if err != nil {
		errMsg := "Invalid user ID or conversation ID"
		log.Println(errMsg + ": " + err.Error())
		http.Error(w, errMsg, http.StatusBadRequest)
		return
	}

	if err = env.DB.DeletePrefsConv(vals[0], vals[1]); err != nil {
		log.Printf(
			"unable to delete preferences for user: %s",
			err.Error(),
		)
		errMsg := InternalServerErrorStr
		responseCode := http.StatusInternalServerError
		if err == models.ErrPrefsConvDNE {
			errMsg = err.Error()
			responseCode = http.StatusNotFound
		}
		http.Error(w, errMsg, responseCode)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

// PatchPrefsHandler updates a user's preferences
func (env *Env) PatchPrefsHandler(w http.ResponseWriter, r *http.Request) {
	defer r.Body.Close()

	reqBody := &models.GlobalPrefs{}
	if err := parseReqBody(w, r.Body, reqBody); err != nil {
		return
	}

	vals, err := parseStringToInt(r.Header.Get("User-ID"))
	if err != nil {
		errMsg := "Invalid user ID or conversation ID"
		log.Println(errMsg + ": " + err.Error())
		http.Error(w, errMsg, http.StatusBadRequest)
		return
	}

	if err := env.DB.PatchPrefs(vals[0], reqBody); err != nil {
		log.Printf(
			"unable to update preferences for user: %s",
			err.Error(),
		)
		errMsg := InternalServerErrorStr
		responseCode := http.StatusInternalServerError
		if err == models.ErrPrefsDNE {
			errMsg = err.Error()
			responseCode = http.StatusNotFound
		}
		http.Error(w, errMsg, responseCode)
		return
	}

	w.Header().Set("Content-Type", ApplicationJSON)
	json.NewEncoder(w).Encode(reqBody)
}

// PatchPrefsConvHandler updates a user's preferences for a specific conversation
func (env *Env) PatchPrefsConvHandler(w http.ResponseWriter, r *http.Request) {
	defer r.Body.Close()

	vars := mux.Vars(r)

	reqBody := &models.ConversationPrefs{}
	if err := parseReqBody(w, r.Body, reqBody); err != nil {
		return
	}

	vals, err := parseStringToInt(r.Header.Get("User-ID"), vars["conversation"])
	if err != nil {
		errMsg := "Invalid user ID or conversation ID"
		log.Println(errMsg + ": " + err.Error())
		http.Error(w, errMsg, http.StatusBadRequest)
		return
	}

	if err := env.DB.PatchPrefsConv(vals[0], vals[1], reqBody); err != nil {
		log.Printf(
			"unable to update preferences for user: %s",
			err.Error(),
		)
		errMsg := InternalServerErrorStr
		responseCode := http.StatusInternalServerError
		if err == models.ErrPrefsConvDNE {
			errMsg = err.Error()
			responseCode = http.StatusNotFound
		}
		http.Error(w, errMsg, responseCode)
		return
	}

	w.Header().Set("Content-Type", ApplicationJSON)
	json.NewEncoder(w).Encode(reqBody)
}
