package middleware

import (
	"strings"
	"sync"

	"github.com/gin-gonic/gin"
	ginprometheus "github.com/zsais/go-gin-prometheus"
)

var (
	prometheusInstance *ginprometheus.Prometheus
	once               sync.Once
)

// GetPrometheusInstance retourne une instance singleton du middleware Prometheus
func GetPrometheusInstance() *ginprometheus.Prometheus {
	once.Do(func() {
		prometheusInstance = ginprometheus.NewPrometheus("api")

		// Configuration optionnelle de la fonction de mapping pour éviter la haute cardinalité
		prometheusInstance.ReqCntURLLabelMappingFn = func(c *gin.Context) string {
			url := c.Request.URL.Path
			for _, param := range c.Params {
				url = strings.Replace(url, param.Value, ":"+param.Key, 1)
			}
			return url
		}
	})
	return prometheusInstance
}

// PrometheusMiddleware retourne le middleware Prometheus
func PrometheusMiddleware() gin.HandlerFunc {
	p := GetPrometheusInstance()
	return p.HandlerFunc()
}

// SetupPrometheusEndpoint configure l'endpoint /metrics
func SetupPrometheusEndpoint(r *gin.Engine) {
	p := GetPrometheusInstance()
	p.SetMetricsPath(r)
}
