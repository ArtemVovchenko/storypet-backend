package server

import (
	"github.com/ArtemVovchenko/storypet-backend/internal/app/configs"
	"os"
)

type Config struct {
	BindAddr     string
	LogPrefix    string
	LogFlags     int
	LogOutStream *os.File
}

func NewConfig() *Config {
	return &Config{
		BindAddr:     configs.PORT,
		LogPrefix:    configs.SRV_LOG_PREFIX,
		LogFlags:     configs.SRV_LOG_FLAGS,
		LogOutStream: configs.SRV_LOG_STREAM,
	}
}
