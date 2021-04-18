package server

import (
	"github.com/stretchr/testify/assert"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestServer_HandleTestRequest(t *testing.T) {
	s := New(NewConfig())
	rec := httptest.NewRecorder()
	req, _ := http.NewRequest(http.MethodGet, "/api/users/test", nil)
	s.handleTest().ServeHTTP(rec, req)
	assert.Equal(t, "Test OK", rec.Body.String())
}
