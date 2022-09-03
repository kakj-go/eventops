/*
 * Copyright (c) 2021 Terminus, Inc.
 *
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 *       http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 *
 */

package main

import (
	"context"
	"eventops/internal/core/dialer"
	"flag"
	"github.com/rancher/remotedialer"
	"github.com/sirupsen/logrus"
	"net/http"
	"time"
)

var serverAddr string
var clientID string
var token string
var user string
var debug bool

func auth(string, string) bool {
	return true
}

func main() {
	flag.StringVar(&serverAddr, "connect", "ws://192.168.0.105:8080/api/dialer/connect", "Address to connect to")
	flag.StringVar(&clientID, "id", "runner_tunnel", "Client ID")
	flag.StringVar(&token, "token", "123456", "Client Token")
	flag.StringVar(&user, "user", "kakj", "Client User")
	flag.BoolVar(&debug, "debug", true, "Debug logging")
	flag.Parse()

	if debug {
		logrus.SetLevel(logrus.DebugLevel)
	}

	headers := http.Header{
		dialer.IdHeader:   []string{clientID},
		dialer.AuthHeader: []string{token},
		dialer.UserHeader: []string{user},
	}

	for {
		remotedialer.ClientConnect(context.Background(), serverAddr, headers, nil, auth, nil)
		logrus.Errorf("failed to client server: %v, reclient after 1 second", serverAddr)
		time.Sleep(time.Second)
	}
}
