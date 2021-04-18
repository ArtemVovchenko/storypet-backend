package server

import (
	"encoding/json"
	"github.com/ArtemVovchenko/storypet-backend/internal/app/configs"
	"github.com/ArtemVovchenko/storypet-backend/internal/app/models"
	"github.com/ArtemVovchenko/storypet-backend/internal/app/store"
	"github.com/gorilla/mux"
	"io"
	"log"
	"net/http"
)

type Server struct {
	config *configs.ServerConfig
	logger *log.Logger
	router *mux.Router
	store  *store.Store
}

func New(config *configs.ServerConfig) *Server {
	return &Server{
		config: config,
		logger: log.New(config.LogOutStream, config.LogPrefix, config.LogFlags),
		router: mux.NewRouter(),
	}
}

func (s *Server) Start() error {
	s.configureRouter()
	if err := s.configureStore(); err != nil {
		return err
	}
	s.logger.Println("starting server")
	return http.ListenAndServe(s.config.BindAddr, s.router)
}

func (s *Server) configureRouter() {
	s.router.Path("/api/users/test").Methods(http.MethodGet).HandlerFunc(s.handleTest()).Name("User Test")
	s.router.Path("/api/users/login").Methods(http.MethodPost).HandlerFunc(nil).Name("User Login ")
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

func (s *Server) handleTest() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		_, err := io.WriteString(w, "Test OK")
		if err != nil {
			s.logger.Printf("Error processing request at `/api/users/test`: %s", err)
		}
	}
}

func (s *Server) handleLogin() http.HandlerFunc {

	return nil
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
			s.respondError(w, r, http.StatusBadRequest, err)
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
			s.respondError(w, r, http.StatusUnprocessableEntity, err)
			return
		}
		u.Sanitise()
		s.respond(w, r, http.StatusCreated, u)
	}
}

func (s *Server) respondError(w http.ResponseWriter, r *http.Request, statusCode int, err error) {
	s.respond(w, r, statusCode, map[string]string{"error": err.Error()})
}

func (s *Server) respond(w http.ResponseWriter, r *http.Request, statusCode int, data interface{}) {
	w.WriteHeader(statusCode)
	if data != nil {
		_ = json.NewEncoder(w).Encode(data)
	}
}
