package api

import (
	"encoding/json"
	"github.com/ArtemVovchenko/storypet-backend/internal/app/middleware"
	"github.com/ArtemVovchenko/storypet-backend/internal/app/models"
	"github.com/gorilla/mux"
	"net/http"
)

type UserAPI struct {
	server server
}

func NewUserAPI(server server) *UserAPI {
	return &UserAPI{server: server}
}

func (a *UserAPI) ConfigureRoutes(router *mux.Router) {
	router.Path("/api/users").
		Name("User Register").
		Methods(http.MethodGet, http.MethodPost).
		HandlerFunc(
			a.ServeRegistrationRequest,
		)
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
			a.server.Logger().Printf("Reqest ID: %s database error: %v", requestUUID, err)
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
