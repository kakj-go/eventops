package flow

import (
	"context"
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
	"tiggerops/internal/flow/client"
)

func NewFlow(ctx context.Context, dbClient *gorm.DB) *Service {
	var register = Service{ctx: ctx, flowDbClient: client.NewFlowClient(dbClient), dbClient: dbClient}
	return &register
}

type Service struct {
	flowDbClient *client.Client
	dbClient     *gorm.DB
	ctx          context.Context
}

func (s *Service) Router(router *gin.RouterGroup) {

}

func (s *Service) Run() error {
	return nil
}

func (s *Service) Name() string {
	return "flow"
}
