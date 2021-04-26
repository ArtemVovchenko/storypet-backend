package api

import (
	"fmt"
	"github.com/ArtemVovchenko/storypet-backend/internal/app/server/api/exceptions"
	"github.com/ArtemVovchenko/storypet-backend/internal/pkg/filesutil"
	"github.com/gorilla/mux"
	"io"
	"net/http"
	"os"
)

type DatabaseAPI struct {
	server server
}

func NewDatabaseAPI(server server) *DatabaseAPI {
	return &DatabaseAPI{server: server}
}

func (a *DatabaseAPI) ConfigureRoutes(router *mux.Router) {
	sb := router.PathPrefix("/api/database").Subrouter()
	sb.Use(a.server.Middleware().Authentication.IsAuthorised)
	sb.Path("/dump/make").
		Name("Make Database Dump").
		Methods(http.MethodGet).
		HandlerFunc(
			a.server.Middleware().AccessPermission.DatabaseAccess(
				a.ServeDumpingRequest,
			),
		)

	sb.Path("/dump").
		Name("Make Database Dump").
		Methods(http.MethodGet, http.MethodPost).
		Handler(
			a.server.Middleware().AccessPermission.DatabaseAccess(
				a.ServeRootRequest,
			),
		)

	sb.Path("/dump/{fileName}").
		Name("Make Database Dump").
		Methods(http.MethodGet, http.MethodPut, http.MethodDelete).
		Handler(
			a.server.Middleware().AccessPermission.DatabaseAccess(
				a.ServeRequestByDumpName,
			),
		)
	sb.Path("/dump/download/{fileName}").
		Name("Download Database Dump").
		Methods(http.MethodGet).
		Handler(
			a.server.Middleware().AccessPermission.DatabaseAccess(
				a.ServeDumpDownloadRequest,
			),
		)
}

func (a *DatabaseAPI) ServeRootRequest(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodGet:
		dumpFiles, err := a.server.DatabaseStore().Dumps().SelectAll()
		if err != nil {
			a.server.RespondError(w, r, http.StatusInternalServerError, err)
			return
		}
		a.server.Respond(w, r, http.StatusOK, dumpFiles)
	case http.MethodPost:
		file, handler, err := r.FormFile("file")
		if err != nil {
			a.server.RespondError(w, r, http.StatusUnprocessableEntity, exceptions.DumpFileNotFoundInRequest)
			return
		}
		newFileModel, err := a.server.DatabaseStore().Dumps().InsertNewDumpFile(a.server.DumpFilesFolder())
		if err != nil {
			a.server.RespondError(w, r, http.StatusServiceUnavailable, exceptions.DumpSaveFailed)
			return
		}
		serverFile, err := os.OpenFile(newFileModel.FilePath, os.O_WRONLY|os.O_CREATE, os.ModePerm)
		if err != nil {
			_, _ = a.server.DatabaseStore().Dumps().DeleteByName(newFileModel.FileName)
			a.server.RespondError(w, r, http.StatusServiceUnavailable, exceptions.DumpSaveFailed)
			return
		}
		defer func() {
			_ = serverFile.Close()
		}()
		written, err := io.Copy(serverFile, file)
		if err != nil || written != handler.Size {
			_, _ = a.server.DatabaseStore().Dumps().DeleteByName(newFileModel.FileName)
			a.server.RespondError(w, r, http.StatusServiceUnavailable, exceptions.DumpSaveFailed)
			return
		}
		a.server.Respond(w, r, http.StatusOK, newFileModel)
	}
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

func (a *DatabaseAPI) ServeDumpingRequest(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodGet:
		dumpRecord, err := a.server.DatabaseStore().Dumps().Make(a.server.DumpFilesFolder())
		if err != nil {
			a.server.RespondError(w, r, http.StatusServiceUnavailable, exceptions.DatabaseDumpFailed)
			return
		}
		a.server.Respond(w, r, http.StatusOK, dumpRecord)
	}
}

func (a *DatabaseAPI) ServeDumpDownloadRequest(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodGet:
		dumpFileName := mux.Vars(r)["fileName"]
		dumpRecord, err := a.server.DatabaseStore().Dumps().SelectByName(dumpFileName)
		if err != nil || !filesutil.Exist(dumpRecord.FilePath) {
			a.server.Respond(w, r, http.StatusNotFound, nil)
			return
		}
		w.Header().Set("Content-Type", "application/octet-stream")
		w.Header().Set("Content-Disposition", fmt.Sprintf("attachment; filename=%s", dumpRecord.FileName))
		http.ServeFile(w, r, dumpRecord.FilePath)
	}
}
