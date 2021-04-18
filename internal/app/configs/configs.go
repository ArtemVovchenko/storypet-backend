package configs

import (
	"log"
	"os"
)

var (
	// SrvPort Server Port Defined as OS Env Var
	SrvPort = os.Getenv("SRV_PORT")
	// SrvLogStream is the output file of the log info. Default is os.Stdout
	SrvLogStream = os.Stdout
	// SrvLogFlags is the go log.Logger instance config flags
	SrvLogFlags = log.LstdFlags
	// SrvLogPrefix is an logging message's prefix
	SrvLogPrefix = "SRV:"

	// DbUrl is the connection string to the database
	DbUrl = os.Getenv("DATABASE_URL")
)
