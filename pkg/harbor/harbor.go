package harbor

import (
	"github.com/gin-gonic/gin"
	"github.com/Gskill75/api2/pkg/config"
	db "github.com/Gskill75/api2/pkg/db/sqlc/harbor"
	harborclient "github.com/Gskill75/api2/pkg/harbor/client"
	"github.com/Gskill75/api2/pkg/harbor/project"
)

type HarborSolution struct {
	client  *harborclient.Client
	queries *db.Queries
	cfg     *config.Config
}

func NewHarborSolution(cfg *config.Config, client *harborclient.Client, queries *db.Queries) *HarborSolution {
	return &HarborSolution{
		client:  client,
		queries: queries,
		cfg:     cfg,
	}
}

func (*HarborSolution) Name() string {
	return "harbor"
}

func (*HarborSolution) Version() string {
	return "v1"
}

func (s *HarborSolution) Endpoint(rg *gin.RouterGroup) {
	{
		prj := rg.Group("/project")
		{
			prj.POST("/create/:name", project.CreateProjectHandler(s.client, s.queries))
			prj.POST("/create/robot/:name")
			prj.GET("/robot/:name")
			prj.GET("/:name")

		}
		robot := rg.Group("/robot")
		{
			robot.GET("/:name")
			robot.POST("/create/:name")
			robot.POST("/:name/members")
		}
	}
}
