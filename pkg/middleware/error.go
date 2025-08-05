package middleware

import (
	"database/sql"
	"errors"
	"net/http"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	apierrors "gitn.sigma.fr/sigma/paas/api/api/pkg/errors"
)

func ErrorHandler() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Next()

		if len(c.Errors) == 0 {
			return
		}

		err := c.Errors.Last().Err
		status := determineStatusCode(err)
		message := extractErrorMessage(err)

		// Format de réponse uniforme pour toute l'API
		c.JSON(status, gin.H{
			"error":      message,
			"request_id": c.GetString("request_id"),
			"timestamp":  getCurrentTimestamp(),
		})
	}
}

// determineStatusCode mappe les erreurs vers les codes HTTP appropriés
func determineStatusCode(err error) int {
	switch {
	case errors.Is(err, sql.ErrNoRows):
		return http.StatusNotFound
	case errors.Is(err, apierrors.ErrBadRequest):
		return http.StatusBadRequest
	case errors.Is(err, apierrors.ErrUnauthorized):
		return http.StatusUnauthorized
	case errors.Is(err, apierrors.ErrForbidden):
		return http.StatusForbidden
	case errors.Is(err, apierrors.ErrConflict):
		return http.StatusConflict
	case errors.Is(err, apierrors.ErrTooManyRequests):
		return http.StatusTooManyRequests
	case strings.HasPrefix(err.Error(), "bad_request:"):
		return http.StatusBadRequest
	case strings.HasPrefix(err.Error(), "unauthorized:"):
		return http.StatusUnauthorized
	case strings.HasPrefix(err.Error(), "forbidden:"):
		return http.StatusForbidden
	case strings.HasPrefix(err.Error(), "not_found:"):
		return http.StatusNotFound
	case strings.HasPrefix(err.Error(), "conflict:"):
		return http.StatusConflict
	default:
		return http.StatusInternalServerError
	}
}

// extractErrorMessage extrait le message d'erreur propre
func extractErrorMessage(err error) string {
	msg := err.Error()

	// Supprime les préfixes de type d'erreur
	prefixes := []string{
		"bad_request: ",
		"unauthorized: ",
		"forbidden: ",
		"not_found: ",
		"conflict: ",
		"internal_error: ",
	}

	for _, prefix := range prefixes {
		if strings.HasPrefix(msg, prefix) {
			return strings.TrimPrefix(msg, prefix)
		}
	}

	return msg
}

func getCurrentTimestamp() string {
	return time.Now().UTC().Format(time.RFC3339)
}
