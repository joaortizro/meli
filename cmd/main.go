package main

import (
	"log"
	"meli/internal/infrastructure/server"
	"meli/internal/infrastructure/redis"
)

func  main()  {
	redisClient := redisServer.CreateRedisClient("redis",6379)

	srv := server.CreateServer("0.0.0.0",8080,redisClient)
	
	err := srv.Start(redisClient)
	
    if err != nil {
        log.Fatalf("Server error: %v", err)
    }

}