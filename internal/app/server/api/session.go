package api

import (
	"database/sql"
	"encoding/json"
	"errors"
	"github.com/ArtemVovchenko/storypet-backend/internal/app/models"
	"github.com/ArtemVovchenko/storypet-backend/internal/app/server/api/exceptions"
	"github.com/ArtemVovchenko/storypet-backend/internal/app/sessions"
	"github.com/ArtemVovchenko/storypet-backend/internal/pkg/auth"
	"github.com/gorilla/mux"
	"net/http"
	"time"
)

type SessionAPI struct {
	server server
}

func NewSessionAPI(server server) *SessionAPI {
	return &SessionAPI{server: server}
}

func (a SessionAPI) ConfigureRoutes(router *mux.Router) {
	router.Path("/api/session/login").
		Name("User Login").
		Methods(http.MethodPost).
		Handler(
			http.HandlerFunc(a.ServeLoginRequest),
		)

	router.Path("/api/session/refresh").
		Name("Refresh token").
		Methods(http.MethodPost).
		Handler(
			http.HandlerFunc(a.ServeRefreshRequest),
		)

	router.Path("/api/session").
		Name("Session info").
		Methods(http.MethodGet).
		Handler(
			a.server.Middleware().Authentication.IsAuthorised(
				http.HandlerFunc(a.ServeSessionInfoRequest),
			),
		)

	router.Path("/api/session/logout").
		Name("User Logout").
		Methods(http.MethodPost).
		Handler(
			a.server.Middleware().Authentication.IsAuthorised(
				http.HandlerFunc(a.ServeLogoutRequest),
			),
		)

	router.Path("api/session/iot/login").
		Name("IoT Device Login").
		Methods(http.MethodPost).
		Handler(
			http.HandlerFunc(a.ServeIoTLoginRequest),
		)

	router.Path("api/session/iot/data").
		Name("IoT Device Data").
		Methods(http.MethodPost).
		Handler(
			http.HandlerFunc(a.ServeIoTDataRequest),
		)
}

func (a *SessionAPI) ServeLoginRequest(w http.ResponseWriter, r *http.Request) {
	if r.Method == http.MethodOptions {
		return
	}
	switch r.Method {
	case http.MethodPost:
		type requestBody struct {
			Email    string `json:"email"`
			Password string `json:"password"`
		}

		rb := &requestBody{}
		if err := json.NewDecoder(r.Body).Decode(rb); err != nil {
			a.server.RespondError(w, r, http.StatusBadRequest, err)
			return
		}

		u, err := a.server.DatabaseStore().Users().FindByAccountEmail(rb.Email)
		if err != nil {
			a.server.Respond(w, r, http.StatusUnauthorized, exceptions.IncorrectAuthData)
			return
		}

		if !u.ComparePasswords(rb.Password) {
			a.server.Respond(w, r, http.StatusUnauthorized, exceptions.IncorrectAuthData)
			return
		}

		token, err := auth.CreateToken(u.UserID)
		if err != nil {
			a.server.RespondError(w, r, http.StatusUnprocessableEntity, err)
			return
		}

		if err := a.createAndSaveSession(token, u.UserID); err != nil {
			a.server.RespondError(w, r, http.StatusInternalServerError, nil)
			return
		}

		a.server.Respond(w, r, http.StatusOK, map[string]string{
			"access":  token.AccessToken,
			"refresh": token.RefreshToken,
		})
	}
}

func (a *SessionAPI) ServeLogoutRequest(w http.ResponseWriter, r *http.Request) {
	if r.Method == http.MethodOptions {
		return
	}
	switch r.Method {
	case http.MethodPost:
		accessInfo, err := auth.ExtractAccessMeta(r)
		if err != nil {
			a.server.Respond(w, r, http.StatusUnauthorized, "Unauthorized")
			return
		}
		if err := a.deleteSession(accessInfo.AccessUUID); err != nil {
			a.server.Respond(w, r, http.StatusUnauthorized, "Unauthorized")
			return
		}
		a.server.Respond(w, r, http.StatusOK, "Logged Out")
	}
}

func (a *SessionAPI) ServeRefreshRequest(w http.ResponseWriter, r *http.Request) {
	if r.Method == http.MethodOptions {
		return
	}
	switch r.Method {
	case http.MethodPost:
		type requestBody struct {
			Refresh string `json:"refresh"`
		}

		rb := &requestBody{}
		if err := json.NewDecoder(r.Body).Decode(rb); err != nil {
			a.server.RespondError(w, r, http.StatusBadRequest, err)
			return
		}

		refreshMeta, err := auth.ExtractRefreshMeta(rb.Refresh)
		if err != nil {
			a.server.RespondError(w, r, http.StatusUnprocessableEntity, exceptions.IncorrectRefreshToken)
			return
		}

		userID, err := a.server.PersistentStore().GetUserIDByRefreshUUID(refreshMeta.RefreshUUID)
		if err != nil {
			a.server.Respond(w, r, http.StatusUnprocessableEntity, exceptions.IncorrectRefreshToken)
			return
		}

		err = a.server.PersistentStore().DeleteRefreshByUUID(refreshMeta.RefreshUUID)
		if err != nil {
			a.server.RespondError(w, r, http.StatusInternalServerError, nil)
			return
		}

		token, err := auth.CreateToken(userID)
		if err != nil {
			a.server.RespondError(w, r, http.StatusInternalServerError, nil)
			return
		}

		if err := a.createAndSaveSession(token, userID); err != nil {
			a.server.RespondError(w, r, http.StatusInternalServerError, nil)
			return
		}

		a.server.Respond(w, r, http.StatusOK, map[string]string{
			"access":  token.AccessToken,
			"refresh": token.RefreshToken,
		})
	}
}

func (a *SessionAPI) ServeSessionInfoRequest(w http.ResponseWriter, r *http.Request) {
	if r.Method == http.MethodOptions {
		return
	}
	switch r.Method {
	case http.MethodGet:
		accessInfo, err := auth.ExtractAccessMeta(r)
		if err != nil {
			a.server.Respond(w, r, http.StatusUnauthorized, "Unauthorized")
			return
		}
		session, err := a.server.PersistentStore().GetSessionInfo(accessInfo.AccessUUID)
		if err != nil {
			a.server.Respond(w, r, http.StatusUnauthorized, "Unauthorized")
			return
		}
		a.server.Respond(w, r, http.StatusOK, session)
	}

}

func (a *SessionAPI) createAndSaveSession(tokenPairMeta *auth.TokenPairInfo, userID int) error {
	userRoles, _ := a.server.DatabaseStore().Roles().SelectUserRoles(userID)
	newSession := &sessions.Session{
		UserID:      userID,
		RefreshUUID: tokenPairMeta.RefreshUUID,
		Roles:       userRoles,
	}
	return a.saveSession(tokenPairMeta, newSession)
}

func (a *SessionAPI) deleteSession(accessUUID string) error {
	session, err := a.server.PersistentStore().DeleteSessionInfo(accessUUID)
	if err != nil {
		return err
	}
	if err := a.server.PersistentStore().DeleteRefreshByUUID(session.RefreshUUID); err != nil {
		return err
	}
	return nil
}

func (a *SessionAPI) saveSession(tokenPairInfo *auth.TokenPairInfo, session *sessions.Session) error {
	if err := a.server.PersistentStore().SaveSessionInfo(
		tokenPairInfo.AccessUUID,
		session,
		time.Unix(tokenPairInfo.AccessExpires, 0)); err != nil {
		return err
	}
	if err := a.server.PersistentStore().SaveRefreshInfo(
		tokenPairInfo.RefreshUUID,
		session.UserID,
		time.Unix(tokenPairInfo.RefreshExpires, 0)); err != nil {
		return err
	}
	return nil
}

func (a *SessionAPI) ServeIoTLoginRequest(w http.ResponseWriter, r *http.Request) {
	if r.Method == http.MethodOptions {
		return
	}
	switch r.Method {
	case http.MethodPost:
		type requestBody struct {
			AccessSecret string `json:"access_secret"`
		}

		rb := &requestBody{}
		if err := json.NewDecoder(r.Body).Decode(rb); err != nil {
			a.server.RespondError(w, r, http.StatusBadRequest, err)
			return
		}

		deviceModel, err := a.server.DatabaseStore().IoTDevicesRepository().GetByAccessSecret(rb.AccessSecret)
		if err != nil {
			if errors.Is(err, sql.ErrNoRows) {
				a.server.RespondError(w, r, http.StatusNotFound, nil)
				return
			}
			a.server.Respond(w, r, http.StatusUnauthorized, nil)
			return
		}

		token, err := auth.CreateIoTToken(deviceModel.PetID)
		if err != nil {
			a.server.RespondError(w, r, http.StatusUnprocessableEntity, err)
			return
		}

		a.server.Respond(w, r, http.StatusOK, map[string]string{
			"access": token.Token,
		})
	}
}

func (a *SessionAPI) ServeIoTDataRequest(w http.ResponseWriter, r *http.Request) {
	if r.Method == http.MethodOptions {
		return
	}
	switch r.Method {
	case http.MethodPost:
		type requestBody struct {
			Distance  float64 `json:"distance"`
			MeanSpeed float64 `json:"mean_speed"`
		}

		rb := &requestBody{}
		if err := json.NewDecoder(r.Body).Decode(rb); err != nil {
			a.server.RespondError(w, r, http.StatusBadRequest, err)
			return
		}

		authTokenMeta, err := auth.ExtractIoTAccessMeta(r)
		if err != nil {
			a.server.Respond(w, r, http.StatusUnauthorized, nil)
			return
		}

		activity := &models.Activity{
			PetID:           authTokenMeta.PetID,
			RecordTimestamp: time.Now(),
			Distance:        rb.Distance,
			MeanSpeed:       rb.MeanSpeed,
		}
		if err := a.server.DatabaseStore().Pets().CreateActivityRecord(activity); err != nil {
			a.server.RespondError(w, r, http.StatusInternalServerError, err)
			return
		}
		a.server.Respond(w, r, http.StatusCreated, nil)
	}
}
