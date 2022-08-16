package register

import (
	"context"
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
	"tiggerops/internal/register/client/pipelinedefinitionclient"
	"tiggerops/internal/register/client/triggerdefinitionclient"
)

func NewRegister(ctx context.Context, dbClient *gorm.DB) *Service {
	pipelineVersionDefinitionClient := pipelinedefinitionclient.NewPipelineDefinitionClient(dbClient)
	triggerDefinitionClient := triggerdefinitionclient.NewTriggerDefinitionClient(dbClient)

	var register = Service{ctx: ctx, pipelineVersionDefinitionClient: pipelineVersionDefinitionClient, triggerDefinitionClient: triggerDefinitionClient, dbClient: dbClient}
	return &register
}

type Service struct {
	pipelineVersionDefinitionClient *pipelinedefinitionclient.Client
	triggerDefinitionClient         *triggerdefinitionclient.Client
	dbClient                        *gorm.DB
	ctx                             context.Context
}

func (r *Service) Router(router *gin.RouterGroup) {
	taskDefinition := router.Group("/pipeline-definition")
	{
		//taskDefinition.GET("/", r.PagePipeline)
		taskDefinition.GET("/:name", r.GetPipeline)
		taskDefinition.GET("/:name/:version", r.GetPipelineVersion)
		taskDefinition.GET("/", r.ListMyPipelineVersion)
		taskDefinition.POST("/apply", r.ApplyPipeline)
		taskDefinition.DELETE("/:name/:version", r.DeletePipeline)
	}

	triggerDefinition := router.Group("/trigger-definition")
	{
		triggerDefinition.POST("/apply", r.ApplyTriggerDefinition)
		triggerDefinition.DELETE("/:name", r.DeleteTriggerDefinition)
		triggerDefinition.GET("/", r.ListMyTriggerDefinition)
		triggerDefinition.GET("/:name/list-event-trigger", r.ListEventTrigger)
	}
}

func (r *Service) Run() error {
	return nil
}

func (r *Service) Name() string {
	return "register"
}
