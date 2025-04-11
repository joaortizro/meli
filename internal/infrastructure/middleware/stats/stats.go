package stats

import (
	"context"
	"log"
	"net/http"
	"github.com/gin-gonic/gin"
	"github.com/redis/go-redis/v9"
)

var ctx = context.Background()

func CalulateRequests(redisClient *redis.Client) gin.HandlerFunc {
	return func(c *gin.Context) {
		amountOfRequest := "stats:requests"

		err := redisClient.Incr(ctx, amountOfRequest).Err()

		if err != nil {
			log.Printf("Error: could not increment the counter '%s': %v", amountOfRequest, err)
		}

		var methodKey string

		switch c.Request.Method {
		case http.MethodGet:
			methodKey = "stats:method:get"
		case http.MethodPost:
			methodKey = "stats:method:post"
		case http.MethodPut:
			methodKey = "stats:method:put"
		case http.MethodDelete:
			methodKey = "stats:method:delete"
		default:
			methodKey = "stats:method:other"
		}

		err = redisClient.Incr(ctx, methodKey).Err()

		if err != nil {
			log.Printf("Error: could not increment the counter '%s': %v", methodKey, err)
		}

		c.Next()
	}
}

func StatsHandler(c *gin.Context, redisClient *redis.Client) {

	keysToFetch := []string{
		"stats:requests",
		"stats:method:get",
		"stats:method:post",
		"stats:method:put",
		"stats:method:delete",
		"stats:method:other",
	}

	statistics, err := redisClient.MGet(ctx, keysToFetch...).Result()

	log.Println(statistics...)
	
	if err != nil {
		log.Print("Error, could not get the stats keys", err)
		c.JSON(http.StatusInternalServerError, gin.H{"message": err})
	}

	statsResponse := make(map[string]string)

	for i, key := range keysToFetch {
		statsResponse[key] = "0"
		if statistics[i]!= nil {
			statsResponse[key] = statistics[i].(string)
		}
		
	}

	c.JSON(http.StatusOK, statsResponse)
}
