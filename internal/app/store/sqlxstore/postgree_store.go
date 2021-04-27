package sqlxstore

import (
	"github.com/ArtemVovchenko/storypet-backend/internal/app/store/repos"
	"github.com/ArtemVovchenko/storypet-backend/internal/app/store/sqlxstore/configs"
	"github.com/ArtemVovchenko/storypet-backend/internal/pkg/url"
	"github.com/jmoiron/sqlx"
	_ "github.com/lib/pq"
	"log"
)

type PostgreDatabaseStore struct {
	config *configs.DatabaseConfig
	db     *sqlx.DB
	logger *log.Logger

	userRepository *UserRepository
	roleRepository *RoleRepository
	petRepository  *PetRepository
	dumpRepository *DumpRepository
}

func NewPostgreDatabaseStore(logger *log.Logger) *PostgreDatabaseStore {
	config := configs.NewDatabaseConfig()
	return &PostgreDatabaseStore{
		config: config,
		logger: logger,
	}
}

func (s *PostgreDatabaseStore) Open() error {
	dbDriverConnectionString, err := url.ParsePostgreConn(s.config.ConnectionString)
	if err != nil {
		s.logger.Println(err)
		return err
	}
	db, err := sqlx.Connect("postgres", dbDriverConnectionString)
	if err != nil {
		s.logger.Println(err)
		return err
	}
	if err := db.Ping(); err != nil {
		s.logger.Println(err)
		return err
	}
	s.db = db
	return nil
}

func (s *PostgreDatabaseStore) Close() {
	defer func(db *sqlx.DB) {
		if err := db.Close(); err != nil {
			s.logger.Println(err)
		}
	}(s.db)
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

func (s *PostgreDatabaseStore) Pets() repos.PetRepository {
	if s.petRepository != nil {
		return s.petRepository
	}
	s.petRepository = &PetRepository{store: s}
	return s.petRepository
}

func (s *PostgreDatabaseStore) Dumps() repos.DumpRepository {
	if s.dumpRepository != nil {
		return s.dumpRepository
	}
	s.dumpRepository = &DumpRepository{
		store: s,
	}
	return s.dumpRepository
}
