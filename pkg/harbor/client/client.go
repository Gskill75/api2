package client

import (
	"context"
	"fmt"

	harbor "github.com/goharbor/go-client/pkg/harbor"
	"github.com/goharbor/go-client/pkg/sdk/v2.0/client/project"
	"gitn.sigma.fr/sigma/paas/api/api/pkg/config"
	db "gitn.sigma.fr/sigma/paas/api/api/pkg/db/sqlc/harbor"
	"k8s.io/klog/v2"
)

type Client struct {
	cfg     *config.Config
	cs      *harbor.ClientSet
	queries *db.Queries
}

// New initializes a new Harbor client using the official ClientSet
func New(cfg *config.Config, queries *db.Queries) (*Client, error) {
	if cfg == nil {
		klog.Error("Harbor config is nil")
		return nil, fmt.Errorf("harbor config is nil")
	}
	if cfg.Harbor.Url == "" || cfg.Harbor.Username == "" || cfg.Harbor.Token == "" {
		klog.Error("Harbor config missing URL, Username or Token")
		return nil, fmt.Errorf("harbor config missing required fields")
	}
	if queries == nil {
		klog.Error("Missing DB queries when initializing Harbor client")
		return nil, fmt.Errorf("db.Queries is required")
	}

	klog.Infof("Initializing Harbor client for %s", cfg.Harbor.Url)

	cs, err := harbor.NewClientSet(&harbor.ClientSetConfig{
		URL:      cfg.Harbor.Url,
		Username: cfg.Harbor.Username,
		Password: cfg.Harbor.Token,
		Insecure: cfg.Harbor.Insecure,
	})
	if err != nil {
		klog.Errorf("failed to create Harbor ClientSet: %v", err)
		return nil, err
	}

	return &Client{
		cfg: cfg,
		cs:  cs,
	}, nil
}

// ClientSet expose le client Harbor pour les appels directs
func (c *Client) ClientSet() *harbor.ClientSet {
	return c.cs
}

func (c *Client) Ping() error {
	klog.V(2).Info("Checking Harbor connectivity via Ping")

	ctx := context.Background()
	page := int64(1)
	size := int64(1)
	params := project.NewListProjectsParams().
		WithPage(&page).
		WithPageSize(&size)

	_, err := c.ClientSet().V2().Project.ListProjects(ctx, params)
	if err != nil {
		klog.Errorf("Harbor ping failed: %v", err)
		return fmt.Errorf("harbor ping failed: %w", err)
	}

	klog.Info("Harbor ping successful")
	return nil
}
