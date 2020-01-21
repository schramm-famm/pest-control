package handlers

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"pest-control/models"
	"reflect"
	"strings"
	"testing"

	"go.mongodb.org/mongo-driver/mongo"
)

func TestPostPrefsHandler(t *testing.T) {
	tests := []struct {
		Name       string
		StatusCode int
		ReqBody    interface{}
		ResBody    models.Preferences
	}{
		{
			Name:       "Successful default preference creation",
			StatusCode: http.StatusOK,
			ReqBody:    map[string]interface{}{},
			ResBody:    *models.NewPreferences(),
		},
		{
			Name:       "Successful custom preference creation",
			StatusCode: http.StatusOK,
			ReqBody: map[string]interface{}{
				"global": map[string]bool{
					"invitation":   false,
					"text_entered": false,
				},
				"conversation": []map[string]bool{{
					"tag": false,
				}},
			},
			ResBody: models.Preferences{
				Global: &models.GlobalPrefs{
					Role:         true,
					Tag:          true,
					TextModified: true,
				},
				Conversation: []*models.ConversationPrefs{{
					Role:         true,
					TextEntered:  true,
					TextModified: true,
				}},
			},
		},
		{
			Name:       "Unsuccessful preference creation with bad request",
			StatusCode: http.StatusBadRequest,
			ReqBody: map[string]string{
				"global": "true",
			},
		},
	}

	for _, test := range tests {
		t.Run(test.Name, func(t *testing.T) {
			test.ResBody.ID = "blah"
			rBody, _ := json.Marshal(test.ReqBody)
			r := httptest.NewRequest("POST", "/api/prefs", bytes.NewReader(rBody))
			w := httptest.NewRecorder()

			mDB := &models.MockDB{
				ID: test.ResBody.ID,
			}

			env := &Env{DB: mDB}
			env.PostPrefsHandler(w, r)

			if w.Code != test.StatusCode {
				t.Errorf("Response has incorrect status code, expected status code %d, got %d", test.StatusCode, w.Code)
			}

			if w.Code == http.StatusOK {
				resBody := models.Preferences{}
				_ = json.NewDecoder(w.Body).Decode(&resBody)
				if !reflect.DeepEqual(test.ResBody, resBody) {
					t.Errorf("Response has incorrect preferences, expected %+v, got %+v", test.ResBody, resBody)
				}
			}
		})
	}
}

func TestPostPrefsConvHandler(t *testing.T) {
	tests := []struct {
		Name       string
		StatusCode int
		ReqBody    map[string]interface{}
		ResBody    models.ConversationPrefs
		Error      error
	}{
		{
			Name:       "Successful default conversation preference creation",
			StatusCode: http.StatusOK,
			ReqBody:    map[string]interface{}{},
			ResBody:    *models.NewConversationPrefs(),
		},
		{
			Name:       "Successful custom conversation preference creation",
			StatusCode: http.StatusOK,
			ReqBody: map[string]interface{}{
				"tag": false,
			},
			ResBody: models.ConversationPrefs{
				Role:         true,
				TextEntered:  true,
				TextModified: true,
			},
		},
		{
			Name:       "Unsuccessful conversation preference creation with bad request",
			StatusCode: http.StatusBadRequest,
			ReqBody: map[string]interface{}{
				"text_modified": 10,
			},
		},
		{
			Name:       "Unsuccessful conversation preference creation with existing conversation",
			StatusCode: http.StatusConflict,
			ReqBody: map[string]interface{}{
				"tag": false,
			},
			Error: models.ErrPrefsExists,
		},
	}

	for _, test := range tests {
		t.Run(test.Name, func(t *testing.T) {
			rBody, _ := json.Marshal(test.ReqBody)
			r := httptest.NewRequest("POST", "/api/prefs", bytes.NewReader(rBody))
			w := httptest.NewRecorder()

			mDB := &models.MockDB{CreateErr: test.Error}

			env := &Env{DB: mDB}
			env.PostPrefsConvHandler(w, r)

			if w.Code != test.StatusCode {
				t.Errorf("Response has incorrect status code, expected status code %d, got %d", test.StatusCode, w.Code)
			}

			if w.Code == http.StatusOK {
				resBody := models.ConversationPrefs{}
				_ = json.NewDecoder(w.Body).Decode(&resBody)
				if !reflect.DeepEqual(test.ResBody, resBody) {
					t.Errorf("Response has incorrect preferences, expected %+v, got %+v", test.ResBody, resBody)
				}
			}
		})
	}
}

func TestGetPrefsHandler(t *testing.T) {
	tests := []struct {
		Name       string
		StatusCode int
		Query      string
		ResBody    models.GlobalPrefs
		Error      error
	}{
		{
			Name:       "Successful preference retrieval with user query",
			StatusCode: http.StatusOK,
			Query:      "?user_id=2",
			ResBody: models.GlobalPrefs{
				TextModified: true,
				TextEntered:  true,
			},
		},
		{
			Name:       "Unsuccessful preference retrieval with non-existent user",
			StatusCode: http.StatusNotFound,
			Query:      "?user_id=2",
			Error:      mongo.ErrNoDocuments,
		},
	}

	for _, test := range tests {
		t.Run(test.Name, func(t *testing.T) {
			r := httptest.NewRequest(
				"",
				"/api/prefs"+test.Query,
				nil,
			)
			w := httptest.NewRecorder()

			env := &Env{DB: &models.MockDB{
				Prefs: &models.Preferences{
					Global: &test.ResBody,
				},
				GetErr: test.Error,
			}}
			env.GetPrefsHandler(w, r)

			if w.Code != test.StatusCode {
				t.Errorf(
					"Response has incorrect status code, expected %d, got %d",
					test.StatusCode,
					w.Code,
				)
			}

			if w.Code == http.StatusOK {
				resBody := models.GlobalPrefs{}
				_ = json.NewDecoder(w.Body).Decode(&resBody)
				if !reflect.DeepEqual(test.ResBody, resBody) {
					t.Errorf(
						"Response has incorrect body, expected %+v, got %+v",
						test.ResBody,
						resBody,
					)
				}
			} else if w.Code == http.StatusNotFound &&
				strings.TrimRight(w.Body.String(), "\n") != test.Error.Error() {
				t.Errorf(
					"Response is incorrect, expected %s, got %s",
					test.Error.Error(),
					w.Body.String(),
				)
			}
		})
	}
}

func TestGetPrefsConvHandler(t *testing.T) {
	tests := []struct {
		Name       string
		StatusCode int
		ResBody    models.ConversationPrefs
		Error      error
	}{
		{
			Name:       "Successful preference retrieval with user query",
			StatusCode: http.StatusOK,
			ResBody: models.ConversationPrefs{
				TextModified: true,
				TextEntered:  true,
			},
		},
		{
			Name:       "Unsuccessful preference retrieval with non-existent user",
			StatusCode: http.StatusNotFound,
			Error:      mongo.ErrNoDocuments,
		},
	}

	for _, test := range tests {
		t.Run(test.Name, func(t *testing.T) {
			r := httptest.NewRequest(
				"",
				"/api/prefs/conversations/1?user_id=2",
				nil,
			)
			w := httptest.NewRecorder()

			env := &Env{DB: &models.MockDB{
				Prefs: &models.Preferences{
					Conversation: []*models.ConversationPrefs{&test.ResBody},
				},
				GetErr: test.Error,
			}}
			env.GetPrefsConvHandler(w, r)

			if w.Code != test.StatusCode {
				t.Errorf(
					"Response has incorrect status code, expected %d, got %d",
					test.StatusCode,
					w.Code,
				)
			}

			if w.Code == http.StatusOK {
				resBody := models.ConversationPrefs{}
				err := json.NewDecoder(w.Body).Decode(&resBody)
				if err != nil {
					t.Errorf(
						"Error occurred while decoding response body: %s",
						err.Error(),
					)
				}
				if !reflect.DeepEqual(test.ResBody, resBody) {
					t.Errorf(
						"Response has incorrect body, expected %+v, got %+v",
						test.ResBody,
						resBody,
					)
				}
			} else if w.Code == http.StatusNotFound &&
				strings.TrimRight(w.Body.String(), "\n") != test.Error.Error() {
				t.Errorf(
					"Response is incorrect, expected %s, got %s",
					test.Error.Error(),
					w.Body.String(),
				)
			}
		})
	}
}

func TestDeletePrefsHandler(t *testing.T) {
	tests := []struct {
		Name       string
		StatusCode int
		Error      error
	}{
		{
			Name:       "Successful preference deletion",
			StatusCode: http.StatusNoContent,
		},
		{
			Name:       "Unsuccessful preference deletion for non-existent resource",
			StatusCode: http.StatusNotFound,
			Error:      models.ErrPrefsDNE,
		},
	}

	for _, test := range tests {
		t.Run(test.Name, func(t *testing.T) {
			r := httptest.NewRequest("DELETE", "/api/prefs?user_id=2", nil)
			w := httptest.NewRecorder()

			env := &Env{DB: &models.MockDB{DeleteErr: test.Error}}
			env.DeletePrefsHandler(w, r)

			if w.Code != test.StatusCode {
				t.Errorf("Response has incorrect status code, expected status code %d, got %d", test.StatusCode, w.Code)
			}

			if w.Code > http.StatusNoContent &&
				strings.TrimRight(w.Body.String(), "\n") != test.Error.Error() {
				t.Errorf(
					"Response has incorrect body, expected %s, got %s",
					test.Error.Error(),
					w.Body.String(),
				)
			}
		})
	}
}

func TestDeletePrefsConvHandler(t *testing.T) {
	tests := []struct {
		Name       string
		StatusCode int
		Error      error
	}{
		{
			Name:       "Successful conversation preference deletion",
			StatusCode: http.StatusNoContent,
		},
		{
			Name:       "Unsuccessful conversation preference deletion for non-existent resource",
			StatusCode: http.StatusNotFound,
			Error:      models.ErrPrefsDNE,
		},
	}

	for _, test := range tests {
		t.Run(test.Name, func(t *testing.T) {
			r := httptest.NewRequest("DELETE", "/api/prefs/conversations/1?user_id=2", nil)
			w := httptest.NewRecorder()

			env := &Env{DB: &models.MockDB{DeleteErr: test.Error}}
			env.DeletePrefsHandler(w, r)

			if w.Code != test.StatusCode {
				t.Errorf("Response has incorrect status code, expected status code %d, got %d", test.StatusCode, w.Code)
			}

			if w.Code > http.StatusNoContent &&
				strings.TrimRight(w.Body.String(), "\n") != test.Error.Error() {
				t.Errorf(
					"Response has incorrect body, expected %s, got %s",
					test.Error.Error(),
					w.Body.String(),
				)
			}
		})
	}
}

func TestPatchPrefsHandler(t *testing.T) {
	tests := []struct {
		Name       string
		StatusCode int
		ReqBody    interface{}
		ResBody    models.Preferences
	}{
		{
			Name:       "Successful default preference update",
			StatusCode: http.StatusOK,
			ReqBody:    map[string]interface{}{},
			ResBody:    *models.NewPreferences(),
		},
		{
			Name:       "Successful custom preference update",
			StatusCode: http.StatusOK,
			ReqBody: map[string]interface{}{
				"global": map[string]bool{
					"invitation":   false,
					"text_entered": false,
				},
				"conversation": []map[string]bool{{
					"tag": false,
				}},
			},
			ResBody: models.Preferences{
				Global: &models.GlobalPrefs{
					Role:         true,
					Tag:          true,
					TextModified: true,
				},
				Conversation: []*models.ConversationPrefs{{
					Role:         true,
					TextEntered:  true,
					TextModified: true,
				}},
			},
		},
		{
			Name:       "Unsuccessful preference update with bad request",
			StatusCode: http.StatusBadRequest,
			ReqBody: map[string]string{
				"global": "true",
			},
		},
	}

	for _, test := range tests {
		t.Run(test.Name, func(t *testing.T) {
			rBody, _ := json.Marshal(test.ReqBody)
			r := httptest.NewRequest("PATCH", "/api/prefs", bytes.NewReader(rBody))
			w := httptest.NewRecorder()

			env := &Env{DB: &models.MockDB{}}
			env.PatchPrefsHandler(w, r)

			if w.Code != test.StatusCode {
				t.Errorf("Response has incorrect status code, expected status code %d, got %d", test.StatusCode, w.Code)
			}

			if w.Code == http.StatusOK {
				resBody := models.Preferences{}
				_ = json.NewDecoder(w.Body).Decode(&resBody)
				if !reflect.DeepEqual(test.ResBody, resBody) {
					t.Errorf("Response has incorrect preferences, expected %+v, got %+v", test.ResBody, resBody)
				}
			}
		})
	}
}
