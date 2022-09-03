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

package uc

import (
	"context"
	"eventops/internal/core/client/userclient"
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

type Service struct {
	ctx          context.Context
	userDbClient *userclient.Client
	dbClient     *gorm.DB
}

func NewService(ctx context.Context, dbClient *gorm.DB) *Service {
	var register = Service{
		ctx:          ctx,
		userDbClient: userclient.NewUserClient(dbClient),
		dbClient:     dbClient,
	}
	return &register
}

func (u *Service) Router(router *gin.RouterGroup) {
	taskDefinition := router.Group("/user")
	{
		taskDefinition.GET("/me", u.me)
		taskDefinition.POST("/register", u.register)
		taskDefinition.POST("/login", u.login)
	}
}

func (u *Service) Run() error {
	return nil
}

func (u *Service) Name() string {
	return "uc"
}
