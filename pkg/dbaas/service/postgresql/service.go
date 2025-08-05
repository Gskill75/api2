package service_postgresql

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/jackc/pgx/v5/pgtype"
	awxclient "gitn.sigma.fr/sigma/paas/api/api/pkg/awx/client"
	"gitn.sigma.fr/sigma/paas/api/api/pkg/config"
	db "gitn.sigma.fr/sigma/paas/api/api/pkg/db/sqlc/postgresql"
	"k8s.io/klog/v2"
)

type PostgresService struct {
	awxClient *awxclient.Client
	queries   *db.Queries
	cfg       *config.Config
}

type PostgresProvisionRequest struct {
	TemplateName string `json:"template_name"`
	InstanceName string `json:"instance_name"`
	Username     string `json:"username"`
	Password     string `json:"password"`
	CustomerID   string `json:"customer_id"`
}

type PostgresProvisionResponse struct {
	InstanceName string `json:"instance_name"`
	Username     string `json:"username"`
	JobID        int    `json:"job_id"`
	Status       string `json:"status"`
	CustomerID   string `json:"customer_id"`
}

func NewPostgresService(awxClient *awxclient.Client, queries *db.Queries, cfg *config.Config) *PostgresService {
	return &PostgresService{
		awxClient: awxClient,
		queries:   queries,
		cfg:       cfg,
	}
}

func (p *PostgresService) ProvisionDatabase(ctx context.Context, req PostgresProvisionRequest, createdBy string) (*PostgresProvisionResponse, error) {
	klog.Infof("Provisioning PostgreSQL instance - starting")

	templateID, err := p.awxClient.JobTemplateService().GetTemplateIDByName(ctx, req.TemplateName)
	if err != nil {
		return nil, fmt.Errorf("template_name_not_found: %w", err)
	}

	klog.Infof("Launching AWX job template '%d' with name '%v'", templateID, req.TemplateName)

	// Prepare extra variables for the job
	extraVars := map[string]interface{}{
		"instance_name": req.InstanceName,
		"custom":        "TESTIII",
		"username":      req.Username,
		"password":      req.Password,
		"customer_id":   req.CustomerID,
	}

	klog.Infof("Calling AWX API to launch job...")
	response, err := p.awxClient.JobTemplateService().LaunchJob(ctx, templateID, extraVars)
	if err != nil {
		klog.Errorf("Failed to launch PostgreSQL provisioning job: %v", err)
		return nil, fmt.Errorf("awx_launch_failed: %w", err)
	}
	klog.Infof("AWX API call successful - Job ID: %d", response.Job)

	// Convert extraVars to JSON for database storage
	extraVarsJSON, err := json.Marshal(extraVars)
	if err != nil {
		klog.Errorf("Failed to marshal extra_vars to JSON: %v", err)
		return nil, fmt.Errorf("failed_to_marshal_extra_vars: %w", err)
	}

	// Insert l'action en db
	historyRecord, err := p.queries.CreateHistory(ctx, db.CreateHistoryParams{
		InstanceName:    req.InstanceName,
		CustomerID:      req.CustomerID,
		AwxJobID:        pgtype.Int8{Int64: int64(response.Job), Valid: true},
		AwxTemplateName: pgtype.Text{String: req.TemplateName, Valid: true},
		AwxTemplateID:   pgtype.Int4{Int32: int32(templateID), Valid: true},
		ActionType:      "create",
		Status:          "running",
		ExtraVars:       extraVarsJSON,
		CreatedBy:       createdBy,
	})
	if err != nil {
		klog.Errorf("Job launched but failed to insert into DB: %v", err)
		return nil, fmt.Errorf("db_insert_failed: %w", err)
	}

	// Start monitoring job for status updates (use background context)
	err = p.DoMonitorJob(context.Background(), response.Job, historyRecord.ID)
	if err != nil {
		klog.Errorf("Failed to start job monitoring for job %d: %v", response.Job, err)
		// Don't return error as job is already launched and recorded
	}

	klog.Infof("PostgreSQL instance '%s' provisioning started for customer %s", req.InstanceName, req.CustomerID)

	return &PostgresProvisionResponse{
		InstanceName: req.InstanceName,
		Username:     req.Username,
		JobID:        response.Job,
		Status:       "running",
		CustomerID:   req.CustomerID,
	}, nil
}

// GetJobStatus retrieves the status of a provisioning job
func (p *PostgresService) GetJobStatus(ctx context.Context, jobID int) (*PostgresProvisionResponse, error) {
	job, err := p.awxClient.JobService().GetJob(ctx, jobID)
	if err != nil {
		return nil, fmt.Errorf("failed_to_get_job_status: %w", err)
	}

	var status string
	switch job.Status {
	case "successful":
		status = "completed"
	case "failed", "error", "canceled":
		status = "failed"
	case "pending", "waiting", "running":
		status = "running"
	default:
		status = job.Status
	}

	return &PostgresProvisionResponse{
		JobID:  job.ID,
		Status: status,
	}, nil
}

// CheckActiveJob pour verifier avant lancement si duplicat
func (p *PostgresService) CheckActiveJob(ctx context.Context, customerID, templateName string) (*PostgresProvisionResponse, error) {
	activeJobs, err := p.awxClient.JobService().GetRunningJobsForTemplate(ctx, templateName)
	if err != nil {
		return nil, fmt.Errorf("failed_to_get_running_job_for_template: %w", err)
	}

	klog.Infof("Found %d jobs for template %s", len(activeJobs), templateName)

	// Check if any jobs are in active states
	for _, job := range activeJobs {
		if job.Status == "running" || job.Status == "waiting" || job.Status == "pending" {
			klog.Infof("Found active job %d with status %s", job.ID, job.Status)

			var status string
			switch job.Status {
			case "running", "waiting", "pending":
				status = "running"
			default:
				status = job.Status
			}

			return &PostgresProvisionResponse{
				JobID:      job.ID,
				Status:     status,
				CustomerID: customerID,
			}, nil
		}
	}

	// No active jobs found
	klog.Infof("No active jobs found for template %s", templateName)
	return nil, nil
}

func (p *PostgresService) DoMonitorJob(ctx context.Context, jobID int, historyID int32) error {
	klog.Infof("Starting job monitoring for job ID: %d", jobID)

	go func() {
		defer func() {
			if r := recover(); r != nil {
				klog.Errorf("Job monitoring panic for job %d: %v", jobID, r)
			}
		}()

		// MonitorJob blocks until job completes, so we call it in goroutine
		finalJob, err := p.awxClient.JobService().MonitorJob(ctx, jobID)
		if err != nil {
			klog.Errorf("Failed to monitor job %d: %v", jobID, err)

			// Update database with failed status even if monitoring failed
			err = p.queries.UpdateHistoryCompletion(ctx, db.UpdateHistoryCompletionParams{
				ID:           historyID,
				Status:       db.StatusEnum(finalJob.Status),
				AwxStatus:    db.NullAwxStatusEnum{AwxStatusEnum: db.AwxStatusEnum("canceled"), Valid: true},
				ErrorMessage: pgtype.Text{String: err.Error(), Valid: true},
			})

			if err != nil {
				klog.Errorf("Failed to update failed status for job %d: %v", jobID, err)
			}
			return
		}

		// Map AWX status to our enum
		var status string
		switch finalJob.Status {
		case "successful":
			status = "completed"
		case "failed", "error", "canceled":
			status = "failed"
		default:
			status = "failed" // Default to failed if unknown
		}

		// Update database with final status
		err = p.queries.UpdateHistoryCompletion(ctx, db.UpdateHistoryCompletionParams{
			ID:           historyID,
			Status:       db.StatusEnum(status),
			AwxStatus:    db.NullAwxStatusEnum{AwxStatusEnum: db.AwxStatusEnum(finalJob.Status), Valid: true},
			ErrorMessage: pgtype.Text{String: "", Valid: finalJob.Status != "successful"},
		})

		if err != nil {
			klog.Errorf("Failed to update completion for job %d: %v", jobID, err)
		} else {
			klog.Infof("Job %d completed with status: %s", jobID, status)
		}
	}()

	return nil
}
