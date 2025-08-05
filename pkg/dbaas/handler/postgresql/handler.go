package handler_postgresql

import (
	"fmt"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	awxclient "github.com/Gskill75/api2/pkg/awx/client"
	db "github.com/Gskill75/api2/pkg/db/sqlc/postgresql"
	service "github.com/Gskill75/api2/pkg/dbaas/service/postgresql"
	"github.com/Gskill75/api2/pkg/utils"
	"k8s.io/klog/v2"
)

type ProvisionPostgresRequest struct {
	TemplateName string `json:"template_name" binding:"required"`
	InstanceName string `json:"instance_name"`
	Username     string `json:"username"`
	Password     string `json:"password"`
	CustomerID   string `json:"customer_id"`
}

// ProvisionPostgresHandler godoc
// @Summary     Provision a PostgreSQL instance
// @Description Provisions a new PostgreSQL instance using AWX automation
// @Tags        dbaas - PostgreSQL
// @Accept      json
// @Produce     json
// @Param       request body ProvisionPostgresRequest true "instance provisioning request"
// @Success     200 {object} string "instance provisioned successfully"
// @Failure     400 {object} map[string]string "Invalid request body"
// @Failure     401 {object} map[string]string "Unauthorized or missing customer_id"
// @Failure     500 {object} map[string]string "Internal server error"
// @Router      /postgres/v1/patroni/instance [post]
// @Security Bearer
func ProvisionPostgresHandler(awxClient *awxclient.Client, queries *db.Queries, postgresService *service.PostgresService) gin.HandlerFunc {
	return func(c *gin.Context) {
		rid := c.GetString("request_id")
		customerID, ok := utils.GetCustomerIDOrAbort(c)
		if !ok {
			return
		}

		var req ProvisionPostgresRequest
		if err := c.ShouldBindJSON(&req); err != nil {
			klog.Warningf("[request_id=%s] Invalid request body: %v", rid, err)
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request body", "request_id": rid})
			return
		}

		// test avec l'authentication desactive
		//customerID := "1"

		serviceReq := service.PostgresProvisionRequest{
			TemplateName: req.TemplateName,
			InstanceName: req.InstanceName,
			Username:     req.Username,
			Password:     req.Password,
			CustomerID:   customerID,
		}

		response, err := postgresService.ProvisionDatabase(c.Request.Context(), serviceReq, c.GetString("sub"))
		if err != nil {
			klog.Errorf("[request_id=%s] Failed to provision PostgreSQL database: %v", rid, err)

			// Gestion des erreurs un peu nul ?
			if strings.Contains(err.Error(), "awx_launch_failed") {
				c.JSON(http.StatusBadGateway, gin.H{"error": "External service unavailable", "request_id": rid})
				return
			}
			if strings.Contains(err.Error(), "template_name_not_found") {
				c.JSON(http.StatusInternalServerError, gin.H{"error": "Template not found", "request_id": rid})
				return
			}
			if strings.Contains(err.Error(), "db_insert_failed") {
				c.JSON(http.StatusInternalServerError, gin.H{"error": "Job launched but DB insert failed", "request_id": rid})
				return
			}

			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to provision database", "request_id": rid})
			return
		}

		klog.Infof("[request_id=%s] PostgreSQL instance provisioning started for customer %s, job_id=%d", rid, "1", response.JobID)
		c.JSON(http.StatusAccepted, gin.H{
			"message":       "PostgreSQL instance provisioning started",
			"job_id":        response.JobID,
			"status":        response.Status,
			"instance_name": response.InstanceName,
			"request_id":    rid,
		})
	}
}

// GetJobStatusHandler godoc
// @Summary     Get job status
// @Description Get the status of a PostgreSQL provisioning job
// @Tags        dbaas - PostgreSQL
// @Accept      json
// @Produce     json
// @Param       job_id path int true "Job ID"
// @Success     200 {object} map[string]interface{} "Job status"
// @Failure     400 {object} map[string]string "Invalid job ID"
// @Failure     404 {object} map[string]string "Job not found"
// @Failure     500 {object} map[string]string "Internal server error"
// @Router      /postgres/v1/patroni/instance/{job_id}/status [get]
// @Security Bearer
func GetJobStatusHandler(postgresService *service.PostgresService) gin.HandlerFunc {
	return func(c *gin.Context) {
		rid := c.GetString("request_id")

		jobIDStr := c.Param("job_id")
		jobID := 0
		if _, err := fmt.Sscanf(jobIDStr, "%d", &jobID); err != nil || jobID <= 0 {
			klog.Warningf("[request_id=%s] Invalid job ID: %s", rid, jobIDStr)
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid job ID", "request_id": rid})
			return
		}

		status, err := postgresService.GetJobStatus(c.Request.Context(), jobID)
		if err != nil {
			klog.Errorf("[request_id=%s] Failed to get job status for job %d: %v", rid, jobID, err)

			if strings.Contains(err.Error(), "failed_to_get_job_status") {
				c.JSON(http.StatusNotFound, gin.H{"error": "Job not found", "request_id": rid})
				return
			}

			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get job status", "request_id": rid})
			return
		}

		klog.Infof("[request_id=%s] Job %d status: %s", rid, jobID, status.Status)
		c.JSON(http.StatusOK, gin.H{
			"job_id":     status.JobID,
			"status":     status.Status,
			"request_id": rid,
		})
	}
}

// CheckActiveJobHandler godoc
// @Summary     Check for active jobs
// @Description Check if customer has an active job for the given template
// @Tags        dbaas - PostgreSQL
// @Accept      json
// @Produce     json
// @Param       template_name query string true "Template name"
// @Success     200 {object} map[string]interface{} "Active job status"
// @Failure     400 {object} map[string]string "Missing template name"
// @Failure     500 {object} map[string]string "Internal server error"
// @Router      /postgres/v1/patroni/instance/check [get]
// @Security Bearer
func CheckActiveJobHandler(postgresService *service.PostgresService) gin.HandlerFunc {
	return func(c *gin.Context) {
		rid := c.GetString("request_id")

		templateName := c.Query("template_name")
		if templateName == "" {
			klog.Warningf("[request_id=%s] Missing template_name parameter", rid)
			c.JSON(http.StatusBadRequest, gin.H{"error": "Missing template_name parameter", "request_id": rid})
			return
		}

		// For now using hardcoded customer ID
		customerID := "1"

		activeJob, err := postgresService.CheckActiveJob(c.Request.Context(), customerID, templateName)
		if err != nil {
			klog.Errorf("[request_id=%s] Failed to check active jobs for template %s: %v", rid, templateName, err)
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to check active jobs", "request_id": rid})
			return
		}

		if activeJob != nil {
			klog.Infof("[request_id=%s] Found active job %d for template %s", rid, activeJob.JobID, templateName)
			c.JSON(http.StatusOK, gin.H{
				"has_active_job": true,
				"job_id":         activeJob.JobID,
				"status":         activeJob.Status,
				"template_name":  templateName,
				"customer_id":    customerID,
				"request_id":     rid,
			})
		} else {
			klog.Infof("[request_id=%s] No active job found for template %s", rid, templateName)
			c.JSON(http.StatusOK, gin.H{
				"has_active_job": false,
				"template_name":  templateName,
				"customer_id":    customerID,
				"request_id":     rid,
			})
		}
	}
}
