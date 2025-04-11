package server

import (
	"fmt"
	"log"
	"meli/internal/infrastructure/middleware/metrics"
	"meli/internal/infrastructure/middleware/rateLimiter"
	"meli/internal/infrastructure/middleware/stats"
	"net/http"
	"net/http/httputil"
	"net/http/httptest"
	"net/url"
	"context"
	"encoding/json"
	"time"
	"github.com/gin-gonic/gin"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/redis/go-redis/v9"
)

const API_URL string = "https://api.mercadolibre.com/"
var ctx = context.Background()
const CACHE_TTL = 10 * time.Minute 
type Server struct {
	host   string
	port   uint
	router *gin.Engine
}

type CachedResponse struct {
	StatusCode  int    `json:"statusCode"`
	ContentType string `json:"contentType"`
	Body        []byte `json:"body"`
}

func CreateServer(host string, port uint, redisClient *redis.Client) *Server {
	gin.SetMode(gin.ReleaseMode)

	router := gin.New()

	//to handle panic
	router.Use(gin.Recovery())
	//simple request logger
	router.Use(gin.Logger())
	//rateLimiter
	router.Use(rateLimiter.RateLimiterMiddleware(redisClient))
	//stats
	router.Use(stats.CalulateRequests(redisClient))
	//metrics
	router.Use(metrics.PrometheusMiddleware())

	return &Server{
		host:   host,
		port:   port,
		router: router,
	}
}

func (s *Server) Start(redisClient *redis.Client) error {
	address := fmt.Sprintf("%s:%d", s.host, s.port)

	log.Println("Server running on", address)

	s.registerRoutes(redisClient)

	err := s.router.Run(address)
	if err != nil {
		return fmt.Errorf("failed to start server: %w", err)
	}

	return nil
}

func (s *Server) registerRoutes(redisClient *redis.Client) {
	s.router.GET("/hello/world", func(c *gin.Context) {
		c.String(200, "Hello, World!")
	})

	s.router.GET("/goodbye/world", func(c *gin.Context) {
		c.String(200, "Goodbye, World!")
	})

	s.router.GET("/stats", func(c *gin.Context) {
		stats.StatsHandler(c, redisClient)
    })

	s.router.GET("/metrics",gin.WrapH(promhttp.Handler()))

	s.router.Any("p/*proxyPath", func(c *gin.Context) {

		if c.Request.Method == http.MethodGet {
			var cacheKey = fmt.Sprintf("cache:%s:%s", c.Request.Method, c.Request.RequestURI)
			cachedVal, err := redisClient.Get(ctx, cacheKey).Result()
			//CACHE HIT
			if err == nil {
				log.Printf("INFO: Cache HIT %s", cacheKey)
				c.Header("X-Cache-Status", "HIT")
				c.Data(http.StatusOK, "text/plain; charset=utf-8", []byte(cachedVal))
				c.Abort()
				return
			}
			//CACHE MISS
			statusCode, headers, body := proxyWithoutCache(c)
			var jsonData interface{}
			errUnmarshal := json.Unmarshal(body, &jsonData)

			if errUnmarshal != nil {
				log.Printf("ERROR: response from %s is not a valid json: %v", c.Request.RequestURI, errUnmarshal)
			}

			if  statusCode == 200 {
				errSet := redisClient.Set(ctx, cacheKey, body, CACHE_TTL).Err()
				if errSet != nil {
					log.Printf("Error: failed to SET cache %s: %v", cacheKey, errSet)
				} else {
					log.Printf("INFO: Response for %s stored in chache.", cacheKey)
				}
			}
			
			for k, v := range headers {
				c.Writer.Header()[k] = v
		    }
		   
		   c.Writer.WriteHeader(statusCode) 
		   c.Writer.Write(body)            
			
		}
	})
}

func proxyWithoutCache (c *gin.Context) (statusCode int, headers http.Header, body []byte){
	proxyPath := c.Param("proxyPath")
	baseURL, _ := url.Parse(API_URL)

	proxy := httputil.NewSingleHostReverseProxy(baseURL)

	proxy.Director = func(req *http.Request) {

		req.URL.Path = proxyPath
		req.Host = baseURL.Host
		req.URL.Host = baseURL.Host
		req.URL.Scheme = baseURL.Scheme

		req.Header.Set("X-Forwarded-For", c.ClientIP())
	}

	proxy.ErrorHandler = func(rw http.ResponseWriter, req *http.Request, err error) {
		fmt.Printf("Error for %s: %v\n", req.URL.Path, err)
		fmt.Printf("Error details: %v\n", err)
		rw.WriteHeader(http.StatusBadGateway)
	}

	recorder := httptest.NewRecorder()

	proxy.ServeHTTP(recorder, c.Request)

	statusCode = recorder.Code
	headers = recorder.Header() 
	body = recorder.Body.Bytes()  

	return

}