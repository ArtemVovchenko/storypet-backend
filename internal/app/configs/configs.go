package configs

import (
	"log"
	"os"
)

var (
	PORT           = os.Getenv("PORT")
	SRV_LOG_STREAM = os.Stdout
	SRV_LOG_FLAGS  = log.LstdFlags
	SRV_LOG_PREFIX = "SRV:"
)
