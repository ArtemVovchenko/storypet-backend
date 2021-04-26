package server

import (
	"encoding/json"
	"github.com/ArtemVovchenko/storypet-backend/internal/app/configs"
	"github.com/ArtemVovchenko/storypet-backend/internal/app/middleware"
	"github.com/ArtemVovchenko/storypet-backend/internal/app/server/api"
	"github.com/ArtemVovchenko/storypet-backend/internal/app/store"
	"github.com/gorilla/handlers"
	"github.com/gorilla/mux"
	"log"
	"net/http"
	"os"
)

type Server struct {
	config    *configs.ServerConfig
	logger    *log.Logger
	errLogger *log.Logger
	router    *mux.Router

	databaseStore   store.DatabaseStore
	persistentStore store.PersistentStore

	middleware middleware.Middleware

	databaseAPI *api.DatabaseAPI
	sessionAPI  *api.SessionAPI
	userAPI     *api.UserAPI
}

func New() *Server {
	config := configs.NewServerConfig()
	server := &Server{
		config:    config,
		logger:    log.New(config.LogOutStream, config.LogPrefix, config.LogFlags),
		errLogger: log.New(config.ErrLogOutStream, configs.SrvErrLogPrefix, configs.SrvErrLogFlags),
		router:    mux.NewRouter(),
	}
	server.middleware = middleware.New(server)
	server.databaseAPI = api.NewDatabaseAPI(server)
	server.sessionAPI = api.NewSessionAPI(server)
	server.userAPI = api.NewUserAPI(server)
	return server
}

func (s *Server) Start() error {
	s.configureRouter()
	if err := s.configureStore(); err != nil {
		return err
	}
	s.logger.Println("starting server")
	if s.config.BindAddr == "" {
		log.Fatalln("$PORT is not specified")
	}
	return http.ListenAndServe(":"+s.config.BindAddr, s.router)
}

func (s *Server) PersistentStore() store.PersistentStore {
	return s.persistentStore
}

func (s *Server) DatabaseStore() store.DatabaseStore {
	return s.databaseStore
}

func (s *Server) Middleware() middleware.Middleware {
	return s.middleware
}

func (s *Server) DumpFilesFolder() string {
	return s.config.DatabaseDumpsDir
}

func (s *Server) Logger() *log.Logger {
	return s.logger
}

func (s *Server) RespondError(w http.ResponseWriter, r *http.Request, statusCode int, err error) {
	if err != nil {
		s.Respond(w, r, statusCode, map[string]string{"error": err.Error()})
	} else {
		s.Respond(w, r, statusCode, err)
	}
}

func (s *Server) Respond(w http.ResponseWriter, _ *http.Request, statusCode int, data interface{}) {
	w.WriteHeader(statusCode)
	if data != nil {
		_ = json.NewEncoder(w).Encode(data)
	}
}

func (s *Server) configureRouter() {
	s.router.Use(handlers.CORS(handlers.AllowedOrigins([]string{"*"})))
	s.router.Use(s.middleware.InfoMiddleware.MarkRequest)
	s.router.Use(s.middleware.InfoMiddleware.LogRequest)
	s.router.Use(s.middleware.ResponseWriting.JSONBody)

	s.databaseAPI.ConfigureRoutes(s.router)
	s.sessionAPI.ConfigureRoutes(s.router)
	s.userAPI.ConfigureRoutes(s.router)
}

func (s *Server) configureStore() error {
	databaseLogger := log.New(os.Stdout, "DATABASE: ", log.LstdFlags)
	database := store.NewDatabaseStore(databaseLogger)
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
