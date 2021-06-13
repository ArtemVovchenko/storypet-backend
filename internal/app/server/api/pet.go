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

	sb.Path("/{id:[0-9]+}/reports").
		Name("Pets health reports Request").
		Methods(http.MethodGet, http.MethodPost).
		HandlerFunc(a.ServeReportRequest)

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

	sb.Path("/{id:[0-9]+}/statistic").
		Name("Pet Statistic").
		Methods(http.MethodGet).
		HandlerFunc(a.ServeStatisticRequest)

	sb.Path("/{id:[0-9]+}/statistic/today").
		Name("Pet this day statistic").
		Methods(http.MethodGet).
		HandlerFunc(a.ServeTodayStatisticRequest)
}

func (a *PetsAPI) ServeRootRequest(w http.ResponseWriter, r *http.Request) {
	if r.Method == http.MethodOptions {
		return
	}
	requestID, _, err := a.server.GetAuthorizedRequestInfo(r)
	if err != nil {
		a.server.RespondError(w, r, http.StatusInternalServerError, nil)
		return
	}

	type responsePetEntity struct {
		PetID          int             `json:"pet_id"`
		Name           string          `json:"name"`
		Owner          *models.User    `json:"owner"`
		PetType        *models.PetType `json:"pet_type"`
		Mother         *models.Pet     `json:"mother,omitempty"`
		Father         *models.Pet     `json:"father,omitempty"`
		MotherVerified bool            `json:"mother_verified"`
		FatherVerified bool            `json:"father_verified"`
		Breed          string          `json:"breed,omitempty"`
		FamilyName     string          `json:"family_name,omitempty"`
	}
	responsePetEntitiesBuilder := func(pets []models.Pet) ([]responsePetEntity, error) {
		responseEntities := make([]responsePetEntity, len(pets), len(pets))
		for idx := range pets {
			owner, err := a.server.DatabaseStore().Users().FindByID(pets[idx].UserID)
			petType, err := a.server.DatabaseStore().Pets().FindTypeByID(pets[idx].PetType)
			var mother, father *models.Pet
			var breed, familyName string
			if pets[idx].FatherID != nil && pets[idx].FatherID.Valid {
				father, err = a.server.DatabaseStore().Pets().FindByID(int(pets[idx].FatherID.Int64))
			}
			if pets[idx].MotherID != nil && pets[idx].MotherID.Valid {
				mother, err = a.server.DatabaseStore().Pets().FindByID(int(pets[idx].MotherID.Int64))
			}
			if pets[idx].Breed != nil && pets[idx].Breed.Valid {
				breed = pets[idx].Breed.String
			}
			if pets[idx].FamilyName != nil && pets[idx].FamilyName.Valid {
				familyName = pets[idx].FamilyName.String
			}
			if err != nil {
				return nil, err
			}

			responseEntities[idx] = responsePetEntity{
				PetID:          pets[idx].PetID,
				Name:           pets[idx].Name,
				Owner:          owner,
				PetType:        petType,
				Mother:         mother,
				Father:         father,
				MotherVerified: pets[idx].MotherVerified,
				FatherVerified: pets[idx].FatherVerified,
				Breed:          breed,
				FamilyName:     familyName,
			}
		}
		return responseEntities, nil
	}

	switch r.Method {
	case http.MethodGet:
		userID := r.URL.Query().Get("user_id")
		if userID != "" {
			rawUserID, err := strconv.ParseInt(userID, 10, 64)
			if err != nil {
				a.server.Logger().Printf("Parsing int err: %v, Request ID: %v", err, requestID)
				a.server.RespondError(w, r, http.StatusBadRequest, nil)
				return
			}
			pets, err := a.server.DatabaseStore().Pets().SelectByUserID(int(rawUserID))
			if err != nil {
				a.server.Logger().Printf("Database err: %v, Request ID: %v", err, requestID)
				a.server.RespondError(w, r, http.StatusInternalServerError, nil)
				return
			}
			responseEntities, err := responsePetEntitiesBuilder(pets)
			if err != nil {
				a.server.Logger().Printf("Database err: %v, Request ID: %v", err, requestID)
				a.server.RespondError(w, r, http.StatusInternalServerError, nil)
				return
			}
			a.server.Respond(w, r, http.StatusOK, responseEntities)
			return
		}
		pets, err := a.server.DatabaseStore().Pets().SelectAll()
		if err != nil {
			a.server.Logger().Printf("Database err: %v, Request ID: %v", err, requestID)
			a.server.RespondError(w, r, http.StatusInternalServerError, nil)
			return
		}
		responseEntities, err := responsePetEntitiesBuilder(pets)
		if err != nil {
			a.server.Logger().Printf("Database err: %v, Request ID: %v", err, requestID)
			a.server.RespondError(w, r, http.StatusInternalServerError, nil)
			return
		}
		a.server.Respond(w, r, http.StatusOK, responseEntities)

	case http.MethodPost:
		type requestBody struct {
			Name       string  `json:"name"`
			PetType    int     `json:"pet_type"`
			Breed      *string `json:"breed"`
			FamilyName *string `json:"family_name"`
			UserID     *int    `json:"owner_id"`
			MotherID   *int    `json:"mother_id"`
			FatherID   *int    `json:"father_id"`
		}
		rb := &requestBody{}
		if err := json.NewDecoder(r.Body).Decode(rb); err != nil {
			a.server.RespondError(w, r, http.StatusBadRequest, err)
			return
		}
		if rb.UserID == nil {
			a.server.RespondError(w, r, http.StatusUnprocessableEntity, errors.New("pet owner is not specified"))
			return
		}
		newPetModel := &models.Pet{
			Name:    rb.Name,
			PetType: rb.PetType,
			UserID:  *rb.UserID,
		}

		newPetModel.SetSpecifiedBreed(rb.Breed)
		newPetModel.SetSpecifiedFamilyName(rb.FamilyName)
		newPetModel.SetSpecifiedMotherID(rb.MotherID)
		newPetModel.SetSpecifiedFatherID(rb.FatherID)
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

	type responsePetEntity struct {
		PetID          int             `json:"pet_id"`
		Name           string          `json:"name"`
		Owner          *models.User    `json:"owner"`
		PetType        *models.PetType `json:"pet_type"`
		Mother         *models.Pet     `json:"mother,omitempty"`
		Father         *models.Pet     `json:"father,omitempty"`
		MotherVerified bool            `json:"mother_verified"`
		FatherVerified bool            `json:"father_verified"`
		Breed          string          `json:"breed,omitempty"`
		FamilyName     string          `json:"family_name,omitempty"`
	}

	responsePetEntityBuilder := func(pet *models.Pet) (*responsePetEntity, error) {
		owner, err := a.server.DatabaseStore().Users().FindByID(pet.UserID)
		petType, err := a.server.DatabaseStore().Pets().FindTypeByID(pet.PetType)
		var mother, father *models.Pet
		var breed, familyName string
		if pet.FatherID != nil && pet.FatherID.Valid {
			father, err = a.server.DatabaseStore().Pets().FindByID(int(pet.FatherID.Int64))
		}
		if pet.MotherID != nil && pet.MotherID.Valid {
			mother, err = a.server.DatabaseStore().Pets().FindByID(int(pet.MotherID.Int64))
		}
		if pet.Breed != nil && pet.Breed.Valid {
			breed = pet.Breed.String
		}
		if pet.FamilyName != nil && pet.FamilyName.Valid {
			familyName = pet.FamilyName.String
		}
		if err != nil {
			return nil, err
		}

		responseEntity := &responsePetEntity{
			PetID:          pet.PetID,
			Name:           pet.Name,
			Owner:          owner,
			PetType:        petType,
			Mother:         mother,
			Father:         father,
			MotherVerified: pet.MotherVerified,
			FatherVerified: pet.FatherVerified,
			Breed:          breed,
			FamilyName:     familyName,
		}
		return responseEntity, nil
	}

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
		responseEntity, err := responsePetEntityBuilder(petModel)
		if err != nil {
			a.server.Logger().Printf("Database err: %v, Request ID: %v", err, requestID)
			a.server.RespondError(w, r, http.StatusInternalServerError, nil)
			return
		}
		a.server.Respond(w, r, http.StatusOK, responseEntity)

	case http.MethodPut:
		type requestBody struct {
			Name       string  `json:"name"`
			PetType    int     `json:"pet_type"`
			Breed      *string `json:"breed"`
			FamilyName *string `json:"family_name"`
			UserID     *int    `json:"owner_id"`
			FatherID   *int    `json:"father_id"`
			MotherID   *int    `json:"mother_id"`
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
		newPetModel.SetSpecifiedFatherID(rb.FatherID)
		newPetModel.SetSpecifiedMotherID(rb.MotherID)
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

func (a *PetsAPI) ServeReportRequest(w http.ResponseWriter, r *http.Request) {
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
	requestedID := int(rawID)

	switch r.Method {
	case http.MethodGet:
		type commentResponseEntity struct {
			Report       *models.PetHealthReport `json:"report"`
			Veterinarian *models.User            `json:"creator"`
		}
		responseEntitiesCreator := func(reportModels []models.PetHealthReport) ([]commentResponseEntity, error) {
			responseEntities := make([]commentResponseEntity, len(reportModels), len(reportModels))
			for idx := range reportModels {
				creator, err := a.server.DatabaseStore().Users().FindByID(reportModels[idx].VeterinarianID)
				if err != nil {
					return nil, err
				}
				responseEntities[idx] = commentResponseEntity{
					Report:       &reportModels[idx],
					Veterinarian: creator,
				}
			}
			return responseEntities, nil
		}

		reportModels, err := a.server.DatabaseStore().Pets().GetAllPetHealthReports(requestedID)
		if err != nil {
			if errors.Is(err, sql.ErrNoRows) {
				a.server.RespondError(w, r, http.StatusNotFound, nil)
				return
			}
			a.server.Logger().Printf("DATABASE Error: %v RequestID: %v", err, requestID)
			a.server.RespondError(w, r, http.StatusInternalServerError, nil)
			return
		}

		responseModels, err := responseEntitiesCreator(reportModels)
		if err != nil {
			if errors.Is(err, sql.ErrNoRows) {
				a.server.RespondError(w, r, http.StatusNotFound, nil)
				return
			}
			a.server.Logger().Printf("DATABASE Error: %v RequestID: %v", err, requestID)
			a.server.RespondError(w, r, http.StatusInternalServerError, nil)
			return
		}

		a.server.Respond(w, r, http.StatusOK, responseModels)

	case http.MethodPost:
		type requestBody struct {
			UserID     int     `json:"creator_id"`
			Conclusion string  `json:"conclusion"`
			Comments   *string `json:"comment"`
		}

		rb := &requestBody{}
		if err := json.NewDecoder(r.Body).Decode(rb); err != nil {
			a.server.RespondError(w, r, http.StatusBadRequest, err)
			return
		}

		commentModel := &models.PetHealthReport{
			PetID:            requestedID,
			ReportTimestamp:  time.Now(),
			VeterinarianID:   rb.UserID,
			ReportConclusion: rb.Conclusion,
		}
		commentModel.SetSpecifiedReportComments(rb.Comments)
		commentModel.BeforeCreate()
		if err := a.server.DatabaseStore().Pets().CreatePetHealthReport(commentModel); err != nil {
			a.server.Logger().Printf("DATABASE Error: %v RequestID: %v", err, requestID)
			a.server.RespondError(w, r, http.StatusInternalServerError, nil)
			return
		}
		a.server.Respond(w, r, http.StatusCreated, nil)
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
			MeanSpeed float64 `json:"mean_speed"`
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
			MeanSpeed:       rb.MeanSpeed,
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

	type eatingsResponseEntity struct {
		PortionWeight float64      `json:"portion_weight"`
		EatingTime    time.Time    `json:"eating_time"`
		Food          *models.Food `json:"food"`
	}

	eatingsResponseEntityBuilder := func(eatingModels []models.Eating) ([]eatingsResponseEntity, error) {
		responseModels := make([]eatingsResponseEntity, len(eatingModels), len(eatingModels))
		for idx := range eatingModels {
			foodModel, err := a.server.DatabaseStore().Foods().FindByID(eatingModels[idx].FoodID)
			if err != nil {
				return nil, err
			}
			responseModels[idx] = eatingsResponseEntity{
				PortionWeight: eatingModels[idx].PortionWeight,
				EatingTime:    eatingModels[idx].Time,
				Food:          foodModel,
			}
		}
		return responseModels, nil
	}

	switch r.Method {
	case http.MethodGet:
		eatings, err := a.server.DatabaseStore().Foods().GetPetsEatings(requestedPetID)
		if err != nil {
			if errors.Is(err, sql.ErrNoRows) {
				a.server.RespondError(w, r, http.StatusOK, nil)
				return
			}
			a.server.Logger().Printf("Database error: %v Request ID: %v", err, requestID)
			a.server.RespondError(w, r, http.StatusInternalServerError, nil)
			return
		}
		responseModels, err := eatingsResponseEntityBuilder(eatings)
		if err != nil {
			if errors.Is(err, sql.ErrNoRows) {
				a.server.RespondError(w, r, http.StatusOK, nil)
				return
			}
			a.server.Logger().Printf("Database error: %v Request ID: %v", err, requestID)
			a.server.RespondError(w, r, http.StatusInternalServerError, nil)
			return
		}
		a.server.Respond(w, r, http.StatusOK, responseModels)

	case http.MethodPost:
		type requestBody struct {
			FoodID        int     `json:"food_id"`
			PortionWeight float64 `json:"portion_weight"`
		}
		rb := &requestBody{}
		if err := json.NewDecoder(r.Body).Decode(rb); err != nil {
			a.server.RespondError(w, r, http.StatusBadRequest, err)
			return
		}
		eatingModel := &models.Eating{
			PetID:         requestedPetID,
			FoodID:        rb.FoodID,
			Time:          time.Now(),
			PortionWeight: rb.PortionWeight,
		}
		if err := a.server.DatabaseStore().Foods().AddPetEating(eatingModel); err != nil {
			a.server.RespondError(w, r, http.StatusUnprocessableEntity, nil)
			return
		}
		a.server.Respond(w, r, http.StatusCreated, nil)
	}
}

func (a *PetsAPI) ServeStatisticRequest(w http.ResponseWriter, r *http.Request) {
	type responseBody struct {
		FoodCalories  []models.FoodCaloriesReport  `json:"food_calories"`
		RERCalories   []models.RERCaloriesReport   `json:"rer_calories"`
		Anthropometry []models.AnthropometryReport `json:"anthropometry"`
		Activity      []models.ActivityReport      `json:"activity"`
	}

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

	foodCalories, rerCalories, anthropometry, activity, err := a.server.DatabaseStore().Pets().GetPetStatistics(requestedPetID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			a.server.RespondError(w, r, http.StatusNotFound, nil)
			return
		}
		a.server.Logger().Printf("Database error: %v RequestID: %v", err, requestID)
		a.server.RespondError(w, r, http.StatusInternalServerError, err)
		return
	}
	rb := &responseBody{
		FoodCalories:  foodCalories,
		RERCalories:   rerCalories,
		Anthropometry: anthropometry,
		Activity:      activity,
	}
	a.server.Respond(w, r, http.StatusOK, rb)
}

func (a *PetsAPI) ServeTodayStatisticRequest(w http.ResponseWriter, r *http.Request) {
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

	dayStatistics, err := a.server.DatabaseStore().Pets().GetPetDateStatistics(requestedPetID, time.Now())
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			a.server.RespondError(w, r, http.StatusNotFound, nil)
			return
		}
		a.server.Logger().Printf("Database error: %v RequestID: %v", err, requestID)
		a.server.RespondError(w, r, http.StatusInternalServerError, err)
		return
	}
	a.server.Respond(w, r, http.StatusOK, dayStatistics)
}
