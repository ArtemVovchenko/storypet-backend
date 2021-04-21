package configs

type DatabaseConfig struct {
	ConnectionString string
}

func NewDatabaseConfig() *DatabaseConfig {
	return &DatabaseConfig{
		ConnectionString: DbUrl,
	}
}
