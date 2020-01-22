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
				"global": map[string]models.Option{
					"invitation":   models.None,
					"text_entered": models.Email,
				},
				"conversation": []map[string]models.Option{{
					"tag": models.Browser,
				}},
			},
			ResBody: models.Preferences{
				Global: &models.GlobalPrefs{
					models.None,
					&models.GeneralPrefs{
						Role:         models.All,
						Tag:          models.All,
						TextEntered:  models.Email,
						TextModified: models.All,
					},
				},
				Conversation: []*models.ConversationPrefs{{
					0,
					&models.GeneralPrefs{
						Role:         models.All,
						Tag:          models.Browser,
						TextEntered:  models.All,
						TextModified: models.All,
					},
				}},
			},
		},
		{
			Name:       "Unsuccessful preference creation with bad request",
			StatusCode: http.StatusBadRequest,
			ReqBody: map[string]interface{}{
				"global": map[string]string{"invitation": "something"},
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
				Prefs: &test.ResBody,
			}

			env := &Env{DB: mDB}
			env.PostPrefsHandler(w, r)

			if w.Code != test.StatusCode {
				t.Errorf("Response has incorrect status code, expected status code %d, got %d", test.StatusCode, w.Code)
			}

			if w.Code == http.StatusOK {
				resBody := models.Preferences{Conversation: []*models.ConversationPrefs{}}
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
				"tag": models.None,
			},
			ResBody: models.ConversationPrefs{
				0,
				&models.GeneralPrefs{
					Role:         models.All,
					Tag:          models.None,
					TextEntered:  models.All,
					TextModified: models.All,
				},
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
				"tag": models.None,
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
		ResBody    models.GlobalPrefs
		Error      error
	}{
		{
			Name:       "Successful preference retrieval",
			StatusCode: http.StatusOK,
			ResBody: models.GlobalPrefs{
				models.All,
				&models.GeneralPrefs{
					TextModified: models.All,
					TextEntered:  models.All,
				},
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
			r := httptest.NewRequest("", "/api/prefs", nil)
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
				0,
				&models.GeneralPrefs{
					TextModified: models.All,
					TextEntered:  models.None,
				},
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
			r := httptest.NewRequest("", "/api/prefs/conversations/1", nil)
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
			r := httptest.NewRequest("DELETE", "/api/prefs", nil)
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
			r := httptest.NewRequest("DELETE", "/api/prefs/conversations/1", nil)
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
		ReqBody    map[string]interface{}
		ResBody    models.GlobalPrefs
		Error      error
	}{
		{
			Name:       "Successful default preference update",
			StatusCode: http.StatusOK,
			ReqBody:    map[string]interface{}{},
			ResBody:    models.GlobalPrefs{},
		},
		{
			Name:       "Successful custom preference update",
			StatusCode: http.StatusOK,
			ReqBody: map[string]interface{}{
				"tag": models.Email,
			},
			ResBody: models.GlobalPrefs{
				models.Option(""),
				&models.GeneralPrefs{
					Tag: models.Email,
				},
			},
		},
		{
			Name:       "Unsuccessful preference update with bad request",
			StatusCode: http.StatusBadRequest,
			ReqBody: map[string]interface{}{
				"text_modified": 10,
			},
		},
		{
			Name:       "Unsuccessful preference update with non-existent resource",
			StatusCode: http.StatusNotFound,
			ReqBody: map[string]interface{}{
				"tag": models.Browser,
			},
			Error: mongo.ErrNoDocuments,
		},
	}

	for _, test := range tests {
		t.Run(test.Name, func(t *testing.T) {
			rBody, _ := json.Marshal(test.ReqBody)
			r := httptest.NewRequest("PATCH", "/api/prefs", bytes.NewReader(rBody))
			w := httptest.NewRecorder()

			mDB := &models.MockDB{PatchErr: test.Error}

			env := &Env{DB: mDB}
			env.PatchPrefsHandler(w, r)

			if w.Code != test.StatusCode {
				t.Errorf("Response has incorrect status code, expected status code %d, got %d", test.StatusCode, w.Code)
			}

			if w.Code == http.StatusOK {
				resBody := models.GlobalPrefs{}
				_ = json.NewDecoder(w.Body).Decode(&resBody)
				if !reflect.DeepEqual(test.ResBody, resBody) {
					t.Errorf("Response has incorrect body, expected %+v, got %+v", test.ResBody, resBody)
				}
			} else if w.Code == http.StatusNotFound &&
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

func TestPatchPrefsConvHandler(t *testing.T) {
	tests := []struct {
		Name       string
		StatusCode int
		ReqBody    map[string]interface{}
		ResBody    models.ConversationPrefs
		Error      error
	}{
		{
			Name:       "Successful default preference update",
			StatusCode: http.StatusOK,
			ReqBody:    map[string]interface{}{},
			ResBody:    models.ConversationPrefs{},
		},
		{
			Name:       "Successful custom preference update",
			StatusCode: http.StatusOK,
			ReqBody: map[string]interface{}{
				"tag": models.All,
			},
			ResBody: models.ConversationPrefs{
				0,
				&models.GeneralPrefs{
					Tag: models.All,
				},
			},
		},
		{
			Name:       "Unsuccessful preference update with bad request",
			StatusCode: http.StatusBadRequest,
			ReqBody: map[string]interface{}{
				"text_modified": 10,
			},
		},
		{
			Name:       "Unsuccessful preference update with non-existent resource",
			StatusCode: http.StatusNotFound,
			ReqBody: map[string]interface{}{
				"tag": models.Browser,
			},
			Error: mongo.ErrNoDocuments,
		},
	}

	for _, test := range tests {
		t.Run(test.Name, func(t *testing.T) {
			rBody, _ := json.Marshal(test.ReqBody)
			r := httptest.NewRequest("PATCH", "/api/prefs/conversations/1", bytes.NewReader(rBody))
			w := httptest.NewRecorder()

			mDB := &models.MockDB{PatchErr: test.Error}

			env := &Env{DB: mDB}
			env.PatchPrefsConvHandler(w, r)

			if w.Code != test.StatusCode {
				t.Errorf("Response has incorrect status code, expected status code %d, got %d", test.StatusCode, w.Code)
			}

			if w.Code == http.StatusOK {
				resBody := models.ConversationPrefs{}
				_ = json.NewDecoder(w.Body).Decode(&resBody)
				if !reflect.DeepEqual(test.ResBody, resBody) {
					t.Errorf("Response has incorrect body, expected %+v, got %+v", test.ResBody, resBody)
				}
			} else if w.Code == http.StatusNotFound &&
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
