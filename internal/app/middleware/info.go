package middleware

import (
	"github.com/twinj/uuid"
	"golang.org/x/net/context"
	"net/http"
	"time"
)

type InfoMiddleware struct {
	server server
}

type responseWriter struct {
	http.ResponseWriter
	statusCode int
}

func (r *responseWriter) WriteHeader(statusCode int) {
	r.statusCode = statusCode
	r.ResponseWriter.WriteHeader(statusCode)
}

func NewInfoMiddleware(server server) *InfoMiddleware {
	return &InfoMiddleware{server: server}
}

func (m *InfoMiddleware) MarkRequest(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		requestUUID := uuid.NewV4().String()
		r.Header.Set(hdrReqUUID, requestUUID)
		next.ServeHTTP(w, r.WithContext(context.WithValue(r.Context(), CtxReqestUUID, requestUUID)))
	})
}

func (m *InfoMiddleware) LogRequest(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		m.server.Logger().Printf("Request %s %s : ID: %s", r.Method, r.RequestURI, r.Header.Get(hdrReqUUID))
		rw := &responseWriter{w, http.StatusOK}

		start := time.Now()
		next.ServeHTTP(rw, r)
		m.server.Logger().Printf("Request %s status=%v %v completed in %v",
			r.Header.Get(hdrReqUUID),
			rw.statusCode,
			http.StatusText(rw.statusCode), time.Now().Sub(start).String(),
		)
	})
}
