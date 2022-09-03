package pipeline

import (
	"context"
	"eventops/internal/core/client/actuatorclient"
	"eventops/internal/core/client/eventclient"
	"eventops/internal/core/client/pipelineclient"
	"eventops/internal/core/client/pipelinedefinitionclient"
	"eventops/internal/core/client/taskclient"
	"eventops/internal/core/client/triggerdefinitionclient"
	"eventops/internal/core/flowmanager"
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

func NewService(ctx context.Context, dbClient *gorm.DB, manager *flowmanager.FlowManager) *Service {
	var register = Service{
		ctx:      ctx,
		dbClient: dbClient,
		manager:  manager,

		pipelineDbClient:         pipelineclient.NewPipelineClient(dbClient),
		taskDbClient:             taskclient.NewTaskClient(dbClient),
		eventClient:              eventclient.NewEventClient(dbClient),
		eventTriggerClient:       eventclient.NewEventClient(dbClient),
		triggerDefinitionClient:  triggerdefinitionclient.NewTriggerDefinitionClient(dbClient),
		actuatorClient:           actuatorclient.NewActuatorsClient(dbClient),
		pipelineDefinitionClient: pipelinedefinitionclient.NewPipelineDefinitionClient(dbClient),
	}
	return &register
}

type Service struct {
	pipelineDbClient         *pipelineclient.Client
	taskDbClient             *taskclient.Client
	eventClient              *eventclient.Client
	eventTriggerClient       *eventclient.Client
	triggerDefinitionClient  *triggerdefinitionclient.Client
	actuatorClient           *actuatorclient.Client
	pipelineDefinitionClient *pipelinedefinitionclient.Client

	dbClient *gorm.DB
	ctx      context.Context
	manager  *flowmanager.FlowManager
}

func (s *Service) Router(router *gin.RouterGroup) {
	clientGroup := router.Group("/pipeline")
	{
		clientGroup.POST("/:id/cancel", s.Cancel)
		clientGroup.GET("/:id", s.Get)
		clientGroup.GET("/", s.List)
		clientGroup.POST("/callback", s.Callback)
	}
}

func (s *Service) Run() error {
	return s.manager.Run()
}

func (s *Service) Name() string {
	return "pipeline"
}
