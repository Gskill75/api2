package client

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"time"

	"k8s.io/klog/v2"
)

// JobService handles job operations
type JobService struct {
	client *Client
}

// NewJobService creates a new job service
func NewJobService(client *Client) *JobService {
	return &JobService{
		client: client,
	}
}

// GetJob retrieves job status by ID
func (js *JobService) GetJob(ctx context.Context, jobID int) (*Job, error) {
	endpoint := fmt.Sprintf("/jobs/%d/", jobID)

	resp, err := js.client.Requester.MakeRequest(ctx, "GET", endpoint, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to get job status: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read job response: %w", err)
	}

	var job Job
	if err := json.Unmarshal(body, &job); err != nil {
		return nil, fmt.Errorf("failed to unmarshal job response: %w", err)
	}

	return &job, nil
}

// GetRunningJobsForTemplate gets running jobs for a specific template
func (js *JobService) GetRunningJobsForTemplate(ctx context.Context, templateName string) ([]Job, error) {
	// First get template ID
	templateID, err := js.client.JobTemplateService().GetTemplateIDByName(ctx, templateName)
	if err != nil {
		return nil, err
	}

	// Get jobs for this template that are running
	endpoint := fmt.Sprintf("/jobs/?job_template=%d&status__in=pending,waiting,running", templateID)

	resp, err := js.client.Requester.MakeRequest(ctx, "GET", endpoint, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to get running jobs: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read jobs response: %w", err)
	}

	var jobsResp JobsResponse
	if err := json.Unmarshal(body, &jobsResp); err != nil {
		return nil, fmt.Errorf("failed to unmarshal jobs response: %w", err)
	}

	return jobsResp.Results, nil
}

// MonitorJob polls job status until completion
func (js *JobService) MonitorJob(ctx context.Context, jobID int) (*Job, error) {
	ticker := time.NewTicker(5 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return nil, ctx.Err()
		case <-ticker.C:
			job, err := js.GetJob(ctx, jobID)
			if err != nil {
				return nil, err
			}

			klog.Infof("Job %d status: %s", jobID, job.Status)

			switch job.Status {
			case "successful":
				return job, nil
			case "failed", "error", "canceled":
				return job, fmt.Errorf("job failed with status: %s", job.Status)
			}

		}
	}
}
