package middleware

import (
	"github.com/ArtemVovchenko/storypet-backend/internal/pkg/auth"
	"net/http"
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
		session, err := m.server.PersistentStore().GetSessionInfo(accessInfo.AccessUUID)
		if err != nil {
			m.server.Respond(w, r, http.StatusUnauthorized, "Unauthorized")
			return
		}
		if accessInfo.UserID != session.UserID {
			m.server.Respond(w, r, http.StatusUnauthorized, "Unauthorized")
			return
		}
		next(w, r)
	}
}
