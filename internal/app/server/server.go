package server

import (
	"github.com/gorilla/mux"
	"io"
	"log"
	"net/http"
)

type Server struct {
	config *Config
	logger *log.Logger
	router *mux.Router
}

func New(config *Config) *Server {
	return &Server{
		config: config,
		logger: log.New(config.LogOutStream, config.LogPrefix, config.LogFlags),
		router: mux.NewRouter(),
	}
}

func (s *Server) Start() error {
	s.ConfigureRouter()
	s.logger.Println("starting server")
	return http.ListenAndServe(s.config.BindAddr, s.router)
}

func (s *Server) ConfigureRouter() {
	s.router.HandleFunc("/api/users/test", s.handleTest())
}

func (s *Server) handleTest() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		_, err := io.WriteString(w, "Test OK")
		if err != nil {
			s.logger.Printf("Error processing request at `/api/users/test`: %s", err)
		}
	}
}
