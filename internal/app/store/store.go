package store

import (
	"github.com/ArtemVovchenko/storypet-backend/internal/app/store/configs"
	"github.com/go-redis/redis/v7"
	"github.com/jmoiron/sqlx"
	_ "github.com/lib/pq"
)

type Store struct {
	config      *configs.DatabaseConfig
	db          *sqlx.DB
	redisClient *redis.Client

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

	s.redisClient = redis.NewClient(&redis.Options{
		Addr: s.config.RedisDNS,
	})

	if _, err := s.redisClient.Ping().Result(); err != nil {
		return err
	}
	return nil
}

func (s *Store) RedisClient() *redis.Client {
	return s.redisClient
}

func (s *Store) Close() {
	s.db.Close()
	s.redisClient.Close()
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
