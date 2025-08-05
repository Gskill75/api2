package namespace

import (
	"errors"

	"github.com/gin-gonic/gin"
	apierrors "gitn.sigma.fr/sigma/paas/api/api/pkg/errors"
	history "gitn.sigma.fr/sigma/paas/api/api/pkg/kubernetes/history"
	"gitn.sigma.fr/sigma/paas/api/api/pkg/kubernetes/service"
	"gitn.sigma.fr/sigma/paas/api/api/pkg/utils"
	"k8s.io/klog/v2"
)

// GetNamespaceHandler godoc
// @Summary      Get your namespace details
// @Description  Retrieves a namespace only if it belongs to your customer_id (from the JWT).
// @Tags         namespaces
// @Produce      json
// @Param        name path string true "Namespace name"
// @Success      200 {object} map[string]interface{} "Namespace details"
// @Failure      400 {object} map[string]string "Namespace name is required"
// @Failure      401 {object} map[string]string "Forbidden access to namespace"
// @Failure      404 {object} map[string]string "Namespace not found in database"
// @Router       /kubernetes/v1/namespaces/{name} [get]
// @Security Bearer
func GetNamespaceHandler(nsService *service.NamespaceService) gin.HandlerFunc {
	return func(c *gin.Context) {
		rid := c.GetString("request_id")
		customerID := c.GetString("customer_id")
		email := c.GetString("email")
		name := c.Param("name")

		// Validation des paramètres d'entrée
		if name == "" {
			klog.Warningf("[request_id=%s] Namespace name is required", rid)
			history.LogNamespaceHistory(
				c.Request.Context(), nsService.Queries, customerID,
				"get", "error", "", email, email,
				"Namespace name is empty", "Namespace name is required",
			)
			c.Error(apierrors.NewBadRequest("Namespace name is required"))
			return
		}

		// Appel du service métier
		ns, err := nsService.GetCustomerNamespace(c.Request.Context(), name, customerID)
		if err != nil {
			// Mapping des erreurs métier vers erreurs API
			switch {
			case errors.Is(err, service.ErrNamespaceNotFound):
				klog.Warningf("[request_id=%s] Namespace '%s' not found in database", rid, name)
				history.LogNamespaceHistory(
					c.Request.Context(), nsService.Queries, customerID,
					"get", "error", name, email, email,
					"Namespace not found", err.Error(),
				)
				c.Error(apierrors.NewNotFound("Namespace not found in your tenant"))
				return

			case errors.Is(err, service.ErrForbiddenAccess):
				klog.Warningf("[request_id=%s] Namespace '%s' access denied for customer '%s'", rid, name, customerID)
				history.LogNamespaceHistory(
					c.Request.Context(), nsService.Queries, customerID,
					"get", "error", name, email, email,
					"Forbidden access", err.Error(),
				)
				c.Error(apierrors.NewNotFound("Namespace not found in your tenant"))
				return

			default:
				klog.Errorf("[request_id=%s] Failed to query database: %v", rid, err)
				history.LogNamespaceHistory(
					c.Request.Context(), nsService.Queries, customerID,
					"get", "error", name, email, email,
					"Database/internal error", err.Error(),
				)
				c.Error(apierrors.NewInternalError("Failed to query database"))
				return
			}
		}

		// logger les lectures a voir si besoins
		/*
		   history.LogNamespaceHistory(
		       c.Request.Context(), nsService.Queries, customerID,
		       "get", "success", name, email, email, "", "",
		   )
		*/

		utils.APISuccess(c, gin.H{
			"name":        ns.Name,
			"customer_id": ns.CustomerID,
			"created_by":  ns.CreatedBy,
			"created_at":  ns.CreatedAt,
			"updated_at":  ns.UpdatedAt,
		})
	}
}

// createNSRequest represents the request body to create a namespace.
// swagger:model
type createNSRequest struct {
	Name string `json:"name" binding:"required,min=2,max=63"` // min=2 pour éviter "a", sinon min=1
}

// CreateNamespaceHandler godoc
// @Summary     Create a new Kubernetes namespace
// @Description Creates a Kubernetes namespace. The namespace is created under the customer ID associated with the JWT.
// @Tags        namespaces
// @Accept      json
// @Produce     json
// @Param       request body createNSRequest true "Namespace creation request"
// @Success     201 {object} map[string]interface{} "Namespace created successfully"
// @Failure     400 {object} map[string]string "Invalid request body"
// @Failure     401 {object} map[string]string "Unauthorized or missing customer_id"
// @Failure     409 {object} map[string]string "Namespace already exists"
// @Failure     500 {object} map[string]string "Internal server error"
// @Router      /kubernetes/v1/namespaces [post]
// @Security Bearer
func CreateNamespaceHandler(nsService *service.NamespaceService) gin.HandlerFunc {
	return func(c *gin.Context) {
		rid := c.GetString("request_id")
		customerID := c.GetString("customer_id")
		email := c.GetString("email")

		var req createNSRequest
		if err := c.ShouldBindJSON(&req); err != nil {
			klog.Warningf("[request_id=%s] Invalid request body: %v", rid, err)
			history.LogNamespaceHistory(
				c.Request.Context(), nsService.Queries, customerID,
				"create", "error", "", email, email, "Invalid request body", err.Error(),
			)
			c.Error(apierrors.NewBadRequest("Invalid request body"))
			return
		}

		// Appel unique à la logique métier
		result, err := nsService.CreateNamespace(c.Request.Context(), service.CreateNamespaceParams{
			Name:       req.Name,
			CustomerID: customerID,
			Email:      email,
		})
		switch {
		case errors.Is(err, service.ErrAlreadyExistsK8s):
			klog.Warningf("[request_id=%s] Namespace '%s' already exists in K8s", rid, req.Name)
			history.LogNamespaceHistory(
				c.Request.Context(), nsService.Queries, customerID,
				"create", "error", req.Name, email, email, "Already exists in K8s", "",
			)
			c.Error(apierrors.NewConflict("The namespace name is not available. Please choose another one."))
			return
		case errors.Is(err, service.ErrAlreadyExistsDB):
			klog.Warningf("[request_id=%s] Namespace '%s' already exists in DB", rid, req.Name)
			history.LogNamespaceHistory(
				c.Request.Context(), nsService.Queries, customerID,
				"create", "error", req.Name, email, email, "Already exists in DB", "",
			)
			c.Error(apierrors.NewConflict("The namespace name is not available. Please choose another one."))
			return
		case err != nil:
			klog.Errorf("[request_id=%s] Creation failed: %v", rid, err)
			history.LogNamespaceHistory(
				c.Request.Context(), nsService.Queries, customerID,
				"create", "error", req.Name, email, email, "Failed to create namespace", err.Error(),
			)
			c.Error(apierrors.NewInternalError("Failed to create namespace"))
			return
		}

		klog.Infof("[request_id=%s] Namespace %s created successfully for customer %s", rid, req.Name, customerID)
		history.LogNamespaceHistory(
			c.Request.Context(), nsService.Queries, customerID,
			"create", "success", req.Name, email, email, "", "",
		)
		utils.APISuccess(c, gin.H{
			"message":     "Namespace created successfully",
			"name":        result.Name,
			"customer_id": result.CustomerID,
			"created_by":  result.CreatedBy,
		})
	}
}

// Helper pour erreurs uniformes
func apiError(c *gin.Context, code int, msg string) {
	c.JSON(code, gin.H{
		"error":      msg,
		"request_id": c.GetString("request_id"),
	})
}

// GetByCustomerHandler godoc
// @Summary      List namespaces by customer
// @Description  Lists all Kubernetes namespaces for the authenticated customer.
// @Tags         namespaces
// @Produce      json
// @Success      200 {object} map[string]interface{} "List of namespaces"
// @Failure      500 {object} map[string]string "Failed to list namespaces"
// @Router       /kubernetes/v1/namespaces/customer [get]
// @Security     Bearer
func GetByCustomerHandler(nsService *service.NamespaceService) gin.HandlerFunc {
	return func(c *gin.Context) {
		rid := c.GetString("request_id")
		customerID := c.GetString("customer_id")

		namespaces, err := nsService.ListNamespacesByCustomer(c.Request.Context(), customerID)
		if err != nil {
			klog.Errorf("[request_id=%s] Failed to list namespaces for customer %s: %v", rid, customerID, err)
			history.LogNamespaceHistory(
				c.Request.Context(), nsService.Queries, customerID,
				"list", "error", "", "", "",
				"Failed to list namespaces", err.Error(),
			)
			c.Error(apierrors.NewInternalError("Failed to list namespaces"))
			return
		}

		var results []gin.H
		for _, ns := range namespaces {
			results = append(results, gin.H{
				"name":        ns.Name,
				"customer_id": ns.CustomerID,
				"created_by":  ns.CreatedBy,
				"created_at":  ns.CreatedAt,
				"updated_at":  ns.UpdatedAt,
			})
		}

		utils.APISuccess(c, gin.H{
			"namespaces": results,
			"count":      len(results),
		})
	}
}

// GetByCustomerAdminHandler godoc
// @Summary      [Admin] List namespaces for any customer
// @Description  Lists all Kubernetes namespaces belonging to the specified customer. Admin-only endpoint. Requires the user to have role "admin" in their JWT.
// @Tags         admin
// @Produce      json
// @Param        customerUniqueId path string true "Customer unique ID"
// @Success      200 {object} map[string]interface{} "List of namespaces"
// @Failure      400 {object} map[string]string "Customer ID is required"
// @Failure      403 {object} map[string]string "Only admin can access this resource"
// @Failure      500 {object} map[string]string "Failed to list namespaces"
// @Router       /kubernetes/v1/admin/customer/{customerUniqueId} [get]
// @Security     Bearer
func GetByCustomerAdminHandler(nsService *service.NamespaceService) gin.HandlerFunc {
	return func(c *gin.Context) {
		rid := c.GetString("request_id")
		customerID := c.Param("customerUniqueId")

		if customerID == "" {
			klog.Warningf("[request_id=%s] Customer ID is required", rid)
			c.Error(apierrors.NewBadRequest("Customer ID is required"))
			return
		}

		namespaces, err := nsService.ListNamespacesByCustomer(c.Request.Context(), customerID)
		if err != nil {
			klog.Errorf("[request_id=%s] Failed to list namespaces for customer %s: %v", rid, customerID, err)
			// Optionnel : ajouter audit admin ici
			c.Error(apierrors.NewInternalError("Failed to list namespaces"))
			return
		}

		var results []gin.H
		for _, ns := range namespaces {
			results = append(results, gin.H{
				"name":        ns.Name,
				"customer_id": ns.CustomerID,
				"created_by":  ns.CreatedBy,
				"created_at":  ns.CreatedAt,
				"updated_at":  ns.UpdatedAt,
			})
		}

		utils.APISuccess(c, gin.H{
			"namespaces":  results,
			"count":       len(results),
			"customer_id": customerID,
		})
	}
}

type deleteNamespaceAdminRequest struct {
	CustomerID string `json:"customer_id" binding:"required"`
}

// DeleteNamespaceAdminHandler godoc
// @Summary      [Admin] Delete namespace for any customer
// @Description  Deletes a namespace for the specified customer. Requires admin role in the JWT.
// @Tags         admin
// @Accept       json
// @Produce      json
// @Param        name path string true "Namespace name to delete"
// @Param        request body deleteNamespaceAdminRequest true "Customer ID"
// @Success      200 {object} map[string]interface{} "Namespace deleted successfully"
// @Failure      400 {object} map[string]string "Missing or invalid input"
// @Failure      403 {object} map[string]string "Unauthorized - admin role required"
// @Failure      404 {object} map[string]string "Namespace not found for the given customer"
// @Failure      500 {object} map[string]string "Internal server error"
// @Router       /kubernetes/v1/admin/namespaces/{name} [delete]
// @Security     Bearer
func DeleteNamespaceAdminHandler(nsService *service.NamespaceService) gin.HandlerFunc {
	return func(c *gin.Context) {
		rid := c.GetString("request_id")
		name := c.Param("name")
		email := c.GetString("email")

		// Validation paramètre path
		if name == "" {
			klog.Warningf("[request_id=%s] Namespace name is required", rid)
			history.LogNamespaceHistory(
				c.Request.Context(), nsService.Queries, "",
				"delete", "error", name, email, email,
				"Missing namespace name", "Namespace name is required",
			)
			c.Error(apierrors.NewBadRequest("Namespace name is required"))
			return
		}

		// Validation body JSON (customerID attendu dans le POST/DELETE)
		var req deleteNamespaceAdminRequest
		if err := c.ShouldBindJSON(&req); err != nil || req.CustomerID == "" {
			klog.Warningf("[request_id=%s] Missing or invalid customer_id in body", rid)
			history.LogNamespaceHistory(
				c.Request.Context(), nsService.Queries, "",
				"delete", "error", name, email, email,
				"Missing or invalid customer_id", "Missing or invalid customer_id in body",
			)
			c.Error(apierrors.NewBadRequest("Missing or invalid customer_id"))
			return
		}

		// Appel du service métier mutualisé
		err := nsService.DeleteNamespace(c.Request.Context(), name, req.CustomerID)

		switch {
		case errors.Is(err, service.ErrNamespaceNotFound):
			klog.Warningf("[request_id=%s] Namespace '%s' not found for customer '%s'", rid, name, req.CustomerID)
			history.LogNamespaceHistory(
				c.Request.Context(), nsService.Queries, req.CustomerID,
				"delete", "error", name, email, email,
				"Namespace not found for customer during delete", "Namespace not found",
			)
			c.Error(apierrors.NewNotFound("Namespace not found for customer"))
			return

		case errors.Is(err, service.ErrForbiddenAccess):
			klog.Warningf("[request_id=%s] Access denied for namespace '%s', customer '%s'", rid, name, req.CustomerID)
			history.LogNamespaceHistory(
				c.Request.Context(), nsService.Queries, req.CustomerID,
				"delete", "error", name, email, email,
				"Access denied for customer", "Access denied",
			)
			c.Error(apierrors.NewNotFound("Namespace not found for customer"))
			return

		case errors.Is(err, service.ErrDeleteK8sFailed):
			klog.Errorf("[request_id=%s] Failed to delete K8s namespace '%s'", rid, name)
			history.LogNamespaceHistory(
				c.Request.Context(), nsService.Queries, req.CustomerID,
				"delete", "error", name, email, email,
				"Failed to delete K8s namespace", err.Error(),
			)
			c.Error(apierrors.NewInternalError("Failed to delete namespace in Kubernetes"))
			return

		case errors.Is(err, service.ErrDeleteDBFailed):
			klog.Errorf("[request_id=%s] Failed to delete namespace from DB: %v", rid, err)
			history.LogNamespaceHistory(
				c.Request.Context(), nsService.Queries, req.CustomerID,
				"delete", "error", name, email, email,
				"Database error on admin delete", err.Error(),
			)
			c.Error(apierrors.NewInternalError("Failed to delete namespace from database"))
			return

		case err != nil:
			klog.Errorf("[request_id=%s] Unexpected admin delete error: %v", rid, err)
			history.LogNamespaceHistory(
				c.Request.Context(), nsService.Queries, req.CustomerID,
				"delete", "error", name, email, email,
				"Unknown error on admin delete", err.Error(),
			)
			c.Error(apierrors.NewInternalError("Failed to delete namespace"))
			return
		}

		klog.Infof("[request_id=%s] Admin '%s' deleted namespace '%s' for customer '%s'", rid, email, name, req.CustomerID)
		history.LogNamespaceHistory(
			c.Request.Context(), nsService.Queries, req.CustomerID,
			"delete", "success", name, email, email,
			"", "",
		)

		utils.APISuccess(c, gin.H{
			"message":     "Namespace deleted successfully",
			"name":        name,
			"customer_id": req.CustomerID,
		})
	}
}

// DeleteMyNamespaceHandler godoc
// @Summary      Delete a namespace for the current user
// @Description  Deletes a namespace belonging to the current authenticated customer.
// @Tags         namespaces
// @Produce      json
// @Param        name path string true "Namespace name to delete"
// @Success      200 {object} map[string]interface{} "Namespace deleted successfully"
// @Failure      400 {object} map[string]string "Namespace name is required"
// @Failure      404 {object} map[string]string "Namespace not found for your tenant"
// @Failure      500 {object} map[string]string "Internal server error"
// @Router       /kubernetes/v1/namespaces/{name} [delete]
// @Security     Bearer
func DeleteNamespaceHandler(nsService *service.NamespaceService) gin.HandlerFunc {
	return func(c *gin.Context) {
		//rid := c.GetString("request_id")
		customerID := c.GetString("customer_id")
		name := c.Param("name")
		email := c.GetString("email")

		if name == "" {
			history.LogNamespaceHistory(
				c.Request.Context(), nsService.Queries, customerID,
				"delete", "error", "", email, email,
				"Missing namespace name", "Namespace name is required",
			)
			c.Error(apierrors.NewBadRequest("Namespace name is required"))
			return
		}

		err := nsService.DeleteNamespace(c.Request.Context(), name, customerID)
		switch {
		case errors.Is(err, service.ErrNamespaceNotFound), errors.Is(err, service.ErrForbiddenAccess):
			history.LogNamespaceHistory(
				c.Request.Context(), nsService.Queries, customerID,
				"delete", "error", name, email, email,
				"Namespace not found or access denied", "Namespace not found for your tenant",
			)
			c.Error(apierrors.NewNotFound("Namespace not found for your tenant"))
			return

		case errors.Is(err, service.ErrDeleteK8sFailed):
			history.LogNamespaceHistory(
				c.Request.Context(), nsService.Queries, customerID,
				"delete", "error", name, email, email,
				"K8s error during delete", err.Error(),
			)
			c.Error(apierrors.NewInternalError("Failed to delete namespace in Kubernetes"))
			return

		case errors.Is(err, service.ErrDeleteDBFailed):
			history.LogNamespaceHistory(
				c.Request.Context(), nsService.Queries, customerID,
				"delete", "error", name, email, email,
				"Database error during delete", err.Error(),
			)
			c.Error(apierrors.NewInternalError("Failed to delete namespace from database"))
			return

		case err != nil:
			// Pour toute autre erreur technique
			history.LogNamespaceHistory(
				c.Request.Context(), nsService.Queries, customerID,
				"delete", "error", name, email, email,
				"Generic error during delete", err.Error(),
			)
			c.Error(apierrors.NewInternalError("Failed to delete namespace"))
			return
		}

		// Log succès
		history.LogNamespaceHistory(
			c.Request.Context(), nsService.Queries, customerID,
			"delete", "success", name, email, email, "", "",
		)
		utils.APISuccess(c, gin.H{
			"message":     "Namespace deleted successfully",
			"name":        name,
			"customer_id": customerID,
		})
	}
}
