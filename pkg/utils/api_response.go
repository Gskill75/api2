package utils

import (
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	apierrors "github.com/Gskill75/api2/pkg/errors"
)

// APIError helper global pour pousser des erreurs dans le contexte Gin
func APIError(c *gin.Context, errType error, msg string) {
	var wrappedErr error

	switch errType {
	case apierrors.ErrBadRequest:
		wrappedErr = apierrors.NewBadRequest(msg)
	case apierrors.ErrUnauthorized:
		wrappedErr = apierrors.NewUnauthorized(msg)
	case apierrors.ErrForbidden:
		wrappedErr = apierrors.NewForbidden(msg)
	case apierrors.ErrNotFound:
		wrappedErr = apierrors.NewNotFound(msg)
	case apierrors.ErrConflict:
		wrappedErr = apierrors.NewConflict(msg)
	default:
		wrappedErr = apierrors.NewInternalError(msg)
	}

	c.Error(wrappedErr)
}

// APISuccess helper pour les réponses de succès uniformes
func APISuccess(c *gin.Context, data gin.H) {
	if data == nil {
		data = gin.H{}
	}
	data["request_id"] = c.GetString("request_id")
	data["timestamp"] = getCurrentTimestamp()
	c.JSON(http.StatusOK, data)
}

func getCurrentTimestamp() string {
	return time.Now().UTC().Format(time.RFC3339)
}
