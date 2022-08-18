package event

import (
	"context"
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
	"tiggerops/internal/event/client/eventclient"
	"tiggerops/pkg/eventprocess"
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
	s.process.Run()
	s.process.LoopLoadEvent()
	return nil
}

func (s *Service) Name() string {
	return "event"
}
