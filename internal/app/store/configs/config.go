package configs

type DatabaseConfig struct {
	ConnectionString string
	RedisDNS         string
}

func NewDatabaseConfig() *DatabaseConfig {
	return &DatabaseConfig{
		ConnectionString: DbUrl,
		RedisDNS:         RedisDNS,
	}
}
