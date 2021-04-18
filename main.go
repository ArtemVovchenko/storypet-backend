package main

import (
	configs2 "github.com/ArtemVovchenko/storypet-backend/internal/app/configs"
	"github.com/ArtemVovchenko/storypet-backend/internal/app/server"
	"log"
)

func main() {
	configs := configs2.NewServerConfig()
	s := server.New(configs)

	if err := s.Start(); err != nil {
		log.Fatal(err)
	}
}
