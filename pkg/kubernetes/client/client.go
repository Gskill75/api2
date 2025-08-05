package client

import (
	"context"
	"fmt"
	"sync"
	"time"

	"gitn.sigma.fr/sigma/paas/api/api/pkg/config"
	db "gitn.sigma.fr/sigma/paas/api/api/pkg/db/sqlc/kubernetes"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	kubeclient "k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/util/flowcontrol"
	"k8s.io/klog/v2"
)

// Client encapsule l'accès au cluster Kubernetes et à la couche BDD.
type Client struct {
	cfg       *config.Config
	clientset *kubeclient.Clientset
	queries   *db.Queries
	// Add shutdown management
	shutdownCtx    context.Context
	shutdownCancel context.CancelFunc
	shutdownWG     sync.WaitGroup
}

// New crée un client Kubernetes configuré à partir du fichier config
func New(cfg *config.Config, queries *db.Queries) (*Client, error) {
	if cfg == nil {
		klog.Error("Missing configuration (cfg) when initializing Kubernetes client")
		return nil, fmt.Errorf("config is required")
	}
	if queries == nil {
		klog.Error("Missing DB queries when initializing Kubernetes client")
		return nil, fmt.Errorf("db.Queries is required")
	}
	if cfg.Kubernetes.Url == "" {
		klog.Error("Missing Kubernetes cluster URL in config")
		return nil, fmt.Errorf("kubernetes url is required in config")
	}
	if cfg.Kubernetes.Token == "" {
		klog.Warning("No Kubernetes token provided in config: access may fail or be limited.")
	}

	insecure := false
	if cfg.Kubernetes.Insecure {
		insecure = true
		klog.Warning("INSECURE mode enabled for Kubernetes client (TLS verification is disabled)!")
	}
	// Conversion
	qps := float32(cfg.Kubernetes.QPS)
	burst := cfg.Kubernetes.Burst
	if qps <= 0 {
		qps = 50
	}
	if burst <= 0 {
		burst = 100
	}

	klog.Infof("Initializing Kubernetes client for cluster: %s (insecure=%v)", cfg.Kubernetes.Url, insecure)

	restConfig := &rest.Config{
		Host:        cfg.Kubernetes.Url,
		BearerToken: cfg.Kubernetes.Token,
		TLSClientConfig: rest.TLSClientConfig{
			Insecure: insecure,
		},
		// Add rate limiting configuration
		QPS:         float32(cfg.Kubernetes.QPS),
		Burst:       cfg.Kubernetes.Burst,
		RateLimiter: flowcontrol.NewTokenBucketRateLimiter(qps, burst),
	}
	// Gestion du timeout global (en secondes → time.Duration)
	if cfg.Kubernetes.Timeout > 0 {
		restConfig.Timeout = time.Duration(cfg.Kubernetes.Timeout) * time.Second
	} else {
		restConfig.Timeout = 30 * time.Second // valeur de secours
	}

	clientset, err := kubeclient.NewForConfig(restConfig)
	if err != nil {
		klog.Errorf("Failed to initialize Kubernetes clientset: %v", err)
		return nil, fmt.Errorf("failed to init Kubernetes clientset: %w", err)
	}
	klog.Info("Kubernetes clientset successfully initialized")
	// Contexte d’arrêt commun.
	sdCtx, sdCancel := context.WithCancel(context.Background())

	return &Client{
		cfg:            cfg,
		clientset:      clientset,
		queries:        queries,
		shutdownCtx:    sdCtx,
		shutdownCancel: sdCancel,
	}, nil
}

// ListNamespaces vérifie que le cluster répond en listant les namespaces
func (k *Client) ListNamespaces(ctx context.Context) error {
	ctx, cancel := context.WithTimeout(k.shutdownCtx, 10*time.Second)
	defer cancel()

	_, err := k.clientset.CoreV1().Namespaces().List(ctx, metav1.ListOptions{})
	return err
}

func (c *Client) Close() error {
	klog.Info("Shutting down Kubernetes client")
	c.shutdownCancel()

	done := make(chan struct{})
	go func() {
		c.shutdownWG.Wait()
		close(done)
	}()
	timeout := 30 * time.Second
	if c.cfg != nil && c.cfg.Kubernetes.Timeout > 0 {
		timeout = time.Duration(c.cfg.Kubernetes.Timeout) * time.Second
	}
	select {
	case <-done:
		klog.Info("Kubernetes client shutdown completed gracefully")
		return nil
	case <-time.After(timeout):
		err := fmt.Errorf("kubernetes client shutdown exceeded %v", timeout)
		klog.Error(err)
		return err
	}
}

// Clientset expose le client brut pour usage avancé
func (k *Client) Clientset() *kubeclient.Clientset {
	return k.clientset
}
