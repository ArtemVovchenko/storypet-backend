package persistentstore

import (
	"encoding/json"
	"github.com/ArtemVovchenko/storypet-backend/internal/app/sessions"
	"github.com/ArtemVovchenko/storypet-backend/internal/app/store/persistentstore/configs"
	"github.com/go-redis/redis/v7"
	"strconv"
	"time"
)

type RedisStore struct {
	configs *configs.PersistentDatabaseConfig
	db      *redis.Client
}

func NewRedisStore() *RedisStore {
	config := configs.NewPersistentDatabaseConfig()
	return &RedisStore{configs: config}
}

func (s *RedisStore) Open() error {
	addr, _ := redis.ParseURL(configs.RedisURL)
	client := redis.NewClient(addr)
	if err := client.Ping().Err(); err != nil {
		return err
	}
	s.db = client
	return nil
}

func (s *RedisStore) Close() {
	s.db.Close()
}

func (s *RedisStore) SaveSessionInfo(accessUUID string, session *sessions.Session, expireTime time.Time) error {
	sessionData, err := json.Marshal(session)
	if err != nil {
		return err
	}
	err = s.db.Set(accessUUID, sessionData, expireTime.Sub(time.Now())).Err()
	if err != nil {
		return err
	}
	return nil
}

func (s *RedisStore) SaveRefreshInfo(refreshUUID string, userID int, expireTime time.Time) error {
	return s.db.Set(refreshUUID, userID, expireTime.Sub(time.Now())).Err()
}

func (s *RedisStore) GetSessionInfo(accessUUID string) (*sessions.Session, error) {
	sessionData, err := s.db.Get(accessUUID).Result()
	if err != nil {
		return nil, err
	}
	var session sessions.Session
	if err := json.Unmarshal([]byte(sessionData), &session); err != nil {
		return nil, err
	}
	return &session, nil
}

func (s *RedisStore) DeleteSessionInfo(accessUUID string) (*sessions.Session, error) {
	sessionData, err := s.db.Get(accessUUID).Result()
	if err != nil {
		return nil, err
	}
	var session sessions.Session
	if err := json.Unmarshal([]byte(sessionData), &session); err != nil {
		return nil, err
	}
	_ = s.db.Del(accessUUID)
	_ = s.db.Del(session.RefreshUUID)

	return &session, nil
}

func (s *RedisStore) GetUserIDByRefreshUUID(refreshUUID string) (int, error) {
	userIDStr, err := s.db.Get(refreshUUID).Result()
	if err != nil {
		return 0, err
	}
	userID, err := strconv.ParseInt(userIDStr, 10, 64)
	if err != nil {
		return 0, err
	}
	return int(userID), nil
}

func (s *RedisStore) DeleteRefreshByUUID(refreshUUID string) error {
	return s.db.Del(refreshUUID).Err()
}
