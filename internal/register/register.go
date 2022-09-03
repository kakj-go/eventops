package register

import (
	"context"
	"eventops/internal/core/client/actuatorclient"
	"eventops/internal/core/client/pipelinedefinitionclient"
	"eventops/internal/core/client/triggerdefinitionclient"
	"eventops/internal/core/dialer"
	"eventops/internal/core/eventprocess"
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

func NewService(ctx context.Context, dbClient *gorm.DB, eventProcess *eventprocess.Process, dialerServer *dialer.Server) *Service {
	pipelineVersionDefinitionClient := pipelinedefinitionclient.NewPipelineDefinitionClient(dbClient)
	triggerDefinitionClient := triggerdefinitionclient.NewTriggerDefinitionClient(dbClient)
	actuatorClient := actuatorclient.NewActuatorsClient(dbClient)

	var register = Service{
		ctx:      ctx,
		dbClient: dbClient,

		pipelineVersionDefinitionClient: pipelineVersionDefinitionClient,
		triggerDefinitionClient:         triggerDefinitionClient,
		actuatorClient:                  actuatorClient,

		dialerServer: dialerServer,
		eventProcess: eventProcess,
	}
	return &register
}

type Service struct {
	eventProcess *eventprocess.Process
	dialerServer *dialer.Server

	pipelineVersionDefinitionClient *pipelinedefinitionclient.Client
	triggerDefinitionClient         *triggerdefinitionclient.Client
	actuatorClient                  *actuatorclient.Client

	dbClient *gorm.DB
	ctx      context.Context
}

func (r *Service) Router(router *gin.RouterGroup) {
	taskDefinition := router.Group("/pipeline-definition")
	{
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
		//triggerDefinition.GET("/:name/list-event-trigger", r.ListEventTrigger)
	}

	clientGroup := router.Group("/actuator")
	{
		clientGroup.POST("/apply", r.ApplyActuator)
		clientGroup.DELETE("/:name", r.DeleteActuator)
		clientGroup.GET("/", r.ListMyActuator)
	}
}

func (r *Service) Run() error {
	return nil
}

func (r *Service) Name() string {
	return "register"
}
