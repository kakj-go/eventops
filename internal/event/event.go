/*
 * Copyright 2022 The kakj-go Authors.
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *     http://www.apache.org/licenses/LICENSE-2.0
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 */

package event

import (
	"context"
	"eventops/internal/core/client/eventclient"
	"eventops/internal/core/eventprocess"
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

type Service struct {
	ctx           context.Context
	eventDbClient *eventclient.Client
	dbClient      *gorm.DB

	process *eventprocess.Process
}

func NewService(ctx context.Context, dbClient *gorm.DB, eventProcess *eventprocess.Process) *Service {
	var register = Service{
		ctx:           ctx,
		eventDbClient: eventclient.NewEventClient(dbClient),
		dbClient:      dbClient,
		process:       eventProcess,
	}
	return &register
}

func (s *Service) Router(router *gin.RouterGroup) {
	event := router.Group("/event")
	{
		event.POST("/send", s.send)
	}
}

func (s *Service) Run() error {
	s.process.ProcessEvent()
	go s.process.LoopLoadProcessingEventToPass()
	return s.process.LoadPassEventTriggerToProcess()
}

func (s *Service) Name() string {
	return "event"
}
