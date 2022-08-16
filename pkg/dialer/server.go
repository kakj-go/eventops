package dialer

import (
	"fmt"
	"github.com/rancher/remotedialer"
	"net/http"
	"strconv"
	"strings"
	"sync"
	"tiggerops/conf"
	"time"
)

type Server struct {
	clients map[string]remotedialer.Dialer
	l       sync.Mutex

	DialerServer *remotedialer.Server
}

func NewServer() *Server {
	if conf.GetActuator().Dialer.PrintTunnelData {
		remotedialer.PrintTunnelData = true
	}

	handler := remotedialer.New(authorizer, remotedialer.DefaultErrorWriter)
	handler.PeerToken = conf.GetActuator().Dialer.PeerToken
	handler.PeerID = conf.GetActuator().Dialer.PeerID

	for _, peer := range strings.Split(conf.GetActuator().Dialer.Peers, ",") {
		parts := strings.SplitN(strings.TrimSpace(peer), ":", 3)
		if len(parts) != 3 {
			continue
		}
		handler.AddPeer(parts[2], parts[0], parts[1])
	}

	return &Server{
		DialerServer: handler,
		l:            sync.Mutex{},
		clients:      map[string]remotedialer.Dialer{},
	}
}

func authorizer(req *http.Request) (string, bool, error) {
	id := req.Header.Get("x-tunnel-id")
	return id, id != "", nil
}

func (server *Server) GetClient(clientKey, timeout string) remotedialer.Dialer {
	server.l.Lock()
	defer server.l.Unlock()

	key := fmt.Sprintf("%s/%s", clientKey, timeout)
	client := server.clients[key]
	if client != nil {
		return client
	}

	var deadline = time.Second * 30
	t, err := strconv.Atoi(timeout)
	if err == nil {
		deadline = time.Duration(t) * time.Second
	}

	dialer := server.DialerServer.Dialer(clientKey, deadline)
	server.clients[key] = dialer
	return client
}
