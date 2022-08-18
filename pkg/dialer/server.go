package dialer

import (
	"fmt"
	"github.com/rancher/remotedialer"
	"net/http"
	"strconv"
	"sync"
	"tiggerops/conf"
	"time"
)

const AuthHeader = "eventops-API-Tunnel-Token"
const IdHeader = "eventops-API-Tunnel-Id"

type Server struct {
	clients map[string]remotedialer.Dialer
	l       sync.Mutex

	AuthList     map[string]string
	RemoteDialer *remotedialer.Server
}

func NewServer() *Server {
	if conf.GetActuator().PrintTunnelData {
		remotedialer.PrintTunnelData = true
	}

	var server = &Server{
		l:        sync.Mutex{},
		clients:  map[string]remotedialer.Dialer{},
		AuthList: map[string]string{},
	}

	handler := remotedialer.New(server.authorizer, remotedialer.DefaultErrorWriter)
	server.RemoteDialer = handler

	return server
}

func (server *Server) AddAuthInfo(id, token string) {
	server.AuthList[id] = token
}

func (server *Server) authorizer(req *http.Request) (string, bool, error) {
	id := req.Header.Get(IdHeader)

	return id, server.AuthList[id] == req.Header.Get(AuthHeader), nil
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

	dialer := server.RemoteDialer.Dialer(clientKey, deadline)
	if dialer != nil {
		server.clients[key] = dialer
	}
	return dialer
}
