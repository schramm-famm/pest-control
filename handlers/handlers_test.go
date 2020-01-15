package handlers

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"pest-control/models"
	"reflect"
	"testing"
)

type mockDB struct{}

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
			rBody, _ := json.Marshal(test.ReqBody)
			r := httptest.NewRequest("POST", "/api/prefs", bytes.NewReader(rBody))
			w := httptest.NewRecorder()

			env := &Env{DB: &mockDB{}}
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

func TestGetPrefsHandler(t *testing.T) {
	tests := []struct {
		Name       string
		StatusCode int
		Query      string
		ResBody    map[string]string
	}{
		{
			Name:       "Successful preference retrieval with no query",
			StatusCode: http.StatusOK,
			ResBody: map[string]string{
				"user_id":         "",
				"conversation_id": "",
			},
		},
		{
			Name:       "Successful preference retrieval with user query",
			StatusCode: http.StatusOK,
			Query:      "?user_id=blah",
			ResBody: map[string]string{
				"user_id":         "blah",
				"conversation_id": "",
			},
		},
		{
			Name:       "Successful preference retrieval with conversation query",
			StatusCode: http.StatusOK,
			Query:      "?user_id=blah&conversation_id=blah",
			ResBody: map[string]string{
				"user_id":         "blah",
				"conversation_id": "blah",
			},
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

			env := &Env{DB: &mockDB{}}
			env.GetPrefsHandler(w, r)

			if w.Code != test.StatusCode {
				t.Errorf("Response has incorrect status code, expected status code %d, got %d", test.StatusCode, w.Code)
			}

			if w.Code == http.StatusOK {
				resBody := make(map[string]string)
				_ = json.NewDecoder(w.Body).Decode(&resBody)
				if !reflect.DeepEqual(test.ResBody, resBody) {
					t.Errorf("Response has incorrect query, expected %+v, got %+v", test.ResBody, resBody)
				}
			}
		})
	}
}

func TestPutPrefsHandler(t *testing.T) {
	tests := []struct {
		Name       string
		StatusCode int
		ReqBody    interface{}
		ResBody    models.Preferences
	}{
		{
			Name:       "Successful default preference replacement",
			StatusCode: http.StatusOK,
			ReqBody:    map[string]interface{}{},
			ResBody:    *models.NewPreferences(),
		},
		{
			Name:       "Successful custom preference replacement",
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
			Name:       "Unsuccessful preference replacement with bad request",
			StatusCode: http.StatusBadRequest,
			ReqBody: map[string]string{
				"global": "true",
			},
		},
	}

	for _, test := range tests {
		t.Run(test.Name, func(t *testing.T) {
			rBody, _ := json.Marshal(test.ReqBody)
			r := httptest.NewRequest("PUT", "/api/prefs", bytes.NewReader(rBody))
			w := httptest.NewRecorder()

			env := &Env{DB: &mockDB{}}
			env.PutPrefsHandler(w, r)

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

func TestDeletePrefsHandler(t *testing.T) {
	tests := []struct {
		Name       string
		StatusCode int
		Query      string
		ResBody    map[string]string
	}{
		{
			Name:       "Successful preference deletion",
			StatusCode: http.StatusOK,
			ResBody: map[string]string{
				"user_id":         "",
				"conversation_id": "",
			},
		},
		{
			Name:       "Successful preference deletion with user query",
			StatusCode: http.StatusOK,
			Query:      "?user_id=blah",
			ResBody: map[string]string{
				"user_id":         "blah",
				"conversation_id": "",
			},
		},
		{
			Name:       "Successful preference deletion with conversation query",
			StatusCode: http.StatusOK,
			Query:      "?user_id=blah&conversation_id=blah",
			ResBody: map[string]string{
				"user_id":         "blah",
				"conversation_id": "blah",
			},
		},
	}

	for _, test := range tests {
		t.Run(test.Name, func(t *testing.T) {
			r := httptest.NewRequest("DELETE", "/api/prefs"+test.Query, nil)
			w := httptest.NewRecorder()

			env := &Env{DB: &mockDB{}}
			env.DeletePrefsHandler(w, r)

			if w.Code != test.StatusCode {
				t.Errorf("Response has incorrect status code, expected status code %d, got %d", test.StatusCode, w.Code)
			}

			if w.Code == http.StatusOK {
				resBody := make(map[string]string)
				_ = json.NewDecoder(w.Body).Decode(&resBody)
				if !reflect.DeepEqual(test.ResBody, resBody) {
					t.Errorf("Response has incorrect query, expected %+v, got %+v", test.ResBody, resBody)
				}
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

			env := &Env{DB: &mockDB{}}
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
