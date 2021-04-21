package configs

import (
	"log"
	"os"
)

var (
	SrvPort      = os.Getenv("PORT")
	SrvLogStream = os.Stdout
	SrvLogFlags  = log.LstdFlags
	SrvLogPrefix = "SRV:"

	SrvErrLogStream = os.Stderr
	SrvErrLogFlags  = log.Llongfile | log.LstdFlags
	SrvErrLogPrefix = "ERR:"
)
