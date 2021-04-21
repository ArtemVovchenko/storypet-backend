package main

import (
	"github.com/ArtemVovchenko/storypet-backend/internal/app/server"
	"log"
)

// Starts the server
func main() {
	s := server.New()
	log.Fatal(s.Start())
}
