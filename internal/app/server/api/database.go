package api

import (
	"github.com/ArtemVovchenko/storypet-backend/internal/app/store"
	"github.com/ArtemVovchenko/storypet-backend/internal/pkg/filesutil"
	"github.com/gorilla/mux"
	"net/http"
)

type server interface {
	Respond(w http.ResponseWriter, h *http.Request, code int, data interface{})
	RespondError(w http.ResponseWriter, h *http.Request, code int, err error)
	PersistentStore() store.PersistentStore
	DatabaseStore() store.DatabaseStore
	DumpFilesFolder() string
}

type DatabaseAPI struct {
	server server
}

func NewDatabaseAPI(server server) *DatabaseAPI {
	return &DatabaseAPI{server: server}
}

func (a *DatabaseAPI) ServeEmptyRequest(w http.ResponseWriter, r *http.Request) {
	dumpFiles, err := a.server.DatabaseStore().Dumps().SelectAll()
	if err != nil {
		a.server.RespondError(w, r, http.StatusInternalServerError, err)
		return
	}
	a.server.Respond(w, r, http.StatusOK, dumpFiles)
}

func (a *DatabaseAPI) ServeRequestByDumpName(w http.ResponseWriter, r *http.Request) {
	dumpFileName := mux.Vars(r)["fileName"]
	switch r.Method {
	case http.MethodGet:
		if !filesutil.Exist(a.server.DumpFilesFolder() + dumpFileName) {
			a.server.RespondError(w, r, http.StatusNotFound, nil)
			return
		}
		dumpFile, err := a.server.DatabaseStore().Dumps().SelectByName(dumpFileName)
		if err != nil {
			a.server.RespondError(w, r, http.StatusNotFound, nil)
			return
		}
		a.server.Respond(w, r, http.StatusOK, dumpFile)

	case http.MethodPut:
		if !filesutil.Exist(a.server.DumpFilesFolder() + dumpFileName) {
			a.server.RespondError(w, r, http.StatusNotFound, nil)
			return
		}
		if err := a.server.DatabaseStore().Dumps().Execute(a.server.DumpFilesFolder() + dumpFileName); err != nil {
			a.server.RespondError(w, r, http.StatusInternalServerError, nil)
			return
		}
		_ = filesutil.ClearDir(a.server.DumpFilesFolder())
		a.server.Respond(w, r, http.StatusOK, map[string]string{"Status": "Succeed rollback"})

	case http.MethodDelete:
		dumpFile, err := a.server.DatabaseStore().Dumps().DeleteByName(dumpFileName)
		if err != nil {
			a.server.RespondError(w, r, http.StatusNotFound, nil)
			return
		}
		filesutil.Delete(dumpFile.FilePath)
		a.server.Respond(w, r, http.StatusNoContent, dumpFile)
	}
}
