package sqlxstore

import (
	"bytes"
	"fmt"
	"github.com/ArtemVovchenko/storypet-backend/internal/app/models"
	"github.com/ArtemVovchenko/storypet-backend/internal/pkg/filesutil"
	"github.com/jmoiron/sqlx"
	"github.com/twinj/uuid"
	"io/ioutil"
	"os/exec"
	"time"
)

type DumpRepository struct {
	store *PostgreDatabaseStore
}

/*
Make creates the .sql dump file for the database
with all public schema definitions and stored data.

It accepts the folder, where created dump would be saved.
*/
func (r *DumpRepository) Make(savePath string) (*models.Dump, error) {
	psqlConnectionAddr := r.store.config.ConnectionString
	fileUUID := uuid.NewV4().String()
	migrationFileName := fmt.Sprintf("%s-dump.sql", fileUUID)

	if !filesutil.Exist(savePath) {
		if err := filesutil.CreateDir(savePath); err != nil {
			r.store.logger.Println(err)
			return nil, err
		}
	}

	var migrationFilePath string
	if savePath[len(savePath)-1] != '/' {
		migrationFilePath = savePath + "/" + migrationFileName
	} else {
		migrationFilePath = savePath + migrationFileName
	}

	var stderr bytes.Buffer
	var stdout bytes.Buffer
	cmd := exec.Command("pg_dump", psqlConnectionAddr, "--column-inserts", "-f", migrationFilePath)
	cmd.Stderr = &stderr
	cmd.Stdout = &stdout
	go func() {
		err := cmd.Run()
		if err != nil {
			r.store.logger.Println(err)
		}
	}()

	dumpFile := models.Dump{
		FilePath:  migrationFilePath,
		CreatedAt: time.Now(),
	}

	transaction, err := r.store.db.Beginx()
	if err != nil {
		r.store.logger.Println(err)
		return nil, err
	}
	defer func(tx *sqlx.Tx) {
		filesutil.Delete(migrationFilePath)
		_ = transaction.Rollback()
	}(transaction)

	if _, err := transaction.NamedExec(
		`INSERT INTO public.database_dumps (dump_filepath, created_at) VALUES (:dump_filepath, :created_at)`,
		dumpFile,
	); err != nil {
		r.store.logger.Println(err)
		return nil, err
	}

	if err := transaction.Commit(); err != nil {
		r.store.logger.Println(err)
		return nil, err
	}
	dumpFile.AfterCreate()
	return &dumpFile, nil
}

func (r *DumpRepository) Execute(dumpFilePath string) error {
	dumpFileContent, err := ioutil.ReadFile(dumpFilePath)
	if err != nil {
		r.store.logger.Println(err)
		return err
	}
	dumpQueries := string(dumpFileContent)

	transaction, err := r.store.db.Beginx()
	if err != nil {
		r.store.logger.Println(err)
		return err
	}
	defer func(transaction *sqlx.Tx) {
		err := transaction.Rollback()
		r.store.logger.Println(err)
	}(transaction)
	if _, err := transaction.Exec(
		`DROP SCHEMA IF EXISTS public CASCADE;
			   CREATE SCHEMA IF NOT EXISTS public;`,
	); err != nil {
		r.store.logger.Println(err)
		return err
	}
	if _, err := transaction.Exec(dumpQueries); err != nil {
		r.store.logger.Println(err)
		return err
	}
	if _, err := transaction.Exec(`TRUNCATE public.database_dumps;`); err != nil {
		r.store.logger.Println(err)
		return err
	}
	if err := transaction.Commit(); err != nil {
		r.store.logger.Println(err)
		return err
	}
	return nil
}

func (r *DumpRepository) InsertNewDumpFile(savePath string) (*models.Dump, error) {
	fileUUID := uuid.NewV4().String()
	dumpFileName := fmt.Sprintf("%s-dump.sql", fileUUID)
	if !filesutil.Exist(savePath) {
		if err := filesutil.CreateDir(savePath); err != nil {
			r.store.logger.Println(err)
			return nil, err
		}
	}

	var dumpFilePath string
	if savePath[len(savePath)-1] != '/' {
		dumpFilePath = savePath + "/" + dumpFileName
	} else {
		dumpFilePath = savePath + dumpFileName
	}

	dumpFileModel := models.Dump{
		FilePath:  dumpFilePath,
		CreatedAt: time.Now(),
	}

	transaction, err := r.store.db.Beginx()
	if err != nil {
		r.store.logger.Println(err)
		return nil, err
	}
	defer func() {
		if err := transaction.Rollback(); err != nil {
			r.store.logger.Println(err)
		}
	}()

	if _, err := transaction.NamedExec(
		`INSERT INTO public.database_dumps (dump_filepath, created_at) VALUES (:dump_filepath, :created_at)`,
		dumpFileModel,
	); err != nil {
		r.store.logger.Println(err)
		return nil, err
	}
	if err := transaction.Commit(); err != nil {
		r.store.logger.Println(err)
		return nil, err
	}

	dumpFileModel.AfterCreate()
	return &dumpFileModel, nil
}

func (r *DumpRepository) SelectAll() ([]models.Dump, error) {
	var dumps []models.Dump
	if err := r.store.db.Select(
		&dumps,
		`SELECT * FROM public.database_dumps ORDER BY created_at DESC;`,
	); err != nil {
		r.store.logger.Println(err)
		return nil, err
	}
	for idx := range dumps {
		dumps[idx].AfterCreate()
	}
	return dumps, nil
}

func (r *DumpRepository) SelectByName(dumpFileName string) (*models.Dump, error) {
	dumpFile := &models.Dump{}
	if err := r.store.db.Get(dumpFile,
		`SELECT * FROM public.database_dumps WHERE dump_filepath LIKE $1;`,
		"%"+dumpFileName,
	); err != nil {
		r.store.logger.Println(err)
		return nil, err
	}
	dumpFile.AfterCreate()
	return dumpFile, nil
}

func (r *DumpRepository) DeleteByName(dumpFileName string) (*models.Dump, error) {
	dumpFile, err := r.SelectByName(dumpFileName)
	if err != nil {
		r.store.logger.Println(err)
		return nil, err
	}

	if _, err := r.store.db.Exec(
		`DELETE FROM public.database_dumps WHERE dump_filepath LIKE $1`,
		"%"+dumpFileName,
	); err != nil {
		r.store.logger.Println(err)
		return nil, err
	}
	return dumpFile, nil
}
