package metrics


import (
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
	"github.com/gin-gonic/gin"
)


var (
	proxyRequestsTotal = promauto.NewCounter(prometheus.CounterOpts{
		Name: "proxy_requests_total",
		Help: "Total number of requests handled by the proxy.",
	})

	proxyRequestsByMethodTotal = promauto.NewCounterVec(prometheus.CounterOpts{
		Name: "proxy_requests_by_method_total",
		Help: "Total number of requests handled by the proxy, partitioned by HTTP method.",
	},[]string{"method"}, 

	)
)

func PrometheusMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		proxyRequestsTotal.Inc()
		proxyRequestsByMethodTotal.WithLabelValues(c.Request.Method).Inc()
		c.Next()
	}
}