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

type PetsAPI struct {
	server server
}

func NewPetsAPI(server server) *PetsAPI {
	return &PetsAPI{server: server}
}

func (a *PetsAPI) ConfigureRouter(router *mux.Router) {
	sb := router.PathPrefix("/api/pets").Subrouter()
	sb.Use(a.server.Middleware().Authentication.IsAuthorised)

	sb.Path("").
		Name("Pets Root Request").
		Methods(http.MethodGet, http.MethodPost).
		HandlerFunc(a.ServeRootRequest)

	sb.Path("/{id:[0-9]+}").
		Name("Pets ID Request").
		Methods(http.MethodGet, http.MethodPut, http.MethodDelete).
		HandlerFunc(a.ServeIDRequest)

	sb.Path("/types").
		Name("Pets types Root Request").
		Methods(http.MethodGet, http.MethodPost).
		HandlerFunc(a.ServeTypesRootRequest)

	sb.Path("/types/{id:[0-9]+}").
		Name("Pets types ID Request").
		Methods(http.MethodGet, http.MethodPut, http.MethodDelete).
		HandlerFunc(a.ServeTypesIDRequest)
}

func (a *PetsAPI) ServeRootRequest(w http.ResponseWriter, r *http.Request) {
	requestID, session, err := a.server.GetAuthorizedRequestInfo(r)
	if err != nil {
		a.server.RespondError(w, r, http.StatusInternalServerError, nil)
		return
	}

	switch r.Method {
	case http.MethodGet:
		pets, err := a.server.DatabaseStore().Pets().SelectAll()
		if err != nil {
			a.server.Logger().Printf("Database err: %v, Request ID: %v", err, requestID)
			a.server.RespondError(w, r, http.StatusInternalServerError, nil)
			return
		}
		a.server.Respond(w, r, http.StatusOK, pets)

	case http.MethodPost:
		type requestBody struct {
			Name           string  `json:"name"`
			PetType        int     `json:"pet_type"`
			VeterinarianID *int    `json:"veterinarian_id"`
			MotherID       *int    `json:"mother_id"`
			FatherID       *int    `json:"father_id"`
			Breed          *string `json:"breed"`
			FamilyName     *string `json:"family_name"`
			UserID         *int    `json:"user_id"`
		}
		rb := &requestBody{}
		if err := json.NewDecoder(r.Body).Decode(rb); err != nil {
			a.server.RespondError(w, r, http.StatusBadRequest, err)
			return
		}
		newPetModel := &models.Pet{
			Name:    rb.Name,
			PetType: rb.PetType,
		}
		petOwner := session.UserID
		if rb.UserID != nil && session.UserID != *rb.UserID {
			if !permissions.AnyRoleHavePermissions(session.Roles, permissions.Permissions().UsersPermission, permissions.Permissions().PetsPermission) {
				a.server.RespondError(w, r, http.StatusForbidden, exceptions.CanNotAssignPetToAnotherUser)
				return
			}
			petOwner = *rb.UserID
		}

		if rb.VeterinarianID != nil {
			specifiedVetRoles, err := a.server.DatabaseStore().Roles().SelectUserRoles(*rb.VeterinarianID)
			if err != nil {
				a.server.Respond(w, r, http.StatusUnprocessableEntity, exceptions.UserIsNotVeterinarian)
				return
			}
			if !permissions.AnyRoleIsVeterinarian(specifiedVetRoles) {
				a.server.Respond(w, r, http.StatusUnprocessableEntity, exceptions.UserIsNotVeterinarian)
				return
			}
		}

		newPetModel.UserID = petOwner
		newPetModel.SetSpecifiedBreed(rb.Breed)
		newPetModel.SetSpecifiedFamilyName(rb.FamilyName)
		newPetModel.SetSpecifiedFatherID(rb.FatherID)
		newPetModel.SetSpecifiedMotherID(rb.MotherID)
		newPetModel.SetSpecifiedVeterinarianID(rb.VeterinarianID)
		newPetModel.BeforeCreate()

		if err := newPetModel.Validate(); err != nil {
			a.server.RespondError(w, r, http.StatusUnprocessableEntity, err)
			return
		}
		newPetModel, err := a.server.DatabaseStore().Pets().CreatePet(newPetModel)
		if err != nil {
			a.server.Logger().Printf("Database err: %v, Request ID: %v", err, requestID)
			a.server.RespondError(w, r, http.StatusUnprocessableEntity, nil)
			return
		}
		a.server.Respond(w, r, http.StatusCreated, newPetModel)
	}
}

func (a *PetsAPI) ServeIDRequest(w http.ResponseWriter, r *http.Request) {
	requestID, session, err := a.server.GetAuthorizedRequestInfo(r)
	if err != nil {
		a.server.RespondError(w, r, http.StatusInternalServerError, nil)
		return
	}
	rawID, err := strconv.ParseInt(mux.Vars(r)["id"], 10, 64)
	if err != nil {
		a.server.RespondError(w, r, http.StatusBadRequest, exceptions.UnprocessableURIParam)
		return
	}
	requestedID := int(rawID)

	switch r.Method {
	case http.MethodGet:
		petModel, err := a.server.DatabaseStore().Pets().FindByID(requestedID)
		if err != nil {
			if errors.Is(err, sql.ErrNoRows) {
				a.server.RespondError(w, r, http.StatusNotFound, nil)
				return
			}
			a.server.Logger().Printf("Database err: %v, Request ID: %v", err, requestID)
			a.server.RespondError(w, r, http.StatusInternalServerError, nil)
			return
		}
		a.server.Respond(w, r, http.StatusOK, petModel)

	case http.MethodPut:
		type requestBody struct {
			Name           string  `json:"name"`
			PetType        int     `json:"pet_type"`
			VeterinarianID *int    `json:"veterinarian_id"`
			MotherID       *int    `json:"mother_id"`
			FatherID       *int    `json:"father_id"`
			Breed          *string `json:"breed"`
			FamilyName     *string `json:"family_name"`
			UserID         *int    `json:"user_id"`
		}
		updatingPet, err := a.server.DatabaseStore().Pets().FindByID(requestedID)
		if err != nil {
			if errors.Is(err, sql.ErrNoRows) {
				a.server.RespondError(w, r, http.StatusNotFound, nil)
				return
			}
			a.server.Logger().Printf("Database err: %v, Request ID: %v", err, requestID)
			a.server.RespondError(w, r, http.StatusInternalServerError, nil)
			return
		}

		if updatingPet.UserID != session.UserID {
			if !permissions.AnyRoleHavePermissions(session.Roles, permissions.Permissions().PetsPermission) {
				a.server.RespondError(w, r, http.StatusForbidden, nil)
				return
			}
		}

		rb := &requestBody{}
		if err := json.NewDecoder(r.Body).Decode(rb); err != nil {
			a.server.RespondError(w, r, http.StatusBadRequest, err)
			return
		}

		newPetModel := &models.Pet{
			Name:    rb.Name,
			PetType: rb.PetType,
		}

		petOwner := session.UserID
		if rb.UserID != nil && session.UserID != *rb.UserID {
			if !permissions.AnyRoleHavePermissions(session.Roles, permissions.Permissions().UsersPermission, permissions.Permissions().PetsPermission) {
				a.server.RespondError(w, r, http.StatusForbidden, exceptions.CanNotAssignPetToAnotherUser)
				return
			}
			petOwner = *rb.UserID
		}

		if rb.VeterinarianID != nil && *rb.VeterinarianID != 0 {
			specifiedVetRoles, err := a.server.DatabaseStore().Roles().SelectUserRoles(*rb.VeterinarianID)
			if err != nil {
				a.server.Respond(w, r, http.StatusUnprocessableEntity, exceptions.UserIsNotVeterinarian)
				return
			}
			if !permissions.AnyRoleIsVeterinarian(specifiedVetRoles) {
				a.server.Respond(w, r, http.StatusUnprocessableEntity, exceptions.UserIsNotVeterinarian)
				return
			}
		}

		newPetModel.UserID = petOwner
		newPetModel.SetSpecifiedBreed(rb.Breed)
		newPetModel.SetSpecifiedFamilyName(rb.FamilyName)
		newPetModel.SetSpecifiedFatherID(rb.FatherID)
		newPetModel.SetSpecifiedMotherID(rb.MotherID)
		newPetModel.SetSpecifiedVeterinarianID(rb.VeterinarianID)
		newPetModel.BeforeCreate()
		if err := newPetModel.Validate(); err != nil {
			a.server.RespondError(w, r, http.StatusUnprocessableEntity, err)
			return
		}

		updatingPet.Update(newPetModel)
		updatingPet.BeforeCreate()

		updated, err := a.server.DatabaseStore().Pets().UpdatePet(updatingPet)
		if err != nil {
			a.server.Logger().Printf("Database err: %v, Request ID: %v", err, requestID)
			a.server.RespondError(w, r, http.StatusUnprocessableEntity, nil)
			return
		}
		a.server.Respond(w, r, http.StatusOK, updated)

	case http.MethodDelete:
		deletingPet, err := a.server.DatabaseStore().Pets().FindByID(requestedID)
		if err != nil {
			if errors.Is(err, sql.ErrNoRows) {
				a.server.RespondError(w, r, http.StatusNotFound, nil)
				return
			}
			a.server.Logger().Printf("Database err: %v, Request ID: %v", err, requestID)
			a.server.RespondError(w, r, http.StatusInternalServerError, nil)
			return
		}

		if deletingPet.UserID != session.UserID {
			if !permissions.AnyRoleHavePermissions(session.Roles, permissions.Permissions().PetsPermission, permissions.Permissions().UsersPermission) {
				a.server.RespondError(w, r, http.StatusForbidden, nil)
				return
			}
		}
		if _, err := a.server.DatabaseStore().Pets().DeleteByID(requestedID); err != nil {
			a.server.Logger().Printf("Database err: %v, Request ID: %v", err, requestID)
			a.server.RespondError(w, r, http.StatusUnprocessableEntity, nil)
			return
		}
		a.server.Respond(w, r, http.StatusNoContent, nil)
	}
}

func (a *PetsAPI) ServeTypesRootRequest(w http.ResponseWriter, r *http.Request) {
	requestID, session, err := a.server.GetAuthorizedRequestInfo(r)
	if err != nil {
		a.server.RespondError(w, r, http.StatusInternalServerError, nil)
		return
	}

	switch r.Method {
	case http.MethodGet:
		typesModels, err := a.server.DatabaseStore().Pets().SelectAllTypes()
		if err != nil {
			a.server.Logger().Printf("Database err: %v, Request ID: %v", err, requestID)
			a.server.RespondError(w, r, http.StatusInternalServerError, nil)
			return
		}
		a.server.Respond(w, r, http.StatusOK, typesModels)

	case http.MethodPost:
		type requestBody struct {
			TypeName       string  `json:"type_name"`
			RERCoefficient float64 `json:"rer_coefficient"`
		}
		if !permissions.AnyRoleHavePermissions(session.Roles, permissions.Permissions().PetsPermission) {
			a.server.RespondError(w, r, http.StatusForbidden, nil)
			return
		}
		rb := &requestBody{}
		if err := json.NewDecoder(r.Body).Decode(rb); err != nil {
			a.server.RespondError(w, r, http.StatusBadRequest, err)
			return
		}
		newPetType := &models.PetType{
			TypeName:       rb.TypeName,
			RERCoefficient: rb.RERCoefficient,
		}
		if err := newPetType.Validate(); err != nil {
			a.server.RespondError(w, r, http.StatusUnprocessableEntity, err)
			return
		}
		newPetType, err := a.server.DatabaseStore().Pets().CreatePetType(newPetType)
		if err != nil {
			a.server.Logger().Printf("Database err: %v, Request ID: %v", err, requestID)
			a.server.RespondError(w, r, http.StatusUnprocessableEntity, nil)
			return
		}
		a.server.Respond(w, r, http.StatusCreated, newPetType)
	}
}

func (a *PetsAPI) ServeTypesIDRequest(w http.ResponseWriter, r *http.Request) {
	requestID, session, err := a.server.GetAuthorizedRequestInfo(r)
	if err != nil {
		a.server.RespondError(w, r, http.StatusInternalServerError, nil)
		return
	}
	rawID, err := strconv.ParseInt(mux.Vars(r)["id"], 10, 64)
	if err != nil {
		a.server.RespondError(w, r, http.StatusBadRequest, exceptions.UnprocessableURIParam)
		return
	}
	requestedID := int(rawID)

	switch r.Method {
	case http.MethodGet:
		typeModel, err := a.server.DatabaseStore().Pets().SelectTypeByID(requestedID)
		if err != nil {
			if errors.Is(err, sql.ErrNoRows) {
				a.server.RespondError(w, r, http.StatusNotFound, nil)
				return
			}
			a.server.Logger().Printf("Database err: %v, Request ID: %v", err, requestID)
			a.server.RespondError(w, r, http.StatusInternalServerError, nil)
			return
		}
		a.server.Respond(w, r, http.StatusOK, typeModel)

	case http.MethodPut:
		type requestBody struct {
			TypeName       string  `json:"type_name"`
			RERCoefficient float64 `json:"rer_coefficient"`
		}
		if !permissions.AnyRoleHavePermissions(session.Roles, permissions.Permissions().PetsPermission) {
			a.server.RespondError(w, r, http.StatusForbidden, nil)
			return
		}
		rb := &requestBody{}
		if err := json.NewDecoder(r.Body).Decode(rb); err != nil {
			a.server.RespondError(w, r, http.StatusBadRequest, err)
			return
		}
		updatedPetType := &models.PetType{
			TypeID:         requestedID,
			TypeName:       rb.TypeName,
			RERCoefficient: rb.RERCoefficient,
		}
		if err := updatedPetType.Validate(); err != nil {
			a.server.RespondError(w, r, http.StatusUnprocessableEntity, err)
			return
		}
		updatedPetType, err = a.server.DatabaseStore().Pets().UpdatePetType(updatedPetType)
		if err != nil {
			if errors.Is(err, sql.ErrNoRows) {
				a.server.RespondError(w, r, http.StatusNotFound, nil)
				return
			}
			a.server.Logger().Printf("Database err: %v, Request ID: %v", err, requestID)
			a.server.RespondError(w, r, http.StatusUnprocessableEntity, nil)
			return
		}
		a.server.Respond(w, r, http.StatusOK, updatedPetType)

	case http.MethodDelete:
		if !permissions.AnyRoleHavePermissions(session.Roles, permissions.Permissions().PetsPermission) {
			a.server.RespondError(w, r, http.StatusForbidden, nil)
			return
		}
		if _, err := a.server.DatabaseStore().Pets().DeleteTypeByID(requestedID); err != nil {
			if errors.Is(err, sql.ErrNoRows) {
				a.server.RespondError(w, r, http.StatusNotFound, nil)
				return
			}
			a.server.Logger().Printf("Database err: %v, Request ID: %v", err, requestID)
			a.server.RespondError(w, r, http.StatusInternalServerError, nil)
			return
		}
	}
}
