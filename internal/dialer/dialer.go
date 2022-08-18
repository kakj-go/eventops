package dialer

import (
	"context"
	"fmt"
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
	"tiggerops/pkg/dialer"
)

func NewService(ctx context.Context, dbClient *gorm.DB, DialerServer *dialer.Server) *Service {
	var actuator = Service{
		ctx:          ctx,
		dbClient:     dbClient,
		DialerServer: DialerServer,
	}
	return &actuator
}

type Service struct {
	dbClient *gorm.DB

	DialerServer *dialer.Server

	ctx context.Context
}

func (s *Service) Router(router *gin.RouterGroup) {
	dialerGroup := router.Group("/dialer")
	{
		dialerGroup.Any("/connect", func(c *gin.Context) {
			s.DialerServer.RemoteDialer.ServeHTTP(c.Writer, c.Request)
		})
		dialerGroup.GET("/test", func(c *gin.Context) {
			clientID := c.Query("clientID")
			dial := s.DialerServer.GetClient(clientID, "60")
			if dial == nil {
				fmt.Println("")
				return
			}

			// docker client dialer

			//dockerClient, err := client.NewClientWithOpts(client.WithHost("tcp://127.0.0.1:2375"), client.WithDialContext(func(ctx context.Context, network, addr string) (net.Conn, error) {
			//	return dial(network, addr)
			//}))
			//if err != nil {
			//	return
			//}
			//fmt.Println(dockerClient.ImageList(context.Background(), types.ImageListOptions{}))

			// ssh client dialer

			//conn, err := dial("tcp", "127.0.0.1:22")
			//if err != nil {
			//	logrus.Infof("%v", err)
			//	return
			//}
			//
			//sshConn, a, b, err := ssh.NewClientConn(conn, "127.0.0.1:22", &ssh.ClientConfig{
			//	User:            "root",
			//	Auth:            []ssh.AuthMethod{ssh.Password("zhang2357")},
			//	HostKeyCallback: ssh.InsecureIgnoreHostKey(),
			//})
			//if err != nil {
			//	logrus.Infof("%v", err)
			//	return
			//}
			//sshClient := ssh.NewClient(sshConn, a, b)
			//
			//gophClient := goph.Client{
			//	Client: sshClient,
			//}
			//out, err := gophClient.Run("ls /tmp/")
			//if err != nil {
			//	logrus.Infof("%v", err)
			//	return
			//}
			//fmt.Println(string(out))
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
