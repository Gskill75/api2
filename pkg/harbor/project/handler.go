package project

import (
	"context"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/go-openapi/runtime"
	"github.com/goharbor/go-client/pkg/sdk/v2.0/client/project"
	"github.com/goharbor/go-client/pkg/sdk/v2.0/models"
	db "gitn.sigma.fr/sigma/paas/api/api/pkg/db/sqlc/harbor"
	harborclient "gitn.sigma.fr/sigma/paas/api/api/pkg/harbor/client"

	"k8s.io/klog/v2"
)

// CreateProjectRequest : structure attendue via la requête API interne
type createProjectRequest struct {
	Name         string `json:"project_name" binding:"required"`
	Public       bool   `json:"public"`
	CustomerID   string `json:"customer_id"` // optionnel sauf pour admin
	CreatedBy    string `json:"created_by" binding:"required"`
	StorageLimit int64  `json:"storage_limit"`
}

// harborAPIProject représente le JSON attendu par Harbor
type harborAPIProject struct {
	ProjectName string `json:"project_name"`
	Public      int    `json:"public,string"` // Harbor attend un int entre 0/1 sous forme string éventuellement
	Metadata    struct {
		Public bool `json:"public,string"`
	} `json:"metadata,omitempty"`
	CountLimit     int  `json:"count_limit"`
	StorageLimit   int  `json:"storage_limit"`
	ClearTextArtif bool `json:"enable_content_trust"`
}

// CreateProjectHandler godoc
// @Summary     Create a new Harbor Project
func CreateProjectHandler(h *harborclient.Client, queries *db.Queries) gin.HandlerFunc {
	return func(c *gin.Context) {
		rid := c.GetString("request_id")

		var req createProjectRequest
		if err := c.ShouldBindJSON(&req); err != nil {
			klog.Warningf("[request_id=%s] Invalid request body: %v", rid, err)
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request body", "request_id": rid})
			return
		}

		role := c.GetString("role")
		customerID := c.GetString("customer_id")
		if role == "admin" && req.CustomerID != "" {
			customerID = req.CustomerID
		}
		if customerID == "" {
			klog.Warningf("[request_id=%s] Missing customer_id", rid)
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Missing customer_id", "request_id": rid})
			return
		}

		// Vérifie si le projet existe déjà
		headParams := project.NewHeadProjectParams().WithProjectName(req.Name)
		_, err := h.ClientSet().V2().Project.HeadProject(context.Background(), headParams)
		if err == nil {
			klog.Warningf("[request_id=%s] Harbor project '%s' already exists", rid, req.Name)
			c.JSON(http.StatusConflict, gin.H{"error": "Project already exists", "request_id": rid})
			return
		} else {
			if apiErr, ok := err.(*runtime.APIError); !ok || apiErr.Code != http.StatusNotFound {
				klog.Errorf("[request_id=%s] Failed to check Harbor project: %v", rid, err)
				c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to check project existence", "request_id": rid})
				return
			}
		}

		// Prépare le projet à créer
		p := &models.ProjectReq{
			ProjectName: req.Name,
			Metadata: &models.ProjectMetadata{
				Public:             boolToString(req.Public),
				AutoScan:           strPtr("true"),
				AutoSbomGeneration: strPtr("true"),
				PreventVul:         strPtr("true"),
				Severity:           strPtr("critical"),
			},
		}

		createParams := project.NewCreateProjectParams().WithProject(p)

		_, err = h.ClientSet().V2().Project.CreateProject(context.Background(), createParams)
		if err != nil {
			klog.Errorf("[request_id=%s] Failed to create Harbor project: %v", rid, err)
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Harbor create failed", "request_id": rid})
			return
		}

		// Enregistrement DB
		err = queries.InsertHarborProject(c.Request.Context(), db.InsertHarborProjectParams{
			Name:       req.Name,
			CustomerID: customerID,
			CreatedBy:  req.CreatedBy,
		})
		if err != nil {
			klog.Errorf("[request_id=%s] Harbor project created but failed to insert into DB: %v", rid, err)
			c.JSON(http.StatusInternalServerError, gin.H{"error": "DB insert failed", "request_id": rid})
			return
		}

		klog.Infof("[request_id=%s] Harbor project '%s' created for customer '%s'", rid, req.Name, customerID)
		c.JSON(http.StatusCreated, gin.H{
			"message":     "Harbor project created successfully",
			"name":        req.Name,
			"customer_id": customerID,
			"created_by":  req.CreatedBy,
			"request_id":  rid,
		})
	}
}

func boolToString(b bool) string {
	if b {
		return "true"
	}
	return "false"
}

func strPtr(s string) *string {
	return &s
}
