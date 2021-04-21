package sqlxstore

import (
	"github.com/ArtemVovchenko/storypet-backend/internal/app/store/repos"
	"github.com/ArtemVovchenko/storypet-backend/internal/app/store/sqlxstore/configs"
	"github.com/jmoiron/sqlx"
	_ "github.com/lib/pq"
	"os/exec"
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
	db, err := sqlx.Connect("postgres", s.config.ConnectionString)
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
	s.db.Close()
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

func (s *PostgreDatabaseStore) MakeDump() {
	exec.Command("pg_dump", "")
}
