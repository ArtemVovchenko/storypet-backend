package middleware

import (
	"github.com/ArtemVovchenko/storypet-backend/internal/app/store"
	"log"
	"net/http"
)

type CtxKeys int

const (
	CtxRequestUUID CtxKeys = iota
	CtxAccessUUID
)

const (
	hdrReqUUID = "X-Request-ID"
)

type Middleware struct {
	Authentication   *AuthenticationMiddleware
	ResponseWriting  *ResponseWriterMiddleware
	AccessPermission *AccessPermissionMiddleware
	InfoMiddleware   *InfoMiddleware
}

type server interface {
	Respond(w http.ResponseWriter, h *http.Request, code int, data interface{})
	RespondError(w http.ResponseWriter, h *http.Request, code int, err error)
	Logger() *log.Logger
	PersistentStore() store.PersistentStore
}

func New(server server) Middleware {
	return Middleware{
		Authentication:   newAuthentication(server),
		ResponseWriting:  NewResponseWriterMiddleware(server),
		AccessPermission: NewAccessPermissionMiddleware(server),
		InfoMiddleware:   NewInfoMiddleware(server),
	}
}
