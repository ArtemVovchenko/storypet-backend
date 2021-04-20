package server

import (
	"encoding/json"
	"github.com/ArtemVovchenko/storypet-backend/internal/app/configs"
	"github.com/ArtemVovchenko/storypet-backend/internal/app/middleware"
	"github.com/ArtemVovchenko/storypet-backend/internal/app/sessions"
	"github.com/ArtemVovchenko/storypet-backend/internal/app/store"
	"github.com/ArtemVovchenko/storypet-backend/internal/pkg/auth"
	"github.com/gorilla/mux"
	"log"
	"net/http"
	"time"
)

type Server struct {
	config          *configs.ServerConfig
	logger          *log.Logger
	errLogger       *log.Logger
	router          *mux.Router
	databaseStore   store.DatabaseStore
	persistentStore store.PersistentStore
	middleware      *middleware.Middleware
}

func New() *Server {
	config := configs.NewServerConfig()
	return &Server{
		config:    config,
		logger:    log.New(config.LogOutStream, config.LogPrefix, config.LogFlags),
		errLogger: log.New(config.ErrLogOutStream, configs.SrvErrLogPrefix, configs.SrvErrLogFlags),
		router:    mux.NewRouter(),
	}
}

func (s *Server) PersistentStore() store.PersistentStore {
	return s.persistentStore
}

func (s *Server) Start() error {
	s.configureMiddleware()
	s.configureRouter()
	if err := s.configureStore(); err != nil {
		return err
	}
	s.logger.Println("starting server")
	return http.ListenAndServe(s.config.BindAddr, s.router)
}

func (s *Server) RespondError(w http.ResponseWriter, r *http.Request, statusCode int, err error) {
	s.Respond(w, r, statusCode, map[string]string{"error": err.Error()})
}

func (s *Server) Respond(w http.ResponseWriter, r *http.Request, statusCode int, data interface{}) {
	w.WriteHeader(statusCode)
	if data != nil {
		_ = json.NewEncoder(w).Encode(data)
	}
}

func (s *Server) configureMiddleware() {
	s.middleware = middleware.New(s)
}

func (s *Server) configureRouter() {
	// Test Requests
	s.router.Path("/api/users/test").
		Methods(http.MethodGet).
		HandlerFunc(s.handleTest()).Name("User Test")
	s.router.Path("/api/users/login/test").
		Methods(http.MethodGet).
		HandlerFunc(s.middleware.Authentication.IsAuthorised(s.handleTest())).
		Name("Authorized User Test")

	s.router.Path("/api/users/login").
		Methods(http.MethodPost).
		HandlerFunc(s.handleLogin()).
		Name("User Login")
	s.router.Path("/api/users/refresh").
		Methods(http.MethodPost).
		HandlerFunc(s.handleRefresh()).
		Name("Refresh token")
	s.router.Path("/api/users/session").
		Methods(http.MethodGet).
		HandlerFunc(s.middleware.Authentication.IsAuthorised(s.handleSessionInfo())).
		Name("Session info")
	s.router.Path("/api/users/logout").
		Methods(http.MethodPost).
		HandlerFunc(s.middleware.Authentication.IsAuthorised(s.handleLogout())).
		Name("User Logout")
	s.router.Path("/api/users").
		Methods(http.MethodPost).
		HandlerFunc(s.handleRegistration()).
		Name("User Register")
}

func (s *Server) configureStore() error {
	database := store.NewDatabaseStore()
	if err := database.Open(); err != nil {
		return err
	}
	s.databaseStore = database

	persistentDatabase := store.NewPersistentStore()
	if err := persistentDatabase.Open(); err != nil {
		return err
	}
	s.persistentStore = persistentDatabase
	return nil
}

func (s *Server) Store() store.DatabaseStore {
	return s.databaseStore
}

func (s *Server) createAndSaveSession(tokenPairMeta *auth.TokenPairInfo, userID int) error {
	userRoles, _ := s.databaseStore.Roles().SelectUserRoles(userID)
	newSession := &sessions.Session{
		UserID:      userID,
		RefreshUUID: tokenPairMeta.RefreshUUID,
		Roles:       userRoles,
	}
	return s.saveSession(tokenPairMeta, newSession)
}

func (s *Server) deleteSession(accessUUID string) error {
	session, err := s.persistentStore.DeleteSessionInfo(accessUUID)
	if err != nil {
		return err
	}
	if err := s.persistentStore.DeleteRefreshByUUID(session.RefreshUUID); err != nil {
		return err
	}
	return nil
}

func (s *Server) saveSession(tokenPairInfo *auth.TokenPairInfo, session *sessions.Session) error {
	if err := s.persistentStore.SaveSessionInfo(
		tokenPairInfo.AccessUUID,
		session,
		time.Unix(tokenPairInfo.AccessExpires, 0)); err != nil {
		return err
	}
	if err := s.persistentStore.SaveRefreshInfo(
		tokenPairInfo.RefreshUUID,
		session.UserID,
		time.Unix(tokenPairInfo.RefreshExpires, 0)); err != nil {
		return err
	}
	return nil
}
