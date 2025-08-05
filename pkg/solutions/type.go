package solutions

import "github.com/gin-gonic/gin"

type Solution interface {
	Name() string
	Version() string
	Endpoint(*gin.RouterGroup)
}
