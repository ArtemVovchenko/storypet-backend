package middleware

import (
	"github.com/ArtemVovchenko/storypet-backend/internal/app/store"
	"net/http"
)

type Middleware struct {
	Authentication *AuthenticationMiddleware
}

type server interface {
	Respond(w http.ResponseWriter, h *http.Request, code int, data interface{})
	PersistentStore() store.PersistentStore
}

func New(server server) *Middleware {
	return &Middleware{
		Authentication: newAuthentication(server),
	}
}
