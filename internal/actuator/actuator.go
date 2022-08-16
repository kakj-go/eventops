package actuator

import (
	"context"
	"fmt"
	"github.com/docker/docker/api/types"
	"github.com/docker/docker/client"
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
	"net"
	"tiggerops/internal/actuator/client/actuatorclient"
	"tiggerops/pkg/dialer"
)

func NewActuator(ctx context.Context, dbClient *gorm.DB, DialerServer *dialer.Server) *Service {
	var actuator = Service{
		ctx:            ctx,
		actuatorClient: actuatorclient.NewActuatorClient(dbClient),
		dbClient:       dbClient,
		DialerServer:   DialerServer,
	}
	return &actuator
}

type Service struct {
	actuatorClient *actuatorclient.Client
	dbClient       *gorm.DB

	DialerServer *dialer.Server

	ctx context.Context
}

func (s *Service) Router(router *gin.RouterGroup) {
	dialerGroup := router.Group("/dialer")
	{
		dialerGroup.Any("/connect", func(c *gin.Context) {
			s.DialerServer.DialerServer.ServeHTTP(c.Writer, c.Request)
		})
		dialerGroup.GET("/test", func(c *gin.Context) {
			clientID := c.Query("clientID")
			dial := s.DialerServer.GetClient(clientID, "60")

			dockerClient, err := client.NewClientWithOpts(client.WithDialContext(func(ctx context.Context, network, addr string) (net.Conn, error) {
				return dial(network, addr)
			}))
			if err != nil {
				return
			}
			fmt.Println(dockerClient.ImageList(context.Background(), types.ImageListOptions{}))
		})
	}
}

func (s *Service) Run() error {
	// todo load from db
	s.DialerServer.AddAuthInfo("docker", "123456")
	return nil
}

func (s *Service) Name() string {
	return "actuator"
}
