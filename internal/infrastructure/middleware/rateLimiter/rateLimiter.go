package rateLimiter

import (
	"context"
	"log"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/redis/go-redis/v9"
)

var ctx = context.Background()

const window = time.Minute 
const limit = 10

func RateLimiterMiddleware(redisClient *redis.Client) gin.HandlerFunc {
	return func(c *gin.Context) {
		userIP := c.ClientIP()
		
		log.Println(c.Request.URL.Path)

		countPerIP, err := redisClient.Incr(ctx, userIP).Result()

		if err != nil {
			log.Printf("Error when incrementing count per IP %s: %v", userIP, err)
			c.AbortWithStatus(http.StatusInternalServerError)
			return
		}

		if countPerIP == 1 {
			redisClient.Expire(ctx, userIP, window).Err()
		}

		if countPerIP > limit {
			log.Printf("Limit reached for IP %s", userIP)
			c.AbortWithStatus(http.StatusTooManyRequests)
			return
		}

		c.Next()
	}
}
