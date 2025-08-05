package client

import (
	"bytes"
	"context"
	"crypto/tls"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

	"k8s.io/klog/v2"
)

// Requester handles HTTP requests to AWX API
type Requester struct {
	httpClient *http.Client
	baseURL    string
	username   string
	password   string
	token      string
	bearer     string
}

// NewRequester creates a new HTTP requester for AWX API
func NewRequester(baseURL, username, password, token, bearer string, insecure bool) *Requester {
	return &Requester{
		httpClient: newHTTPClient(insecure),
		baseURL:    baseURL,
		username:   username,
		password:   password,
		token:      token,
		bearer:     bearer,
	}
}

// setCredentials sets the appropriate authentication headers based on available credentials
func (r *Requester) setCredentials(req *http.Request) {
	if r.bearer != "" {
		req.Header.Set("Authorization", "Bearer "+r.bearer)
		klog.V(4).Info("Using Bearer token authentication")
	} else if r.token != "" {
		req.Header.Set("Authorization", "Token "+r.token)
		klog.V(4).Info("Using Token authentication")
	} else if r.username != "" && r.password != "" {
		req.SetBasicAuth(r.username, r.password)
		klog.V(4).Info("Using Basic authentication")
	}
}

// MakeRequest performs HTTP request to AWX API
func (r *Requester) MakeRequest(ctx context.Context, method, endpoint string, body interface{}) (*http.Response, error) {
	var url string

	// Check if baseURL already contains /api/v2/
	if !strings.HasSuffix(r.baseURL, "/api/v2/") {
		url = fmt.Sprintf("%s/api/v2%s", r.baseURL, endpoint)
	}

	var reqBody io.Reader
	if body != nil {
		jsonBody, err := json.Marshal(body)
		if err != nil {
			return nil, fmt.Errorf("failed to marshal request body: %w", err)
		}
		reqBody = bytes.NewBuffer(jsonBody)
	}

	klog.Infof("Making AWX API request: %s %s", method, url)

	req, err := http.NewRequestWithContext(ctx, method, url, reqBody)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	// Set authentication headers
	r.setCredentials(req)
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Accept", "application/json")

	resp, err := r.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to execute request: %w", err)
	}

	if resp.StatusCode >= 400 {
		body, _ := io.ReadAll(resp.Body)
		resp.Body.Close()
		return nil, fmt.Errorf("AWX API error (status %d): %s", resp.StatusCode, string(body))
	}

	return resp, nil
}

// newHTTPClient creates a new HTTP client with TLS configuration
func newHTTPClient(insecure bool) *http.Client {
	transport := &http.Transport{
		TLSClientConfig: &tls.Config{
			InsecureSkipVerify: insecure,
		},
	}

	return &http.Client{
		Transport: transport,
		Timeout:   30 * time.Second,
	}
}
