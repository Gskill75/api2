package client

import (
	"context"
	"encoding/json"
	"fmt"
	"io"

	"k8s.io/klog/v2"
)

// JobTemplateService handles job template operations
type JobTemplateService struct {
	client *Client
}

// NewJobTemplateService creates a new job template service
func NewJobTemplateService(client *Client) *JobTemplateService {
	return &JobTemplateService{
		client: client,
	}
}

// GetTemplateIDByName retrieves job template ID by name
func (jts *JobTemplateService) GetTemplateIDByName(ctx context.Context, jobTemplate string) (int, error) {
	resp, err := jts.client.Requester.MakeRequest(ctx, "GET", "/job_templates/", nil)
	if err != nil {
		klog.Errorf("failed to get job templates: %v", err)
		return 0, fmt.Errorf("failed to get job templates: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return 0, fmt.Errorf("failed to read response body: %w", err)
	}

	var templatesResp JobTemplatesResponse
	if err := json.Unmarshal(body, &templatesResp); err != nil {
		return 0, fmt.Errorf("failed to unmarshal response: %w", err)
	}

	for _, template := range templatesResp.Results {
		if template.Name == jobTemplate {
			return template.ID, nil
		}
	}

	klog.Errorf("template name '%s' not found", jobTemplate)
	return 0, fmt.Errorf("template_name_not_found")
}

// LaunchJob launches a job template with optional extra variables
func (jts *JobTemplateService) LaunchJob(ctx context.Context, templateID int, extraVars map[string]any) (*JobLaunchResponse, error) {
	endpoint := fmt.Sprintf("/job_templates/%d/launch/", templateID)

	launchReq := JobLaunchRequest{
		ExtraVars: extraVars,
	}

	resp, err := jts.client.Requester.MakeRequest(ctx, "POST", endpoint, launchReq)
	if err != nil {
		return nil, fmt.Errorf("failed to launch job: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read launch response: %w", err)
	}

	var launchResp JobLaunchResponse
	if err := json.Unmarshal(body, &launchResp); err != nil {
		return nil, fmt.Errorf("failed to unmarshal launch response: %w", err)
	}

	return &launchResp, nil
}
