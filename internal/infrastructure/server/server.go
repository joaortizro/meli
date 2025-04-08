package server

import (
	"fmt"
	"log"
	"meli/internal/infrastructure/middleware"
	"net/http"
	"net/http/httputil"
	"net/url"

	"github.com/gin-gonic/gin"
	"github.com/redis/go-redis/v9"
)

const API_URL string = "https://api.mercadolibre.com/"

type Server struct {
	host   string
	port   uint
	router *gin.Engine
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

	return &Server{
		host:   host,
		port:   port,
		router: router,
	}
}

func (s *Server) Start() error {
	address := fmt.Sprintf("%s:%d", s.host, s.port)

	log.Println("Server running on", address)

	s.registerRoutes()

	err := s.router.Run(address)
	if err != nil {
		return fmt.Errorf("failed to start server: %w", err)
	}

	return nil
}

func (s *Server) registerRoutes() {
	s.router.GET("/hello/world", func(c *gin.Context) {
		c.String(200, "Hello, World!")
	})

	s.router.GET("/goodbye/world", func(c *gin.Context) {
		c.String(200, "Goodbye, World!")
	})

	s.router.Any("p/*proxyPath", func(c *gin.Context) {

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
			c.JSON(http.StatusBadGateway, gin.H{"error": "Bad Gateway", "message": err.Error()})
		}

		proxy.ServeHTTP(c.Writer, c.Request)
	})
}
