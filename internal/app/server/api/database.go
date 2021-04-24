package api

import (
	"errors"
	"github.com/ArtemVovchenko/storypet-backend/internal/pkg/filesutil"
	"github.com/gorilla/mux"
	"net/http"
)

var errDatabaseDumpFailed = errors.New("could not create database dump")

type DatabaseAPI struct {
	server server
}

func NewDatabaseAPI(server server) *DatabaseAPI {
	return &DatabaseAPI{server: server}
}

func (a *DatabaseAPI) ConfigureRoutes(router *mux.Router) {
	router.Path("/api/database/dump/make").
		Name("Make Database Dump").
		Methods(http.MethodGet).
		HandlerFunc(
			a.server.Middleware().ResponseWriting.JSONBody(
				a.server.Middleware().Authentication.IsAuthorised(
					a.server.Middleware().AccessPermission.DatabaseAccess(
						a.ServeDumpingRequest,
					),
				),
			),
		)

	router.Path("/api/database/dump").
		Name("Make Database Dump").
		Methods(http.MethodGet).
		Handler(
			a.server.Middleware().ResponseWriting.JSONBody(
				a.server.Middleware().Authentication.IsAuthorised(
					a.server.Middleware().AccessPermission.DatabaseAccess(
						a.ServeEmptyRequest,
					),
				),
			),
		)

	router.Path("/api/database/dump/{fileName}").
		Name("Make Database Dump").
		Methods(http.MethodGet, http.MethodPut, http.MethodDelete).
		Handler(
			a.server.Middleware().ResponseWriting.JSONBody(
				a.server.Middleware().Authentication.IsAuthorised(
					a.server.Middleware().AccessPermission.DatabaseAccess(
						a.ServeRequestByDumpName,
					),
				),
			),
		)
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

// TODO: Вернуть обьект дампа
func (a *DatabaseAPI) ServeDumpingRequest(w http.ResponseWriter, r *http.Request) {
	if err := a.server.DatabaseStore().Dumps().Make(a.server.DumpFilesFolder()); err != nil {
		a.server.RespondError(w, r, http.StatusServiceUnavailable, errDatabaseDumpFailed)
		return
	}
	a.server.Respond(w, r, http.StatusOK, nil)
}
