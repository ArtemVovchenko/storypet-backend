package configs

import (
	"os"
)

type ServerConfig struct {
	BindAddr     string
	LogPrefix    string
	LogFlags     int
	LogOutStream *os.File
	Database     *DatabaseConfig
}

type DatabaseConfig struct {
	ConnectionString string
}

func NewServerConfig() *ServerConfig {
	return &ServerConfig{
		BindAddr:     SrvPort,
		LogPrefix:    SrvLogPrefix,
		LogFlags:     SrvLogFlags,
		LogOutStream: SrvLogStream,
		Database:     newDatabaseConfig(),
	}
}

func newDatabaseConfig() *DatabaseConfig {
	return &DatabaseConfig{
		ConnectionString: DbUrl,
	}
}
