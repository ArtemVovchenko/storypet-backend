package middleware

import (
	"github.com/go-redis/redis/v7"
	"net/http"
)

type Middleware struct {
	Authentication *AuthenticationMiddleware
}

type server interface {
	Respond(w http.ResponseWriter, h *http.Request, code int, data interface{})
	RedisStorage() *redis.Client
}

func New(server server) *Middleware {
	return &Middleware{
		Authentication: newAuthentication(server),
	}
}
