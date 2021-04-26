package api

import (
	"database/sql"
	"encoding/json"
	"errors"
	"github.com/ArtemVovchenko/storypet-backend/internal/app/models"
	"github.com/ArtemVovchenko/storypet-backend/internal/app/server/api/exceptions"
	"github.com/gorilla/mux"
	"net/http"
	"strconv"
)

type RolesAPI struct {
	server server
}

func NewRolesAPI(server server) *RolesAPI {
	return &RolesAPI{server: server}
}

func (a *RolesAPI) ConfigureRouter(router *mux.Router) {
	sb := router.PathPrefix("/api/roles").Subrouter()

	sb.Use(a.server.Middleware().Authentication.IsAuthorised)
	sb.Use(a.server.Middleware().AccessPermission.RolesAccess)

	sb.Path("").
		Name("Roles Root").
		Methods(http.MethodGet, http.MethodPost, http.MethodOptions).
		HandlerFunc(a.ServeRootRequest)

	sb.Path("").
		Name("Roles by ID").
		Methods(http.MethodGet, http.MethodPut, http.MethodDelete, http.MethodOptions).
		HandlerFunc(a.ServeRootRequest)
}

func (a *RolesAPI) ServeRootRequest(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodGet:
		roleModels, err := a.server.DatabaseStore().Roles().SelectAll()
		if err != nil {
			a.server.Logger().Println(err)
			a.server.RespondError(w, r, http.StatusInternalServerError, nil)
			return
		}
		a.server.Respond(w, r, http.StatusOK, roleModels)

	case http.MethodPost:
		type requestBody struct {
			RoleName               string  `json:"role_name"`
			RoleDescription        *string `json:"role_description"`
			AllowRolesCrud         bool    `json:"allow_roles_crud"`
			AllowUsersCrud         bool    `json:"allow_users_crud"`
			AllowVeterinariansCrud bool    `json:"allow_veterinarians_crud"`
			AllowVaccinesCrud      bool    `json:"allow_vaccines_crud"`
			AllowFoodCrud          bool    `json:"allow_food_crud"`
			AllowPetsCrud          bool    `json:"allow_pets_crud"`
			AllowDatabaseDump      bool    `json:"allow_database_dump"`
		}
		rb := &requestBody{}
		if err := json.NewDecoder(r.Body).Decode(rb); err != nil {
			a.server.RespondError(w, r, http.StatusBadRequest, err)
			return
		}
		roleModel := &models.Role{
			RoleName:               rb.RoleName,
			AllowRolesCrud:         rb.AllowRolesCrud,
			AllowUsersCrud:         rb.AllowUsersCrud,
			AllowVeterinariansCrud: rb.AllowVeterinariansCrud,
			AllowVaccinesCrud:      rb.AllowVaccinesCrud,
			AllowFoodCrud:          rb.AllowFoodCrud,
			AllowPetsCrud:          rb.AllowPetsCrud,
			AllowDatabaseDump:      rb.AllowDatabaseDump,
		}
		roleModel.SetDescription(rb.RoleDescription)
		roleModel.BeforeCreate()
		roleModel, err := a.server.DatabaseStore().Roles().Create(roleModel)
		if err != nil {
			a.server.Logger().Println(err)
			a.server.RespondError(w, r, http.StatusUnprocessableEntity, nil)
			return
		}
		a.server.Respond(w, r, http.StatusCreated, roleModel)
	}
}

func (a *RolesAPI) ServeIDRequest(w http.ResponseWriter, r *http.Request) {
	rawID, err := strconv.ParseInt(mux.Vars(r)["id"], 10, 64)
	if err != nil {
		a.server.RespondError(w, r, http.StatusBadRequest, exceptions.UnprocessableURIParam)
		return
	}
	roleID := int(rawID)

	switch r.Method {
	case http.MethodGet:
		roleModel, err := a.server.DatabaseStore().Roles().FindByID(roleID)
		if err != nil {
			if errors.Is(err, sql.ErrNoRows) {
				a.server.RespondError(w, r, http.StatusNotFound, nil)
				return
			}
			a.server.Logger().Println(err)
			a.server.RespondError(w, r, http.StatusInternalServerError, nil)
			return
		}
		a.server.Respond(w, r, http.StatusOK, roleModel)

	case http.MethodPut:
		type requestBody struct {
			RoleName               string  `json:"role_name"`
			RoleDescription        *string `json:"role_description"`
			AllowRolesCrud         bool    `json:"allow_roles_crud"`
			AllowUsersCrud         bool    `json:"allow_users_crud"`
			AllowVeterinariansCrud bool    `json:"allow_veterinarians_crud"`
			AllowVaccinesCrud      bool    `json:"allow_vaccines_crud"`
			AllowFoodCrud          bool    `json:"allow_food_crud"`
			AllowPetsCrud          bool    `json:"allow_pets_crud"`
			AllowDatabaseDump      bool    `json:"allow_database_dump"`
		}
		rb := &requestBody{}
		if err := json.NewDecoder(r.Body).Decode(rb); err != nil {
			a.server.RespondError(w, r, http.StatusBadRequest, err)
			return
		}
		roleModel := &models.Role{
			RoleID:                 roleID,
			RoleName:               rb.RoleName,
			AllowRolesCrud:         rb.AllowRolesCrud,
			AllowUsersCrud:         rb.AllowUsersCrud,
			AllowVeterinariansCrud: rb.AllowVeterinariansCrud,
			AllowVaccinesCrud:      rb.AllowVaccinesCrud,
			AllowFoodCrud:          rb.AllowFoodCrud,
			AllowPetsCrud:          rb.AllowPetsCrud,
			AllowDatabaseDump:      rb.AllowDatabaseDump,
		}
		roleModel.SetDescription(rb.RoleDescription)
		roleModel.BeforeCreate()
		newModel, err := a.server.DatabaseStore().Roles().Update(roleModel)
		if err != nil {
			a.server.Logger().Println(err)
			a.server.RespondError(w, r, http.StatusUnprocessableEntity, nil)
			return
		}
		a.server.Respond(w, r, http.StatusOK, newModel)

	case http.MethodDelete:
		_, err := a.server.DatabaseStore().Roles().DeleteByID(roleID)
		if err != nil {
			if errors.Is(err, sql.ErrNoRows) {
				a.server.RespondError(w, r, http.StatusNotFound, nil)
				return
			}
			a.server.Logger().Println(err)
			a.server.RespondError(w, r, http.StatusInternalServerError, nil)
			return
		}
		a.server.Respond(w, r, http.StatusNoContent, nil)
	}
}
