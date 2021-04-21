package middleware

import (
	"github.com/ArtemVovchenko/storypet-backend/internal/app/store"
	"net/http"
)

type Middleware struct {
	Authentication   *AuthenticationMiddleware
	ResponseWriting  *ResponseWriterMiddleware
	AccessPermission *AccessPermissionMiddleware
}

type server interface {
	Respond(w http.ResponseWriter, h *http.Request, code int, data interface{})
	RespondError(w http.ResponseWriter, h *http.Request, code int, err error)
	PersistentStore() store.PersistentStore
}

func New(server server) *Middleware {
	return &Middleware{
		Authentication:   newAuthentication(server),
		ResponseWriting:  NewResponseWriterMiddleware(server),
		AccessPermission: NewAccessPermissionMiddleware(server),
	}
}
