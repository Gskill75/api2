package client

import (
	"context"
	"fmt"

	"github.com/Gskill75/api2/pkg/config"
	"k8s.io/klog/v2"
)

// Client represents an AWX API client
type Client struct {
	BaseURL   string
	Requester *Requester
}

// New initialise un nouveau client AWX en utilisant l'API
func New(cfg *config.Config) (*Client, error) {
	if cfg == nil {
		klog.Error("AWX config is nil")
		return nil, fmt.Errorf("awx config is nil")
	}
	if cfg.Awx.Url == "" {
		klog.Error("AWX config missing URL")
		return nil, fmt.Errorf("awx config missing required URL fields")
	}

	// Check authentication credentials - exactly one method must be provided
	authMethods := 0
	var providedMethods []string

	if cfg.Awx.Username != "" || cfg.Awx.Password != "" {
		if cfg.Awx.Username == "" || cfg.Awx.Password == "" {
			klog.Error("Basic authentication requires both username and password")
			return nil, fmt.Errorf("basic authentication requires both username and password")
		}
		authMethods++
		providedMethods = append(providedMethods, "basic auth (username/password)")
	}

	if cfg.Awx.Token != "" {
		authMethods++
		providedMethods = append(providedMethods, "token")
	}

	if cfg.Awx.Bearer != "" {
		authMethods++
		providedMethods = append(providedMethods, "bearer")
	}

	if authMethods == 0 {
		klog.Error("No authentication method provided")
		return nil, fmt.Errorf("exactly one authentication method is required: username/password, token, or bearer")
	}

	if authMethods > 1 {
		klog.Errorf("Multiple authentication methods provided: %v", providedMethods)
		return nil, fmt.Errorf("exactly one authentication method is required, but %d were provided: %v", authMethods, providedMethods)
	}

	klog.Infof("Initializing AWX client for %s", cfg.Awx.Url)

	requester := NewRequester(
		cfg.Awx.Url,
		cfg.Awx.Username,
		cfg.Awx.Password,
		cfg.Awx.Token,
		cfg.Awx.Bearer,
		cfg.Awx.Insecure,
	)

	return &Client{
		BaseURL:   cfg.Awx.Url,
		Requester: requester,
	}, nil
}

// JobService returns a new JobService
func (c *Client) JobService() *JobService {
	return NewJobService(c)
}

// JobTemplateService returns a new JobTemplateService
func (c *Client) JobTemplateService() *JobTemplateService {
	return NewJobTemplateService(c)
}

// TestConnection tests AWX API connectivity
func (c *Client) TestConnection(ctx context.Context) error {
	resp, err := c.Requester.MakeRequest(ctx, "GET", "/ping/", nil)
	if err != nil {
		return fmt.Errorf("failed to ping AWX: %w", err)
	}
	defer resp.Body.Close()

	klog.Infof("AWX ping successful, status: %d", resp.StatusCode)
	return nil
}
