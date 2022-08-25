package dialer

import (
	"context"
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
	"tiggerops/internal/register/client/actuatorclient"
	"tiggerops/pkg/dialer"
)

func NewService(ctx context.Context, dbClient *gorm.DB, DialerServer *dialer.Server) *Service {
	var actuator = Service{
		ctx:            ctx,
		dbClient:       dbClient,
		actuatorClient: actuatorclient.NewActuatorsClient(dbClient),
		dialerServer:   DialerServer,
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