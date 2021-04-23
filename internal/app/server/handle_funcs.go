package server

import (
	"encoding/json"
	"errors"
	"github.com/ArtemVovchenko/storypet-backend/internal/app/models"
	"github.com/ArtemVovchenko/storypet-backend/internal/pkg/auth"
	"net/http"
)

var (
	errIncorrectAuthData     = errors.New("incorrect email or password")
	errIncorrectRefreshToken = errors.New("incorrect refresh token")
	errDatabaseDumpFailed    = errors.New("could not create database dump")
)

func (s *Server) handleTest() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if err := json.NewEncoder(w).Encode(map[string]string{"Status": "Test OK"}); err != nil {
			s.logger.Printf("Error processing request at `/api/users/test`: %s", err)
		}
	}
}

func (s *Server) handleLogin() http.HandlerFunc {
	type requestBody struct {
		Email    string `json:"email"`
		Password string `json:"password"`
	}
	return func(w http.ResponseWriter, r *http.Request) {
		rb := &requestBody{}
		if err := json.NewDecoder(r.Body).Decode(rb); err != nil {
			s.RespondError(w, r, http.StatusBadRequest, err)
			return
		}

		u, err := s.databaseStore.Users().FindByAccountEmail(rb.Email)
		if err != nil {
			s.Respond(w, r, http.StatusUnauthorized, errIncorrectAuthData)
			return
		}

		if !u.ComparePasswords(rb.Password) {
			s.Respond(w, r, http.StatusUnauthorized, errIncorrectAuthData)
			return
		}

		token, err := auth.CreateToken(u.UserID)
		if err != nil {
			s.RespondError(w, r, http.StatusUnprocessableEntity, err)
			return
		}

		if err := s.createAndSaveSession(token, u.UserID); err != nil {
			s.RespondError(w, r, http.StatusInternalServerError, nil)
			return
		}

		s.Respond(w, r, http.StatusOK, map[string]string{
			"access":  token.AccessToken,
			"refresh": token.RefreshToken,
		})
	}
}

func (s *Server) handleRefresh() http.HandlerFunc {
	type requestBody struct {
		Refresh string `json:"refresh"`
	}

	return func(w http.ResponseWriter, r *http.Request) {
		rb := &requestBody{}
		if err := json.NewDecoder(r.Body).Decode(rb); err != nil {
			s.RespondError(w, r, http.StatusBadRequest, err)
			return
		}

		refreshMeta, err := auth.ExtractRefreshMeta(rb.Refresh)
		if err != nil {
			s.RespondError(w, r, http.StatusUnprocessableEntity, errIncorrectRefreshToken)
			return
		}

		userID, err := s.persistentStore.GetUserIDByRefreshUUID(refreshMeta.RefreshUUID)
		if err != nil {
			s.Respond(w, r, http.StatusUnprocessableEntity, errIncorrectRefreshToken)
			return
		}

		err = s.persistentStore.DeleteRefreshByUUID(refreshMeta.RefreshUUID)
		if err != nil {
			s.errLogger.Println(err)
			s.RespondError(w, r, http.StatusInternalServerError, nil)
			return
		}

		token, err := auth.CreateToken(userID)
		if err != nil {
			s.errLogger.Println(err)
			s.RespondError(w, r, http.StatusInternalServerError, nil)
			return
		}

		if err := s.createAndSaveSession(token, userID); err != nil {
			s.errLogger.Println(err)
			s.RespondError(w, r, http.StatusInternalServerError, nil)
			return
		}

		s.Respond(w, r, http.StatusOK, map[string]string{
			"access":  token.AccessToken,
			"refresh": token.RefreshToken,
		})
	}
}

func (s *Server) handleLogout() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		accessInfo, err := auth.ExtractAccessMeta(r)
		if err != nil {
			s.Respond(w, r, http.StatusUnauthorized, "Unauthorized")
			return
		}
		if err := s.deleteSession(accessInfo.AccessUUID); err != nil {
			s.Respond(w, r, http.StatusUnauthorized, "Unauthorized")
			return
		}
		s.Respond(w, r, http.StatusOK, "Logged Out")
	}
}

func (s *Server) handleSessionInfo() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		accessInfo, err := auth.ExtractAccessMeta(r)
		if err != nil {
			s.Respond(w, r, http.StatusUnauthorized, "Unauthorized")
			return
		}
		session, err := s.persistentStore.GetSessionInfo(accessInfo.AccessUUID)
		if err != nil {
			s.Respond(w, r, http.StatusUnauthorized, "Unauthorized")
			return
		}
		s.Respond(w, r, http.StatusOK, session)
	}
}

func (s *Server) handleRegistration() http.HandlerFunc {
	type requestBody struct {
		AccountEmail string  `json:"account_email"`
		Password     string  `json:"password"`
		Username     string  `json:"username"`
		FullName     string  `json:"full_name"`
		BackupEmail  *string `json:"backup_email"`
		Location     *string `json:"location"`
	}

	return func(w http.ResponseWriter, r *http.Request) {
		rb := &requestBody{}
		if err := json.NewDecoder(r.Body).Decode(rb); err != nil {
			s.RespondError(w, r, http.StatusBadRequest, err)
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

		if _, err := s.databaseStore.Users().Create(u); err != nil {
			s.RespondError(w, r, http.StatusUnprocessableEntity, err)
			return
		}
		u.Sanitise()
		s.Respond(w, r, http.StatusCreated, u)
	}
}

func (s *Server) handleMakingDump() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if err := s.databaseStore.Dumps().Make(s.config.DatabaseDumpsDir); err != nil {
			s.RespondError(w, r, http.StatusServiceUnavailable, errDatabaseDumpFailed)
			return
		}
		s.Respond(w, r, http.StatusOK, nil)
	}
}
