package api

import (
	"database/sql"
	"encoding/json"
	"errors"
	"github.com/ArtemVovchenko/storypet-backend/internal/app/models"
	"github.com/ArtemVovchenko/storypet-backend/internal/app/permissions"
	"github.com/ArtemVovchenko/storypet-backend/internal/app/server/api/exceptions"
	"github.com/gorilla/mux"
	"net/http"
	"strconv"
)

type FoodsAPI struct {
	server server
}

func NewFoodsAPI(server server) *FoodsAPI {
	return &FoodsAPI{server: server}
}

func (a *FoodsAPI) ConfigureRouter(router *mux.Router) {
	sb := router.PathPrefix("/api/foods").Subrouter()
	sb.Use(a.server.Middleware().Authentication.IsAuthorised)

	sb.Path("").
		Name("Foods Root Request").
		Methods(http.MethodGet, http.MethodPost).
		HandlerFunc(a.ServeRootRequest)
	sb.Path("/{id:[0-9]+}").
		Name("Foods ID Request").
		Methods(http.MethodGet, http.MethodPut, http.MethodDelete).
		HandlerFunc(a.ServeIDRequest)
}

func (a *FoodsAPI) ServeRootRequest(w http.ResponseWriter, r *http.Request) {
	if r.Method == http.MethodOptions {
		return
	}
	requestID, session, err := a.server.GetAuthorizedRequestInfo(r)
	if err != nil {
		a.server.RespondError(w, r, http.StatusInternalServerError, nil)
		return
	}

	switch r.Method {
	case http.MethodGet:
		if queriedNamePattern := r.URL.Query().Get("name"); queriedNamePattern != "" {
			foodModels, err := a.server.DatabaseStore().Foods().SelectByNameSimilarity(queriedNamePattern)
			if err != nil {
				a.server.Logger().Printf("Database error: %v Request ID: %v", err, requestID)
				a.server.RespondError(w, r, http.StatusInternalServerError, nil)
				return
			}
			a.server.Respond(w, r, http.StatusOK, foodModels)
			return
		}

		foodModels, err := a.server.DatabaseStore().Foods().SelectAll()
		if err != nil {
			a.server.Logger().Printf("Database error: %v Request ID: %v", err, requestID)
			a.server.RespondError(w, r, http.StatusInternalServerError, err)
			return
		}
		a.server.Respond(w, r, http.StatusOK, foodModels)

	case http.MethodPost:
		type requestBody struct {
			FoodName     string  `json:"food_name"`
			Calories     float64 `json:"calories"`
			Description  *string `json:"description"`
			Manufacturer *string `json:"manufacturer"`
		}
		rb := &requestBody{}
		if err := json.NewDecoder(r.Body).Decode(rb); err != nil {
			a.server.RespondError(w, r, http.StatusBadRequest, err)
			return
		}

		foodModel := &models.Food{
			FoodName:           rb.FoodName,
			Calories:           rb.Calories,
			SpecifiedCreatorID: session.UserID,
		}
		foodModel.SetSpecifiedDescription(rb.Description)
		foodModel.SetSpecifiedManufacturer(rb.Manufacturer)
		foodModel.BeforeCreate()
		if err := foodModel.Validate(); err != nil {
			a.server.RespondError(w, r, http.StatusUnprocessableEntity, err)
			return
		}
		foodModel, err = a.server.DatabaseStore().Foods().Create(foodModel)
		if err != nil {
			a.server.RespondError(w, r, http.StatusUnprocessableEntity, nil)
			return
		}
		a.server.Respond(w, r, http.StatusCreated, foodModel)
	}
}

func (a *FoodsAPI) ServeIDRequest(w http.ResponseWriter, r *http.Request) {
	if r.Method == http.MethodOptions {
		return
	}
	requestID, session, err := a.server.GetAuthorizedRequestInfo(r)
	if err != nil {
		a.server.RespondError(w, r, http.StatusInternalServerError, nil)
		return
	}

	requestedID, err := strconv.ParseInt(mux.Vars(r)["id"], 10, 64)
	if err != nil {
		a.server.RespondError(w, r, http.StatusBadRequest, exceptions.UnprocessableURIParam)
		return
	}
	requestedModel, err := a.server.DatabaseStore().Foods().FindByID(int(requestedID))
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			a.server.RespondError(w, r, http.StatusNotFound, nil)
			return
		}
		a.server.Logger().Printf("Database error: %v Request ID: %v", err, requestID)
		a.server.RespondError(w, r, http.StatusInternalServerError, nil)
		return
	}

	switch r.Method {
	case http.MethodGet:
		a.server.Respond(w, r, http.StatusOK, requestedModel)

	case http.MethodPut:
		type requestBody struct {
			FoodName     string  `json:"food_name"`
			Calories     float64 `json:"calories"`
			Description  *string `json:"description"`
			Manufacturer *string `json:"manufacturer"`
		}
		if !permissions.AnyRoleHavePermissions(session.Roles, permissions.Permissions().FoodPermission) {
			if requestedModel.CreatorID == nil || !requestedModel.CreatorID.Valid {
				a.server.RespondError(w, r, http.StatusForbidden, nil)
				return
			}
			if session.UserID != int(requestedModel.CreatorID.Int64) {
				a.server.RespondError(w, r, http.StatusForbidden, nil)
				return
			}
		}
		rb := &requestBody{}
		if err := json.NewDecoder(r.Body).Decode(rb); err != nil {
			a.server.RespondError(w, r, http.StatusBadRequest, err)
			return
		}
		foodModel := &models.Food{
			FoodID:                requestedModel.FoodID,
			FoodName:              rb.FoodName,
			Calories:              rb.Calories,
			SpecifiedDescription:  requestedModel.SpecifiedDescription,
			SpecifiedManufacturer: requestedModel.SpecifiedManufacturer,
			SpecifiedCreatorID:    requestedModel.SpecifiedCreatorID,
		}
		foodModel.SetSpecifiedManufacturer(rb.Manufacturer)
		foodModel.SetSpecifiedDescription(rb.Description)
		foodModel.BeforeCreate()
		if err := foodModel.Validate(); err != nil {
			a.server.RespondError(w, r, http.StatusUnprocessableEntity, err)
			return
		}
		foodModel, err = a.server.DatabaseStore().Foods().Update(foodModel)
		if err != nil {
			a.server.Logger().Printf("Database error: %v Request ID: %v", err, requestID)
			a.server.RespondError(w, r, http.StatusUnprocessableEntity, nil)
			return
		}
		a.server.Respond(w, r, http.StatusOK, foodModel)

	case http.MethodDelete:
		if !permissions.AnyRoleHavePermissions(session.Roles, permissions.Permissions().FoodPermission) {
			if requestedModel.CreatorID == nil || !requestedModel.CreatorID.Valid {
				a.server.RespondError(w, r, http.StatusForbidden, nil)
				return
			}
			if session.UserID != int(requestedModel.CreatorID.Int64) {
				a.server.RespondError(w, r, http.StatusForbidden, nil)
				return
			}
		}
		if _, err := a.server.DatabaseStore().Foods().DeleteByID(requestedModel.FoodID); err != nil {
			a.server.Logger().Printf("Database error: %v Request ID: %v", err, requestID)
			a.server.RespondError(w, r, http.StatusInternalServerError, err)
			return
		}
		a.server.Respond(w, r, http.StatusNoContent, nil)
	}
}
