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
