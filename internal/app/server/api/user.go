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
	"github.com/lib/pq"
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

	sb.Path("/password").
		Name("Change Password").
		Methods(http.MethodPost).
		HandlerFunc(a.ServePasswordChangeRequest)

	sb.Path("/{id:[0-9]+}").
		Name("User By ID").
		Methods(http.MethodGet, http.MethodPut, http.MethodDelete).
		HandlerFunc(a.ServeRequestByID)

	sb.Path("/{id:[0-9]+}/role").
		Name("User Roles By ID").
		Methods(http.MethodGet, http.MethodPost, http.MethodDelete).
		HandlerFunc(a.ServeRoleRequest)

	sb.Path("/{id:[0-9]+}/clinic").
		Name("Veterinarian clinic By ID").
		Methods(http.MethodGet, http.MethodPost, http.MethodPut, http.MethodDelete).
		HandlerFunc(a.ServeClinicRequest)
}

func (a *UserAPI) ServeRegistrationRequest(w http.ResponseWriter, r *http.Request) {
	if r.Method == http.MethodOptions {
		return
	}
	switch r.Method {
	case http.MethodPost:
		requestUUID := r.Context().Value(middleware.CtxReqestUUID).(string)
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
			a.server.Logger().Printf("Database err: %v, Request ID: %s", requestUUID)
			a.server.RespondError(w, r, http.StatusUnprocessableEntity, err)
			return
		}
		u.Sanitise()
		a.server.Respond(w, r, http.StatusCreated, u)
	}
}

func (a *UserAPI) ServeRootRequest(w http.ResponseWriter, r *http.Request) {
	if r.Method == http.MethodOptions {
		return
	}
	switch r.Method {
	case http.MethodGet:
		requestID := r.Context().Value(middleware.CtxReqestUUID).(string)
		users, err := a.server.DatabaseStore().Users().SelectAll()
		if err != nil {
			a.server.Logger().Printf("Database err: %v, Request ID: %s", requestID)
			a.server.RespondError(w, r, http.StatusInternalServerError, nil)
			return
		}
		a.server.Respond(w, r, http.StatusOK, users)
	}
}

func (a *UserAPI) ServeRequestByID(w http.ResponseWriter, r *http.Request) {
	if r.Method == http.MethodOptions {
		return
	}
	requestID, session, err := a.server.GetAuthorizedRequestInfo(r)
	if err != nil {
		a.server.RespondError(w, r, http.StatusInternalServerError, nil)
		return
	}

	rawUserID, err := strconv.ParseInt(mux.Vars(r)["id"], 10, 64)
	if err != nil {
		a.server.RespondError(w, r, http.StatusBadRequest, exceptions.UnprocessableURIParam)
		return
	}
	requestedUserID := int(rawUserID)

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

func (a *UserAPI) ServeRoleRequest(w http.ResponseWriter, r *http.Request) {
	if r.Method == http.MethodOptions {
		return
	}
	requestID, session, err := a.server.GetAuthorizedRequestInfo(r)
	if err != nil {
		a.server.RespondError(w, r, http.StatusInternalServerError, nil)
		return
	}
	rawUserID, err := strconv.ParseInt(mux.Vars(r)["id"], 10, 64)
	if err != nil {
		a.server.RespondError(w, r, http.StatusBadRequest, exceptions.UnprocessableURIParam)
		return
	}
	requestedUserID := int(rawUserID)

	switch r.Method {
	case http.MethodGet:
		userRoles, err := a.server.DatabaseStore().Roles().SelectUserRoles(requestedUserID)
		if err != nil {
			if errors.Is(err, sql.ErrNoRows) {
				a.server.RespondError(w, r, http.StatusNotFound, exceptions.RequestedUserNotFound)
				return
			}
			a.server.Logger().Println(err)
			a.server.RespondError(w, r, http.StatusInternalServerError, nil)
			return
		}
		a.server.Respond(w, r, http.StatusOK, userRoles)

	case http.MethodPost:
		type requestBody struct {
			RoleID int `json:"role_id"`
		}
		type responseBody struct {
			User  *models.User  `json:"user"`
			Roles []models.Role `json:"roles"`
		}

		if !permissions.AnyRoleHavePermissions(
			session.Roles, permissions.Permissions().RolesPermission,
			permissions.Permissions().UsersPermission,
		) {
			a.server.RespondError(w, r, http.StatusForbidden, nil)
			return
		}

		rb := &requestBody{}
		if err := json.NewDecoder(r.Body).Decode(rb); err != nil {
			a.server.RespondError(w, r, http.StatusUnprocessableEntity, err)
			return
		}
		if err := a.server.DatabaseStore().Users().AssignRole(requestedUserID, rb.RoleID); err != nil {
			a.server.Logger().Printf("Database err: %v, Request ID: %s", requestID)
			a.server.RespondError(w, r, http.StatusUnprocessableEntity, nil)
			return
		}
		userModel, err := a.server.DatabaseStore().Users().FindByID(requestedUserID)
		if err != nil {
			a.server.Logger().Printf("Database err: %v, Request ID: %s", requestID)
			a.server.RespondError(w, r, http.StatusInternalServerError, nil)
			return
		}
		userRoles, err := a.server.DatabaseStore().Roles().SelectUserRoles(requestedUserID)
		if err != nil {
			a.server.Logger().Printf("Database err: %v, Request ID: %s", requestID)
			a.server.RespondError(w, r, http.StatusInternalServerError, nil)
			return
		}
		a.server.Respond(w, r, http.StatusOK, &responseBody{User: userModel, Roles: userRoles})

	case http.MethodDelete:
		type requestBody struct {
			RoleID int `json:"role_id"`
		}
		type responseBody struct {
			User  *models.User  `json:"user"`
			Roles []models.Role `json:"roles"`
		}

		if !permissions.AnyRoleHavePermissions(
			session.Roles, permissions.Permissions().RolesPermission,
			permissions.Permissions().UsersPermission,
		) {
			a.server.RespondError(w, r, http.StatusForbidden, nil)
			return
		}

		rb := &requestBody{}
		if err := json.NewDecoder(r.Body).Decode(rb); err != nil {
			a.server.RespondError(w, r, http.StatusUnprocessableEntity, err)
			return
		}
		if err := a.server.DatabaseStore().Users().DeleteRole(requestedUserID, rb.RoleID); err != nil {
			a.server.Logger().Printf("Database err: %v, Request ID: %s", requestID)
			a.server.RespondError(w, r, http.StatusUnprocessableEntity, nil)
			return
		}
		userModel, err := a.server.DatabaseStore().Users().FindByID(requestedUserID)
		if err != nil {
			a.server.Logger().Printf("Database err: %v, Request ID: %s", requestID)
			a.server.RespondError(w, r, http.StatusInternalServerError, nil)
			return
		}
		userRoles, err := a.server.DatabaseStore().Roles().SelectUserRoles(requestedUserID)
		if err != nil {
			a.server.Logger().Printf("Database err: %v, Request ID: %s", requestID)
			a.server.RespondError(w, r, http.StatusInternalServerError, nil)
			return
		}
		a.server.Respond(w, r, http.StatusOK, &responseBody{User: userModel, Roles: userRoles})

	}
}

func (a *UserAPI) ServePasswordChangeRequest(w http.ResponseWriter, r *http.Request) {
	if r.Method == http.MethodOptions {
		return
	}
	switch r.Method {
	case http.MethodPost:
		type requestBody struct {
			OldPassword string `json:"old_password"`
			NewPassword string `json:"new_password"`
		}

		requestID, session, err := a.server.GetAuthorizedRequestInfo(r)
		if err != nil {
			a.server.RespondError(w, r, http.StatusInternalServerError, nil)
			return
		}

		rb := &requestBody{}
		if err := json.NewDecoder(r.Body).Decode(rb); err != nil {
			a.server.RespondError(w, r, http.StatusBadRequest, err)
			return
		}

		userModel, err := a.server.DatabaseStore().Users().FindByID(session.UserID)
		if err != nil {
			a.server.Logger().Printf("Database err: %v, Request ID: %s", requestID)
			a.server.RespondError(w, r, http.StatusInternalServerError, nil)
			return
		}
		if !userModel.ComparePasswords(rb.OldPassword) {
			a.server.RespondError(w, r, http.StatusForbidden, exceptions.IncorrectOldPassword)
			return
		}
		if err := a.server.DatabaseStore().Users().ChangePassword(userModel.UserID, rb.NewPassword); err != nil {
			var pqErr pq.Error
			if errors.As(err, &pqErr) {
				a.server.Logger().Printf("Database err: %v, Request ID: %s", requestID)
				a.server.RespondError(w, r, http.StatusInternalServerError, nil)
				return
			}
			a.server.RespondError(w, r, http.StatusUnprocessableEntity, err)
			return
		}
		a.server.Respond(w, r, http.StatusOK, nil)
	}
}

func (a *UserAPI) ServeClinicRequest(w http.ResponseWriter, r *http.Request) {
	if r.Method == http.MethodOptions {
		return
	}
	requestID, session, err := a.server.GetAuthorizedRequestInfo(r)
	if err != nil {
		a.server.RespondError(w, r, http.StatusInternalServerError, nil)
		return
	}

	rawUserID, err := strconv.ParseInt(mux.Vars(r)["id"], 10, 64)
	if err != nil {
		a.server.RespondError(w, r, http.StatusBadRequest, exceptions.UnprocessableURIParam)
		return
	}
	requestedUserID := int(rawUserID)

	switch r.Method {
	case http.MethodGet:
		vetClinic, err := a.server.DatabaseStore().Users().SelectClinicByUserID(requestedUserID)
		if err != nil {
			if errors.Is(err, sql.ErrNoRows) {
				a.server.RespondError(w, r, http.StatusNotFound, nil)
				return
			}
			a.server.Logger().Printf("Database err: %v, Request ID: %s", requestID)
			a.server.RespondError(w, r, http.StatusInternalServerError, nil)
			return
		}
		a.server.Respond(w, r, http.StatusOK, vetClinic)

	case http.MethodPost:
		type requestBody struct {
			ClinicId   string `json:"clinic_id"`
			ClinicName string `json:"clinic_name"`
		}
		if !permissions.AnyRoleHavePermissions(session.Roles, permissions.Permissions().VeterinariansPermission) {
			a.server.RespondError(w, r, http.StatusForbidden, nil)
			return
		}
		requestedUserRoles, err := a.server.DatabaseStore().Roles().SelectUserRoles(requestedUserID)
		if err != nil {
			if errors.Is(err, sql.ErrNoRows) {
				a.server.RespondError(w, r, http.StatusNotFound, nil)
				return
			}
			a.server.Logger().Printf("Database err: %v, Request ID: %s", requestID)
			a.server.RespondError(w, r, http.StatusInternalServerError, nil)
			return
		}
		if !permissions.AnyRoleIsVeterinarian(requestedUserRoles) {
			a.server.RespondError(w, r, http.StatusBadRequest, exceptions.UserIsNotVeterinarian)
			return
		}
		if _, err := a.server.DatabaseStore().Users().SelectClinicByUserID(requestedUserID); !errors.Is(err, sql.ErrNoRows) {
			a.server.RespondError(w, r, http.StatusBadRequest, exceptions.RecordAlreadyExist)
			return
		}
		rb := &requestBody{}
		if err := json.NewDecoder(r.Body).Decode(rb); err != nil {
			a.server.RespondError(w, r, http.StatusBadRequest, err)
			return
		}
		newVetModel := &models.VetClinic{
			UserID:     requestedUserID,
			ClinicID:   rb.ClinicId,
			ClinicName: rb.ClinicName,
		}
		if _, err := a.server.DatabaseStore().Users().CreateClinic(newVetModel); err != nil {
			a.server.Logger().Printf("Database err: %v, Request ID: %s", requestID)
			a.server.RespondError(w, r, http.StatusInternalServerError, nil)
			return
		}
		a.server.Respond(w, r, http.StatusCreated, nil)

	case http.MethodPut:
		type requestBody struct {
			ClinicId   string `json:"clinic_id"`
			ClinicName string `json:"clinic_name"`
		}
		if !permissions.AnyRoleHavePermissions(session.Roles, permissions.Permissions().VeterinariansPermission) {
			a.server.RespondError(w, r, http.StatusForbidden, nil)
			return
		}
		rb := &requestBody{}
		if err := json.NewDecoder(r.Body).Decode(rb); err != nil {
			a.server.RespondError(w, r, http.StatusBadRequest, err)
			return
		}
		newVetModel := &models.VetClinic{
			UserID:     requestedUserID,
			ClinicID:   rb.ClinicId,
			ClinicName: rb.ClinicName,
		}
		newVetModel, err = a.server.DatabaseStore().Users().UpdateClinic(newVetModel)
		if err != nil {
			if errors.Is(err, sql.ErrNoRows) {
				a.server.RespondError(w, r, http.StatusNotFound, nil)
				return
			}
			a.server.Logger().Printf("Database err: %v, Request ID: %s", requestID)
			a.server.RespondError(w, r, http.StatusInternalServerError, nil)
			return
		}
		a.server.Respond(w, r, http.StatusOK, newVetModel)

	case http.MethodDelete:
		if !permissions.AnyRoleHavePermissions(session.Roles, permissions.Permissions().VeterinariansPermission) {
			a.server.RespondError(w, r, http.StatusForbidden, nil)
			return
		}
		if _, err := a.server.DatabaseStore().Users().DeleteClinic(requestedUserID); err != nil {
			if errors.Is(err, sql.ErrNoRows) {
				a.server.RespondError(w, r, http.StatusNotFound, nil)
				return
			}
			a.server.Logger().Printf("Database err: %v, Request ID: %s", requestID)
			a.server.RespondError(w, r, http.StatusInternalServerError, nil)
			return
		}
		a.server.RespondError(w, r, http.StatusNoContent, nil)
	}
}
