package configs

type PersistentDatabaseConfig struct {
	ConnectionString string
}

func NewPersistentDatabaseConfig() *PersistentDatabaseConfig {
	return &PersistentDatabaseConfig{
		ConnectionString: RedisURL,
	}
}
