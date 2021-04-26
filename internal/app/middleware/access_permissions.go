package middleware

import (
	"errors"
	"github.com/ArtemVovchenko/storypet-backend/internal/app/permissions"
	"net/http"
)

type AccessPermissionMiddleware struct {
	server server
}

var errAccessDenied = errors.New("access denied")

func NewAccessPermissionMiddleware(server server) *AccessPermissionMiddleware {
	return &AccessPermissionMiddleware{server: server}
}

func (m *AccessPermissionMiddleware) FullAccess(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		accessUUID := r.Context().Value(CtxAccessUUID).(string)
		session, err := m.server.PersistentStore().GetSessionInfo(accessUUID)
		if err != nil {
			m.server.RespondError(w, r, http.StatusForbidden, errAccessDenied)
			return
		}
		if !permissions.AnyRoleHavePermissions(session.Roles, permissions.All()...) {
			m.server.RespondError(w, r, http.StatusForbidden, errAccessDenied)
			return
		}
		next(w, r)
	}
}

func (m *AccessPermissionMiddleware) DatabaseAccess(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		accessUUID := r.Context().Value(CtxAccessUUID).(string)
		session, err := m.server.PersistentStore().GetSessionInfo(accessUUID)
		if err != nil {
			m.server.RespondError(w, r, http.StatusForbidden, errAccessDenied)
			return
		}
		if !permissions.AnyRoleHavePermissions(session.Roles, permissions.Permissions().DatabasePermission) {
			m.server.RespondError(w, r, http.StatusForbidden, errAccessDenied)
			return
		}
		next(w, r)
	}
}
