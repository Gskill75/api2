package health

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

type Options struct {
	Checkers []ReadinessChecker
}

func RegisterRoutes(r *gin.Engine, opts *Options) {
	r.GET("/healthz", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"status": "ok"})
	})

	r.GET("/readyz", func(c *gin.Context) {
		failed := map[string]string{}

		if opts != nil {
			for _, checker := range opts.Checkers {
				if err := checker.Check(); err != nil {
					failed[checker.Name()] = err.Error()
				}
			}
		}

		if len(failed) > 0 {
			c.JSON(http.StatusServiceUnavailable, gin.H{
				"status": "not ready",
				"errors": failed,
			})
			return
		}

		c.JSON(http.StatusOK, gin.H{"status": "ready"})
	})
}
