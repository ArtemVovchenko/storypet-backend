package main

import (
	"github.com/ArtemVovchenko/storypet-backend/internal/app/configs"
	"github.com/ArtemVovchenko/storypet-backend/internal/app/server"
	"log"
)

func main() {
	config := configs.NewServerConfig()
	s := server.New(config)

	if err := s.Start(); err != nil {
		log.Fatal(err)
	}
}
