package main

import (
	"context"
	"fmt"
	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
	"gorm.io/gorm"
	"log"
	"net/http"
	"os"
	"os/signal"
	"tiggerops/conf"
	"tiggerops/internal/actuator"
	"tiggerops/internal/event"
	"tiggerops/internal/register"
	"tiggerops/internal/uc"
	"tiggerops/pkg/dbclient"
	"tiggerops/pkg/dialer"
	"tiggerops/pkg/token"
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

	ucService := uc.NewService(ctx, dbClient)
	registerService := register.NewRegister(ctx, dbClient)
	eventService := event.NewService(ctx, dbClient)
	actuatorService := actuator.NewActuator(ctx, dbClient, dialerServer)

	var services []Service
	services = append(services, ucService)
	services = append(services, registerService)
	services = append(services, eventService)
	services = append(services, actuatorService)

	return &server{
		ginEngine: router,
		db:        dbClient,
		ctx:       ctx,
		services:  services,
	}, nil
}

func (srv *server) run() {
	srv.runServices()
	srv.routing()

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
