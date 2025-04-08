package main

import (
	"log"
	"meli/internal/infrastructure/server"
)

func  main()  {
	srv := server.CreateServer("localhost",8080)
	err := srv.Start()
	
    if err != nil {
        log.Fatalf("Server error: %v", err)
    }

}