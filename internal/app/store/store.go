package store

import (
	"github.com/ArtemVovchenko/storypet-backend/internal/app/store/configs"
	"github.com/jmoiron/sqlx"
	_ "github.com/lib/pq"
)

type Store struct {
	config         *configs.DatabaseConfig
	db             *sqlx.DB
	userRepository *UserRepository
}

func New(config *configs.DatabaseConfig) *Store {
	return &Store{
		config: config,
	}
}

func (s *Store) Open() error {
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

func (s *Store) Close() {
	s.db.Close()
}

func (s *Store) Users() *UserRepository {
	if s.userRepository != nil {
		return s.userRepository
	}
	s.userRepository = &UserRepository{
		store: s,
	}
	return s.userRepository
}
