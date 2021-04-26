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

	DatabaseLogStream = os.Stderr
	DatabaseLogFlags  = log.Llongfile | log.LstdFlags
	DatabaseLogPrefix = "DATABASE: "

	DatabaseDumpsDir = getCWD() + os.Getenv("DATABASE_DUMP_DIR")
)

func getCWD() string {
	cwd, _ := os.Getwd()
	if cwd[len(cwd)-1] != '/' {
		return cwd + "/"
	}
	return cwd
}
