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

package main

import (
	"context"
	"eventops/conf"
	"eventops/internal/core/dialer"
	"eventops/internal/core/eventprocess"
	"eventops/internal/core/flowmanager"
	"eventops/internal/core/token"
	dialerservice "eventops/internal/dialer"
	"eventops/internal/event"
	"eventops/internal/pipeline"
	"eventops/internal/register"
	"eventops/internal/uc"
	"eventops/pkg/dbclient"
	"fmt"
	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
	"gorm.io/gorm"
	"log"
	"net/http"
	"os"
	"os/signal"
	"time"
)

type server struct {
	services  []Service
	ginEngine *gin.Engine

	ctx context.Context
	db  *gorm.DB
}

func newServer() (*server, error) {
	ctx := context.Background()
	router := gin.Default()

	mysqlConf := conf.GetMysql()
	dbClient, err := dbclient.DBClient(mysqlConf.User, mysqlConf.Password, mysqlConf.Address, mysqlConf.Post, mysqlConf.Db)
	if err != nil {
		return nil, err
	}

	dialerServer := dialer.NewServer()
	pipelineManager := flowmanager.NewFlowManager(ctx, dbClient, dialerServer)
	eventProcess := eventprocess.NewProcess(dbClient, ctx, pipelineManager)

	ucService := uc.NewService(ctx, dbClient)
	registerService := register.NewService(ctx, dbClient, eventProcess, dialerServer)
	eventService := event.NewService(ctx, dbClient, eventProcess)
	actuatorService := dialerservice.NewService(ctx, dbClient, dialerServer)
	pipelineService := pipeline.NewService(ctx, dbClient, pipelineManager)

	var services []Service
	services = append(services, ucService)
	services = append(services, registerService)
	services = append(services, eventService)
	services = append(services, actuatorService)
	services = append(services, pipelineService)

	return &server{
		ginEngine: router,
		db:        dbClient,
		ctx:       ctx,
		services:  services,
	}, nil
}

func (srv *server) run() {
	srv.routing()
	srv.runServices()

	server := srv.ListenAndServe()
	srv.ListenShutdown(server)
}

func (srv *server) ListenShutdown(server *http.Server) {
	quit := make(chan os.Signal)
	signal.Notify(quit, os.Interrupt)
	<-quit
	log.Println("Shutdown Server ...")

	ctx, cancel := context.WithTimeout(srv.ctx, 5*time.Second)
	defer cancel()
	if err := server.Shutdown(ctx); err != nil {
		log.Fatal("Server Shutdown:", err)
	}
	log.Println("Server exiting")
}

func (srv *server) ListenAndServe() *http.Server {
	server := &http.Server{
		Addr:    ":" + conf.GetPort(),
		Handler: srv.ginEngine,
	}

	go func() {
		logrus.Infof("server start port: %v", conf.GetPort())
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("listen: %s\n", err)
		}
	}()
	return server
}

func (srv *server) runServices() {
	for _, service := range srv.services {
		err := service.Run()
		if err != nil {
			panic(fmt.Errorf("run service %v", service.Name()))
		}
	}
}

func (srv *server) routing() {

	routeGroup := srv.ginEngine.Group("/api")
	routeGroup.Use(token.LoginAuth)

	for _, router := range srv.services {
		router.Router(routeGroup)
	}
}
