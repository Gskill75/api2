package hello

import (
	"github.com/gin-gonic/gin"
	"github.com/Gskill75/api2/pkg/kubernetes_v2/service"
)

// HelloHandler godoc
// @Summary      Hello world message
// @Description  Retourne un "hello world" pour tester l'API v2 Kubernetes
// @Tags         kubernetes-v2
// @Produce      json
// @Success      200 {object} map[string]string "Hello message"
// @Router       /kubernetes/v2/hello [get]
// @Security     Bearer
func HelloHandler(helloService *service.HelloService) gin.HandlerFunc {
	return func(c *gin.Context) {
		msg := helloService.GetHelloWorld()
		c.JSON(200, gin.H{
			"message": msg,
		})
	}
}
