package middleware

import (
	"context"
	"net/http"
	"strings"

	"github.com/coreos/go-oidc/v3/oidc"
	"github.com/gin-gonic/gin"
)

type OIDCConfig struct {
	Issuer   string
	Audience string
}

func NewOIDCMiddleware(cfg OIDCConfig) (gin.HandlerFunc, error) {
	ctx := context.Background()
	provider, err := oidc.NewProvider(ctx, cfg.Issuer)
	if err != nil {
		return nil, err
	}

	verifier := provider.Verifier(&oidc.Config{
		ClientID: cfg.Audience,
	})

	return func(c *gin.Context) {
		authHeader := c.GetHeader("Authorization")
		if authHeader == "" || !strings.HasPrefix(authHeader, "Bearer ") {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "missing or invalid Authorization header"})
			return
		}
		rawToken := strings.TrimPrefix(authHeader, "Bearer ")

		idToken, err := verifier.Verify(c.Request.Context(), rawToken)
		if err != nil {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "invalid or expired token"})
			return
		}
		var claims struct {
			Sub        string `json:"sub"`
			Email      string `json:"email"`
			CustomerID string `json:"customer_id"`
		}
		if err := idToken.Claims(&claims); err != nil {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "failed to parse token claims"})
			return
		}

		// Récupérer les claims bruts pour lire les rôles
		var rawClaims map[string]interface{}
		if err := idToken.Claims(&rawClaims); err != nil {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "failed to parse raw token claims"})
			return
		}

		// Extraire les rôles depuis resource_access.<audience>.roles
		var roles []string
		if ra, ok := rawClaims["resource_access"].(map[string]interface{}); ok {
			if client, ok := ra[cfg.Audience].(map[string]interface{}); ok {
				if rlist, ok := client["roles"].([]interface{}); ok {
					for _, r := range rlist {
						if rstr, ok := r.(string); ok {
							roles = append(roles, rstr)
						}
					}
				}
			}
		}

		// Injecter dans le contexte Gin
		c.Set("sub", claims.Sub)
		c.Set("email", claims.Email)
		c.Set("roles", roles)
		c.Set("customer_id", claims.CustomerID)

		c.Next()
	}, nil

}
