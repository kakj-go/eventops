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

package dialer

import (
	"context"
	"eventops/internal/core/client/actuatorclient"
	"eventops/internal/core/dialer"
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

func NewService(ctx context.Context, dbClient *gorm.DB, dialerServer *dialer.Server) *Service {
	var actuator = Service{
		ctx:            ctx,
		dbClient:       dbClient,
		actuatorClient: actuatorclient.NewActuatorsClient(dbClient),
		dialerServer:   dialerServer,
	}
	return &actuator
}

type Service struct {
	dbClient       *gorm.DB
	actuatorClient *actuatorclient.Client

	dialerServer *dialer.Server

	ctx context.Context
}

func (s *Service) Router(router *gin.RouterGroup) {
	dialerGroup := router.Group("/dialer")
	{
		dialerGroup.Any("/connect", func(c *gin.Context) {
			s.dialerServer.RemoteDialer.ServeHTTP(c.Writer, c.Request)
		})
	}
}

func (s *Service) Run() error {
	s.loadActuatorToDialerServer()
	return nil
}

func (s *Service) loadActuatorToDialerServer() {
	actuatorList, err := s.actuatorClient.ListActuator(nil, actuatorclient.ListActuatorQuery{})
	if err != nil {
		panic(err)
	}

	for _, actuator := range actuatorList {
		s.dialerServer.AddAuthInfo(actuator.ClientId, actuator.Creater, actuator.ClientToken)
	}
}

func (s *Service) Name() string {
	return "dialer"
}
