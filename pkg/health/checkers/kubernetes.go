package checkers

import (
	"context"
	"fmt"
	"time"

	k8sclient "github.com/Gskill75/api2/pkg/kubernetes/client"
	"k8s.io/klog/v2"
)

type KubernetesChecker struct {
	NameStr string
	Client  *k8sclient.Client
}

func (k KubernetesChecker) Name() string {
	if k.NameStr == "" {
		return "kubernetes"
	}
	return k.NameStr
}

func (k KubernetesChecker) Check() error {
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	start := time.Now()

	err := k.Client.ListNamespaces(ctx)
	duration := time.Since(start)
	if err != nil {
		// Gestion d'erreurs simple mais informative
		if ctx.Err() == context.DeadlineExceeded {
			klog.ErrorS(err, "kubernetes API timeout", "duration", duration)
			return fmt.Errorf("kubernetes API timeout after %v", duration)
		}

		// Erreur de permissions ou connectivité
		klog.ErrorS(err, "kubernetes API check failed", "duration", duration)
		return fmt.Errorf("kubernetes API unavailable: %w", err)
	}

	// Succès - API répond et permissions OK
	klog.V(2).InfoS("kubernetes health check success", "duration", duration)
	return nil
}
