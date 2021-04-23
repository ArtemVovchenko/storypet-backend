package api

import (
	"github.com/ArtemVovchenko/storypet-backend/internal/app/middleware"
	"github.com/ArtemVovchenko/storypet-backend/internal/app/store"
	"net/http"
)

type server interface {
	Respond(w http.ResponseWriter, h *http.Request, code int, data interface{})
	RespondError(w http.ResponseWriter, h *http.Request, code int, err error)

	PersistentStore() store.PersistentStore
	DatabaseStore() store.DatabaseStore

	Middleware() middleware.Middleware

	DumpFilesFolder() string
}
