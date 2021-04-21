package middleware

import (
	"net/http"
)

type ResponseWriterMiddleware struct {
	server server
}

func NewResponseWriterMiddleware(server server) *ResponseWriterMiddleware {
	return &ResponseWriterMiddleware{server: server}
}

func (r *ResponseWriterMiddleware) JSONBody(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		next(w, r)
	}
}
