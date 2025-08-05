package middleware

import (
	"time"

	"github.com/gin-gonic/gin"
	"k8s.io/klog/v2"
)

func RequestLogger() gin.HandlerFunc {
	return func(c *gin.Context) {
		start := time.Now()

		// Extraction sécurisée des données de base
		requestID := getRequestID(c)
		method := c.Request.Method
		path := c.Request.URL.Path
		clientIP := c.ClientIP()

		// Processing de la requête
		c.Next()

		// Calcul des métriques
		status := c.Writer.Status()
		duration := time.Since(start)
		responseSize := c.Writer.Size()

		// Détermination du niveau de log
		logFunc := getLogFunctionForStatus(status)

		// Log structuré final
		logFunc("http_request",
			"request_id", requestID,
			"method", method,
			"path", path,
			"status", status,
			"duration_ms", duration.Milliseconds(),
			"client_ip", clientIP,
			"response_size", responseSize,
			"user_agent", c.Request.UserAgent(),
		)
	}
}

// Helpers functions
func getRequestID(c *gin.Context) string {
	if id, exists := c.Get(RequestIDKey); exists {
		if strID, ok := id.(string); ok {
			return strID
		}
	}
	return "unknown"
}

func getLogFunctionForStatus(status int) func(string, ...interface{}) {
	switch {
	case status >= 500:
		return func(msg string, keysAndValues ...interface{}) {
			klog.V(0).InfoS(msg, keysAndValues...)
		}
	case status >= 400:
		return func(msg string, keysAndValues ...interface{}) {
			klog.V(1).InfoS(msg, keysAndValues...)
		}
	default:
		return func(msg string, keysAndValues ...interface{}) {
			klog.V(2).InfoS(msg, keysAndValues...)
		}
	}
}
