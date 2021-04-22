package configs

import "os"

var (
	// DbUrl is the connection string to the database
	DbUrl        = os.Getenv("DATABASE_URL")
	TmpDumpFiles = "/tmp/"
)
