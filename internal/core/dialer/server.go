/*
 * Copyright 2022 The kakj-go Authors.
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *     http://www.apache.org/licenses/LICENSE-2.0
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 */

package dialer

import (
	"eventops/conf"
	"fmt"
	"github.com/rancher/remotedialer"
	"net/http"
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

var deadline = 20 * 365 * 24 * time.Hour

func (server *Server) GetClient(creater, clientKey string) remotedialer.Dialer {
	server.l.Lock()
	defer server.l.Unlock()

	sign := signBuild(clientKey, creater)

	hasSession := server.RemoteDialer.HasSession(sign)
	if !hasSession {
		return nil
	}

	client := server.clients[sign]
	if client != nil {
		return client
	}

	dialer := server.RemoteDialer.Dialer(sign, deadline)
	if dialer != nil {
		server.clients[sign] = dialer
	}
	return dialer
}
