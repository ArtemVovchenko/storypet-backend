package sqlxstore

import (
	"fmt"
	"github.com/ArtemVovchenko/storypet-backend/internal/app/store/sqlxstore/configs"
	"github.com/jmoiron/sqlx"
	"github.com/twinj/uuid"
	"os"
	"os/exec"
)

func (s *PostgreDatabaseStore) MakeDump() (string, error) {
	psqlConnectionAddr := s.config.ConnectionString

	workDir, err := os.Getwd()
	if err != nil {
		return "", err
	}
	fileUUID := uuid.NewV4().String()
	migrationFileName := fmt.Sprintf("%s-dump.sql", fileUUID)

	if _, err := os.Stat(workDir + configs.TmpDumpFiles); os.IsNotExist(err) {
		if err := os.MkdirAll(workDir+configs.TmpDumpFiles, os.ModePerm); err != nil {
			return "", err
		}
	}
	migrationFilePath := workDir + configs.TmpDumpFiles + migrationFileName

	cmd := exec.Command("pg_dump", psqlConnectionAddr, "--column-inserts", "-f", migrationFilePath)
	if err := cmd.Run(); err != nil {
		return "", err
	}
	return migrationFilePath, nil
}

func (s *PostgreDatabaseStore) ExecuteDump(dumpQueries string) error {
	transaction, err := s.db.Beginx()
	if err != nil {
		return err
	}
	defer func(transaction *sqlx.Tx) {
		// TODO: log the potential rollback error
		_ = transaction.Rollback()
	}(transaction)
	if _, err := transaction.Exec(
		`DROP SCHEMA IF EXISTS public CASCADE;
			   CREATE SCHEMA IF NOT EXISTS public;`,
	); err != nil {
		return err
	}
	if _, err := transaction.Exec(dumpQueries); err != nil {
		return err
	}
	if err := transaction.Commit(); err != nil {
		return err
	}
	return nil
}
