package kubernetesv2

import (
	"fmt"

	"github.com/gin-gonic/gin"
	"gitn.sigma.fr/sigma/paas/api/api/pkg/config"
	db "gitn.sigma.fr/sigma/paas/api/api/pkg/db/sqlc/kubernetes"
	kubeclient "gitn.sigma.fr/sigma/paas/api/api/pkg/kubernetes/client"
	hellohandler "gitn.sigma.fr/sigma/paas/api/api/pkg/kubernetes_v2/handler/hello"
	"gitn.sigma.fr/sigma/paas/api/api/pkg/kubernetes_v2/service"
	"k8s.io/klog/v2"
)

// KubernetesSolutionV2 : struct principale pour ta v2
type KubernetesSolutionV2 struct {
	client        *kubeclient.Client
	queries       *db.Queries
	cfg           *config.Config
	service_hello *service.HelloService
}

// NewKubernetesSolution : constructeur avec validation des dépendances
func NewKubernetesSolution(cfg *config.Config, client *kubeclient.Client, queries *db.Queries) (*KubernetesSolutionV2, error) {
	if cfg == nil {
		klog.Errorf("V2: config is required")
		return nil, fmt.Errorf("config is required")
	}
	if client == nil {
		klog.Errorf("V2: kubernetes client is required")
		return nil, fmt.Errorf("kubernetes client is required")
	}
	if queries == nil {
		klog.Errorf("V2: database queries are required")
		return nil, fmt.Errorf("database queries are required")
	}
	helloSvc := service.NewHelloService()
	return &KubernetesSolutionV2{
		client:        client,
		queries:       queries,
		cfg:           cfg,
		service_hello: helloSvc,
	}, nil
}

func (s *KubernetesSolutionV2) Name() string    { return "kubernetes" }
func (s *KubernetesSolutionV2) Version() string { return "v2" }

// Endpoint : wiring des endpoints v2
func (s *KubernetesSolutionV2) Endpoint(rg *gin.RouterGroup) {
	v2 := rg.Group("")
	v2.GET("/hello", hellohandler.HelloHandler(s.service_hello))
	// Ici tu branches tes futurs endpoints v2 (/namespaces, /admin, ...)
}
