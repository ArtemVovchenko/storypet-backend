package middleware

import (
	"github.com/ArtemVovchenko/storypet-backend/internal/pkg/auth"
	"net/http"
	"strconv"
)

type AuthenticationMiddleware struct {
	server server
}

func newAuthentication(server server) *AuthenticationMiddleware {
	return &AuthenticationMiddleware{server: server}
}

func (m *AuthenticationMiddleware) IsAuthorised(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		accessInfo, err := auth.ExtractAccessMeta(r)
		if err != nil {
			m.server.Respond(w, r, http.StatusUnauthorized, "Unauthorized")
			return
		}
		redisInternalUserID, err := m.server.RedisStorage().Get(accessInfo.AccessUUID).Result()
		if err != nil {
			m.server.Respond(w, r, http.StatusUnauthorized, "Unauthorized")
			return
		}
		verifiedID, err := strconv.ParseInt(redisInternalUserID, 10, 64)
		if err != nil {
			m.server.Respond(w, r, http.StatusUnauthorized, "Unauthorized")
			return
		}
		if accessInfo.UserID != int(verifiedID) {
			m.server.Respond(w, r, http.StatusUnauthorized, "Unauthorized")
			return
		}
		next(w, r)
	}
}
