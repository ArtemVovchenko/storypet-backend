package store

import (
	"github.com/ArtemVovchenko/storypet-backend/internal/app/configs"
	"github.com/jmoiron/sqlx"
	_ "github.com/lib/pq"
	"log"
)

type Store struct {
	config *configs.DatabaseConfig
	db     *sqlx.DB
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

func (s Store) Close(logger *log.Logger) {
	if err := s.db.Close(); err != nil {
		logger.Printf("Error while closing the database %s", err)
	}
}
