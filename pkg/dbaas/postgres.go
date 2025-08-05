package dbaas

import (
	"github.com/gin-gonic/gin"
	awxclient "gitn.sigma.fr/sigma/paas/api/api/pkg/awx/client"
	"gitn.sigma.fr/sigma/paas/api/api/pkg/config"
	db "gitn.sigma.fr/sigma/paas/api/api/pkg/db/sqlc/postgresql"
	handler "gitn.sigma.fr/sigma/paas/api/api/pkg/dbaas/handler/postgresql"
	service "gitn.sigma.fr/sigma/paas/api/api/pkg/dbaas/service/postgresql"
)

type DbaasSolution struct {
	awxclient *awxclient.Client
	queries   *db.Queries
	cfg       *config.Config
	service   *service.PostgresService
}

func NewDbaasSolution(cfg *config.Config, awxclient *awxclient.Client, queries *db.Queries) *DbaasSolution {
	postgresService := service.NewPostgresService(awxclient, queries, cfg)
	return &DbaasSolution{
		awxclient: awxclient,
		queries:   queries,
		cfg:       cfg,
		service:   postgresService,
	}
}

func (*DbaasSolution) Name() string {
	return "postgres"
}
func (*DbaasSolution) Version() string {
	return "v1"
}

/*
func (s *DbaasSolution) Endpoint(rg *gin.RouterGroup) {
	rg.POST("/provision", handler.ProvisionPostgresHandler(s.awxclient, s.queries, s.service))
	rg.GET("/provision/:job_id/status", handler.GetJobStatusHandler(s.service))
	rg.GET("/provision/check", handler.CheckActiveJobHandler(s.service))
}
*/

func (s *DbaasSolution) Endpoint(rg *gin.RouterGroup) {

	userGroup := rg.Group("/patroni")
	{
		userGroup.POST("/instance", handler.ProvisionPostgresHandler(s.awxclient, s.queries, s.service))
		userGroup.GET("/instance/:job_id/status", handler.GetJobStatusHandler(s.service))
		userGroup.GET("/instance/check", handler.CheckActiveJobHandler(s.service))
	}
}
