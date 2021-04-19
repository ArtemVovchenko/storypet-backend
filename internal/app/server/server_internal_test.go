package server

import (
	"bytes"
	"encoding/json"
	"github.com/ArtemVovchenko/storypet-backend/internal/app/configs"
	"github.com/stretchr/testify/assert"
	"log"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestServer_HandleTestRequest(t *testing.T) {
	s := New(configs.NewServerConfig())
	rec := httptest.NewRecorder()
	req, _ := http.NewRequest(http.MethodGet, "/api/users/test", nil)
	s.handleTest().ServeHTTP(rec, req)
	assert.Equal(t, "Test OK", rec.Body.String())
}

func TestServer_HandleRegisterRequest(t *testing.T) {
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
		s := New(configs.NewServerConfig())
		err := s.configureStore()
		assert.NoError(t, err)
		t.Run(tc.name, func(t *testing.T) {
			rec := httptest.NewRecorder()
			b := &bytes.Buffer{}
			_ = json.NewEncoder(b).Encode(tc.payload)
			req, _ := http.NewRequest(http.MethodGet, "/api/users/create", b)
			s.handleRegistration().ServeHTTP(rec, req)
			log.Printf("Response: %s", rec.Body)
			assert.Equal(t, tc.expectedResponseCode, rec.Code)
		})
	}
}
