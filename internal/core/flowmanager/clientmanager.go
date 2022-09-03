/*
 * Copyright (c) 2021 Terminus, Inc.
 *
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 *       http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 *
 */

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
