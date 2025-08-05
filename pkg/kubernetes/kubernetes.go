package kubernetes

import (
	"fmt"
	"time"

	"github.com/gin-gonic/gin"
	"gitn.sigma.fr/sigma/paas/api/api/pkg/config"
	db "gitn.sigma.fr/sigma/paas/api/api/pkg/db/sqlc/kubernetes"
	kubeclient "gitn.sigma.fr/sigma/paas/api/api/pkg/kubernetes/client"
	namespacehandler "gitn.sigma.fr/sigma/paas/api/api/pkg/kubernetes/handler/namespace"
	"gitn.sigma.fr/sigma/paas/api/api/pkg/kubernetes/service"
	"gitn.sigma.fr/sigma/paas/api/api/pkg/utils"
	"k8s.io/klog/v2"
)

// KubernetesSolution gère l'API Kubernetes avec validation et gestion d'erreurs centralisée
type KubernetesSolution struct {
	client     *kubeclient.Client
	queries    *db.Queries
	cfg        *config.Config
	service_ns *service.NamespaceService
}

// NewKubernetesSolution initialise la solution avec validation des dépendances
func NewKubernetesSolution(cfg *config.Config, client *kubeclient.Client, queries *db.Queries) (*KubernetesSolution, error) {
	if cfg == nil {
		klog.Errorf("KubernetesSolution: config is required")
		return nil, fmt.Errorf("config is required")
	}
	if client == nil {
		klog.Errorf("KubernetesSolution: kubernetes client is required")
		return nil, fmt.Errorf("kubernetes client is required")
	}
	if queries == nil {
		klog.Errorf("KubernetesSolution: database queries are required")
		return nil, fmt.Errorf("database queries are required")
	}

	nsService := service.NewNamespaceService(queries, client)
	return &KubernetesSolution{
		client:     client,
		queries:    queries,
		cfg:        cfg,
		service_ns: nsService,
	}, nil
}

func (s *KubernetesSolution) Name() string {
	return "kubernetes"
}

func (s *KubernetesSolution) Version() string {
	return "v1"
}

// Endpoint configure les routes avec middleware d'authentification et de logging
func (s *KubernetesSolution) Endpoint(rg *gin.RouterGroup) {
	// Middleware de logging spécifique Kubernetes
	rg.Use(s.kubernetesLoggingMiddleware())

	// Routes pour les utilisateurs standards
	s.setupNamespaceRoutes(rg)

	// Routes d'administration
	s.setupAdminRoutes(rg)
}

// kubernetesLoggingMiddleware ajoute un logging spécifique aux opérations Kubernetes
func (s *KubernetesSolution) kubernetesLoggingMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		start := time.Now()
		path := c.Request.URL.Path

		c.Next()

		latency := time.Since(start)
		status := c.Writer.Status()

		klog.InfoS("kubernetes_request",
			"method", c.Request.Method,
			"path", path,
			"status", status,
			"latency", latency,
			"customer_id", c.GetString("customer_id"),
			"request_id", c.GetString("request_id"),
		)
	}
}

// setupNamespaceRoutes configure les routes pour les utilisateurs standards
func (s *KubernetesSolution) setupNamespaceRoutes(rg *gin.RouterGroup) {
	nsGroup := rg.Group("/namespaces")

	// Middleware d'authentification utilisateur standard pour ce groupe
	nsGroup.Use(func(c *gin.Context) {
		if _, ok := utils.GetCustomerIDOrAbort(c); !ok {
			return // utils gère déjà la réponse d'erreur
		}
		c.Next()
	})

	{
		nsGroup.GET("/customer", namespacehandler.GetByCustomerHandler(s.service_ns))
		nsGroup.POST("", namespacehandler.CreateNamespaceHandler(s.service_ns))
		nsGroup.GET("/:name", namespacehandler.GetNamespaceHandler(s.service_ns))
		nsGroup.DELETE("/:name", namespacehandler.DeleteNamespaceHandler(s.service_ns))
	}
}

// setupAdminRoutes configure les routes d'administration
func (s *KubernetesSolution) setupAdminRoutes(rg *gin.RouterGroup) {
	adminGroup := rg.Group("/admin")

	// Middleware d'authentification admin pour ce groupe
	adminGroup.Use(func(c *gin.Context) {
		// Vérifier d'abord l'authentification de base
		if _, ok := utils.GetCustomerIDOrAbort(c); !ok {
			return
		}
		// Puis vérifier le rôle admin
		if !utils.IsAdminOrAbort(c) {
			return
		}
		c.Next()
	})

	{
		adminGroup.GET("/customer/:customerUniqueId", namespacehandler.GetByCustomerAdminHandler(s.service_ns))
		// adminGroup.POST("/", namespacehandler.CreateNamespaceHandler(s.client, s.queries))
		adminGroup.DELETE("/namespaces/:name", namespacehandler.DeleteNamespaceAdminHandler(s.service_ns))
	}
}
