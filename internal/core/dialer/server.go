package dialer

import (
	"eventops/conf"
	"fmt"
	"github.com/rancher/remotedialer"
	"net/http"
	"strconv"
	"sync"
	"time"
)

const AuthHeader = "eventops-API-Tunnel-Token"
const IdHeader = "eventops-API-Tunnel-Id"
const UserHeader = "eventops-API-Tunnel-user"

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

func signBuild(clientId, user string) string {
	return fmt.Sprintf("%v/%v", user, clientId)
}

func (server *Server) AddAuthInfo(clientId, user, token string) {
	server.l.Lock()
	defer server.l.Unlock()

	server.AuthList[signBuild(clientId, user)] = token
}

func (server *Server) DeleteAuthInfo(clientId, user string) {
	server.l.Lock()
	defer server.l.Unlock()

	delete(server.AuthList, fmt.Sprintf("%s/%s", user, clientId))
}

func (server *Server) authorizer(req *http.Request) (string, bool, error) {
	server.l.Lock()
	defer server.l.Unlock()

	id := req.Header.Get(IdHeader)
	user := req.Header.Get(UserHeader)

	sign := signBuild(id, user)
	return sign, server.AuthList[sign] == req.Header.Get(AuthHeader), nil
}

func (server *Server) GetClient(creater, clientKey, timeout string) remotedialer.Dialer {
	server.l.Lock()
	defer server.l.Unlock()

	key := fmt.Sprintf("%s-%s/%s", creater, clientKey, timeout)
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
