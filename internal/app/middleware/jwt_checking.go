package middleware

import (
	"errors"
	"github.com/ArtemVovchenko/storypet-backend/internal/pkg/auth"
	"golang.org/x/net/context"
	"net/http"
)

type AuthenticationMiddleware struct {
	server server
}

func newAuthentication(server server) *AuthenticationMiddleware {
	return &AuthenticationMiddleware{server: server}
}

var errUnauthorized = errors.New("unauthorized")

func (m *AuthenticationMiddleware) IsAuthorised(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		accessInfo, err := auth.ExtractAccessMeta(r)
		if err != nil {
			m.server.RespondError(w, r, http.StatusUnauthorized, errUnauthorized)
			return
		}
		session, err := m.server.PersistentStore().GetSessionInfo(accessInfo.AccessUUID)
		if err != nil {
			m.server.RespondError(w, r, http.StatusUnauthorized, errUnauthorized)
			return
		}
		if accessInfo.UserID != session.UserID {
			m.server.RespondError(w, r, http.StatusUnauthorized, errUnauthorized)
			return
		}
		next.ServeHTTP(w, r.WithContext(context.WithValue(r.Context(), CtxAccessUUID, accessInfo.AccessUUID)))
	})
}
