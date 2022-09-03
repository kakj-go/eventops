package flowmanager

import (
	"eventops/internal/core/client/actuatorclient"
	"eventops/internal/core/client/eventclient"
	"eventops/internal/core/client/pipelineclient"
	"eventops/internal/core/client/pipelinedefinitionclient"
	"eventops/internal/core/client/taskclient"
	"eventops/internal/core/client/triggerdefinitionclient"
	"gorm.io/gorm"
)

type clientManager struct {
	db *gorm.DB

	eventClient              *eventclient.Client
	eventTriggerClient       *eventclient.Client
	triggerDefinitionClient  *triggerdefinitionclient.Client
	actuatorClient           *actuatorclient.Client
	pipelineDefinitionClient *pipelinedefinitionclient.Client
	pipelineClient           *pipelineclient.Client
	taskClient               *taskclient.Client
}

func newClientManager(dbClient *gorm.DB) *clientManager {
	return &clientManager{
		db:                       dbClient,
		eventClient:              eventclient.NewEventClient(dbClient),
		eventTriggerClient:       eventclient.NewEventClient(dbClient),
		triggerDefinitionClient:  triggerdefinitionclient.NewTriggerDefinitionClient(dbClient),
		actuatorClient:           actuatorclient.NewActuatorsClient(dbClient),
		pipelineDefinitionClient: pipelinedefinitionclient.NewPipelineDefinitionClient(dbClient),
		pipelineClient:           pipelineclient.NewPipelineClient(dbClient),
		taskClient:               taskclient.NewTaskClient(dbClient),
	}
}
