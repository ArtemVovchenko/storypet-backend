package server

import (
	"encoding/json"
	"github.com/ArtemVovchenko/storypet-backend/internal/app/configs"
	"github.com/ArtemVovchenko/storypet-backend/internal/app/middleware"
	"github.com/ArtemVovchenko/storypet-backend/internal/app/store"
	"github.com/ArtemVovchenko/storypet-backend/internal/pkg/auth"
	"github.com/go-redis/redis/v7"
	"github.com/gorilla/mux"
	"log"
	"net/http"
	"strconv"
	"time"
)

type Server struct {
	config     *configs.ServerConfig
	logger     *log.Logger
	errLogger  *log.Logger
	router     *mux.Router
	store      *store.Store
	middleware *middleware.Middleware
}

func New(config *configs.ServerConfig) *Server {
	return &Server{
		config:    config,
		logger:    log.New(config.LogOutStream, config.LogPrefix, config.LogFlags),
		errLogger: log.New(config.ErrLogOutStream, configs.SrvErrLogPrefix, configs.SrvErrLogFlags),
		router:    mux.NewRouter(),
	}
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
	s.router.Path("/api/users/test").Methods(http.MethodGet).HandlerFunc(s.handleTest()).Name("User Test")
	s.router.Path("/api/users/login/test").Methods(http.MethodGet).HandlerFunc(s.middleware.Authentication.IsAuthorised(s.handleTest())).Name("Authorized User Test")

	s.router.Path("/api/users/login").Methods(http.MethodPost).HandlerFunc(s.handleLogin()).Name("User Login")
	s.router.Path("/api/users/refresh").Methods(http.MethodPost).HandlerFunc(s.handleRefresh()).Name("Refresh token")
	s.router.Path("/api/users/logout").Methods(http.MethodPost).HandlerFunc(s.middleware.Authentication.IsAuthorised(s.handleLogout())).Name("User Logout")
	s.router.Path("/api/users").Methods(http.MethodPost).HandlerFunc(s.handleRegistration()).Name("User Register")
}

func (s *Server) configureStore() error {
	st := store.New(s.config.Database)
	if err := st.Open(); err != nil {
		return err
	}

	s.store = st
	return nil
}

func (s *Server) saveJWTTokens(tInfo *auth.TokenPairInfo, userID int) error {
	at := time.Unix(tInfo.AccessExpires, 0)
	rt := time.Unix(tInfo.RefreshExpires, 0)
	if err := s.store.RedisClient().Set(tInfo.AccessUUID, strconv.Itoa(userID), at.Sub(time.Now())).Err(); err != nil {
		return err
	}
	if err := s.store.RedisClient().Set(tInfo.RefreshUUID, strconv.Itoa(userID), rt.Sub(time.Now())).Err(); err != nil {
		return err
	}
	return nil
}

func (s *Server) RedisStorage() *redis.Client {
	return s.store.RedisClient()
}
