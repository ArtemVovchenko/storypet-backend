package configs

import (
	database "github.com/ArtemVovchenko/storypet-backend/internal/app/store/configs"
	"os"
)

type ServerConfig struct {
	BindAddr     string
	LogPrefix    string
	LogFlags     int
	LogOutStream *os.File
	Database     *database.DatabaseConfig
}

func NewServerConfig() *ServerConfig {
	return &ServerConfig{
		BindAddr:     SrvPort,
		LogPrefix:    SrvLogPrefix,
		LogFlags:     SrvLogFlags,
		LogOutStream: SrvLogStream,
		Database:     database.NewDatabaseConfig(),
	}
}
