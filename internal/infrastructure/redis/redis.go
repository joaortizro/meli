package  redisServer

import (
	"fmt"
	"context"
	"log"
    "github.com/redis/go-redis/v9"
)

func CreateRedisClient (host string,port uint)  *redis.Client {
	redisAddr := fmt.Sprintf("%s:%d", host, port)

	rdb := redis.NewClient(&redis.Options{
		Addr:     redisAddr,
		Password: "",
		DB:       0, 
	})

	_, err := rdb.Ping(context.Background()).Result()

	if err != nil {
		log.Fatalf("Error when connecting to redis: %v", err)
	}

	return rdb
}