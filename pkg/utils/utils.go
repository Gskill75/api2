package utils

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"k8s.io/klog/v2"
)

func GetCustomerIDOrAbort(c *gin.Context) (string, bool) {
	rid := c.GetString("request_id")
	customerID := c.GetString("customer_id")
	if customerID == "" {
		klog.Warningf("[request_id=%s] Missing customer_id", rid)
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Missing customer_id", "request_id": rid})
		c.Abort()
		return "", false
	}
	return customerID, true
}

// IsAdminOrAbort vérifie que le rôle du token est "admin", sinon répond 403
func IsAdminOrAbort(c *gin.Context) bool {
	rid := c.GetString("request_id")
	email := c.GetString("email")

	val, exists := c.Get("roles")
	if !exists {
		klog.Warningf("[request_id=%s] Access denied: no roles found", rid)
		c.JSON(http.StatusForbidden, gin.H{
			"error":      "Access denied: missing roles",
			"request_id": rid,
		})
		c.Abort()
		return false
	}

	roles, ok := val.([]string)
	if !ok {
		klog.Warningf("[request_id=%s] Access denied: invalid roles format", rid)
		c.JSON(http.StatusForbidden, gin.H{
			"error":      "Access denied: invalid roles format",
			"request_id": rid,
		})
		c.Abort()
		return false
	}

	for _, role := range roles {
		if role == "api:sigma-admin" {
			return true
		}
	}

	// Affichage des rôles dans les logs d'erreur
	klog.Warningf("[request_id=%s] Access denied for user '%s': required role 'api:sigma-admin' missing — user roles: %v", rid, email, roles)

	c.JSON(http.StatusForbidden, gin.H{
		"error":      "Access denied: admin role required",
		"request_id": rid,
		"roles":      roles, // (optionnel : expose aussi dans la réponse)
	})
	c.Abort()
	return false
}
