package event

import (
	"context"
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
	"tiggerops/internal/event/client/eventclient"
	"tiggerops/internal/register/client/triggerdefinitionclient"
)

type Service struct {
	ctx           context.Context
	eventDbClient *eventclient.Client
	dbClient      *gorm.DB

	Process Process
}

func NewService(ctx context.Context, dbClient *gorm.DB) *Service {
	var register = Service{
		ctx:           ctx,
		eventDbClient: eventclient.NewEventClient(dbClient),
		dbClient:      dbClient,
	}
	register.Process = NewProcess(register.eventDbClient, triggerdefinitionclient.NewTriggerDefinitionClient(dbClient), dbClient, ctx)
	return &register
}

func (s *Service) Router(router *gin.RouterGroup) {
	event := router.Group("/event")
	{
		event.POST("/send", s.send)
	}
}

func (s *Service) Run() error {
	s.Process.Run()
	s.Process.loopLoadEvent()
	return nil
}

func (s *Service) Name() string {
	return "event"
}
