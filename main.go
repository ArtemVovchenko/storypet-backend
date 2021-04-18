package main

import (
	"github.com/ArtemVovchenko/storypet-backend/internal/app/server"
	"log"
)

func main() {
	configs := server.NewConfig()
	s := server.New(configs)

	if err := s.Start(); err != nil {
		log.Fatal(err)
	}
}
