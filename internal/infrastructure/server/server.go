package server

import (
	"fmt"
	"log"
	"net/http"
	"net/http/httputil"
	"net/url"
	"github.com/gin-gonic/gin"
)

type Server struct {
	host string
	port uint
	router *gin.Engine
}

func CreateServer (host string,port uint) *Server {
	gin.SetMode(gin.ReleaseMode)
	
	router := gin.New()

	router.Use(gin.Recovery())
	router.Use(gin.Logger())
	
	return &Server{
		host: host,
		port: port,
		router:router,
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
	s.router.Any("/*proxyPath", func(c *gin.Context) {

		proxyPath := c.Param("proxyPath")
		baseURL, _ := url.Parse("https://api.mercadolibre.com/")

		proxy := httputil.NewSingleHostReverseProxy(baseURL)

		proxy.Director = func(req *http.Request) {
            req.URL.Path = proxyPath
            req.Host = baseURL.Host
            req.URL.Host = baseURL.Host
            req.URL.Scheme = baseURL.Scheme
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