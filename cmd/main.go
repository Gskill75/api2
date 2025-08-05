package main

import (
	"context"
	"errors"
	"flag"
	"fmt"
	"net/http"
	"os/signal"
	"syscall"
	"time"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/spf13/cobra"
	swaggerfiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"
	"gitn.sigma.fr/sigma/paas/api/api/pkg/config"
	harbordb "gitn.sigma.fr/sigma/paas/api/api/pkg/db/sqlc/harbor"
	k8sdb "gitn.sigma.fr/sigma/paas/api/api/pkg/db/sqlc/kubernetes"
	postgresdb "gitn.sigma.fr/sigma/paas/api/api/pkg/db/sqlc/postgresql"

	awxclient "gitn.sigma.fr/sigma/paas/api/api/pkg/awx/client"
	"gitn.sigma.fr/sigma/paas/api/api/pkg/dbaas"
	"gitn.sigma.fr/sigma/paas/api/api/pkg/docs"
	"gitn.sigma.fr/sigma/paas/api/api/pkg/harbor"
	harborclient "gitn.sigma.fr/sigma/paas/api/api/pkg/harbor/client"
	"gitn.sigma.fr/sigma/paas/api/api/pkg/health"
	"gitn.sigma.fr/sigma/paas/api/api/pkg/health/checkers"
	"gitn.sigma.fr/sigma/paas/api/api/pkg/kubernetes"
	k8sclient "gitn.sigma.fr/sigma/paas/api/api/pkg/kubernetes/client"
	kubernetesv2 "gitn.sigma.fr/sigma/paas/api/api/pkg/kubernetes_v2"
	"gitn.sigma.fr/sigma/paas/api/api/pkg/middleware"
	"gitn.sigma.fr/sigma/paas/api/api/pkg/solutions"
	"k8s.io/klog/v2"
)

// @title API Self service for cloud
// @version 1.0
// @description Generic API for self-service cloud resources
// @BasePath /api
// @securityDefinitions.apikey Bearer
// @in header
// @name Authorization
// @description Type "Bearer" followed by a space and JWT token.
func main() {
	klog.InitFlags(nil)
	var rootCmd = &cobra.Command{
		Use:   "api",
		Short: "Self-service API for cloud resources",
		Run:   runApp,
	}
	var cfgFile string
	rootCmd.Flags().StringVar(&cfgFile, "config", "", "Path to config file (YAML)")
	rootCmd.PersistentFlags().AddGoFlagSet(flag.CommandLine)

	cobra.CheckErr(rootCmd.Execute())
}
func runApp(cmd *cobra.Command, args []string) {
	cfg, err := config.Load(cmd)
	if err != nil {
		klog.Fatalf("Unable to load config: %v", err)
	}
	ctx := context.Background()
	dsn := fmt.Sprintf(
		"postgres://%s:%s@%s:%s/%s?sslmode=%s",
		cfg.DB.User, cfg.DB.Password, cfg.DB.Host, cfg.DB.Port, cfg.DB.DbName, cfg.DB.SslMode,
	)
	// fmt.Println("hello db ", cfg.DB.User, cfg.DB.Password, cfg.DB.Host, cfg.DB.Port, cfg.DB.DbName, cfg.DB.SslMode)

	pool, err := pgxpool.New(ctx, dsn)
	if err != nil {
		klog.Fatalf("Unable to connect to database: %v", err)
	}
	defer pool.Close()

	if err := pool.Ping(ctx); err != nil {
		klog.Fatalf("Unable to ping database: %v", err)
	}
	//initialize database pool
	postgresQueries := postgresdb.New(pool)
	k8sQueries := k8sdb.New(pool)
	harborQueries := harbordb.New(pool)

	dbConn := pool // Pour les health checks

	// Kubernetes client
	kubeClient, err := k8sclient.New(cfg, k8sQueries)
	if err != nil {
		klog.Fatalf("Unable to init Kubernetes client: %v", err)
	}
	defer func() {
		if err := kubeClient.Close(); err != nil {
			klog.ErrorS(err, "unable to close kubernetes client gracefully")
		}
	}()
	k8sSol, err := kubernetes.NewKubernetesSolution(cfg, kubeClient, k8sQueries)
	cobra.CheckErr(err)

	// Harbor client
	harborClient, err := harborclient.New(cfg, harborQueries)
	if err != nil {
		klog.Fatalf("Unable to init Harbor client: %v", err)
	}

	// Awx client
	awxClient, err := awxclient.New(cfg)
	if err != nil {
		klog.Fatalf("Unable to init Awx client: %v", err)
	}

	// kubernetes v2 service
	k8sSolV2, err := kubernetesv2.NewKubernetesSolution(cfg, kubeClient, k8sQueries)
	cobra.CheckErr(err)

	ss := []solutions.Solution{
		harbor.NewHarborSolution(cfg, harborClient, harborQueries),
		dbaas.NewDbaasSolution(cfg, awxClient, postgresQueries),
		k8sSol,
		k8sSolV2,
	}

	r := gin.Default()

	// CORS middleware
	r.Use(cors.New(cors.Config{
		AllowOrigins:     []string{"*"}, // minute papillon
		AllowMethods:     []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
		AllowHeaders:     []string{"Origin", "Content-Type", "Authorization"},
		AllowCredentials: false,
	}))

	r.Use(middleware.RequestID())
	r.Use(middleware.RequestLogger())
	r.Use(middleware.ErrorHandler())
	docs.SwaggerInfo.BasePath = "/api"
	api := r.Group("/api")

	// OIDC Middleware
	authMiddleware, err := middleware.NewOIDCMiddleware(middleware.OIDCConfig{
		Issuer:   cfg.OIDC.Issuer,
		Audience: cfg.OIDC.Audience,
	})
	if err != nil {
		klog.Fatalf("Unable to init OIDC middleware: %v", err)
	}
	api.Use(authMiddleware)

	for _, s := range ss {
		s.Endpoint(api.Group(fmt.Sprintf("/%s/%s", s.Name(), s.Version())))
	}

	// Health endpoints
	health.RegisterRoutes(r, &health.Options{
		Checkers: []health.ReadinessChecker{
			checkers.DBChecker{DB: dbConn},
			checkers.KubernetesChecker{Client: kubeClient},
			checkers.HarborChecker{NameStr: "harbor", Client: harborClient},
		},
	})

	r.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerfiles.Handler))
	// ---- NEW graceful shutdown section ----
	srv := &http.Server{
		Addr:              cfg.Server.Port, // ex. ":8080"
		Handler:           r,
		ReadHeaderTimeout: 10 * time.Second,
		IdleTimeout:       120 * time.Second,
	}

	// Lancement non bloquant
	go func() {
		if err := srv.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			klog.Fatalf("listen: %v", err)
		}
	}()

	// Capture SIGINT/SIGTERM
	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	<-ctx.Done() // attente du signal
	klog.Info("Shutdown signal received")

	// deadline de 15 s pour finir les requêtes
	shutCtx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()
	if err := srv.Shutdown(shutCtx); err != nil {
		klog.Errorf("Server shutdown error: %v", err)
	}

	pool.Close()       // fermeture pgxpool
	kubeClient.Close() // supposé implémenter io.Closer
	// harborClient.Close() // idem
	klog.Info("Server exiting")

}
