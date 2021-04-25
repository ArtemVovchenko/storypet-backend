package middleware

import (
	"errors"
	"github.com/ArtemVovchenko/storypet-backend/internal/app/permissions"
	"github.com/ArtemVovchenko/storypet-backend/internal/pkg/auth"
	"golang.org/x/net/context"
	"net/http"
)

type CtxKeys int

type AccessPermissionMiddleware struct {
	server server
}

const (
	CtxUserRoles = iota
)

var errAccessDenied = errors.New("access denied")

func NewAccessPermissionMiddleware(server server) *AccessPermissionMiddleware {
	return &AccessPermissionMiddleware{server: server}
}

func (m *AccessPermissionMiddleware) FullAccess(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		accessMeta, err := auth.ExtractAccessMeta(r)
		if err != nil {
			m.server.RespondError(w, r, http.StatusForbidden, errAccessDenied)
			return
		}
		session, err := m.server.PersistentStore().GetSessionInfo(accessMeta.AccessUUID)
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
		accessMeta, err := auth.ExtractAccessMeta(r)
		if err != nil {
			m.server.RespondError(w, r, http.StatusForbidden, errAccessDenied)
			return
		}
		session, err := m.server.PersistentStore().GetSessionInfo(accessMeta.AccessUUID)
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

func (m *AccessPermissionMiddleware) AddRolesToContext(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		accessMeta, err := auth.ExtractAccessMeta(r)
		if err != nil {
			m.server.RespondError(w, r, http.StatusForbidden, errAccessDenied)
			return
		}
		session, err := m.server.PersistentStore().GetSessionInfo(accessMeta.AccessUUID)
		if err != nil {
			m.server.RespondError(w, r, http.StatusForbidden, errAccessDenied)
			return
		}

		next(w, r.WithContext(context.WithValue(r.Context(), CtxUserRoles, session.Roles)))
	}
}
