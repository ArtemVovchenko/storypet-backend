package configs

import "os"

var (
	// RedisURL is the connection string to the database
	RedisURL = os.Getenv("REDIS_URL")
)
