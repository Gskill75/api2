package checkers

import (
	"context"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
	"k8s.io/klog/v2"
)

// DBChecker vérifie la santé de la connexion PostgreSQL
type DBChecker struct {
	NameStr string
	DB      *pgxpool.Pool
	Timeout time.Duration
}

// NewDBChecker crée une nouvelle instance avec timeout configurable
func NewDBChecker(db *pgxpool.Pool, timeout time.Duration) *DBChecker {
	if timeout <= 0 {
		timeout = 10 * time.Second // valeur par défaut
	}

	return &DBChecker{
		NameStr: "database",
		DB:      db,
		Timeout: timeout,
	}
}

// Name retourne le nom du checker
func (d DBChecker) Name() string {
	if d.NameStr != "" {
		return d.NameStr
	}
	return "database"
}

func (d DBChecker) Check() error {
	return d.CheckWithContext(context.Background())
}

// CheckWithContext effectue le health check avec un contexte spécifique
func (d DBChecker) CheckWithContext(parentCtx context.Context) error {
	// Créer un contexte avec timeout
	timeout := d.Timeout
	if timeout <= 0 {
		timeout = 10 * time.Second
	}

	ctx, cancel := context.WithTimeout(parentCtx, timeout)
	defer cancel()

	start := time.Now()

	// Vérifier la connectivité de base avec Ping
	if err := d.DB.Ping(ctx); err != nil {
		duration := time.Since(start)

		// Classification des erreurs pour un meilleur diagnostic
		if ctx.Err() == context.DeadlineExceeded {
			klog.ErrorS(err, "database health check timeout",
				"timeout", timeout,
				"duration", duration)
			return fmt.Errorf("database timeout after %v: %w", duration, err)
		}

		klog.ErrorS(err, "database health check failed",
			"duration", duration)
		return fmt.Errorf("database connection failed: %w", err)
	}

	duration := time.Since(start)
	klog.V(2).InfoS("database health check success",
		"duration", duration,
		"timeout", timeout)

	return nil
}
