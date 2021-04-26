package configs

import (
	"os"
)

type ServerConfig struct {
	BindAddr              string
	LogPrefix             string
	LogFlags              int
	LogOutStream          *os.File
	DatabaseLogPrefix     string
	DatabaseLogFlags      int
	DatabaseLogsOutStream *os.File

	DatabaseDumpsDir string
}

func NewServerConfig() *ServerConfig {
	return &ServerConfig{
		BindAddr:              SrvPort,
		LogPrefix:             SrvLogPrefix,
		LogFlags:              SrvLogFlags,
		LogOutStream:          SrvLogStream,
		DatabaseLogFlags:      DatabaseLogFlags,
		DatabaseLogsOutStream: DatabaseLogStream,
		DatabaseLogPrefix:     DatabaseLogPrefix,

		DatabaseDumpsDir: DatabaseDumpsDir,
	}
}
