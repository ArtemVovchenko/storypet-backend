package configs

import (
	"os"
)

type ServerConfig struct {
	BindAddr        string
	LogPrefix       string
	LogFlags        int
	LogOutStream    *os.File
	ErrLogPrefix    string
	ErrLogFlags     int
	ErrLogOutStream *os.File

	DatabaseDumpsDir string
}

func NewServerConfig() *ServerConfig {
	return &ServerConfig{
		BindAddr:        SrvPort,
		LogPrefix:       SrvLogPrefix,
		LogFlags:        SrvLogFlags,
		LogOutStream:    SrvLogStream,
		ErrLogFlags:     SrvErrLogFlags,
		ErrLogOutStream: SrvErrLogStream,
		ErrLogPrefix:    SrvErrLogPrefix,

		DatabaseDumpsDir: DatabaseDumpsDir,
	}
}
