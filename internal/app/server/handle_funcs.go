package server

import (
	"encoding/json"
	"github.com/ArtemVovchenko/storypet-backend/internal/app/models"
	"github.com/ArtemVovchenko/storypet-backend/internal/pkg/auth"
	"io"
	"net/http"
	"strconv"
)

func (s *Server) handleTest() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		_, err := io.WriteString(w, "Test OK")
		if err != nil {
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

		u, err := s.store.Users().FindByAccountEmail(rb.Email)
		if err != nil {
			s.Respond(w, r, http.StatusUnauthorized, map[string]string{"error": "Invalid email address or password"})
			return
		}

		if !u.ComparePasswords(rb.Password) {
			s.Respond(w, r, http.StatusUnauthorized, map[string]string{"error": "Invalid email address or password"})
			return
		}

		token, err := auth.CreateToken(u.UserID)
		if err != nil {
			s.RespondError(w, r, http.StatusUnprocessableEntity, err)
			return
		}

		if err := s.saveJWTTokens(token, u.UserID); err != nil {
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
			s.Respond(w, r, http.StatusUnprocessableEntity, "Can not validate refresh token")
			return
		}

		userIDStr, err := s.RedisStorage().Get(refreshMeta.RefreshUUID).Result()
		if err != nil {
			s.Respond(w, r, http.StatusUnprocessableEntity, "Invalid refresh token")
			return
		}

		deleted, err := s.RedisStorage().Del(refreshMeta.RefreshUUID).Result()
		if err != nil || deleted == 0 {
			s.errLogger.Println(err)
			s.RespondError(w, r, http.StatusInternalServerError, nil)
			return
		}

		userID, err := strconv.ParseInt(userIDStr, 10, 64)
		token, err := auth.CreateToken(int(userID))
		if err != nil {
			s.errLogger.Println(err)
			s.RespondError(w, r, http.StatusInternalServerError, nil)
			return
		}

		if err := s.saveJWTTokens(token, int(userID)); err != nil {
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
		deleted, err := s.RedisStorage().Del(accessInfo.AccessUUID).Result()
		if err != nil || deleted == 0 {
			s.Respond(w, r, http.StatusUnauthorized, "Unauthorized")
			return
		}
		s.Respond(w, r, http.StatusOK, "Logged Out")
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

		if _, err := s.store.Users().Create(u); err != nil {
			s.RespondError(w, r, http.StatusUnprocessableEntity, err)
			return
		}
		u.Sanitise()
		s.Respond(w, r, http.StatusCreated, u)
	}
}
