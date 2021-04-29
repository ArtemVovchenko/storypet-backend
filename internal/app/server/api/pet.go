package api

import (
	"database/sql"
	"encoding/json"
	"errors"
	"github.com/ArtemVovchenko/storypet-backend/internal/app/models"
	"github.com/ArtemVovchenko/storypet-backend/internal/app/permissions"
	"github.com/ArtemVovchenko/storypet-backend/internal/app/server/api/exceptions"
	"github.com/gorilla/mux"
	jsontime "github.com/liamylian/jsontime/v2/v2"
	"net/http"
	"strconv"
	"time"
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

	sb.Path("/{id:[0-9]+}/veterinarian").
		Name("Pets veterinarian request").
		Methods(http.MethodGet, http.MethodPost, http.MethodDelete).
		HandlerFunc(a.ServeVeterinarianRequest)

	sb.Path("/{id:[0-9]+}/parents").
		Name("Pets parents Request").
		Methods(http.MethodGet, http.MethodPost, http.MethodDelete).
		HandlerFunc(a.ServeParentsRequest)

	sb.Path("/{id:[0-9]+}/parents/verify/{parent:father|mother}").
		Name("Pets parent verification request").
		Methods(http.MethodPost).
		HandlerFunc(a.ServeParentsVerificationRequest)

	sb.Path("/{id:[0-9]+}/vaccines").
		Name("Pets vaccines Request").
		Methods(http.MethodGet, http.MethodPost).
		HandlerFunc(a.ServePetsVaccinesRequest)

	sb.Path("/{id:[0-9]+}/vaccines/{vaccine:[0-9]+}").
		Name("Pets vaccines Request").
		Methods(http.MethodGet, http.MethodPut, http.MethodDelete).
		HandlerFunc(a.ServePetsVaccineRequest)

	sb.Path("/{id:[0-9]+}/stats").
		Name("Pets Stats Request").
		Methods(http.MethodGet, http.MethodPost).
		HandlerFunc(a.ServePetStatsRequest)

	sb.Path("/{id:[0-9]+}/stats/{recID:[0-9]+}").
		Name("Pets Stat ID Request").
		Methods(http.MethodGet, http.MethodPut, http.MethodDelete).
		HandlerFunc(a.ServePetStatsIDRequest)

	sb.Path("/{id:[0-9]+}/activity").
		Name("Pets Activity Request").
		Methods(http.MethodGet, http.MethodPost).
		HandlerFunc(a.ServePetActivityRequest)

	sb.Path("/{id:[0-9]+}/eating").
		Name("Pets Eating Request").
		Methods(http.MethodGet, http.MethodPost).
		HandlerFunc(a.ServeEatingRequest)
}

func (a *PetsAPI) ServeRootRequest(w http.ResponseWriter, r *http.Request) {
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
		pets, err := a.server.DatabaseStore().Pets().SelectAll()
		if err != nil {
			a.server.Logger().Printf("Database err: %v, Request ID: %v", err, requestID)
			a.server.RespondError(w, r, http.StatusInternalServerError, nil)
			return
		}
		a.server.Respond(w, r, http.StatusOK, pets)

	case http.MethodPost:
		type requestBody struct {
			Name       string  `json:"name"`
			PetType    int     `json:"pet_type"`
			Breed      *string `json:"breed"`
			FamilyName *string `json:"family_name"`
			UserID     *int    `json:"user_id"`
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

		newPetModel.UserID = petOwner
		newPetModel.SetSpecifiedBreed(rb.Breed)
		newPetModel.SetSpecifiedFamilyName(rb.FamilyName)
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
	if r.Method == http.MethodOptions {
		return
	}
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
			Name       string  `json:"name"`
			PetType    int     `json:"pet_type"`
			Breed      *string `json:"breed"`
			FamilyName *string `json:"family_name"`
			UserID     *int    `json:"user_id"`
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

		newPetModel.UserID = petOwner
		newPetModel.SetSpecifiedBreed(rb.Breed)
		newPetModel.SetSpecifiedFamilyName(rb.FamilyName)
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
	if r.Method == http.MethodOptions {
		return
	}
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
		typeModel, err := a.server.DatabaseStore().Pets().FindTypeByID(requestedID)
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

func (a *PetsAPI) ServeVeterinarianRequest(w http.ResponseWriter, r *http.Request) {
	if r.Method == http.MethodOptions {
		return
	}
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

	switch r.Method {
	case http.MethodGet:
		if petModel.VeterinarianID == nil || !petModel.VeterinarianID.Valid {
			a.server.RespondError(w, r, http.StatusNotFound, nil)
			return
		}
		veterinarianModel, err := a.server.DatabaseStore().Users().FindByID(int(petModel.VeterinarianID.Int64))
		if err != nil {
			a.server.Logger().Printf("Database err: %v, Request ID: %v", err, requestID)
			a.server.RespondError(w, r, http.StatusInternalServerError, nil)
			return
		}
		a.server.Respond(w, r, http.StatusOK, veterinarianModel)

	case http.MethodPost:
		type requestBody struct {
			VeterinarianID int `json:"veterinarian_id"`
		}
		rb := &requestBody{}
		if err := json.NewDecoder(r.Body).Decode(rb); err != nil {
			a.server.RespondError(w, r, http.StatusBadRequest, nil)
			return
		}
		userRoles, err := a.server.DatabaseStore().Roles().SelectUserRoles(rb.VeterinarianID)
		if err != nil {
			a.server.Logger().Printf("Database error: %v Request ID %v", err, requestID)
			a.server.RespondError(w, r, http.StatusInternalServerError, nil)
			return
		}
		if !permissions.AnyRoleIsVeterinarian(userRoles) {
			a.server.RespondError(w, r, http.StatusUnprocessableEntity, exceptions.UserIsNotVeterinarian)
			return
		}
		if rb.VeterinarianID != session.UserID {
			if !permissions.AnyRoleHavePermissions(session.Roles, permissions.Permissions().UsersPermission) {
				a.server.RespondError(w, r, http.StatusForbidden, nil)
				return
			}
		}
		if petModel.VeterinarianID != nil && petModel.VeterinarianID.Valid {
			a.server.RespondError(w, r, http.StatusUnprocessableEntity, exceptions.PetHasVeterinarian)
			return
		}
		if err := a.server.DatabaseStore().Pets().AssignVeterinarian(requestedID, rb.VeterinarianID); err != nil {
			a.server.Logger().Printf("Database Error: %v Request ID: %v", err, requestID)
			a.server.RespondError(w, r, http.StatusInternalServerError, nil)
			return
		}
		a.server.Respond(w, r, http.StatusOK, nil)

	case http.MethodDelete:
		if petModel.VeterinarianID == nil || !petModel.VeterinarianID.Valid {
			a.server.RespondError(w, r, http.StatusBadRequest, exceptions.PetHasNoVeterinarian)
			return
		}
		if int(petModel.VeterinarianID.Int64) != session.UserID {
			if !permissions.AnyRoleHavePermissions(session.Roles, permissions.Permissions().UsersPermission) {
				a.server.RespondError(w, r, http.StatusForbidden, nil)
				return
			}
		}
		if err := a.server.DatabaseStore().Pets().DeleteVeterinarian(petModel.PetID); err != nil {
			a.server.Logger().Printf("Database error: %v Request ID: %v", err, requestID)
			a.server.RespondError(w, r, http.StatusInternalServerError, nil)
			return
		}
		a.server.Respond(w, r, http.StatusOK, nil)
	}
}

func (a *PetsAPI) ServeParentsRequest(w http.ResponseWriter, r *http.Request) {
	if r.Method == http.MethodOptions {
		return
	}
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

	switch r.Method {
	case http.MethodGet:
		type responseBody struct {
			Father *models.Pet `json:"father"`
			Mother *models.Pet `json:"mother"`
		}

		var father, mother *models.Pet
		if petModel.FatherID != nil && petModel.FatherID.Valid {
			if father, err = a.server.DatabaseStore().Pets().FindByID(int(petModel.FatherID.Int64)); err != nil {
				a.server.Logger().Printf("Database error: %v Request ID %v", err, requestID)
				a.server.RespondError(w, r, http.StatusInternalServerError, nil)
				return
			}
		}
		if petModel.MotherID != nil && petModel.MotherID.Valid {
			if mother, err = a.server.DatabaseStore().Pets().FindByID(int(petModel.MotherID.Int64)); err != nil {
				a.server.Logger().Printf("Database error: %v Request ID %v", err, requestID)
				a.server.RespondError(w, r, http.StatusInternalServerError, nil)
				return
			}
		}
		rb := &responseBody{Father: father, Mother: mother}
		a.server.Respond(w, r, http.StatusOK, rb)

	case http.MethodPost:
		type requestBody struct {
			FatherID *int `json:"father_id"`
			MotherID *int `json:"mother_id"`
		}
		if petModel.UserID != session.UserID {
			if !permissions.AnyRoleHavePermissions(session.Roles, permissions.Permissions().UsersPermission, permissions.Permissions().PetsPermission) {
				a.server.RespondError(w, r, http.StatusForbidden, nil)
				return
			}
		}
		rb := &requestBody{}
		if err := json.NewDecoder(r.Body).Decode(rb); err != nil {
			a.server.RespondError(w, r, http.StatusBadRequest, err)
			return
		}
		if rb.FatherID == nil && rb.MotherID == nil {
			a.server.RespondError(w, r, http.StatusUnprocessableEntity, exceptions.NoParentsSpecified)
			return
		}
		if err := a.server.DatabaseStore().Pets().SpecifyParents(rb.FatherID, rb.MotherID, petModel.PetID); err != nil {
			a.server.Logger().Printf("Database error: %v Request ID: $v", err, requestID)
			a.server.RespondError(w, r, http.StatusInternalServerError, nil)
			return
		}
		a.server.Respond(w, r, http.StatusOK, nil)

	case http.MethodDelete:
		if petModel.UserID != session.UserID {
			if !permissions.AnyRoleHavePermissions(session.Roles, permissions.Permissions().UsersPermission, permissions.Permissions().PetsPermission) {
				a.server.RespondError(w, r, http.StatusForbidden, nil)
				return
			}
		}
		if err := a.server.DatabaseStore().Pets().RemoveParents(petModel.PetID); err != nil {
			a.server.Logger().Printf("Database error: %v Request ID: $v", err, requestID)
			a.server.RespondError(w, r, http.StatusInternalServerError, nil)
			return
		}
		a.server.Respond(w, r, http.StatusOK, nil)
	}
}

func (a *PetsAPI) ServeParentsVerificationRequest(w http.ResponseWriter, r *http.Request) {
	if r.Method == http.MethodOptions {
		return
	}
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
	parent := mux.Vars(r)["parent"]

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

	switch parent {
	case "father":
		if petModel.FatherID == nil || !petModel.FatherID.Valid {
			a.server.RespondError(w, r, http.StatusBadRequest, exceptions.NoParentsSpecified)
			return
		}
		fatherPet, err := a.server.DatabaseStore().Pets().FindByID(int(petModel.FatherID.Int64))
		if err != nil {
			a.server.Logger().Printf("Database error: %v Request ID: %v", err, requestID)
			a.server.RespondError(w, r, http.StatusInternalServerError, nil)
			return
		}
		owner, err := a.server.DatabaseStore().Users().FindByID(fatherPet.UserID)
		if err != nil {
			a.server.Logger().Printf("Database error: %v Request ID: %v", err, requestID)
			a.server.RespondError(w, r, http.StatusInternalServerError, nil)
			return
		}
		if session.UserID != owner.UserID {
			if !permissions.AnyRoleHavePermissions(session.Roles, permissions.Permissions().UsersPermission, permissions.Permissions().PetsPermission) {
				a.server.RespondError(w, r, http.StatusForbidden, nil)
				return
			}
		}
		if err := a.server.DatabaseStore().Pets().VerifyFather(petModel.PetID); err != nil {
			a.server.Logger().Printf("Database error: %v Request ID: %v", err, requestID)
			a.server.RespondError(w, r, http.StatusInternalServerError, nil)
			return
		}
		a.server.Respond(w, r, http.StatusOK, nil)

	case "mother":
		if petModel.MotherID == nil || !petModel.MotherID.Valid {
			a.server.RespondError(w, r, http.StatusBadRequest, exceptions.NoParentsSpecified)
			return
		}
		fatherPet, err := a.server.DatabaseStore().Pets().FindByID(int(petModel.MotherID.Int64))
		if err != nil {
			a.server.Logger().Printf("Database error: %v Request ID: %v", err, requestID)
			a.server.RespondError(w, r, http.StatusInternalServerError, nil)
			return
		}
		owner, err := a.server.DatabaseStore().Users().FindByID(fatherPet.UserID)
		if err != nil {
			a.server.Logger().Printf("Database error: %v Request ID: %v", err, requestID)
			a.server.RespondError(w, r, http.StatusInternalServerError, nil)
			return
		}
		if session.UserID != owner.UserID {
			if !permissions.AnyRoleHavePermissions(session.Roles, permissions.Permissions().UsersPermission, permissions.Permissions().PetsPermission) {
				a.server.RespondError(w, r, http.StatusForbidden, nil)
				return
			}
		}
		if err := a.server.DatabaseStore().Pets().VerifyMother(petModel.PetID); err != nil {
			a.server.Logger().Printf("Database error: %v Request ID: %v", err, requestID)
			a.server.RespondError(w, r, http.StatusInternalServerError, nil)
			return
		}
		a.server.Respond(w, r, http.StatusOK, nil)
	}
}

func (a *PetsAPI) ServePetsVaccinesRequest(w http.ResponseWriter, r *http.Request) {
	if r.Method == http.MethodOptions {
		return
	}
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
	requestedPetID := int(rawID)

	petModel, err := a.server.DatabaseStore().Pets().FindByID(requestedPetID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			a.server.RespondError(w, r, http.StatusNotFound, nil)
			return
		}
		a.server.Logger().Printf("Database err: %v, Request ID: %v", err, requestID)
		a.server.RespondError(w, r, http.StatusInternalServerError, nil)
		return
	}

	switch r.Method {
	case http.MethodGet:
		vaccines, err := a.server.DatabaseStore().Vaccines().SelectByPetID(requestedPetID)
		if err != nil {
			if errors.Is(err, sql.ErrNoRows) {
				a.server.Respond(w, r, http.StatusOK, nil)
				return
			}
			a.server.Logger().Printf("Database error: $v Request ID: %v", err, requestID)
			a.server.RespondError(w, r, http.StatusInternalServerError, nil)
			return
		}
		a.server.Respond(w, r, http.StatusOK, vaccines)

	case http.MethodPost:
		type requestBody struct {
			VaccineName        string    `json:"name"`
			VaccinationDate    time.Time `json:"vaccination_date" time_format:"date"`
			VaccineDescription *string   `json:"vaccine_description"`
		}
		if petModel.VeterinarianID == nil ||
			!petModel.VeterinarianID.Valid ||
			int(petModel.VeterinarianID.Int64) != session.UserID {
			if !permissions.AnyRoleHavePermissions(session.Roles, permissions.All()...) {
				a.server.RespondError(w, r, http.StatusForbidden, nil)
				return
			}
		}
		rb := &requestBody{}
		var jsonTime = jsontime.ConfigWithCustomTimeFormat
		jsontime.AddTimeFormatAlias("date", "2006-01-02")

		if err := jsonTime.NewDecoder(r.Body).Decode(rb); err != nil {
			a.server.RespondError(w, r, http.StatusBadRequest, err)
			return
		}
		vaccineModel := &models.Vaccine{
			PetID:           petModel.PetID,
			Name:            rb.VaccineName,
			VaccinationDate: rb.VaccinationDate,
		}
		vaccineModel.SetSpecifiedDescription(rb.VaccineDescription)
		vaccineModel.BeforeCreate()
		if err := vaccineModel.Validate(); err != nil {
			a.server.RespondError(w, r, http.StatusUnprocessableEntity, err)
			return
		}
		vaccineModel, err = a.server.DatabaseStore().Vaccines().Create(vaccineModel)
		if err != nil {
			a.server.Logger().Printf("Database error: %v Request ID: %v", err, requestID)
			a.server.RespondError(w, r, http.StatusUnprocessableEntity, nil)
			return
		}
		a.server.Respond(w, r, http.StatusCreated, vaccineModel)
	}
}

func (a *PetsAPI) ServePetsVaccineRequest(w http.ResponseWriter, r *http.Request) {
	if r.Method == http.MethodOptions {
		return
	}
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
	requestedPetID := int(rawID)
	rawID, err = strconv.ParseInt(mux.Vars(r)["vaccine"], 10, 64)
	if err != nil {
		a.server.RespondError(w, r, http.StatusBadRequest, exceptions.UnprocessableURIParam)
		return
	}
	requestedVaccineID := int(rawID)

	petModel, err := a.server.DatabaseStore().Pets().FindByID(requestedPetID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			a.server.RespondError(w, r, http.StatusNotFound, nil)
			return
		}
		a.server.Logger().Printf("Database err: %v, Request ID: %v", err, requestID)
		a.server.RespondError(w, r, http.StatusInternalServerError, nil)
		return
	}

	switch r.Method {
	case http.MethodGet:
		vaccines, err := a.server.DatabaseStore().Vaccines().FindByID(requestedVaccineID)
		if err != nil {
			if errors.Is(err, sql.ErrNoRows) {
				a.server.Respond(w, r, http.StatusNotFound, nil)
				return
			}
			a.server.Logger().Printf("Database error: $v Request ID: %v", err, requestID)
			a.server.RespondError(w, r, http.StatusInternalServerError, nil)
			return
		}
		a.server.Respond(w, r, http.StatusOK, vaccines)

	case http.MethodPut:
		type requestBody struct {
			VaccineName        string    `json:"name"`
			VaccinationDate    time.Time `json:"vaccination_date" time_format:"date"`
			VaccineDescription *string   `json:"vaccine_description"`
		}
		if petModel.VeterinarianID == nil ||
			!petModel.VeterinarianID.Valid ||
			int(petModel.VeterinarianID.Int64) != session.UserID {
			if !permissions.AnyRoleHavePermissions(session.Roles, permissions.All()...) {
				a.server.RespondError(w, r, http.StatusForbidden, nil)
				return
			}
		}
		rb := &requestBody{}

		var jsonTime = jsontime.ConfigWithCustomTimeFormat
		jsontime.AddTimeFormatAlias("date", "2006-01-02")

		if err := jsonTime.NewDecoder(r.Body).Decode(rb); err != nil {
			a.server.RespondError(w, r, http.StatusBadRequest, err)
			return
		}
		vaccineModel := &models.Vaccine{
			VaccineID:       requestedVaccineID,
			PetID:           petModel.PetID,
			Name:            rb.VaccineName,
			VaccinationDate: rb.VaccinationDate,
		}
		vaccineModel.SetSpecifiedDescription(rb.VaccineDescription)
		vaccineModel.BeforeCreate()
		if err := vaccineModel.Validate(); err != nil {
			a.server.RespondError(w, r, http.StatusUnprocessableEntity, err)
			return
		}
		vaccineModel, err = a.server.DatabaseStore().Vaccines().Update(vaccineModel)
		if err != nil {
			if errors.Is(err, sql.ErrNoRows) {
				a.server.Respond(w, r, http.StatusNotFound, nil)
				return
			}
			a.server.Logger().Printf("Database error: %v Request ID: %v", err, requestID)
			a.server.RespondError(w, r, http.StatusUnprocessableEntity, nil)
			return
		}
		a.server.Respond(w, r, http.StatusOK, vaccineModel)

	case http.MethodDelete:
		if petModel.VeterinarianID == nil ||
			!petModel.VeterinarianID.Valid ||
			int(petModel.VeterinarianID.Int64) != session.UserID {
			if !permissions.AnyRoleHavePermissions(session.Roles, permissions.All()...) {
				a.server.RespondError(w, r, http.StatusForbidden, nil)
				return
			}
		}
		if _, err := a.server.DatabaseStore().Vaccines().DeleteByID(requestedVaccineID); err != nil {
			if errors.Is(err, sql.ErrNoRows) {
				a.server.Respond(w, r, http.StatusNotFound, nil)
				return
			}
			a.server.Logger().Printf("Database error: %v Request ID: %v", err, requestID)
			a.server.RespondError(w, r, http.StatusUnprocessableEntity, nil)
			return
		}
		a.server.Respond(w, r, http.StatusNoContent, nil)
	}
}

func (a *PetsAPI) ServePetStatsRequest(w http.ResponseWriter, r *http.Request) {
	if r.Method == http.MethodOptions {
		return
	}
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
		petStats, err := a.server.DatabaseStore().Pets().SelectPetAnthropometryRecords(requestedID)
		if err != nil {
			if errors.Is(err, sql.ErrNoRows) {
				a.server.RespondError(w, r, http.StatusNotFound, nil)
				return
			}
			a.server.Logger().Printf("Database error: %v Request ID %v", err, requestID)
			a.server.RespondError(w, r, http.StatusInternalServerError, nil)
			return
		}
		a.server.Respond(w, r, http.StatusOK, petStats)

	case http.MethodPost:
		type requestBody struct {
			Height float64 `json:"height"`
			Weight float64 `json:"weight"`
		}
		petModel, err := a.server.DatabaseStore().Pets().FindByID(requestedID)
		if err != nil {
			if errors.Is(err, sql.ErrNoRows) {
				a.server.RespondError(w, r, http.StatusNotFound, nil)
				return
			}
			a.server.Logger().Printf("Database error: %v Request ID %v", err, requestID)
			a.server.RespondError(w, r, http.StatusInternalServerError, nil)
			return
		}
		if session.UserID != petModel.UserID {
			if petModel.VeterinarianID != nil && int(petModel.VeterinarianID.Int64) != session.UserID {
				if !permissions.AnyRoleHavePermissions(
					session.Roles,
					permissions.Permissions().PetsPermission,
					permissions.Permissions().UsersPermission) {
					a.server.RespondError(w, r, http.StatusForbidden, nil)
					return
				}
			}
		}
		rb := &requestBody{}
		if err := json.NewDecoder(r.Body).Decode(rb); err != nil {
			a.server.RespondError(w, r, http.StatusBadRequest, err)
			return
		}
		recordModel := &models.Anthropometry{
			PetID:  requestedID,
			Time:   time.Now(),
			Height: rb.Height,
			Weight: rb.Weight,
		}
		if err := recordModel.Validate(); err != nil {
			a.server.RespondError(w, r, http.StatusUnprocessableEntity, err)
			return
		}
		recordModel, err = a.server.DatabaseStore().Pets().SpecifyAnthropometry(recordModel)
		if err != nil {
			a.server.Logger().Printf("Database error %v Request ID: %v", err, requestID)
			a.server.RespondError(w, r, http.StatusInternalServerError, nil)
			return
		}
		a.server.Respond(w, r, http.StatusCreated, recordModel)
	}
}

func (a *PetsAPI) ServePetStatsIDRequest(w http.ResponseWriter, r *http.Request) {
	if r.Method == http.MethodOptions {
		return
	}
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
	requestedPetID := int(rawID)
	rawID, err = strconv.ParseInt(mux.Vars(r)["recID"], 10, 64)
	if err != nil {
		a.server.RespondError(w, r, http.StatusBadRequest, exceptions.UnprocessableURIParam)
		return
	}
	requestedRecID := int(rawID)

	record, err := a.server.DatabaseStore().Pets().FindAnthropometryRecordByID(requestedRecID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			a.server.RespondError(w, r, http.StatusNotFound, nil)
			return
		}
		a.server.Logger().Printf("Database error: %v Request ID $v", err, requestID)
		a.server.RespondError(w, r, http.StatusInternalServerError, nil)
		return
	}

	switch r.Method {
	case http.MethodGet:
		a.server.Respond(w, r, http.StatusOK, record)

	case http.MethodPut:
		type requestBody struct {
			Height float64 `json:"height"`
			Weight float64 `json:"weight"`
		}

		petModel, err := a.server.DatabaseStore().Pets().FindByID(requestedPetID)
		if err != nil {
			if errors.Is(err, sql.ErrNoRows) {
				a.server.RespondError(w, r, http.StatusNotFound, nil)
				return
			}
			a.server.Logger().Printf("Database error: %v Request ID %v", err, requestID)
			a.server.RespondError(w, r, http.StatusInternalServerError, nil)
			return
		}
		if session.UserID != petModel.UserID {
			if petModel.VeterinarianID != nil && int(petModel.VeterinarianID.Int64) != session.UserID {
				if !permissions.AnyRoleHavePermissions(
					session.Roles,
					permissions.Permissions().PetsPermission,
					permissions.Permissions().UsersPermission) {
					a.server.RespondError(w, r, http.StatusForbidden, nil)
					return
				}
			}
		}
		rb := &requestBody{}
		if err := json.NewDecoder(r.Body).Decode(rb); err != nil {
			a.server.RespondError(w, r, http.StatusBadRequest, err)
			return
		}
		newRecord := &models.Anthropometry{
			RecordID: record.RecordID,
			PetID:    record.PetID,
			Time:     record.Time,
			Height:   rb.Height,
			Weight:   rb.Weight,
		}
		if err := newRecord.Validate(); err != nil {
			a.server.RespondError(w, r, http.StatusUnprocessableEntity, err)
			return
		}
		newRecord, err = a.server.DatabaseStore().Pets().UpdateAnthropometry(newRecord)
		if err != nil {
			a.server.Logger().Printf("Database error: %v RequestID: %v", err, requestID)
			a.server.RespondError(w, r, http.StatusInternalServerError, nil)
			return
		}
		a.server.Respond(w, r, http.StatusOK, newRecord)

	case http.MethodDelete:
		petModel, err := a.server.DatabaseStore().Pets().FindByID(requestedPetID)
		if err != nil {
			if errors.Is(err, sql.ErrNoRows) {
				a.server.RespondError(w, r, http.StatusNotFound, nil)
				return
			}
			a.server.Logger().Printf("Database error: %v Request ID %v", err, requestID)
			a.server.RespondError(w, r, http.StatusInternalServerError, nil)
			return
		}
		if session.UserID != petModel.UserID {
			if petModel.VeterinarianID != nil && int(petModel.VeterinarianID.Int64) != session.UserID {
				if !permissions.AnyRoleHavePermissions(
					session.Roles,
					permissions.Permissions().PetsPermission,
					permissions.Permissions().UsersPermission) {
					a.server.RespondError(w, r, http.StatusForbidden, nil)
					return
				}
			}
		}
		if _, err := a.server.DatabaseStore().Pets().DeleteAnthropometryByID(requestedRecID); err != nil {
			if errors.Is(err, sql.ErrNoRows) {
				a.server.RespondError(w, r, http.StatusNotFound, nil)
				return
			}
			a.server.Logger().Printf("Database error: %v RequestID: %v", err, requestID)
			a.server.RespondError(w, r, http.StatusInternalServerError, nil)
			return
		}
		a.server.Respond(w, r, http.StatusNoContent, nil)
	}
}

func (a *PetsAPI) ServePetActivityRequest(w http.ResponseWriter, r *http.Request) {
	if r.Method == http.MethodOptions {
		return
	}
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
	requestedPetID := int(rawID)

	switch r.Method {
	case http.MethodGet:
		var recordModels []models.Activity
		startDate := r.URL.Query().Get("start")
		endDate := r.URL.Query().Get("end")
		var startDateT time.Time
		var endDateT time.Time

		if startDate == "" && endDate == "" {
			recordModels, err = a.server.DatabaseStore().Pets().SelectPetActivityRecords(requestedPetID)
			if err != nil {
				if errors.Is(err, sql.ErrNoRows) {
					a.server.RespondError(w, r, http.StatusNotFound, nil)
					return
				}
				a.server.Logger().Printf("Database error: %v RequestID: %v", err, requestID)
				a.server.RespondError(w, r, http.StatusInternalServerError, nil)
				return
			}
			a.server.Respond(w, r, http.StatusOK, recordModels)
			return
		}

		if startDate == "" {
			endDateT, err = time.Parse("2006-01-02", endDate)
			if err != nil {
				a.server.RespondError(w, r, http.StatusBadRequest, exceptions.UnprocessableURLQuery)
				return
			}
			recordModels, err = a.server.DatabaseStore().Pets().SelectPetActivityRecordsToTime(requestedPetID, endDateT)
			if err != nil {
				if errors.Is(err, sql.ErrNoRows) {
					a.server.RespondError(w, r, http.StatusNotFound, nil)
					return
				}
				a.server.Logger().Printf("Database error: %v RequestID: %v", err, requestID)
				a.server.RespondError(w, r, http.StatusInternalServerError, nil)
				return
			}
			a.server.Respond(w, r, http.StatusOK, recordModels)
			return
		}

		if endDate == "" {
			startDateT, err = time.Parse("2006-01-02", startDate)
			if err != nil {
				a.server.RespondError(w, r, http.StatusBadRequest, exceptions.UnprocessableURLQuery)
				return
			}
			recordModels, err = a.server.DatabaseStore().Pets().SelectPetActivityRecordsInInterval(requestedPetID, startDateT, time.Now())
			if err != nil {
				if errors.Is(err, sql.ErrNoRows) {
					a.server.RespondError(w, r, http.StatusNotFound, nil)
					return
				}
				a.server.Logger().Printf("Database error: %v RequestID: %v", err, requestID)
				a.server.RespondError(w, r, http.StatusInternalServerError, nil)
				return
			}
			a.server.Respond(w, r, http.StatusOK, recordModels)
			return
		}

		startDateT, err = time.Parse("2006-01-02", startDate)
		endDateT, err = time.Parse("2006-01-02", endDate)
		if err != nil {
			a.server.RespondError(w, r, http.StatusBadRequest, exceptions.UnprocessableURLQuery)
			return
		}

		recordModels, err = a.server.DatabaseStore().Pets().SelectPetActivityRecordsInInterval(requestedPetID, startDateT, endDateT)
		if err != nil {
			if errors.Is(err, sql.ErrNoRows) {
				a.server.RespondError(w, r, http.StatusNotFound, nil)
				return
			}
			a.server.Logger().Printf("Database error: %v RequestID: %v", err, requestID)
			a.server.RespondError(w, r, http.StatusInternalServerError, nil)
			return
		}
		a.server.Respond(w, r, http.StatusOK, recordModels)
		return

	case http.MethodPost:
		type requestBody struct {
			Distance  float64 `json:"distance"`
			PeakSpeed float64 `json:"peak_speed"`
		}

		petModel, err := a.server.DatabaseStore().Pets().FindByID(requestedPetID)
		if err != nil {
			if errors.Is(err, sql.ErrNoRows) {
				a.server.RespondError(w, r, http.StatusNotFound, nil)
				return
			}
			a.server.Logger().Printf("Database err: %v, Request ID: %v", err, requestID)
			a.server.RespondError(w, r, http.StatusInternalServerError, nil)
			return
		}

		if session.UserID != petModel.UserID {
			if !permissions.AnyRoleHavePermissions(
				session.Roles, permissions.Permissions().PetsPermission,
				permissions.Permissions().RolesPermission) {
				a.server.RespondError(w, r, http.StatusForbidden, nil)
				return
			}
		}

		rb := &requestBody{}
		if err := json.NewDecoder(r.Body).Decode(rb); err != nil {
			a.server.RespondError(w, r, http.StatusBadRequest, err)
			return
		}
		model := &models.Activity{
			PetID:           requestedPetID,
			RecordTimestamp: time.Now(),
			Distance:        rb.Distance,
			PeakSpeed:       rb.PeakSpeed,
		}
		if err := model.Validate(); err != nil {
			a.server.RespondError(w, r, http.StatusUnprocessableEntity, err)
			return
		}
		if err := a.server.DatabaseStore().Pets().CreateActivityRecord(model); err != nil {
			a.server.Logger().Printf("Database error: %v Request ID: %v", err, requestID)
			a.server.RespondError(w, r, http.StatusInternalServerError, nil)
			return
		}
		a.server.Respond(w, r, http.StatusCreated, nil)
	}
}

func (a *PetsAPI) ServeEatingRequest(w http.ResponseWriter, r *http.Request) {
	if r.Method == http.MethodOptions {
		return
	}
	requestID, _, err := a.server.GetAuthorizedRequestInfo(r)
	if err != nil {
		a.server.RespondError(w, r, http.StatusInternalServerError, nil)
		return
	}
	rawID, err := strconv.ParseInt(mux.Vars(r)["id"], 10, 64)
	if err != nil {
		a.server.RespondError(w, r, http.StatusBadRequest, exceptions.UnprocessableURIParam)
		return
	}
	requestedPetID := int(rawID)

	switch r.Method {
	case http.MethodGet:
		var dateT time.Time
		date := r.URL.Query().Get("date")
		if date == "" {
			dateT = time.Now()
		} else {
			dateT, err = time.Parse("2006-01-02", date)
			if err != nil {
				a.server.RespondError(w, r, http.StatusBadRequest, exceptions.UnprocessableURLQuery)
				return
			}
		}
		foodModels, err := a.server.DatabaseStore().Foods().GetPetsEatingsForDate(requestedPetID, dateT)
		if err != nil {
			if errors.Is(err, sql.ErrNoRows) {
				a.server.RespondError(w, r, http.StatusOK, nil)
				return
			}
			a.server.Logger().Printf("Database error: %v Request ID: %v", err, requestID)
			a.server.RespondError(w, r, http.StatusInternalServerError, nil)
			return
		}
		a.server.Respond(w, r, http.StatusOK, foodModels)

	case http.MethodPost:
		type requestBody struct {
			PetID  int `json:"pet_id"`
			FoodID int `json:"food_id"`
		}
		rb := &requestBody{}
		if err := json.NewDecoder(r.Body).Decode(rb); err != nil {
			a.server.RespondError(w, r, http.StatusBadRequest, err)
			return
		}
		eatingModel := &models.Eating{
			PetID:  rb.PetID,
			FoodID: rb.FoodID,
			Time:   time.Now(),
		}
		if err := a.server.DatabaseStore().Foods().AddPetEating(eatingModel); err != nil {
			a.server.RespondError(w, r, http.StatusUnprocessableEntity, nil)
			return
		}
		a.server.Respond(w, r, http.StatusCreated, nil)
	}
}
