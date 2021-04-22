package sqlxstore

import (
	"bytes"
	"fmt"
	"github.com/ArtemVovchenko/storypet-backend/internal/app/store/repos"
	"github.com/ArtemVovchenko/storypet-backend/internal/app/store/sqlxstore/configs"
	"github.com/ArtemVovchenko/storypet-backend/internal/pkg/url"
	"github.com/jmoiron/sqlx"
	_ "github.com/lib/pq"
	"os"
	"os/exec"
	"time"
)

type PostgreDatabaseStore struct {
	config *configs.DatabaseConfig
	db     *sqlx.DB

	userRepository *UserRepository
	roleRepository *RoleRepository
}

func NewPostgreDatabaseStore() *PostgreDatabaseStore {
	config := configs.NewDatabaseConfig()
	return &PostgreDatabaseStore{
		config: config,
	}
}

func (s *PostgreDatabaseStore) Open() error {
	dbDriverConnectionString, err := url.ParsePostgreConn(s.config.ConnectionString)
	if err != nil {
		return err
	}
	db, err := sqlx.Connect("postgres", dbDriverConnectionString)
	if err != nil {
		return err
	}
	if err := db.Ping(); err != nil {
		return err
	}
	s.db = db
	return nil
}

func (s *PostgreDatabaseStore) Close() {
	_ = s.db.Close()
}

func (s *PostgreDatabaseStore) Users() repos.UserRepository {
	if s.userRepository != nil {
		return s.userRepository
	}
	s.userRepository = &UserRepository{
		store: s,
	}
	return s.userRepository
}

func (s *PostgreDatabaseStore) Roles() repos.RoleRepository {
	if s.roleRepository != nil {
		return s.roleRepository
	}
	s.roleRepository = &RoleRepository{
		store: s,
	}
	return s.roleRepository
}

func (s *PostgreDatabaseStore) MakeDump() error {
	psqlConnectionAddr := s.config.ConnectionString

	workDir, err := os.Getwd()
	if err != nil {
		return err
	}
	migrationFileName := fmt.Sprintf("%s_dump.sql", time.Now().Format("02.01.2006:15:04:05"))
	migrationFilePath := workDir + "/local/" + migrationFileName

	var out bytes.Buffer
	var stderr bytes.Buffer
	cmd := exec.Command("pg_dump", psqlConnectionAddr, "-f", migrationFilePath)
	cmd.Stdout = &out
	cmd.Stderr = &stderr
	if err := cmd.Run(); err != nil {
		return err
	}
	return nil
}
