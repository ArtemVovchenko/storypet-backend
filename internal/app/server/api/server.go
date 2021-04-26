package api

import (
	"github.com/ArtemVovchenko/storypet-backend/internal/app/middleware"
	"github.com/ArtemVovchenko/storypet-backend/internal/app/store"
	"log"
	"net/http"
)

type server interface {
	Respond(w http.ResponseWriter, h *http.Request, code int, data interface{})
	RespondError(w http.ResponseWriter, h *http.Request, code int, err error)

	Logger() *log.Logger

	PersistentStore() store.PersistentStore
	DatabaseStore() store.DatabaseStore

	Middleware() middleware.Middleware

	DumpFilesFolder() string
}
