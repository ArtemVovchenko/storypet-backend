package server

import (
	"github.com/ArtemVovchenko/storypet-backend/internal/app/configs"
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
	s.router.HandleFunc("/api/users/test", s.handleTest())
	//s.router.HandleFunc("/api/user/login", s.handleLogin())
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
