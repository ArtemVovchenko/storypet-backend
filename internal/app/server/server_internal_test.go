package server

import (
	"bytes"
	"encoding/json"
	"fmt"
	"github.com/stretchr/testify/assert"
	"log"
	"net/http"
	"net/http/httptest"
	"testing"
)

var (
	s *Server
)

func TestMain(m *testing.M) {
	s = New()
	s.configureMiddleware()
	s.configureRouter()
	if err := s.configureStore(); err != nil {
		log.Fatalln("could not configure databaseStore")
	}
	m.Run()
}

func TestServer_handleTestRequest(t *testing.T) {
	rec := httptest.NewRecorder()
	req, _ := http.NewRequest(http.MethodGet, "/api/users/test", nil)
	s.handleTest().ServeHTTP(rec, req)
	assert.Equal(t, "Test OK", rec.Body.String())
}

func TestServer_handleLoginLifeCycle(t *testing.T) {
	var user_id int

	type registerUserID struct {
		UserID int `json:"user_id"`
	}

	testCases := []struct {
		name                 string
		payload              interface{}
		expectedResponseCode int
	}{
		{
			name: "Valid",
			payload: map[string]string{
				"account_email": "something@gmail.com",
				"backup_email":  "something.else@gmail.com",
				"full_name":     "John Smith",
				"password":      "SuPerPass123",
				"username":      "GoodEvil",
			},
			expectedResponseCode: http.StatusCreated,
		},

		{
			name:                 "Empty JSON",
			payload:              map[string]string{},
			expectedResponseCode: http.StatusUnprocessableEntity,
		},

		{
			name:                 "Empty JSON",
			payload:              "map[string]string{}",
			expectedResponseCode: http.StatusBadRequest,
		},

		{
			name: "Invalid Username",
			payload: map[string]string{
				"account_email": "something@gmail.com",
				"backup_email":  "something.else@gmail.com",
				"full_name":     "John Smith",
				"password":      "SuPerPass123",
				"username":      "_i",
			},
			expectedResponseCode: http.StatusUnprocessableEntity,
		},

		{
			name: "No Password",
			payload: map[string]string{
				"account_email": "something@gmail.com",
				"backup_email":  "something.else@gmail.com",
				"full_name":     "John Smith",
				"username":      "_i",
			},
			expectedResponseCode: http.StatusUnprocessableEntity,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			rec := httptest.NewRecorder()
			b := &bytes.Buffer{}
			_ = json.NewEncoder(b).Encode(tc.payload)
			req, _ := http.NewRequest(http.MethodPost, "/api/users/create", b)
			s.handleRegistration().ServeHTTP(rec, req)
			log.Printf("Response: %s", rec.Body)
			assert.Equal(t, tc.expectedResponseCode, rec.Code)
			if rec.Code == http.StatusCreated {
				r := &registerUserID{}
				_ = json.NewDecoder(rec.Body).Decode(r)
				user_id = r.UserID
			}
		})
	}

	var accessToken string
	//var refreshToken string

	type loginResponse struct {
		Access  string `json:"access"`
		Refresh string `json:"refresh"`
	}

	var lgr loginResponse

	loginTestCases := []struct {
		name                 string
		loginCredentials     interface{}
		expectedResponseCode int
	}{
		{
			name:                 "No Data",
			loginCredentials:     nil,
			expectedResponseCode: http.StatusUnauthorized,
		},
		{
			name: "No Password",
			loginCredentials: map[string]string{
				"email": "something@gmail.com",
			},
			expectedResponseCode: http.StatusUnauthorized,
		},
		{
			name: "No email",
			loginCredentials: map[string]string{
				"password": "SuPerPass15sdvs32t",
			},
			expectedResponseCode: http.StatusUnauthorized,
		},
		{
			name: "Invalid password",
			loginCredentials: map[string]string{
				"email":    "something@gmail.com",
				"password": "SuPerPass15sdvs32t",
			},
			expectedResponseCode: http.StatusUnauthorized,
		},
		{
			name: "Invalid email address",
			loginCredentials: map[string]string{
				"email":    "something@gmail.ua",
				"password": "SuPerPass123",
			},
			expectedResponseCode: http.StatusUnauthorized,
		},
		{
			name: "Valid",
			loginCredentials: map[string]string{
				"email":    "something@gmail.com",
				"password": "SuPerPass123",
			},
			expectedResponseCode: http.StatusOK,
		},
	}

	for _, tc := range loginTestCases {
		t.Run(tc.name, func(t *testing.T) {
			rec := httptest.NewRecorder()
			b := &bytes.Buffer{}
			_ = json.NewEncoder(b).Encode(tc.loginCredentials)
			req, _ := http.NewRequest(http.MethodPost, "/api/users/login", b)
			s.handleLogin().ServeHTTP(rec, req)
			log.Printf("Response: %s", rec.Body)
			assert.Equal(t, tc.expectedResponseCode, rec.Code)
			if rec.Code == http.StatusOK {
				_ = json.NewDecoder(rec.Body).Decode(&lgr)
				//refreshToken = lgr.Refresh
				accessToken = lgr.Access
			}
		})
	}

	testCasesAsLoggedIn := []struct {
		name                 string
		access               interface{}
		expectedResponseCode int
	}{
		{
			name:                 "Valid",
			access:               accessToken,
			expectedResponseCode: http.StatusOK,
		},
		{
			name:                 "No access  token",
			access:               nil,
			expectedResponseCode: http.StatusUnauthorized,
		},
		{
			name:                 "Invalid access  token",
			access:               accessToken + "er",
			expectedResponseCode: http.StatusUnauthorized,
		},
	}

	for _, tc := range testCasesAsLoggedIn {
		t.Run(tc.name, func(t *testing.T) {
			rec := httptest.NewRecorder()
			b := &bytes.Buffer{}
			req, _ := http.NewRequest(http.MethodGet, "/api/users/login", b)
			req.Header.Set("Authorization", fmt.Sprintf("Bear %v", tc.access))
			s.middleware.Authentication.IsAuthorised(s.handleTest()).ServeHTTP(rec, req)
			log.Printf("Response: %s", rec.Body)
			assert.Equal(t, tc.expectedResponseCode, rec.Code)
		})
	}

}
