package service

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"time"

	kubernetesdb "github.com/Gskill75/api2/pkg/db/sqlc/kubernetes"
	k8sclient "github.com/Gskill75/api2/pkg/kubernetes/client"
	v1 "k8s.io/api/core/v1"
	k8serrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

type NamespaceService struct {
	Queries *kubernetesdb.Queries
	Client  *k8sclient.Client
}

var (
	ErrNamespaceNotFound = errors.New("namespace not found or not owned by customer")
	ErrForbiddenAccess   = errors.New("forbidden access to namespace")
	ErrDeleteK8sFailed   = errors.New("failed to delete in k8s")
	ErrDeleteDBFailed    = errors.New("failed to delete from db")
)

func NewNamespaceService(queries *kubernetesdb.Queries, client *k8sclient.Client) *NamespaceService {
	return &NamespaceService{
		Queries: queries,
		Client:  client,
	}
}

func (s *NamespaceService) GetCustomerNamespace(ctx context.Context, name, customerID string) (*kubernetesdb.Namespace, error) {
	if name == "" {
		return nil, errors.New("namespace name is required")
	}

	ns, err := s.Queries.GetNamespace(ctx, kubernetesdb.GetNamespaceParams{
		Name:       name,
		CustomerID: customerID,
	})
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, ErrNamespaceNotFound
		}
		return nil, err // Erreur technique (DB, etc.)
	}

	// Vérification supplémentaire de propriété
	if ns.CustomerID != customerID {
		return nil, ErrForbiddenAccess
	}

	return &ns, nil
}

type CreateNamespaceParams struct {
	Name       string
	CustomerID string
	Email      string
}

type CreateNamespaceResult struct {
	Name       string
	CustomerID string
	CreatedBy  string
	CreatedAt  time.Time
}

// Fonction métier principal :
func (s *NamespaceService) CreateNamespace(ctx context.Context, p CreateNamespaceParams) (*CreateNamespaceResult, error) {
	// (1) Vérification existence côté K8s
	_, err := s.Client.Clientset().CoreV1().Namespaces().Get(ctx, p.Name, metav1.GetOptions{})
	if err == nil {
		return nil, ErrAlreadyExistsK8s
	}
	if !k8serrors.IsNotFound(err) {
		return nil, fmt.Errorf("k8s_api_error: %w", err)
	}

	// (2) Existence côté DB
	_, dbErr := s.Queries.GetNamespace(ctx, kubernetesdb.GetNamespaceParams{
		Name:       p.Name,
		CustomerID: p.CustomerID,
	})
	if dbErr == nil {
		return nil, ErrAlreadyExistsDB
	}
	if !errors.Is(dbErr, sql.ErrNoRows) {
		return nil, fmt.Errorf("db_error: %w", dbErr)
	}

	// (3) Création K8s
	ns := &v1.Namespace{
		ObjectMeta: metav1.ObjectMeta{
			Name: p.Name,
			Annotations: map[string]string{
				"customer-id":                p.CustomerID,
				"created-by":                 p.Email,
				"openshift.io/node-selector": "client=mut",
			},
		},
	}
	_, err = s.Client.Clientset().CoreV1().Namespaces().Create(ctx, ns, metav1.CreateOptions{})
	if err != nil {
		return nil, fmt.Errorf("k8s_create_error: %w", err)
	}

	// (4) Création DB
	err = s.Queries.InsertNamespace(ctx, kubernetesdb.InsertNamespaceParams{
		Name:       p.Name,
		CustomerID: p.CustomerID,
		CreatedBy:  p.Email,
	})
	if err != nil {
		return nil, fmt.Errorf("db_create_error: %w", err)
	}

	return &CreateNamespaceResult{
		Name:       p.Name,
		CustomerID: p.CustomerID,
		CreatedBy:  p.Email,
		CreatedAt:  time.Now(), // ou mieux: retourne la vraie date si dispo
	}, nil
}

// Erreurs métiers
var (
	ErrAlreadyExistsK8s = errors.New("namespace already exists in k8s")
	ErrAlreadyExistsDB  = errors.New("namespace already exists in db")
)

func (s *NamespaceService) ListNamespacesByCustomer(ctx context.Context, customerID string) ([]kubernetesdb.Namespace, error) {
	// Ne fait que déléguer au repo/db, mais centralise le métier
	return s.Queries.ListNamespacesByCustomerID(ctx, customerID)
}

func (s *NamespaceService) DeleteNamespace(ctx context.Context, name, customerID string) error {
	if name == "" {
		return errors.New("namespace name is required")
	}

	ns, err := s.Queries.GetNamespace(ctx, kubernetesdb.GetNamespaceParams{
		Name: name, CustomerID: customerID,
	})
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return ErrNamespaceNotFound
		}
		return err
	}

	if ns.CustomerID != customerID {
		return ErrForbiddenAccess
	}

	// Suppression du namespace en K8s (ignore erreur NotFound)
	err = s.Client.Clientset().CoreV1().Namespaces().Delete(ctx, name, metav1.DeleteOptions{})
	if err != nil && !k8serrors.IsNotFound(err) {
		return ErrDeleteK8sFailed
	}

	// Suppression en base
	_, err = s.Queries.DeleteNamespace(ctx, kubernetesdb.DeleteNamespaceParams{
		Name: name, CustomerID: customerID,
	})
	if err != nil {
		return ErrDeleteDBFailed
	}

	return nil
}
