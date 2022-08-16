package uc

import (
	"context"
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
	"tiggerops/internal/uc/client/userclient"
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
