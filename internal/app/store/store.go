package store

import (
	"github.com/ArtemVovchenko/storypet-backend/internal/app/sessions"
	"github.com/ArtemVovchenko/storypet-backend/internal/app/store/persistentstore"
	"github.com/ArtemVovchenko/storypet-backend/internal/app/store/repos"
	"github.com/ArtemVovchenko/storypet-backend/internal/app/store/sqlxstore"
	"log"
	"time"
)

type DatabaseStore interface {
	Open() error
	Close()
	Users() repos.UserRepository
	Roles() repos.RoleRepository
	Pets() repos.PetRepository
	Vaccines() repos.VaccineRepository
	Dumps() repos.DumpRepository
}

type PersistentStore interface {
	Open() error
	Close()
	SaveSessionInfo(accessUUID string, session *sessions.Session, expireTime time.Time) error
	SaveRefreshInfo(refreshUUID string, userID int, expireTime time.Time) error
	GetSessionInfo(accessUUID string) (*sessions.Session, error)
	DeleteSessionInfo(accessUUID string) (*sessions.Session, error)
	GetUserIDByRefreshUUID(refreshUUID string) (int, error)
	DeleteRefreshByUUID(refreshUUID string) error
}

func NewDatabaseStore(logger *log.Logger) DatabaseStore {
	return sqlxstore.NewPostgreDatabaseStore(logger)
}

func NewPersistentStore() PersistentStore {
	return persistentstore.NewRedisStore()
}
