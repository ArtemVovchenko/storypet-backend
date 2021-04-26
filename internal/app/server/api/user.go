package api

import (
	"database/sql"
	"encoding/json"
	"errors"
	"github.com/ArtemVovchenko/storypet-backend/internal/app/middleware"
	"github.com/ArtemVovchenko/storypet-backend/internal/app/models"
	"github.com/ArtemVovchenko/storypet-backend/internal/app/permissions"
	"github.com/ArtemVovchenko/storypet-backend/internal/app/server/api/exceptions"
	"github.com/gorilla/mux"
	"net/http"
	"strconv"
)

type UserAPI struct {
	server server
}

func NewUserAPI(server server) *UserAPI {
	return &UserAPI{server: server}
}

func (a *UserAPI) ConfigureRoutes(router *mux.Router) {
	router.Path("/api/register").
		Name("User Register").
		Methods(http.MethodPost).
		HandlerFunc(a.ServeRegistrationRequest)

	sb := router.PathPrefix("/api/users").Subrouter()
	sb.Use(a.server.Middleware().Authentication.IsAuthorised)

	sb.Path("").
		Name("Get All users").
		Methods(http.MethodGet).
		HandlerFunc(a.ServeRootRequest)

	sb.Path("/{id:[0-9]+}").
		Name("User By ID").
		Methods(http.MethodGet, http.MethodPut, http.MethodDelete).
		HandlerFunc(a.ServeRequestByID)
}

func (a *UserAPI) ServeRegistrationRequest(w http.ResponseWriter, r *http.Request) {
	requestUUID := r.Context().Value(middleware.CtxReqestUUID).(string)

	switch r.Method {

	case http.MethodGet:
		if r.Context().Value(middleware.CtxAccessUUID) == nil {
			a.server.RespondError(w, r, http.StatusUnauthorized, nil)
			return
		}
		userModels, err := a.server.DatabaseStore().Users().SelectAll()
		if err != nil {
			a.server.Logger().Printf("Request ID: %s database error: %v", requestUUID, err)
			a.server.RespondError(w, r, http.StatusInternalServerError, nil)
			return
		}
		a.server.Respond(w, r, http.StatusOK, userModels)

	case http.MethodPost:
		type requestBody struct {
			AccountEmail string  `json:"account_email"`
			Password     string  `json:"password"`
			Username     string  `json:"username"`
			FullName     string  `json:"full_name"`
			BackupEmail  *string `json:"backup_email"`
			Location     *string `json:"location"`
		}

		rb := &requestBody{}
		if err := json.NewDecoder(r.Body).Decode(rb); err != nil {
			a.server.RespondError(w, r, http.StatusBadRequest, err)
			return
		}

		u := &models.User{
			AccountEmail: rb.AccountEmail,
			Password:     rb.Password,
			Username:     rb.Username,
			FullName:     rb.FullName,
		}
		u.SetBackupEmail(rb.BackupEmail)
		u.SetLocation(rb.Location)

		if _, err := a.server.DatabaseStore().Users().Create(u); err != nil {
			a.server.RespondError(w, r, http.StatusUnprocessableEntity, err)
			return
		}
		u.Sanitise()
		a.server.Respond(w, r, http.StatusCreated, u)
	}
}

func (a *UserAPI) ServeRootRequest(w http.ResponseWriter, r *http.Request) {
	requestID := r.Context().Value(middleware.CtxReqestUUID).(string)
	users, err := a.server.DatabaseStore().Users().SelectAll()
	if err != nil {
		a.server.Logger().Printf("Database err: %v, Request ID: %s", requestID)
		a.server.RespondError(w, r, http.StatusInternalServerError, nil)
		return
	}
	a.server.Respond(w, r, http.StatusOK, users)
}

func (a *UserAPI) ServeRequestByID(w http.ResponseWriter, r *http.Request) {
	requestID := r.Context().Value(middleware.CtxReqestUUID).(string)
	accessID := r.Context().Value(middleware.CtxAccessUUID).(string)
	rawUserID, err := strconv.ParseInt(mux.Vars(r)["id"], 10, 64)
	if err != nil {
		a.server.RespondError(w, r, http.StatusInternalServerError, exceptions.UnprocessableURIParam)
	}
	requestedUserID := int(rawUserID)

	session, err := a.server.PersistentStore().GetSessionInfo(accessID)
	if err != nil {
		a.server.Logger().Printf("Persistent database err: %v, Request ID: %s", requestID)
		a.server.RespondError(w, r, http.StatusInternalServerError, nil)
		return
	}

	switch r.Method {
	case http.MethodGet:
		user, err := a.server.DatabaseStore().Users().FindByID(requestedUserID)
		if err != nil {
			if errors.Is(err, sql.ErrNoRows) {
				a.server.RespondError(w, r, http.StatusNotFound, nil)
				return
			}
			a.server.Logger().Printf("Database err: %v, Request ID: %s", requestID)
			a.server.RespondError(w, r, http.StatusInternalServerError, nil)
			return
		}
		a.server.Respond(w, r, http.StatusOK, user)

	case http.MethodPut:
		type requestBody struct {
			UserID               int     `db:"user_id" json:"user_id"`
			AccountEmail         string  `db:"account_email" json:"account_email"`
			PasswordSHA256       string  `db:"password_sha256" json:"-"`
			Username             string  `db:"username" json:"username"`
			FullName             string  `db:"full_name" json:"full_name"`
			SpecifiedBackupEmail *string `json:"backup_email"`
			SpecifiedLocation    *string `json:"location"`
		}

		rb := &requestBody{}
		if err := json.NewDecoder(r.Body).Decode(rb); err != nil {
			a.server.RespondError(w, r, http.StatusBadRequest, err)
			return
		}

		if requestedUserID != session.UserID {
			if !permissions.AnyRoleHavePermissions(session.Roles, permissions.Permissions().UsersPermission) {
				a.server.RespondError(w, r, http.StatusForbidden, nil)
				return
			}
		}

		userModel := &models.User{
			UserID:       requestedUserID,
			AccountEmail: rb.AccountEmail,
			Username:     rb.Username,
			FullName:     rb.PasswordSHA256,
		}
		userModel.SetBackupEmail(rb.SpecifiedBackupEmail)
		userModel.SetLocation(rb.SpecifiedLocation)

		newModel, err := a.server.DatabaseStore().Users().Update(userModel)
		if err != nil {
			a.server.Logger().Printf("Persistent database err: %v, Request ID: %s", requestID)
			a.server.RespondError(w, r, http.StatusUnprocessableEntity, err)
			return
		}

		a.server.Respond(w, r, http.StatusOK, newModel)

	case http.MethodDelete:
		if requestedUserID != session.UserID {
			if !permissions.AnyRoleHavePermissions(session.Roles, permissions.Permissions().UsersPermission) {
				a.server.RespondError(w, r, http.StatusForbidden, nil)
				return
			}
		}

		if _, err := a.server.DatabaseStore().Users().DeleteByID(requestedUserID); err != nil {
			a.server.Logger().Printf("Database err: %v, Request ID: %s", requestID)
			a.server.RespondError(w, r, http.StatusInternalServerError, err)
			return
		}

		a.server.Respond(w, r, http.StatusNoContent, nil)
	}
}
